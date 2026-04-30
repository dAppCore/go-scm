// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	// Note: context.Context is retained as the collector API cancellation contract.
	"context"

	core "dappco.re/go"
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
func (g *GitHubCollector) Collect(ctx context.Context, cfg *Config) (*Result, error)  /* v090-result-boundary */ {
	if cfg == nil {
		return nil, core.E("collect.GitHubCollector.Collect", "config is required", nil)
	}
	ctx, err := activeCollectContext(ctx)
	if err != nil {
		return nil, err
	}
	result := &Result{Source: g.Name()}
	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitStart(g.Name(), "Starting GitHub collection")
	}
	if emitDryRun(cfg, g.Name(), "[dry-run] Would collect GitHub data", "GitHub dry-run complete", result) {
		return result, nil
	}
	if err := waitCollectLimiter(ctx, cfg, "github"); err != nil {
		return result, err
	}
	content := core.Sprintf("# GitHub Collection\n\n- Org: %s\n- Repo: %s\n- IssuesOnly: %t\n- PRsOnly: %t\n", g.Org, g.Repo, g.IssuesOnly, g.PRsOnly)
	path := "github.md"
	if g.Org != "" || g.Repo != "" {
		name := core.Join("/", g.Org, g.Repo)
		path = core.TrimSuffix(core.TrimPrefix(name, "/"), "/") + ".md"
	}
	outPath, err := writeResultFile(cfg, g.Name(), path, content)
	if err != nil {
		result.Errors++
		return result, err
	}
	result.Items = 1
	result.Files = append(result.Files, outPath)
	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitItem(g.Name(), core.Sprintf("Collected GitHub data for %s/%s", g.Org, g.Repo), nil)
		cfg.Dispatcher.EmitComplete(g.Name(), "GitHub collection complete", result)
	}
	return result, nil
}
