// SPDX-License-Identifier: EUPL-1.2

package manifest

import (
	"crypto/ed25519" // intrinsic
	"encoding/base64"

	core "dappco.re/go"
	coreio "dappco.re/go/io"
	"gopkg.in/yaml.v3"
)

func Load(medium coreio.Medium, root string) (*Manifest, error)  /* v090-result-boundary */ {
	if medium == nil {
		return nil, core.E("manifest.Load", "medium is required", nil)
	}
	raw, err := medium.Read(core.PathJoin(root, ".core", "manifest.yaml"))
	if err != nil {
		return nil, err
	}
	return Parse([]byte(raw))
}

func LoadVerified(medium coreio.Medium, root string, pub ed25519.PublicKey) (*Manifest, error)  /* v090-result-boundary */ {
	m, err := Load(medium, root)
	if err != nil {
		return nil, err
	}
	verifyManifest := m
	if len(pub) > 0 {
		cp := *m
		cp.SignKey = base64.StdEncoding.EncodeToString(pub)
		verifyManifest = &cp
	}
	payload, err := canonicalManifestBytes(m)
	if err != nil {
		return nil, err
	}
	if err := Verify(verifyManifest, payload); err != nil {
		return nil, err
	}
	return m, nil
}

func MarshalYAML(m *Manifest) ([]byte, error)  /* v090-result-boundary */ {
	if m == nil {
		return nil, core.E("manifest.MarshalYAML", "manifest is required", nil)
	}
	return yaml.Marshal(m)
}
