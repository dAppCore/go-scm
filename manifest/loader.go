// SPDX-License-Identifier: EUPL-1.2

package manifest

import (
	"crypto/ed25519" // intrinsic
	"encoding/base64"
	"errors"

	coreio "dappco.re/go/io"
	"dappco.re/go/scm/internal/ax/filepathx"
	"gopkg.in/yaml.v3"
)

func Load(medium coreio.Medium, root string) (*Manifest, error) {
	if medium == nil {
		return nil, errors.New("manifest.Load: medium is required")
	}
	raw, err := medium.Read(filepathx.Join(root, ".core", "manifest.yaml"))
	if err != nil {
		return nil, err
	}
	return Parse([]byte(raw))
}

func LoadVerified(medium coreio.Medium, root string, pub ed25519.PublicKey) (*Manifest, error) {
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

func MarshalYAML(m *Manifest) ([]byte, error) {
	if m == nil {
		return nil, errors.New("manifest.MarshalYAML: manifest is required")
	}
	return yaml.Marshal(m)
}
