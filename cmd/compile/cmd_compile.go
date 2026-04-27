// SPDX-License-Identifier: EUPL-1.2

// Package compile wires manifest compilation into the scm CLI.
package compile

import (
	"os"
	"path/filepath"
	"strings"

	core "dappco.re/go/core"
	"dappco.re/go/scm/manifest"
)

const usage = "usage: scm compile [--root=DIR] [--manifest=FILE] [--out=FILE] [--commit=SHA] [--tag=TAG] [--built-by=NAME] [--targets=GOOS/GOARCH,...] [--checksums=FILE] [--sha256=HEX]"

// Register attaches the compile command to the parent Core command tree.
//
// The command compiles .core/manifest.yaml into core.json, preserving package
// metadata and adding build fields understood by the manifest package.
func Register(app *core.Core) core.Result {
	if app == nil {
		return core.Result{Value: core.E("cmd.compile.Register", "core app is required", nil), OK: false}
	}
	return app.Command("compile", core.Command{Action: run})
}

func run(opts core.Options) core.Result {
	if wantsHelp(opts) {
		core.Print(nil, usage)
		return core.Result{OK: true}
	}

	root := option(opts, "root", ".")
	manifestPath := option(opts, "manifest", filepath.Join(root, ".core", "manifest.yaml"))
	outPath := option(opts, "out", filepath.Join(root, "core.json"))

	raw, err := os.ReadFile(manifestPath)
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
	if err := mkdirParent(outPath); err != nil {
		return failed(err)
	}
	if err := os.WriteFile(outPath, compiled, 0o600); err != nil {
		return failed(err)
	}

	core.Print(nil, "%s", outPath)
	return core.Result{OK: true}
}

func option(opts core.Options, key, fallback string) string {
	if value := strings.TrimSpace(opts.String(key)); value != "" {
		return value
	}
	return fallback
}

func splitList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

func mkdirParent(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

func wantsHelp(opts core.Options) bool {
	return opts.Bool("help") || opts.Bool("h")
}

func failed(err error) core.Result {
	return core.Result{Value: err, OK: false}
}
