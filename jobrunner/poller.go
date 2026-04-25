// SPDX-License-Identifier: EUPL-1.2

package jobrunner

import (
	// Note: AX-6 — Poller accepts caller cancellation and deadline propagation through context.Context.
	"context"
	// Note: AX-6 — sync.RWMutex is structural here because the pinned Core module does not export core.RWMutex.
	"sync"
	// Note: AX-6 — Poller cadence is expressed with time.Duration and time.Ticker.
	"time"

	// Note: AX-6 — Core supplies structured errors and formatting primitives.
	core "dappco.re/go/core"
)

// Poller discovers signals from sources and dispatches them to handlers.
type Poller struct {
	mu       sync.RWMutex
	sources  []JobSource
	handlers []JobHandler
	journal  *Journal
	interval time.Duration
	dryRun   bool
	cycle    int
}

// PollerConfig configures a Poller.
type PollerConfig struct {
	Sources      []JobSource
	Handlers     []JobHandler
	Journal      *Journal
	PollInterval time.Duration
	DryRun       bool
}

// NewPoller creates a Poller from the given config.
func NewPoller(cfg PollerConfig) *Poller {
	interval := cfg.PollInterval
	if interval <= 0 {
		interval = time.Minute
	}
	return &Poller{
		sources:  append([]JobSource(nil), cfg.Sources...),
		handlers: append([]JobHandler(nil), cfg.Handlers...),
		journal:  cfg.Journal,
		interval: interval,
		dryRun:   cfg.DryRun,
	}
}

// AddHandler appends a handler to the poller.
func (p *Poller) AddHandler(h JobHandler) {
	if p == nil || h == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.handlers = append(p.handlers, h)
}

// AddSource appends a source to the poller.
func (p *Poller) AddSource(s JobSource) {
	if p == nil || s == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sources = append(p.sources, s)
}

// Cycle returns the number of completed poll-dispatch cycles.
func (p *Poller) Cycle() int {
	if p == nil {
		return 0
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.cycle
}

// DryRun returns whether dry-run mode is enabled.
func (p *Poller) DryRun() bool {
	if p == nil {
		return false
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.dryRun
}

// SetDryRun enables or disables dry-run mode.
func (p *Poller) SetDryRun(v bool) {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.dryRun = v
}

// Run starts a blocking poll-dispatch loop.
func (p *Poller) Run(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := p.RunOnce(ctx); err != nil {
		return err
	}

	interval := p.pollInterval()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := p.RunOnce(ctx); err != nil {
				return err
			}
		}
	}
}

// RunOnce performs a single poll-dispatch cycle.
func (p *Poller) RunOnce(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	sources, handlers, journal, dryRun := p.snapshot()
	for _, source := range sources {
		if source == nil {
			continue
		}

		signals, err := source.Poll(ctx)
		if err != nil {
			return core.E("jobrunner.Poller.RunOnce", core.Sprintf("poll %s", source.Name()), err)
		}

		for _, signal := range signals {
			if err := ctx.Err(); err != nil {
				return err
			}
			if signal == nil {
				continue
			}

			handler := firstMatchingHandler(handlers, signal)
			if handler == nil {
				continue
			}
			if dryRun {
				continue
			}

			result, err := handler.Execute(ctx, signal)
			if err != nil {
				return core.E("jobrunner.Poller.RunOnce", core.Sprintf("execute %s", handler.Name()), err)
			}
			if result == nil {
				return core.E("jobrunner.Poller.RunOnce", "handler returned nil result", nil)
			}
			if journal != nil {
				if err := journal.Append(signal, result); err != nil {
					return err
				}
			}
			if err := source.Report(ctx, result); err != nil {
				return core.E("jobrunner.Poller.RunOnce", core.Sprintf("report %s", source.Name()), err)
			}
		}
	}

	p.mu.Lock()
	p.cycle++
	p.mu.Unlock()
	return nil
}

func (p *Poller) snapshot() ([]JobSource, []JobHandler, *Journal, bool) {
	if p == nil {
		return nil, nil, nil, false
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	sources := append([]JobSource(nil), p.sources...)
	handlers := append([]JobHandler(nil), p.handlers...)
	return sources, handlers, p.journal, p.dryRun
}

func (p *Poller) pollInterval() time.Duration {
	if p == nil {
		return time.Minute
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.interval <= 0 {
		return time.Minute
	}
	return p.interval
}

func firstMatchingHandler(handlers []JobHandler, signal *PipelineSignal) JobHandler {
	for _, handler := range handlers {
		if handler == nil {
			continue
		}
		if handler.Match(signal) {
			return handler
		}
	}
	return nil
}
