// SPDX-License-Identifier: EUPL-1.2

package scm

import (
	fmt "dappco.re/go/core/scm/internal/ax/fmtx"
	os "dappco.re/go/core/scm/internal/ax/osx"

	"dappco.re/go/core/io"
	"dappco.re/go/core/scm/manifest"
	"dappco.re/go/core/cli/pkg/cli"
)

func addExportCommand(parent *cli.Command) {
	var dir string

	cmd := &cli.Command{
		Use:   "export",
		Short: "Export compiled manifest as JSON",
		Long:  "Read core.json from the project root and print it to stdout. Falls back to compiling .core/manifest.yaml only when core.json is missing.",
		RunE: func(cmd *cli.Command, args []string) error {
			return runExport(dir)
		},
	}

	cmd.Flags().StringVarP(&dir, "dir", "d", ".", "Project root directory")

	parent.AddCommand(cmd)
}

func runExport(dir string) error {
	medium, err := io.NewSandboxed(dir)
	if err != nil {
		return cli.WrapVerb(err, "open", dir)
	}

	var cm *manifest.CompiledManifest

	// Prefer core.json if it exists and is valid.
	if raw, readErr := medium.Read("core.json"); readErr == nil {
		cm, err = manifest.ParseCompiled([]byte(raw))
		if err != nil {
			return err
		}
	} else if !os.IsNotExist(readErr) {
		return cli.WrapVerb(readErr, "read", "core.json")
	} else {
		// Fall back to compiling from source only when the compiled artifact is absent.
		m, loadErr := manifest.Load(medium, ".")
		if loadErr != nil {
			return cli.WrapVerb(loadErr, "load", "manifest")
		}
		cm, err = manifest.Compile(m, manifest.CompileOptions{
			Commit:  gitCommit(dir),
			Tag:     gitTag(dir),
			BuiltBy: "core scm export",
		})
		if err != nil {
			return err
		}
	}

	data, err := manifest.MarshalJSON(cm)
	if err != nil {
		return cli.WrapVerb(err, "marshal", "manifest")
	}

	fmt.Println(string(data))
	return nil
}
