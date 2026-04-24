// SPDX-License-Identifier: EUPL-1.2

package main

import (
	"context"

	core "dappco.re/go/core"
	"dappco.re/go/scm/collect"
)

func main() {
	newApp().Run()
}

func newApp() *core.Core {
	app := core.New(core.WithOption("name", "collect"))
	app.App().Version = "dev"

	app.Command("github", core.Command{Action: github})
	app.Command("market", core.Command{Action: market})
	app.Command("papers", core.Command{Action: papers})

	return app
}

func github(opts core.Options) core.Result {
	if wantsHelp(opts) {
		core.Print(nil, "usage: collect github [--org=ORG] [--repo=REPO] [--out=DIR] [--dry-run]")
		return core.Result{OK: true}
	}
	return runCollector(opts, &collect.GitHubCollector{
		Org:        opts.String("org"),
		Repo:       opts.String("repo"),
		IssuesOnly: opts.Bool("issues-only"),
		PRsOnly:    opts.Bool("prs-only"),
	})
}

func market(opts core.Options) core.Result {
	if wantsHelp(opts) {
		core.Print(nil, "usage: collect market [--coin=COIN] [--from=YYYY-MM-DD] [--historical] [--out=DIR] [--dry-run]")
		return core.Result{OK: true}
	}
	return runCollector(opts, &collect.MarketCollector{
		CoinID:     option(opts, "coin", "bitcoin"),
		Historical: opts.Bool("historical"),
		FromDate:   opts.String("from"),
	})
}

func papers(opts core.Options) core.Result {
	if wantsHelp(opts) {
		core.Print(nil, "usage: collect papers [--source=all|iacr|arxiv] [--category=CAT] [--query=QUERY] [--out=DIR] [--dry-run]")
		return core.Result{OK: true}
	}
	return runCollector(opts, &collect.PapersCollector{
		Source:   opts.String("source"),
		Category: opts.String("category"),
		Query:    opts.String("query"),
	})
}

func runCollector(opts core.Options, collector collect.Collector) core.Result {
	cfg := collect.NewConfig(option(opts, "out", "collect"))
	cfg.DryRun = opts.Bool("dry-run") || opts.Bool("n")
	cfg.Verbose = opts.Bool("verbose") || opts.Bool("v")

	result, err := collector.Collect(context.Background(), cfg)
	if err != nil {
		return core.Result{Value: err, OK: false}
	}
	if result != nil {
		core.Print(nil, "%s: %d items, %d errors, %d skipped", result.Source, result.Items, result.Errors, result.Skipped)
		for _, file := range result.Files {
			core.Print(nil, "%s", file)
		}
	}
	return core.Result{OK: true}
}

func option(opts core.Options, key, fallback string) string {
	if value := opts.String(key); value != "" {
		return value
	}
	return fallback
}

func wantsHelp(opts core.Options) bool {
	return opts.Bool("help") || opts.Bool("h")
}
