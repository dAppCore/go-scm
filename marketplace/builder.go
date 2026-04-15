// SPDX-License-Identifier: EUPL-1.2

package marketplace

import (
	"errors"
	"path/filepath"
	"strings"

	"dappco.re/go/scm/manifest"
	"dappco.re/go/scm/internal/ax/jsonx"
	"dappco.re/go/scm/internal/ax/osx"
)

type Builder struct {
	BaseURL string
	Org     string
}

func BuildFromManifests(manifests []*manifest.Manifest) *Index { return BuildIndexFromManifests(manifests) }

func (b *Builder) BuildFromDirs(dirs ...string) (*Index, error) {
	var manifests []*manifest.Manifest
	for _, dir := range dirs {
		entries, err := osx.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry == nil || !entry.IsDir() {
				continue
			}
			root := filepath.Join(dir, entry.Name())
			if m, err := loadManifestFromRoot(root); err == nil && m != nil {
				if b.BaseURL != "" && m.Code != "" {
					// repo path is derived below in BuildFromManifests
				}
				manifests = append(manifests, m)
				continue
			}
		}
	}
	idx := BuildIndexFromManifests(manifests)
	for i := range idx.Modules {
		if idx.Modules[i].Repo == "" && idx.Modules[i].Code != "" && b.BaseURL != "" {
			org := b.Org
			if org == "" {
				org = "core"
			}
			idx.Modules[i].Repo = strings.TrimRight(b.BaseURL, "/") + "/" + org + "/" + idx.Modules[i].Code
		}
	}
	return idx, nil
}

func loadManifestFromRoot(root string) (*manifest.Manifest, error) {
	if raw, err := osx.ReadFile(filepath.Join(root, "core.json")); err == nil {
		if cm, err := manifest.ParseCompiled(raw); err == nil {
			m := cm.Manifest
			return &m, nil
		}
	}
	raw, err := osx.ReadFile(filepath.Join(root, ".core", "manifest.yaml"))
	if err != nil {
		return nil, err
	}
	return manifest.Parse(raw)
}

func WriteIndex(path string, idx *Index) error {
	if idx == nil {
		return errors.New("marketplace.WriteIndex: index is required")
	}
	raw, err := jsonx.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	return osx.WriteFile(path, raw, 0o600)
}
