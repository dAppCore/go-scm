package scm

import (
	"fmt"
	"path/filepath"

	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/go-scm/marketplace"
)

func addIndexCommand(parent *cli.Command) {
	var (
		dirs    []string
		output  string
		baseURL string
		org     string
	)

	cmd := &cli.Command{
		Use:   "index",
		Short: "Build marketplace index from directories",
		Long:  "Scan directories for core.json or .core/manifest.yaml files and generate a marketplace index.json.",
		RunE: func(cmd *cli.Command, args []string) error {
			if len(dirs) == 0 {
				dirs = []string{"."}
			}
			return runIndex(dirs, output, baseURL, org)
		},
	}

	cmd.Flags().StringArrayVarP(&dirs, "dir", "d", nil, "Directories to scan (repeatable, default: current directory)")
	cmd.Flags().StringVarP(&output, "output", "o", "index.json", "Output path for the index file")
	cmd.Flags().StringVar(&baseURL, "base-url", "", "Base URL for repo links (e.g. https://forge.lthn.ai)")
	cmd.Flags().StringVar(&org, "org", "", "Organisation for repo links")

	parent.AddCommand(cmd)
}

func runIndex(dirs []string, output, baseURL, org string) error {
	b := &marketplace.Builder{
		BaseURL: baseURL,
		Org:     org,
	}

	idx, err := b.BuildFromDirs(dirs...)
	if err != nil {
		return cli.WrapVerb(err, "build", "index")
	}

	absOutput, _ := filepath.Abs(output)
	if err := marketplace.WriteIndex(absOutput, idx); err != nil {
		return err
	}

	cli.Blank()
	cli.Print("  %s %s\n", successStyle.Render("index built"), valueStyle.Render(output))
	cli.Print("  %s %s\n", dimStyle.Render("modules:"), numberStyle.Render(fmt.Sprintf("%d", len(idx.Modules))))
	cli.Blank()

	return nil
}
