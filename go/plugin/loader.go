// SPDX-License-Identifier: EUPL-1.2

package plugin

import (
	core "dappco.re/go"
	coreio "dappco.re/go/io"
)

type Loader struct {
	medium  coreio.Medium
	baseDir string
}

func NewLoader(m coreio.Medium, baseDir string) *Loader {
	return &Loader{medium: m, baseDir: baseDir}
}

func (l *Loader) Discover() ([]*Manifest, error)  /* v090-result-boundary */ {
	if l == nil || l.medium == nil {
		return nil, nil
	}
	entries, err := l.medium.List(l.baseDir)
	if err != nil {
		return nil, err
	}
	var out []*Manifest
	for _, entry := range entries {
		if entry == nil || !entry.IsDir() {
			continue
		}
		manifest, err := LoadManifest(l.medium, core.PathJoin(l.baseDir, entry.Name(), "plugin.json"))
		if err != nil {
			continue
		}
		out = append(out, manifest)
	}
	return out, nil
}

func (l *Loader) LoadPlugin(name string) (*Manifest, error)  /* v090-result-boundary */ {
	if l == nil || l.medium == nil {
		return nil, core.E("plugin.Loader.LoadPlugin", "loader is required", nil)
	}
	return LoadManifest(l.medium, core.PathJoin(l.baseDir, name, "plugin.json"))
}
