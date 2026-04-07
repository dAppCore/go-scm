// SPDX-License-Identifier: EUPL-1.2

package scm

import (
	"crypto/ed25519"
	"encoding/hex"

	"dappco.re/go/core/io"
	"dappco.re/go/core/scm/manifest"
	"forge.lthn.ai/core/cli/pkg/cli"
)

func addSignCommand(parent *cli.Command) {
	var (
		dir     string
		signKey string
	)

	cmd := &cli.Command{
		Use:   "sign",
		Short: "Sign manifest.yaml with a private key",
		Long:  "Read .core/manifest.yaml, attach an ed25519 signature, and write the signed manifest back to disk.",
		RunE: func(cmd *cli.Command, args []string) error {
			return runSign(dir, signKey)
		},
	}

	cmd.Flags().StringVarP(&dir, "dir", "d", ".", "Project root directory")
	cmd.Flags().StringVar(&signKey, "sign-key", "", "Hex-encoded ed25519 private key")

	parent.AddCommand(cmd)
}

func runSign(dir, signKeyHex string) error {
	if signKeyHex == "" {
		return cli.Err("sign key is required")
	}

	medium, err := io.NewSandboxed(dir)
	if err != nil {
		return cli.WrapVerb(err, "open", dir)
	}

	m, err := manifest.Load(medium, ".")
	if err != nil {
		return cli.WrapVerb(err, "load", "manifest")
	}

	keyBytes, err := hex.DecodeString(signKeyHex)
	if err != nil {
		return cli.WrapVerb(err, "decode", "sign key")
	}
	if len(keyBytes) != ed25519.PrivateKeySize {
		return cli.Err("sign key must be %d bytes when decoded", ed25519.PrivateKeySize)
	}

	if err := manifest.Sign(m, ed25519.PrivateKey(keyBytes)); err != nil {
		return err
	}

	data, err := manifest.MarshalYAML(m)
	if err != nil {
		return cli.WrapVerb(err, "marshal", "manifest")
	}

	if err := medium.Write(".core/manifest.yaml", string(data)); err != nil {
		return cli.WrapVerb(err, "write", ".core/manifest.yaml")
	}

	cli.Blank()
	cli.Print("  %s %s\n", successStyle.Render("signed"), valueStyle.Render(m.Code))
	cli.Print("  %s %s\n", dimStyle.Render("output:"), valueStyle.Render(".core/manifest.yaml"))
	cli.Blank()

	return nil
}

func addVerifyCommand(parent *cli.Command) {
	var (
		dir       string
		publicKey string
	)

	cmd := &cli.Command{
		Use:   "verify",
		Short: "Verify manifest signature with a public key",
		Long:  "Read .core/manifest.yaml and verify its ed25519 signature against a public key.",
		RunE: func(cmd *cli.Command, args []string) error {
			return runVerify(dir, publicKey)
		},
	}

	cmd.Flags().StringVarP(&dir, "dir", "d", ".", "Project root directory")
	cmd.Flags().StringVar(&publicKey, "public-key", "", "Hex-encoded ed25519 public key")

	parent.AddCommand(cmd)
}

func runVerify(dir, publicKeyHex string) error {
	if publicKeyHex == "" {
		return cli.Err("public key is required")
	}

	medium, err := io.NewSandboxed(dir)
	if err != nil {
		return cli.WrapVerb(err, "open", dir)
	}

	m, err := manifest.Load(medium, ".")
	if err != nil {
		return cli.WrapVerb(err, "load", "manifest")
	}

	keyBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return cli.WrapVerb(err, "decode", "public key")
	}
	if len(keyBytes) != ed25519.PublicKeySize {
		return cli.Err("public key must be %d bytes when decoded", ed25519.PublicKeySize)
	}

	valid, err := manifest.Verify(m, ed25519.PublicKey(keyBytes))
	if err != nil {
		return cli.WrapVerb(err, "verify", "manifest")
	}
	if !valid {
		return cli.Err("signature verification failed for %s", m.Code)
	}

	cli.Blank()
	cli.Success("Signature verified")
	cli.Print("  %s %s\n", dimStyle.Render("code:"), valueStyle.Render(m.Code))
	cli.Blank()

	return nil
}
