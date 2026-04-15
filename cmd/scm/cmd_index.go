// SPDX-License-Identifier: EUPL-1.2

package scm

import (
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
	fmt "dappco.re/go/core/scm/internal/ax/fmtx"
	os "dappco.re/go/core/scm/internal/ax/osx"

	"dappco.re/go/core/io"
	"dappco.re/go/core/scm/marketplace"
	"dappco.re/go/core/cli/pkg/cli"
)

func addIndexCommand(parent *cli.Command) {
	var (
		dirs     []string
		output   string
		forgeURL string
		org      string
	)

	cmd := &cli.Command{
		Use:   "index",
		Short: "Build marketplace index from directories",
		Long:  "Scan directories for core.json or .core/manifest.yaml files and generate a marketplace index.json.",
		RunE: func(cmd *cli.Command, args []string) error {
			if len(dirs) == 0 {
				dirs = []string{"."}
			}
			return runIndex(dirs, output, forgeURL, org)
		},
	}

	cmd.Flags().StringArrayVarP(&dirs, "dir", "d", nil, "Directories to scan (repeatable, default: current directory)")
	cmd.Flags().StringVarP(&output, "output", "o", "index.json", "Output path for the index file")
	cmd.Flags().StringVar(&forgeURL, "forge-url", "", "Forge base URL for repo links (e.g. https://forge.lthn.ai)")
	cmd.Flags().StringVar(&forgeURL, "base-url", "", "Deprecated alias for --forge-url")
	cmd.Flags().StringVar(&org, "org", "", "Organisation for repo links")

	parent.AddCommand(cmd)
}

func runIndex(dirs []string, output, forgeURL, org string) error {
	repoPaths, err := expandIndexRepoPaths(dirs)
	if err != nil {
		return err
	}

	idx, err := marketplace.BuildIndex(io.Local, repoPaths, marketplace.IndexOptions{
		ForgeURL: forgeURL,
		Org:      org,
	})
	if err != nil {
		return cli.WrapVerb(err, "build", "index")
	}

	absOutput, err := filepath.Abs(output)
	if err != nil {
		return cli.WrapVerb(err, "resolve", output)
	}
	if err := marketplace.WriteIndex(io.Local, absOutput, idx); err != nil {
		return err
	}

	cli.Blank()
	cli.Print("  %s %s\n", successStyle.Render("index built"), valueStyle.Render(output))
	cli.Print("  %s %s\n", dimStyle.Render("modules:"), numberStyle.Render(fmt.Sprintf("%d", len(idx.Modules))))
	cli.Blank()

	return nil
}

func expandIndexRepoPaths(dirs []string) ([]string, error) {
	var repoPaths []string

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, cli.WrapVerb(err, "read", dir)
		}

		repoPaths = append(repoPaths, dir)

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			repoPaths = append(repoPaths, filepath.Join(dir, entry.Name()))
		}
	}

	return repoPaths, nil
}
