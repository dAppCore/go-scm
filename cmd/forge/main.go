// SPDX-License-Identifier: EUPL-1.2

package main

import (
	core "dappco.re/go/core"
	"dappco.re/go/scm/forge"
)

func main() {
	newApp().Run()
}

func newApp() *core.Core {
	app := core.New(core.WithOption("name", "forge"))
	app.App().Version = "dev"

	app.Command("auth", core.Command{Action: auth})
	app.Command("repos", core.Command{Action: repos})

	return app
}

func auth(opts core.Options) core.Result {
	if wantsHelp(opts) {
		core.Print(nil, "usage: forge auth [--url=URL] [--token=TOKEN]")
		return core.Result{OK: true}
	}

	client, err := forge.NewFromConfig(opts.String("url"), opts.String("token"))
	if err != nil {
		return core.Result{Value: err, OK: false}
	}
	user, err := client.GetCurrentUser()
	if err != nil {
		return core.Result{Value: err, OK: false}
	}
	if user != nil {
		core.Print(nil, "%s", user.UserName)
	}
	return core.Result{OK: true}
}

func repos(opts core.Options) core.Result {
	if wantsHelp(opts) {
		core.Print(nil, "usage: forge repos [--org=ORG] [--url=URL] [--token=TOKEN]")
		return core.Result{OK: true}
	}

	client, err := forge.NewFromConfig(opts.String("url"), opts.String("token"))
	if err != nil {
		return core.Result{Value: err, OK: false}
	}

	if org := opts.String("org"); org != "" {
		repositories, err := client.ListOrgRepos(org)
		if err != nil {
			return core.Result{Value: err, OK: false}
		}
		for _, repo := range repositories {
			if repo == nil {
				continue
			}
			core.Print(nil, "%s", repo.FullName)
		}
		return core.Result{OK: true}
	}

	repositories, err := client.ListUserRepos()
	if err != nil {
		return core.Result{Value: err, OK: false}
	}
	for _, repo := range repositories {
		if repo == nil {
			continue
		}
		core.Print(nil, "%s", repo.FullName)
	}
	return core.Result{OK: true}
}

func wantsHelp(opts core.Options) bool {
	return opts.Bool("help") || opts.Bool("h")
}
