// SPDX-License-Identifier: EUPL-1.2

package scm

import (
	"crypto/ed25519"
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
	os "dappco.re/go/core/scm/internal/ax/osx"
	"encoding/hex"
	"testing"

	"dappco.re/go/core/io"
	"dappco.re/go/core/scm/manifest"
	"dappco.re/go/core/cli/pkg/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunSign_Good_WritesSignedManifest_Good(t *testing.T) {
	dir := t.TempDir()
	coreDir := filepath.Join(dir, ".core")
	require.NoError(t, os.MkdirAll(coreDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(coreDir, "manifest.yaml"), []byte(`
code: signed-cli
name: Signed CLI
version: 1.0.0
`), 0644))

	pub, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	err = runSign(dir, hex.EncodeToString(priv))
	require.NoError(t, err)

	raw, err := io.Local.Read(filepath.Join(dir, ".core", "manifest.yaml"))
	require.NoError(t, err)

	m, err := manifest.Parse([]byte(raw))
	require.NoError(t, err)
	assert.Equal(t, "signed-cli", m.Code)
	assert.NotEmpty(t, m.Sign)

	valid, err := manifest.Verify(m, pub)
	require.NoError(t, err)
	assert.True(t, valid)
}

func TestRunVerify_Good_ValidSignature_Good(t *testing.T) {
	dir := t.TempDir()
	coreDir := filepath.Join(dir, ".core")
	require.NoError(t, os.MkdirAll(coreDir, 0755))

	pub, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	m := &manifest.Manifest{
		Code:    "verified-cli",
		Name:    "Verified CLI",
		Version: "1.0.0",
	}
	require.NoError(t, manifest.Sign(m, priv))

	data, err := manifest.MarshalYAML(m)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(coreDir, "manifest.yaml"), data, 0644))

	err = runVerify(dir, hex.EncodeToString(pub))
	require.NoError(t, err)
}

func TestAddScmCommands_Good_SignAndVerifyRegistered_Good(t *testing.T) {
	root := &cli.Command{Use: "root"}

	AddScmCommands(root)

	var scmCmd *cli.Command
	for _, cmd := range root.Commands() {
		if cmd.Name() == "scm" {
			scmCmd = cmd
			break
		}
	}
	require.NotNil(t, scmCmd)

	var signCmd *cli.Command
	var verifyCmd *cli.Command
	for _, cmd := range scmCmd.Commands() {
		switch cmd.Name() {
		case "sign":
			signCmd = cmd
		case "verify":
			verifyCmd = cmd
		}
	}
	require.NotNil(t, signCmd)
	require.NotNil(t, verifyCmd)
	assert.NotNil(t, signCmd.Flags().Lookup("sign-key"))
	assert.NotNil(t, verifyCmd.Flags().Lookup("public-key"))
}
