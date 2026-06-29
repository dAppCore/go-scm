// SPDX-License-Identifier: EUPL-1.2

package manifest

import (
	"crypto/ed25519" // intrinsic
	"encoding/base64"

	core "dappco.re/go"
)

const (
	sonarSignManifestVerify = "manifest.Verify"
)

func canonicalManifestBytes(m *Manifest) ([]byte, error)  /* v090-result-boundary */ {
	if err := validateManifest(m); err != nil {
		return nil, err
	}
	cp := *m
	cp.Sign = ""
	cp.SignKey = ""
	r := core.JSONMarshal(cp)
	if !r.OK {
		err, _ := r.Value.(error)
		return nil, err
	}
	data, _ := r.Value.([]byte)
	return data, nil
}

func Sign(m *Manifest, payload []byte, priv ed25519.PrivateKey) error  /* v090-result-boundary */ {
	if m == nil {
		return core.E("manifest.Sign", "manifest is required", nil)
	}
	if len(priv) != ed25519.PrivateKeySize {
		return core.E("manifest.Sign", "private key is required", nil)
	}
	sig := ed25519.Sign(priv, payload)
	m.Sign = base64.StdEncoding.EncodeToString(sig)
	return nil
}

func Verify(m *Manifest, payload []byte) error  /* v090-result-boundary */ {
	if m == nil {
		return core.E(sonarSignManifestVerify, "manifest is required", nil)
	}
	if m.SignKey == "" {
		return core.E(sonarSignManifestVerify, "sign key is required", nil)
	}
	if m.Sign == "" {
		return core.E(sonarSignManifestVerify, "signature is required", nil)
	}
	pub, err := base64.StdEncoding.DecodeString(m.SignKey)
	if err != nil {
		return core.E(sonarSignManifestVerify, "decode sign key", err)
	}
	sig, err := base64.StdEncoding.DecodeString(m.Sign)
	if err != nil {
		return core.E(sonarSignManifestVerify, "decode signature", err)
	}
	if len(pub) != ed25519.PublicKeySize {
		return core.E(sonarSignManifestVerify, "invalid sign key", nil)
	}
	if len(sig) != ed25519.SignatureSize {
		return core.E(sonarSignManifestVerify, "invalid signature", nil)
	}
	if !ed25519.Verify(ed25519.PublicKey(pub), payload, sig) {
		return core.E(sonarSignManifestVerify, "signature verification failed", nil)
	}
	return nil
}
