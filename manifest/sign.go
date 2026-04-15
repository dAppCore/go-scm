// SPDX-License-Identifier: EUPL-1.2

package manifest

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
)

func canonicalManifestBytes(m *Manifest) ([]byte, error) {
	if err := validateManifest(m); err != nil {
		return nil, err
	}
	cp := *m
	cp.Sign = ""
	return json.Marshal(cp)
}

func Sign(m *Manifest, priv ed25519.PrivateKey) error {
	if m == nil {
		return errors.New("manifest.Sign: manifest is required")
	}
	if len(priv) == 0 {
		return errors.New("manifest.Sign: private key is required")
	}
	payload, err := canonicalManifestBytes(m)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(payload)
	sig := ed25519.Sign(priv, sum[:])
	m.Sign = base64.StdEncoding.EncodeToString(sig)
	return nil
}

func Verify(m *Manifest, pub ed25519.PublicKey) (bool, error) {
	if m == nil {
		return false, errors.New("manifest.Verify: manifest is required")
	}
	if len(pub) == 0 {
		return false, errors.New("manifest.Verify: public key is required")
	}
	if m.Sign == "" {
		return false, nil
	}
	payload, err := canonicalManifestBytes(m)
	if err != nil {
		return false, err
	}
	sum := sha256.Sum256(payload)
	sig, err := base64.StdEncoding.DecodeString(m.Sign)
	if err != nil {
		return false, fmt.Errorf("manifest.Verify: decode signature: %w", err)
	}
	return ed25519.Verify(pub, sum[:], sig), nil
}
