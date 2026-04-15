// SPDX-License-Identifier: EUPL-1.2

package manifest

import (
	"crypto/ed25519"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignAndVerify_Good(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	m := &Manifest{
		Code:    "test-app",
		Name:    "Test App",
		Version: "1.0.0",
		Layout:  "HLCRF",
		Slots:   map[string]string{"C": "main"},
	}

	err = Sign(m, priv)
	require.NoError(t, err)
	assert.NotEmpty(t, m.Sign)
	assert.Equal(t, hex.EncodeToString(pub), m.SignKey)

	ok, err := Verify(m, pub)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestVerify_Good_UsesEmbeddedSignKey_Good(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	m := &Manifest{
		Code:    "test-app",
		Name:    "Test App",
		Version: "1.0.0",
	}

	require.NoError(t, Sign(m, priv))

	ok, err := Verify(m, nil)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestVerify_Bad_Tampered_Good(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	m := &Manifest{Code: "test-app", Version: "1.0.0"}
	_ = Sign(m, priv)

	m.Code = "evil-app" // tamper

	ok, err := Verify(m, pub)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestVerify_Bad_Unsigned_Good(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(nil)
	m := &Manifest{Code: "test-app"}

	ok, err := Verify(m, pub)
	assert.Error(t, err)
	assert.False(t, ok)
}

func TestSign_Bad_InvalidPrivateKey_Good(t *testing.T) {
	m := &Manifest{Code: "test-app", Version: "1.0.0"}

	err := Sign(m, ed25519.PrivateKey([]byte("short")))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid private key length")
	assert.Empty(t, m.Sign)
}

func TestSign_Bad_NilManifest_Good(t *testing.T) {
	err := Sign(nil, ed25519.PrivateKey(make([]byte, ed25519.PrivateKeySize)))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil manifest")
}

func TestVerify_Bad_NilManifest_Good(t *testing.T) {
	ok, err := Verify(nil, ed25519.PublicKey(make([]byte, ed25519.PublicKeySize)))
	assert.Error(t, err)
	assert.False(t, ok)
	assert.Contains(t, err.Error(), "nil manifest")
}

func TestVerify_Bad_InvalidPublicKey_Good(t *testing.T) {
	m := &Manifest{Code: "test-app", Sign: "c2ln"}

	ok, err := Verify(m, ed25519.PublicKey([]byte("short")))
	assert.Error(t, err)
	assert.False(t, ok)
	assert.Contains(t, err.Error(), "invalid public key length")
}
