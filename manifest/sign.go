// SPDX-License-Identifier: EUPL-1.2

package manifest

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"strings"

	coreerr "dappco.re/go/core/log"
	"gopkg.in/yaml.v3"
)

// signable returns the canonical bytes to sign (manifest without sign field).
func signable(m *Manifest) ([]byte, error) {
	tmp := *m
	tmp.Sign = ""
	tmp.SignKey = ""
	return yaml.Marshal(&tmp)
}

// Sign computes the ed25519 signature and stores it in m.Sign (base64).
// Usage: Sign(...)
func Sign(m *Manifest, priv ed25519.PrivateKey) error {
	if m == nil {
		return coreerr.E("manifest.Sign", "nil manifest", nil)
	}
	if len(priv) != ed25519.PrivateKeySize {
		return coreerr.E("manifest.Sign", "invalid private key length", nil)
	}

	msg, err := signable(m)
	if err != nil {
		return coreerr.E("manifest.Sign", "marshal failed", err)
	}
	sig := ed25519.Sign(priv, msg)
	m.Sign = base64.StdEncoding.EncodeToString(sig)
	if pub, ok := priv.Public().(ed25519.PublicKey); ok {
		m.SignKey = hex.EncodeToString(pub)
	}
	return nil
}

// Verify checks the ed25519 signature in m.Sign against the public key.
// Usage: Verify(...)
func Verify(m *Manifest, pub ed25519.PublicKey) (bool, error) {
	if m == nil {
		return false, coreerr.E("manifest.Verify", "nil manifest", nil)
	}
	if m.Sign == "" {
		return false, coreerr.E("manifest.Verify", "no signature present", nil)
	}

	if len(pub) == 0 {
		signKey := strings.TrimSpace(m.SignKey)
		if signKey == "" {
			return false, coreerr.E("manifest.Verify", "no public key provided", nil)
		}
		keyBytes, err := hex.DecodeString(signKey)
		if err != nil {
			return false, coreerr.E("manifest.Verify", "decode sign_key failed", err)
		}
		pub = ed25519.PublicKey(keyBytes)
	}
	if len(pub) != ed25519.PublicKeySize {
		return false, coreerr.E("manifest.Verify", "invalid public key length", nil)
	}
	sig, err := base64.StdEncoding.DecodeString(m.Sign)
	if err != nil {
		return false, coreerr.E("manifest.Verify", "decode failed", err)
	}
	msg, err := signable(m)
	if err != nil {
		return false, coreerr.E("manifest.Verify", "marshal failed", err)
	}
	return ed25519.Verify(pub, msg, sig), nil
}
