// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	"context"
	"errors"
	"time"
)

// Excavator runs multiple collectors as a coordinated operation.
type Excavator struct {
	Collectors []Collector
	ScanOnly   bool
	Resume     bool
}

func (e *Excavator) Name() string { return "excavator" }

// Run executes all collectors sequentially, respecting rate limits and using state for resume support.
func (e *Excavator) Run(ctx context.Context, cfg *Config) (*Result, error) {
	if cfg == nil {
		return nil, errors.New("collect.Excavator.Run: config is required")
	}
	merged := &Result{Source: e.Name()}
	for _, collector := range e.Collectors {
		if collector == nil {
			continue
		}
		if ctx != nil {
			if err := ctx.Err(); err != nil {
				return merged, err
			}
		}
		if cfg.Limiter != nil {
			if err := cfg.Limiter.Wait(ctx, collector.Name()); err != nil {
				return merged, err
			}
		}
		if e.ScanOnly {
			merged.Skipped++
			continue
		}
		result, err := collector.Collect(ctx, cfg)
		if err != nil && result == nil {
			return merged, err
		}
		if result != nil {
			merged = MergeResults(e.Name(), merged, result)
			if cfg.State != nil {
				cfg.State.Set(collector.Name(), &StateEntry{
					Source:  collector.Name(),
					LastRun: nowUTC(),
					Items:   result.Items,
				})
			}
		}
	}
	if cfg.State != nil {
		_ = cfg.State.Save()
	}
	return merged, nil
}

func nowUTC() time.Time { return time.Now().UTC() }
