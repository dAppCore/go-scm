// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	// Note: context.Context is retained as the collector API cancellation contract.
	"context"

	core "dappco.re/go"
	coreio "dappco.re/go/io"
)

// Collector is the interface all collection sources implement.
type Collector interface {
	Name() string
	Collect(ctx context.Context, cfg *Config) (*Result, error)
}

// Config holds shared configuration for all collectors.
type Config struct {
	Output     coreio.Medium
	OutputDir  string
	Limiter    *RateLimiter
	State      *State
	Dispatcher *Dispatcher
	Verbose    bool
	DryRun     bool
}

// Result holds the output of a collection run.
type Result struct {
	Source  string
	Items   int
	Errors  int
	Skipped int
	Files   []string
}

// NewConfig creates a Config with sensible defaults.
func NewConfig(outputDir string) *Config {
	return NewConfigWithMedium(coreio.NewMockMedium(), outputDir)
}

// NewConfigWithMedium creates a Config using the specified storage medium.
func NewConfigWithMedium(m coreio.Medium, outputDir string) *Config {
	if m == nil {
		m = coreio.NewMockMedium()
	}
	if core.Trim(outputDir) == "" {
		outputDir = "collect"
	}
	outputDir = core.CleanPath(outputDir, core.Env("DS"))
	statePath := core.JoinPath(outputDir, ".collect-state.json")
	return &Config{
		Output:     m,
		OutputDir:  outputDir,
		Limiter:    NewRateLimiter(),
		State:      NewState(m, statePath),
		Dispatcher: NewDispatcher(),
	}
}

// MergeResults combines multiple results into a single aggregated result.
func MergeResults(source string, results ...*Result) *Result {
	merged := &Result{Source: source}
	for _, result := range results {
		if result == nil {
			continue
		}
		merged.Items += result.Items
		merged.Errors += result.Errors
		merged.Skipped += result.Skipped
		merged.Files = append(merged.Files, result.Files...)
	}
	return merged
}

func writeResultFile(cfg *Config, source, name, content string) (string, error) {
	if cfg == nil || cfg.Output == nil {
		return "", core.E("collect.writeResultFile", "output medium is required", nil)
	}
	path := core.JoinPath(cfg.OutputDir, source, name)
	if err := cfg.Output.Write(path, content); err != nil {
		return "", err
	}
	return path, nil
}
