// SPDX-License-Identifier: EUPL-1.2

package scm

import (
	"crypto/ed25519"
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
	strings "dappco.re/go/core/scm/internal/ax/stringsx"
	"encoding/hex"
	exec "golang.org/x/sys/execabs"

	"dappco.re/go/core/io"
	"dappco.re/go/core/scm/manifest"
	"dappco.re/go/core/cli/pkg/cli"
)

func addCompileCommand(parent *cli.Command) {
	var (
		version string
		dir     string
		signKey string
		builtBy string
		output  string
	)

	cmd := &cli.Command{
		Use:   "compile",
		Short: "Compile manifest.yaml into core.json",
		Long:  "Read .core/manifest.yaml, attach build metadata (commit, tag), and write core.json to the project root or a custom output path.",
		RunE: func(cmd *cli.Command, args []string) error {
			return runCompile(dir, version, signKey, builtBy, output)
		},
	}

	cmd.Flags().StringVarP(&dir, "dir", "d", ".", "Project root directory")
	cmd.Flags().StringVar(&version, "version", "", "Override the manifest version")
	cmd.Flags().StringVar(&signKey, "sign-key", "", "Hex-encoded ed25519 private key for signing")
	cmd.Flags().StringVar(&builtBy, "built-by", "core scm compile", "Builder identity")
	cmd.Flags().StringVarP(&output, "output", "o", "core.json", "Output path for the compiled manifest")

	parent.AddCommand(cmd)
}

func runCompile(dir, version, signKeyHex, builtBy, output string) error {
	medium, err := io.NewSandboxed(dir)
	if err != nil {
		return cli.WrapVerb(err, "open", dir)
	}

	m, err := manifest.Load(medium, ".")
	if err != nil {
		return cli.WrapVerb(err, "load", "manifest")
	}

	opts := manifest.CompileOptions{
		Version: version,
		Commit:  gitCommit(dir),
		Tag:     gitTag(dir),
		BuiltBy: builtBy,
	}

	if signKeyHex != "" {
		keyBytes, err := hex.DecodeString(signKeyHex)
		if err != nil {
			return cli.WrapVerb(err, "decode", "sign key")
		}
		opts.SignKey = ed25519.PrivateKey(keyBytes)
	}

	cm, err := manifest.Compile(m, opts)
	if err != nil {
		return err
	}

	data, err := manifest.MarshalJSON(cm)
	if err != nil {
		return cli.WrapVerb(err, "marshal", "manifest")
	}

	if err := medium.EnsureDir(filepath.Dir(output)); err != nil {
		return cli.WrapVerb(err, "create", filepath.Dir(output))
	}
	if err := medium.Write(output, string(data)); err != nil {
		return cli.WrapVerb(err, "write", output)
	}

	cli.Blank()
	cli.Print("  %s %s\n", successStyle.Render("compiled"), valueStyle.Render(m.Code))
	cli.Print("  %s %s\n", dimStyle.Render("version:"), valueStyle.Render(cm.Version))
	if opts.Commit != "" {
		cli.Print("  %s %s\n", dimStyle.Render("commit:"), valueStyle.Render(opts.Commit))
	}
	if opts.Tag != "" {
		cli.Print("  %s %s\n", dimStyle.Render("tag:"), valueStyle.Render(opts.Tag))
	}
	cli.Print("  %s %s\n", dimStyle.Render("output:"), valueStyle.Render(output))
	cli.Blank()

	return nil
}

// gitCommit returns the current HEAD commit hash, or empty on error.
func gitCommit(dir string) string {
	out, err := exec.Command("git", "-C", dir, "rev-parse", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// gitTag returns the tag pointing at HEAD, or empty if none.
func gitTag(dir string) string {
	out, err := exec.Command("git", "-C", dir, "describe", "--tags", "--exact-match", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
