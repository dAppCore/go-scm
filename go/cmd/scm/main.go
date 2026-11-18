// SPDX-License-Identifier: EUPL-1.2

package main

import (
	core "github.com/dappcore/go"
	scm "github.com/dappcore/go-scm/go"
	compilecmd "github.com/dappcore/go-scm/go/cmd/compile"
	pkgcmd "github.com/dappcore/go-scm/go/cmd/pkg"
	signcmd "github.com/dappcore/go-scm/go/cmd/sign"
	verifycmd "github.com/dappcore/go-scm/go/cmd/verify"
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
		core.WithCli(),
		core.WithService(scm.NewService(scm.Options{})),
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
