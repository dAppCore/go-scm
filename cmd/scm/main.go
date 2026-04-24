// SPDX-License-Identifier: EUPL-1.2

package main

import (
	core "dappco.re/go/core"
	scm "dappco.re/go/scm"
)

func main() {
	newApp().Run()
}

func newApp() *core.Core {
	app := core.New(
		core.WithOption("name", "scm"),
		core.WithService(scm.NewCoreService(scm.Options{})),
	)
	app.App().Version = "dev"

	app.Command("health", core.Command{Action: health(app)})
	app.Command("dev/health", core.Command{Action: health(app)})

	return app
}

func health(app *core.Core) core.CommandAction {
	return func(_ core.Options) core.Result {
		core.Print(nil, "scm %s", app.App().Version)
		core.Print(nil, "services: %d", len(app.Services()))
		core.Print(nil, "commands: %d", len(app.Commands()))
		return core.Result{OK: true}
	}
}
