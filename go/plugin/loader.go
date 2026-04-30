// SPDX-License-Identifier: EUPL-1.2

package plugin

import (
	// Note: errors.New is retained for stable loader validation errors.
	`errors`
	// Note: filepath.Join is retained because plugin manifests are stored on an OS-specific local path layout.
	`path/filepath`

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
		manifest, err := LoadManifest(l.medium, filepath.Join(l.baseDir, entry.Name(), "plugin.json"))
		if err != nil {
			continue
		}
		out = append(out, manifest)
	}
	return out, nil
}

func (l *Loader) LoadPlugin(name string) (*Manifest, error)  /* v090-result-boundary */ {
	if l == nil || l.medium == nil {
		return nil, errors.New("plugin.Loader.LoadPlugin: loader is required")
	}
	return LoadManifest(l.medium, filepath.Join(l.baseDir, name, "plugin.json"))
}
