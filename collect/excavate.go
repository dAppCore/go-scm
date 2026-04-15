// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	"context"
	fmt "dappco.re/go/core/scm/internal/ax/fmtx"
	stdstrings "strings"
	"time"

	core "dappco.re/go/core/log"
)

// Excavator runs multiple collectors as a coordinated operation.
// It provides sequential execution with rate limit respect, state tracking
// for resume support, and aggregated results.
type Excavator struct {
	// Collectors is the list of collectors to run.
	Collectors []Collector

	// ScanOnly reports what would be collected without performing collection.
	ScanOnly bool

	// Resume enables incremental collection using saved state.
	Resume bool
}

// Name returns the orchestrator name.
// Usage: Name(...)
func (e *Excavator) Name() string {
	return "excavator"
}

// Run executes all collectors sequentially, respecting rate limits and
// using state for resume support. Results are aggregated from all collectors.
// Usage: Run(...)
func (e *Excavator) Run(ctx context.Context, cfg *Config) (*Result, error) {
	result := &Result{Source: e.Name()}

	if cfg == nil {
		return result, core.E("collect.Excavator.Run", "config is required", nil)
	}
	if len(e.Collectors) == 0 {
		return result, nil
	}

	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitStart(e.Name(), fmt.Sprintf("Starting excavation with %d collectors", len(e.Collectors)))
	}
	verboseProgress(cfg, e.Name(), fmt.Sprintf("queueing %d collectors", len(e.Collectors)))

	// Load state if resuming
	if e.Resume && cfg.State != nil {
		verboseProgress(cfg, e.Name(), "loading resume state")
		if err := cfg.State.Load(); err != nil {
			return result, core.E("collect.Excavator.Run", "failed to load state", err)
		}
	}

	// If scan-only, just report what would be collected
	if e.ScanOnly {
		for _, c := range e.Collectors {
			if cfg.Dispatcher != nil {
				cfg.Dispatcher.EmitProgress(e.Name(), fmt.Sprintf("[scan] Would run collector: %s", c.Name()), nil)
			}
			verboseProgress(cfg, e.Name(), fmt.Sprintf("scan-only collector: %s", c.Name()))
		}
		if cfg.Dispatcher != nil {
			cfg.Dispatcher.EmitComplete(e.Name(), "Scan-only excavation complete", result)
		}
		return result, nil
	}

	var runErr error

	for i, c := range e.Collectors {
		if ctx.Err() != nil {
			runErr = core.E("collect.Excavator.Run", "context cancelled", ctx.Err())
			break
		}

		if cfg.Dispatcher != nil {
			cfg.Dispatcher.EmitProgress(e.Name(),
				fmt.Sprintf("Running collector %d/%d: %s", i+1, len(e.Collectors), c.Name()), nil)
		}
		verboseProgress(cfg, e.Name(), fmt.Sprintf("dispatching collector %d/%d: %s", i+1, len(e.Collectors), c.Name()))

		// Check if we should skip (already completed in a previous run)
		if e.Resume && cfg.State != nil {
			if entry, ok := cfg.State.Get(c.Name()); ok {
				if entry.Items > 0 && !entry.LastRun.IsZero() {
					if cfg.Dispatcher != nil {
						cfg.Dispatcher.EmitProgress(e.Name(),
							fmt.Sprintf("Skipping %s (already collected %d items on %s)",
								c.Name(), entry.Items, entry.LastRun.Format(time.RFC3339)), nil)
					}
					verboseProgress(cfg, e.Name(), fmt.Sprintf("resume skip: %s", c.Name()))
					result.Skipped++
					continue
				}
			}
		}

		if cfg.Limiter != nil {
			sourceKey := collectorRateLimitKey(c.Name())
			if err := cfg.Limiter.Wait(ctx, sourceKey); err != nil {
				runErr = core.E("collect.Excavator.Run", "rate limit wait failed", err)
				break
			}
		}

		collectorResult, err := c.Collect(ctx, cfg)
		if err != nil {
			result.Errors++
			if cfg.Dispatcher != nil {
				cfg.Dispatcher.EmitError(e.Name(),
					fmt.Sprintf("Collector %s failed: %v", c.Name(), err), nil)
			}
			continue
		}

		if collectorResult != nil {
			result.Items += collectorResult.Items
			result.Errors += collectorResult.Errors
			result.Skipped += collectorResult.Skipped
			result.Files = append(result.Files, collectorResult.Files...)

			// Update state
			if cfg.State != nil {
				cfg.State.Set(c.Name(), &StateEntry{
					Source:  c.Name(),
					LastRun: time.Now(),
					Items:   collectorResult.Items,
				})
			}
		}
	}

	// Save state
	if cfg.State != nil {
		verboseProgress(cfg, e.Name(), "saving resume state")
		if err := cfg.State.Save(); err != nil {
			if cfg.Dispatcher != nil {
				cfg.Dispatcher.EmitError(e.Name(), fmt.Sprintf("Failed to save state: %v", err), nil)
			}
		}
	}

	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitComplete(e.Name(),
			fmt.Sprintf("Excavation complete: %d items, %d errors, %d skipped",
				result.Items, result.Errors, result.Skipped), result)
	}

	return result, runErr
}

func collectorRateLimitKey(name string) string {
	if key, _, ok := stdstrings.Cut(name, ":"); ok && key != "" {
		return key
	}
	return name
}
