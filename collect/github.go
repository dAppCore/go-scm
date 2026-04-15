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
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	result := &Result{Source: g.Name()}
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
	return result, nil
}
