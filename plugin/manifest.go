// SPDX-License-Identifier: EUPL-1.2

package plugin

import (
	// Note: errors.New is retained for stable manifest validation errors.
	"errors"

	coreio "dappco.re/go/io"
	"dappco.re/go/scm/internal/ax/jsonx"
)

type Manifest struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Description  string   `json:"description"`
	Author       string   `json:"author"`
	Entrypoint   string   `json:"entrypoint"`
	Dependencies []string `json:"dependencies,omitempty"`
	MinVersion   string   `json:"min_version,omitempty"`
}

func (m *Manifest) Validate() error {
	if m == nil {
		return errors.New("plugin.Manifest.Validate: manifest is required")
	}
	if m.Name == "" || m.Version == "" || m.Entrypoint == "" {
		return errors.New("plugin.Manifest.Validate: name, version, and entrypoint are required")
	}
	return nil
}

func LoadManifest(m coreio.Medium, path string) (*Manifest, error) {
	if m == nil {
		return nil, errors.New("plugin.LoadManifest: medium is required")
	}
	raw, err := m.Read(path)
	if err != nil {
		return nil, err
	}
	var manifest Manifest
	if err := jsonx.Unmarshal([]byte(raw), &manifest); err != nil {
		return nil, err
	}
	return &manifest, manifest.Validate()
}
