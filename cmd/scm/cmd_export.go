package scm

import (
	"fmt"

	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/go-io"
	"forge.lthn.ai/core/go-scm/manifest"
)

func addExportCommand(parent *cli.Command) {
	var dir string

	cmd := &cli.Command{
		Use:   "export",
		Short: "Export compiled manifest as JSON",
		Long:  "Read core.json from the project root and print it to stdout. Falls back to compiling .core/manifest.yaml if core.json is not found.",
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

	// Try core.json first.
	cm, err := manifest.LoadCompiled(medium, ".")
	if err != nil {
		// Fall back to compiling from source.
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
