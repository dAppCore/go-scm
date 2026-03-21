package manifest

import (
	"crypto/ed25519"
	"path/filepath"

	coreerr "dappco.re/go/core/log"
	"dappco.re/go/core/io"
	"gopkg.in/yaml.v3"
)

const manifestPath = ".core/manifest.yaml"

// MarshalYAML serializes a manifest to YAML bytes.
func MarshalYAML(m *Manifest) ([]byte, error) {
	return yaml.Marshal(m)
}

// Load reads and parses a .core/manifest.yaml from the given root directory.
func Load(medium io.Medium, root string) (*Manifest, error) {
	path := filepath.Join(root, manifestPath)
	data, err := medium.Read(path)
	if err != nil {
		return nil, coreerr.E("manifest.Load", "read failed", err)
	}
	return Parse([]byte(data))
}

// LoadVerified reads, parses, and verifies the ed25519 signature.
func LoadVerified(medium io.Medium, root string, pub ed25519.PublicKey) (*Manifest, error) {
	m, err := Load(medium, root)
	if err != nil {
		return nil, err
	}
	ok, err := Verify(m, pub)
	if err != nil {
		return nil, coreerr.E("manifest.LoadVerified", "verification error", err)
	}
	if !ok {
		return nil, coreerr.E("manifest.LoadVerified", "signature verification failed for "+m.Code, nil)
	}
	return m, nil
}
