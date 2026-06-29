// SPDX-License-Identifier: EUPL-1.2

package plugin

import (
	core "dappco.re/go"
	coreio "dappco.re/go/io"
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

func (m *Manifest) Validate() error  /* v090-result-boundary */ {
	if m == nil {
		return core.E("plugin.Manifest.Validate", "manifest is required", nil)
	}
	if m.Name == "" || m.Version == "" || m.Entrypoint == "" {
		return core.E("plugin.Manifest.Validate", "name, version, and entrypoint are required", nil)
	}
	return nil
}

func LoadManifest(m coreio.Medium, path string) (*Manifest, error)  /* v090-result-boundary */ {
	if m == nil {
		return nil, core.E("plugin.LoadManifest", "medium is required", nil)
	}
	raw, err := m.Read(path)
	if err != nil {
		return nil, err
	}
	var manifest Manifest
	if r := core.JSONUnmarshal([]byte(raw), &manifest); !r.OK {
		return nil, core.E("plugin.LoadManifest", "decode manifest", nil)
	}
	return &manifest, manifest.Validate()
}
