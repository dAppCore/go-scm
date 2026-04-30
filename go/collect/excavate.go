// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	// Note: context.Context is retained as the excavator API cancellation contract.
	"context"
	// Note: time.Now is retained behind nowUTC for collection state timestamps.
	"time"

	core "dappco.re/go"
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
func (e *Excavator) Run(ctx context.Context, cfg *Config) (*Result, error)  /* v090-result-boundary */ {
	result := &Result{Source: e.Name()}
	if cfg == nil {
		return nil, core.E("collect.Excavator.Run", "config is required", nil)
	}
	ctx, err := activeCollectContext(ctx)
	if err != nil {
		return result, err
	}

	if len(e.Collectors) == 0 {
		return result, nil
	}

	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitStart(e.Name(), core.Sprintf("Starting excavation with %d collectors", len(e.Collectors)))
	}

	if err := e.loadResumeState(cfg); err != nil {
		return result, err
	}

	if e.ScanOnly {
		e.emitScan(cfg, result)
		return result, nil
	}

	for i, c := range e.Collectors {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		e.runCollector(ctx, cfg, i, c, result)
	}

	e.saveState(cfg)
	e.emitComplete(cfg, result)

	return result, nil
}

func (e *Excavator) loadResumeState(cfg *Config) error  /* v090-result-boundary */ {
	if !e.Resume || cfg.State == nil {
		return nil
	}
	return cfg.State.Load()
}

func (e *Excavator) saveState(cfg *Config) {
	if cfg.State == nil {
		return
	}
	if err := cfg.State.Save(); err != nil && cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitError(e.Name(), core.Sprintf("Failed to save state: %v", err), nil)
	}
}

func (e *Excavator) emitComplete(cfg *Config, result *Result) {
	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitComplete(e.Name(), core.Sprintf("Excavation complete: %d items, %d errors, %d skipped", result.Items, result.Errors, result.Skipped), result)
	}
}

func (e *Excavator) emitScan(cfg *Config, result *Result) {
	if cfg.Dispatcher == nil {
		return
	}
	for _, c := range e.Collectors {
		cfg.Dispatcher.EmitProgress(e.Name(), core.Sprintf("[scan] Would run collector: %s", c.Name()), nil)
	}
	cfg.Dispatcher.EmitComplete(e.Name(), "Excavation scan complete", result)
}

func (e *Excavator) runCollector(ctx context.Context, cfg *Config, index int, c Collector, result *Result) {
	if c == nil {
		return
	}
	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitProgress(e.Name(), core.Sprintf("Running collector %d/%d: %s", index+1, len(e.Collectors), c.Name()), nil)
	}
	if e.skipCollected(cfg, c, result) {
		return
	}
	if err := waitCollectLimiter(ctx, cfg, c.Name()); err != nil {
		result.Errors++
		if cfg.Dispatcher != nil {
			cfg.Dispatcher.EmitError(e.Name(), core.Sprintf("Rate limit wait failed for %s: %v", c.Name(), err), nil)
		}
		return
	}
	collectorResult, err := c.Collect(ctx, cfg)
	if err != nil {
		result.Errors++
		if cfg.Dispatcher != nil {
			cfg.Dispatcher.EmitError(e.Name(), core.Sprintf("Collector %s failed: %v", c.Name(), err), nil)
		}
		return
	}
	e.mergeCollectorResult(cfg, c, collectorResult, result)
}

func (e *Excavator) skipCollected(cfg *Config, c Collector, result *Result) bool {
	if !e.Resume || cfg.State == nil {
		return false
	}
	entry, ok := cfg.State.Get(c.Name())
	if !ok || entry == nil || entry.Items <= 0 || entry.LastRun.IsZero() {
		return false
	}
	result.Skipped++
	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitProgress(
			e.Name(),
			core.Sprintf("Skipping %s (already collected %d items on %s)", c.Name(), entry.Items, entry.LastRun.Format("2006-01-02T15:04:05Z07:00")),
			nil,
		)
	}
	return true
}

func (e *Excavator) mergeCollectorResult(cfg *Config, c Collector, collectorResult *Result, result *Result) {
	if collectorResult == nil {
		return
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

func nowUTC() time.Time {
	return time.Now().UTC()
}
