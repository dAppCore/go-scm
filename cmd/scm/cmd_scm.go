// SPDX-License-Identifier: EUPL-1.2

// Package scm provides CLI commands for manifest compilation and marketplace
// index generation.
//
// Commands:
//   - compile: Compile .core/manifest.yaml into core.json
//   - dev:     Workspace registry operations (health, pull, push, impact)
//   - index:   Build marketplace index from repository directories
//   - export:  Export a compiled manifest as JSON to stdout
//   - pkg:     Marketplace search/install/update/list/publish
//   - sign:    Sign .core/manifest.yaml with an ed25519 private key
//   - verify:  Verify a manifest signature with an ed25519 public key
package scm

import (
	"dappco.re/go/core/cli/pkg/cli"
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
	addDevCommand(scmCmd)
	addIndexCommand(scmCmd)
	addExportCommand(scmCmd)
	addPackageCommand(scmCmd)
	addSignCommand(scmCmd)
	addVerifyCommand(scmCmd)

	// RFC-facing aliases live at the root CLI as `core dev ...` and
	// `core pkg ...`. Preserve the nested `core scm ...` surface for
	// compatibility, but only add the root aliases when they are unused.
	if !hasCommand(root, "dev") {
		addDevCommand(root)
	}
	if !hasCommand(root, "pkg") {
		addPackageCommand(root)
	}
}

func hasCommand(parent *cli.Command, name string) bool {
	for _, cmd := range parent.Commands() {
		if cmd != nil && cmd.Name() == name {
			return true
		}
	}
	return false
}
