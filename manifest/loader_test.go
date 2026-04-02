// SPDX-License-Identifier: EUPL-1.2

package manifest

import (
	"crypto/ed25519"
	"testing"

	"dappco.re/go/core/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Good(t *testing.T) {
	fs := io.NewMockMedium()
	fs.Files[".core/manifest.yaml"] = `
code: test-app
name: Test App
version: 1.0.0
layout: HLCRF
slots:
  C: main-content
`
	m, err := Load(fs, ".")
	require.NoError(t, err)
	assert.Equal(t, "test-app", m.Code)
	assert.Equal(t, "main-content", m.Slots["C"])
}

func TestLoad_Bad_NoManifest_Good(t *testing.T) {
	fs := io.NewMockMedium()
	_, err := Load(fs, ".")
	assert.Error(t, err)
}

func TestLoadVerified_Good(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	m := &Manifest{
		Code: "signed-app", Name: "Signed", Version: "1.0.0",
		Layout: "HLCRF", Slots: map[string]string{"C": "main"},
	}
	_ = Sign(m, priv)

	raw, _ := MarshalYAML(m)
	fs := io.NewMockMedium()
	fs.Files[".core/manifest.yaml"] = string(raw)

	loaded, err := LoadVerified(fs, ".", pub)
	require.NoError(t, err)
	assert.Equal(t, "signed-app", loaded.Code)
}

func TestLoadVerified_Bad_Tampered_Good(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	m := &Manifest{Code: "app", Version: "1.0.0"}
	_ = Sign(m, priv)

	raw, _ := MarshalYAML(m)
	tampered := "code: evil\n" + string(raw)[6:]
	fs := io.NewMockMedium()
	fs.Files[".core/manifest.yaml"] = tampered

	_, err := LoadVerified(fs, ".", pub)
	assert.Error(t, err)
}

func TestLoadVerified_Bad_InvalidPublicKey_Good(t *testing.T) {
	fs := io.NewMockMedium()
	fs.Files[".core/manifest.yaml"] = `
code: signed-app
name: Signed
version: 1.0.0
sign: c2ln
`

	_, err := LoadVerified(fs, ".", ed25519.PublicKey([]byte("short")))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid public key length")
}
