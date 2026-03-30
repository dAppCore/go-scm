// SPDX-License-Identifier: EUPL-1.2

// Package scm provides CLI commands for manifest compilation and marketplace
// index generation.
//
// Commands:
//   - compile: Compile .core/manifest.yaml into core.json
//   - index:   Build marketplace index from repository directories
//   - export:  Export a compiled manifest as JSON to stdout
package scm

import (
	"forge.lthn.ai/core/cli/pkg/cli"
)

func init() {
	cli.RegisterCommands(AddScmCommands)
}

// Style aliases from shared package.
var (
	successStyle = cli.SuccessStyle
	errorStyle   = cli.ErrorStyle
	dimStyle     = cli.DimStyle
	valueStyle   = cli.ValueStyle
	numberStyle  = cli.NumberStyle
)

// AddScmCommands registers the 'scm' command and all subcommands.
// Usage: AddScmCommands(...)
func AddScmCommands(root *cli.Command) {
	scmCmd := &cli.Command{
		Use:   "scm",
		Short: "SCM manifest and marketplace operations",
		Long:  "Compile manifests, build marketplace indexes, and export distribution metadata.",
	}
	root.AddCommand(scmCmd)

	addCompileCommand(scmCmd)
	addIndexCommand(scmCmd)
	addExportCommand(scmCmd)
}
