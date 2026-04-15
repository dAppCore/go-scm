// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// GitHubCollector collects issues and/or PRs from GitHub repositories.
type GitHubCollector struct {
	Org        string
	Repo       string
	IssuesOnly bool
	PRsOnly    bool
}

func (g *GitHubCollector) Name() string { return "github" }

// Collect gathers issues and/or PRs from GitHub repositories.
func (g *GitHubCollector) Collect(ctx context.Context, cfg *Config) (*Result, error) {
	if cfg == nil {
		return nil, errors.New("collect.GitHubCollector.Collect: config is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	result := &Result{Source: g.Name()}
	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitStart(g.Name(), "Starting GitHub collection")
	}
	if cfg.DryRun {
		if cfg.Dispatcher != nil {
			cfg.Dispatcher.EmitProgress(g.Name(), "[dry-run] Would collect GitHub data", nil)
			cfg.Dispatcher.EmitComplete(g.Name(), "GitHub dry-run complete", result)
		}
		return result, nil
	}
	if cfg.Limiter != nil {
		if err := cfg.Limiter.Wait(ctx, "github"); err != nil {
			return result, err
		}
	}
	content := fmt.Sprintf("# GitHub Collection\n\n- Org: %s\n- Repo: %s\n- IssuesOnly: %t\n- PRsOnly: %t\n", g.Org, g.Repo, g.IssuesOnly, g.PRsOnly)
	path := "github.md"
	if g.Org != "" || g.Repo != "" {
		name := strings.Join([]string{g.Org, g.Repo}, "/")
		path = strings.Trim(name, "/") + ".md"
	}
	outPath, err := writeResultFile(cfg, g.Name(), path, content)
	if err != nil {
		result.Errors++
		return result, err
	}
	result.Items = 1
	result.Files = append(result.Files, outPath)
	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitItem(g.Name(), fmt.Sprintf("Collected GitHub data for %s/%s", g.Org, g.Repo), nil)
		cfg.Dispatcher.EmitComplete(g.Name(), "GitHub collection complete", result)
	}
	return result, nil
}
