// SPDX-License-Identifier: EUPL-1.2

package main

import (
	core "dappco.re/go"
	scm "dappco.re/go/scm"
	compilecmd "dappco.re/go/scm/cmd/compile"
	pkgcmd "dappco.re/go/scm/cmd/pkg"
	signcmd "dappco.re/go/scm/cmd/sign"
	verifycmd "dappco.re/go/scm/cmd/verify"
)

func main() {
	result := newApp()
	if !result.OK {
		core.Error("scm setup failed", "err", result.Value)
		core.Exit(1)
		return
	}
	result.Value.(*core.Core).Run()
}

func newApp() core.Result {
	app := core.New(
		core.WithOption("name", "scm"),
		core.WithService(scm.NewCoreService(scm.Options{})),
	)
	app.App().Version = "dev"

	if r := app.Command("health", core.Command{Action: health(app)}); !r.OK {
		return r
	}
	if r := app.Command("dev/health", core.Command{Action: health(app)}); !r.OK {
		return r
	}
	if r := compilecmd.Register(app); !r.OK {
		return r
	}
	if r := signcmd.Register(app); !r.OK {
		return r
	}
	if r := verifycmd.Register(app); !r.OK {
		return r
	}
	if r := pkgcmd.Register(app); !r.OK {
		return r
	}

	return core.Ok(app)
}

func health(app *core.Core) core.CommandAction {
	return func(_ core.Options) core.Result {
		core.Print(nil, "scm %s", app.App().Version)
		core.Print(nil, "services: %d", len(app.Services()))
		core.Print(nil, "commands: %d", len(app.Commands()))
		return core.Ok(nil)
	}
}
