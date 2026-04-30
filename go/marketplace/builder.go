// SPDX-License-Identifier: EUPL-1.2

package marketplace

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"

	"dappco.re/go/scm/internal/ax/jsonx"
	"dappco.re/go/scm/internal/ax/osx"
	"dappco.re/go/scm/manifest"
)

type Builder struct {
	BaseURL string
	Org     string
}

func BuildFromManifests(manifests []*manifest.Manifest) *Index {
	return BuildIndexFromManifests(manifests)
}

func (b *Builder) BuildFromDirs(dirs ...string) (*Index, error) {
	manifests, err := loadManifestsFromDirs(dirs)
	if err != nil {
		return nil, err
	}
	idx := BuildIndexFromManifests(manifests)
	b.fillModuleRepos(idx)
	return idx, nil
}

func loadManifestsFromDirs(dirs []string) ([]*manifest.Manifest, error) {
	var manifests []*manifest.Manifest
	for _, dir := range dirs {
		entries, err := osx.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		manifests = append(manifests, loadManifestsFromEntries(dir, entries)...)
	}
	return manifests, nil
}

func loadManifestsFromEntries(dir string, entries []fs.DirEntry) []*manifest.Manifest {
	var manifests []*manifest.Manifest
	for _, entry := range entries {
		if entry == nil || !entry.IsDir() {
			continue
		}
		root := filepath.Join(dir, entry.Name())
		if m, err := loadManifestFromRoot(root); err == nil && m != nil {
			manifests = append(manifests, m)
		}
	}
	return manifests
}

func (b *Builder) fillModuleRepos(idx *Index) {
	for i := range idx.Modules {
		if idx.Modules[i].Repo != "" || idx.Modules[i].Code == "" || b.BaseURL == "" {
			continue
		}
		idx.Modules[i].Repo = b.moduleRepo(idx.Modules[i].Code)
	}
}

func (b *Builder) moduleRepo(code string) string {
	org := b.Org
	if org == "" {
		org = "core"
	}
	return strings.TrimRight(b.BaseURL, "/") + "/" + org + "/" + code
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
