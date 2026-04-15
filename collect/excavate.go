// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	"context"
	"errors"
	"fmt"
	"time"
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
func (e *Excavator) Name() string {
	return "excavator"
}

// Run executes all collectors sequentially, respecting rate limits and
// using state for resume support. Results are aggregated from all collectors.
func (e *Excavator) Run(ctx context.Context, cfg *Config) (*Result, error) {
	result := &Result{Source: e.Name()}
	if cfg == nil {
		return nil, errors.New("collect.Excavator.Run: config is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	if len(e.Collectors) == 0 {
		return result, nil
	}

	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitStart(e.Name(), fmt.Sprintf("Starting excavation with %d collectors", len(e.Collectors)))
	}

	if e.Resume && cfg.State != nil {
		if err := cfg.State.Load(); err != nil {
			return result, err
		}
	}

	if e.ScanOnly {
		for _, c := range e.Collectors {
			if cfg.Dispatcher != nil {
				cfg.Dispatcher.EmitProgress(e.Name(), fmt.Sprintf("[scan] Would run collector: %s", c.Name()), nil)
			}
		}
		if cfg.Dispatcher != nil {
			cfg.Dispatcher.EmitComplete(e.Name(), "Excavation scan complete", result)
		}
		return result, nil
	}

	for i, c := range e.Collectors {
		if c == nil {
			continue
		}
		if err := ctx.Err(); err != nil {
			return result, err
		}

		if cfg.Dispatcher != nil {
			cfg.Dispatcher.EmitProgress(e.Name(), fmt.Sprintf("Running collector %d/%d: %s", i+1, len(e.Collectors), c.Name()), nil)
		}

		if e.Resume && cfg.State != nil {
			if entry, ok := cfg.State.Get(c.Name()); ok && entry != nil && entry.Items > 0 && !entry.LastRun.IsZero() {
				result.Skipped++
				if cfg.Dispatcher != nil {
					cfg.Dispatcher.EmitProgress(
						e.Name(),
						fmt.Sprintf("Skipping %s (already collected %d items on %s)", c.Name(), entry.Items, entry.LastRun.Format("2006-01-02T15:04:05Z07:00")),
						nil,
					)
				}
				continue
			}
		}

		if cfg.Limiter != nil {
			if err := cfg.Limiter.Wait(ctx, c.Name()); err != nil {
				result.Errors++
				if cfg.Dispatcher != nil {
					cfg.Dispatcher.EmitError(e.Name(), fmt.Sprintf("Rate limit wait failed for %s: %v", c.Name(), err), nil)
				}
				continue
			}
		}

		collectorResult, err := c.Collect(ctx, cfg)
		if err != nil {
			result.Errors++
			if cfg.Dispatcher != nil {
				cfg.Dispatcher.EmitError(e.Name(), fmt.Sprintf("Collector %s failed: %v", c.Name(), err), nil)
			}
			continue
		}
		if collectorResult == nil {
			continue
		}

		result.Items += collectorResult.Items
		result.Errors += collectorResult.Errors
		result.Skipped += collectorResult.Skipped
		result.Files = append(result.Files, collectorResult.Files...)

		if cfg.State != nil {
			cfg.State.Set(c.Name(), &StateEntry{
				Source:  c.Name(),
				LastRun: nowUTC(),
				Items:   collectorResult.Items,
			})
		}
	}

	if cfg.State != nil {
		if err := cfg.State.Save(); err != nil && cfg.Dispatcher != nil {
			cfg.Dispatcher.EmitError(e.Name(), fmt.Sprintf("Failed to save state: %v", err), nil)
		}
	}

	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitComplete(e.Name(), fmt.Sprintf("Excavation complete: %d items, %d errors, %d skipped", result.Items, result.Errors, result.Skipped), result)
	}

	return result, nil
}

func nowUTC() time.Time {
	return time.Now().UTC()
}
