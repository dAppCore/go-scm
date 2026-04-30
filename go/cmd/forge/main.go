// SPDX-License-Identifier: EUPL-1.2

package main

import (
	core "dappco.re/go"
	"dappco.re/go/scm/forge"
)

func main() {
	result := newApp()
	if !result.OK {
		core.Error("forge setup failed", "err", result.Value)
		core.Exit(1)
		return
	}
	result.Value.(*core.Core).Run()
}

func newApp() core.Result {
	app := core.New(core.WithOption("name", "forge"))
	app.App().Version = "dev"

	if r := app.Command("auth", core.Command{Action: auth}); !r.OK {
		return r
	}
	if r := app.Command("repos", core.Command{Action: repos}); !r.OK {
		return r
	}

	return core.Ok(app)
}

func auth(opts core.Options) core.Result {
	if wantsHelp(opts) {
		core.Print(nil, "usage: forge auth [--url=URL] [--token=TOKEN]")
		return core.Ok(nil)
	}

	client, err := forge.NewFromConfig(opts.String("url"), opts.String("token"))
	if err != nil {
		return core.Fail(err)
	}
	user, err := client.GetCurrentUser()
	if err != nil {
		return core.Fail(err)
	}
	if user != nil {
		core.Print(nil, "%s", user.UserName)
	}
	return core.Ok(nil)
}

func repos(opts core.Options) core.Result {
	if wantsHelp(opts) {
		core.Print(nil, "usage: forge repos [--org=ORG] [--url=URL] [--token=TOKEN]")
		return core.Ok(nil)
	}

	client, err := forge.NewFromConfig(opts.String("url"), opts.String("token"))
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

func wantsHelp(opts core.Options) bool {
	return opts.Bool("help") || opts.Bool("h")
}
