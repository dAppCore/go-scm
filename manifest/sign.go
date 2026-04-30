// SPDX-License-Identifier: EUPL-1.2

package manifest

import (
	"crypto/ed25519" // intrinsic
	"encoding/base64"
	"encoding/json"
	"errors"

	core "dappco.re/go"
)

func canonicalManifestBytes(m *Manifest) ([]byte, error) {
	if err := validateManifest(m); err != nil {
		return nil, err
	}
	cp := *m
	cp.Sign = ""
	cp.SignKey = ""
	return json.Marshal(cp)
}

func Sign(m *Manifest, payload []byte, priv ed25519.PrivateKey) error {
	if m == nil {
		return errors.New("manifest.Sign: manifest is required")
	}
	if len(priv) != ed25519.PrivateKeySize {
		return errors.New("manifest.Sign: private key is required")
	}
	sig := ed25519.Sign(priv, payload)
	m.Sign = base64.StdEncoding.EncodeToString(sig)
	return nil
}

func Verify(m *Manifest, payload []byte) error {
	if m == nil {
		return core.E("manifest.Verify", "manifest is required", nil)
	}
	if m.SignKey == "" {
		return core.E("manifest.Verify", "sign key is required", nil)
	}
	if m.Sign == "" {
		return core.E("manifest.Verify", "signature is required", nil)
	}
	pub, err := base64.StdEncoding.DecodeString(m.SignKey)
	if err != nil {
		return core.E("manifest.Verify", "decode sign key", err)
	}
	sig, err := base64.StdEncoding.DecodeString(m.Sign)
	if err != nil {
		return core.E("manifest.Verify", "decode signature", err)
	}
	if len(pub) != ed25519.PublicKeySize {
		return core.E("manifest.Verify", "invalid sign key", nil)
	}
	if len(sig) != ed25519.SignatureSize {
		return core.E("manifest.Verify", "invalid signature", nil)
	}
	if !ed25519.Verify(ed25519.PublicKey(pub), payload, sig) {
		return core.E("manifest.Verify", "signature verification failed", nil)
	}
	return nil
}
