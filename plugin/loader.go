// SPDX-License-Identifier: EUPL-1.2

package plugin

import (
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"

	"dappco.re/go/core/io"
	coreerr "dappco.re/go/core/log"
)

// Loader loads plugins from the filesystem.
type Loader struct {
	medium  io.Medium
	baseDir string
}

// NewLoader creates a new plugin loader.
// Usage: NewLoader(...)
func NewLoader(m io.Medium, baseDir string) *Loader {
	return &Loader{
		medium:  m,
		baseDir: baseDir,
	}
}

// Discover finds all plugin directories under baseDir and returns their manifests.
// Directories without a valid plugin.json are silently skipped.
// Usage: Discover(...)
func (l *Loader) Discover() ([]*Manifest, error) {
	if !l.medium.Exists(l.baseDir) {
		return []*Manifest{}, nil
	}

	entries, err := l.medium.List(l.baseDir)
	if err != nil {
		return nil, coreerr.E("plugin.Loader.Discover", "failed to list plugin directory", err)
	}

	var manifests []*Manifest
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		manifest, err := l.LoadPlugin(entry.Name())
		if err != nil {
			// Skip directories without valid manifests
			continue
		}

		manifests = append(manifests, manifest)
	}

	return manifests, nil
}

// LoadPlugin loads a single plugin's manifest by name.
// Usage: LoadPlugin(...)
func (l *Loader) LoadPlugin(name string) (*Manifest, error) {
	manifestPath := filepath.Join(l.baseDir, name, "plugin.json")
	manifest, err := LoadManifest(l.medium, manifestPath)
	if err != nil {
		return nil, coreerr.E("plugin.Loader.LoadPlugin", "failed to load plugin: "+name, err)
	}

	if err := manifest.Validate(); err != nil {
		return nil, coreerr.E("plugin.Loader.LoadPlugin", "invalid plugin manifest: "+name, err)
	}

	return manifest, nil
}
