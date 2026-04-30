// SPDX-License-Identifier: EUPL-1.2

package main

import (
	"strconv"
	"strings"

	core "dappco.re/go"
	"dappco.re/go/scm/gitea"
)

func main() {
	newApp().Run()
}

func newApp() *core.Core {
	app := core.New(core.WithOption("name", "gitea"))
	app.App().Version = "dev"

	_ = app.Command("repos", core.Command{Action: repos})
	_ = app.Command("issues", core.Command{Action: issues})

	return app
}

func repos(opts core.Options) core.Result {
	if wantsHelp(opts) {
		core.Print(nil, "usage: gitea repos [--org=ORG] [--url=URL] [--token=TOKEN]")
		return core.Ok(nil)
	}

	client, err := gitea.NewFromConfig(opts.String("url"), opts.String("token"))
	if err != nil {
		return core.Fail(err)
	}

	if org := opts.String("org"); org != "" {
		repositories, err := client.ListOrgRepos(org)
		if err != nil {
			return core.Fail(err)
		}
		for _, repo := range repositories {
			if repo == nil {
				continue
			}
			core.Print(nil, "%s", repo.FullName)
		}
		return core.Ok(nil)
	}

	repositories, err := client.ListUserRepos()
	if err != nil {
		return core.Fail(err)
	}
	for _, repo := range repositories {
		if repo == nil {
			continue
		}
		core.Print(nil, "%s", repo.FullName)
	}
	return core.Ok(nil)
}

func issues(opts core.Options) core.Result {
	if wantsHelp(opts) {
		core.Print(nil, "usage: gitea issues OWNER/REPO [--state=open|closed|all] [--limit=N] [--url=URL] [--token=TOKEN]")
		return core.Ok(nil)
	}

	owner, repo := splitRepo(opts.String("_arg"))
	if owner == "" || repo == "" {
		owner, repo = opts.String("owner"), opts.String("repo")
	}
	if owner == "" || repo == "" {
		return core.Fail(core.E("gitea.issues", "repository must be OWNER/REPO", nil))
	}

	client, err := gitea.NewFromConfig(opts.String("url"), opts.String("token"))
	if err != nil {
		return core.Fail(err)
	}
	list, err := client.ListIssues(owner, repo, gitea.ListIssuesOpts{
		State: opts.String("state"),
		Limit: intOption(opts, "limit"),
	})
	if err != nil {
		return core.Fail(err)
	}
	for _, issue := range list {
		if issue == nil {
			continue
		}
		core.Print(nil, "#%d %s", issue.Index, issue.Title)
	}
	return core.Ok(nil)
}

func splitRepo(value string) (string, string) {
	parts := strings.SplitN(value, "/", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

func intOption(opts core.Options, key string) int {
	if value := opts.Int(key); value != 0 {
		return value
	}
	value := opts.String(key)
	if value == "" {
		return 0
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return n
}

func wantsHelp(opts core.Options) bool {
	return opts.Bool("help") || opts.Bool("h")
}
