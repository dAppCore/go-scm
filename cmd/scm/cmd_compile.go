package scm

import (
	"crypto/ed25519"
	"encoding/hex"
	"os/exec"
	"strings"

	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/go-io"
	"forge.lthn.ai/core/go-scm/manifest"
)

func addCompileCommand(parent *cli.Command) {
	var (
		dir     string
		signKey string
		builtBy string
	)

	cmd := &cli.Command{
		Use:   "compile",
		Short: "Compile manifest.yaml into core.json",
		Long:  "Read .core/manifest.yaml, attach build metadata (commit, tag), and write core.json to the project root.",
		RunE: func(cmd *cli.Command, args []string) error {
			return runCompile(dir, signKey, builtBy)
		},
	}

	cmd.Flags().StringVarP(&dir, "dir", "d", ".", "Project root directory")
	cmd.Flags().StringVar(&signKey, "sign-key", "", "Hex-encoded ed25519 private key for signing")
	cmd.Flags().StringVar(&builtBy, "built-by", "core scm compile", "Builder identity")

	parent.AddCommand(cmd)
}

func runCompile(dir, signKeyHex, builtBy string) error {
	medium, err := io.NewSandboxed(dir)
	if err != nil {
		return cli.WrapVerb(err, "open", dir)
	}

	m, err := manifest.Load(medium, ".")
	if err != nil {
		return cli.WrapVerb(err, "load", "manifest")
	}

	opts := manifest.CompileOptions{
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

	if err := manifest.WriteCompiled(medium, ".", cm); err != nil {
		return err
	}

	cli.Blank()
	cli.Print("  %s %s\n", successStyle.Render("compiled"), valueStyle.Render(m.Code))
	cli.Print("  %s %s\n", dimStyle.Render("version:"), valueStyle.Render(m.Version))
	if opts.Commit != "" {
		cli.Print("  %s %s\n", dimStyle.Render("commit:"), valueStyle.Render(opts.Commit))
	}
	if opts.Tag != "" {
		cli.Print("  %s %s\n", dimStyle.Render("tag:"), valueStyle.Render(opts.Tag))
	}
	cli.Print("  %s %s\n", dimStyle.Render("output:"), valueStyle.Render("core.json"))
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
