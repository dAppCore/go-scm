// SPDX-License-Identifier: EUPL-1.2

// Package compile wires manifest compilation into the scm CLI.
package compile

import (
	core "dappco.re/go"
	"dappco.re/go/scm/manifest"
)

const usage = "usage: scm compile [--root=DIR] [--manifest=FILE] [--out=FILE] [--commit=SHA] [--tag=TAG] [--built-by=NAME] [--targets=GOOS/GOARCH,...] [--checksums=FILE] [--sha256=HEX]"

// Register attaches the compile command to the parent Core command tree.
//
// The command compiles .core/manifest.yaml into core.json, preserving package
// metadata and adding build fields understood by the manifest package.
func Register(app *core.Core) core.Result {
	if app == nil {
		return core.Fail(core.E("cmd.compile.Register", "core app is required", nil))
	}
	return app.Command("compile", core.Command{Action: run(app)})
}

func run(app *core.Core) core.CommandAction {
	return func(opts core.Options) core.Result {
		if wantsHelp(opts) {
			core.Print(nil, usage)
			return core.Ok(nil)
		}

		root := option(opts, "root", ".")
		manifestPath := option(opts, "manifest", core.PathJoin(root, ".core", "manifest.yaml"))
		outPath := option(opts, "out", core.PathJoin(root, "core.json"))

		raw, err := readFile(app, manifestPath)
		if err != nil {
			return failed(err)
		}
		m, err := manifest.Parse(raw)
		if err != nil {
			return failed(err)
		}

		cm, err := manifest.CompileWithOptions(m, manifest.CompileOptions{
			Commit:  opts.String("commit"),
			Tag:     opts.String("tag"),
			BuiltBy: opts.String("built-by"),
			Build: manifest.BuildInfo{
				Targets:   splitList(option(opts, "targets", opts.String("target"))),
				Checksums: opts.String("checksums"),
				SHA256:    opts.String("sha256"),
			},
		})
		if err != nil {
			return failed(err)
		}

		compiled, err := manifest.MarshalJSON(cm)
		if err != nil {
			return failed(err)
		}
		if r := app.Fs().WriteMode(outPath, string(compiled), 0o600); !r.OK {
			return failed(resultError("cmd.compile.run", "write compiled manifest", r))
		}

		core.Print(nil, "%s", outPath)
		return core.Ok(nil)
	}
}

func option(opts core.Options, key, fallback string) string {
	if value := core.Trim(opts.String(key)); value != "" {
		return value
	}
	return fallback
}

func splitList(value string) []string {
	if core.Trim(value) == "" {
		return nil
	}
	parts := core.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = core.Trim(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

func wantsHelp(opts core.Options) bool {
	return opts.Bool("help") || opts.Bool("h")
}

func failed(err error) core.Result {
	return core.Fail(err)
}

func readFile(app *core.Core, path string) ([]byte, error)  /* v090-result-boundary */ {
	if app == nil {
		return nil, core.E("cmd.compile.readFile", "core app is required", nil)
	}
	r := app.Fs().Read(path)
	if !r.OK {
		return nil, resultError("cmd.compile.readFile", "read file", r)
	}
	raw, ok := r.Value.(string)
	if !ok {
		return nil, core.E("cmd.compile.readFile", "read returned invalid payload", nil)
	}
	return []byte(raw), nil
}

func resultError(op, msg string, r core.Result) error  /* v090-result-boundary */ {
	if err, ok := r.Value.(error); ok {
		return core.E(op, msg, err)
	}
	return core.E(op, msg, nil)
}
