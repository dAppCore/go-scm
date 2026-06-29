// SPDX-License-Identifier: EUPL-1.2

package marketplace

import (
	"io/fs"

	core "dappco.re/go"
	"dappco.re/go/scm/manifest"
)

type Builder struct {
	BaseURL string
	Org     string
}

func BuildFromManifests(manifests []*manifest.Manifest) *Index {
	return BuildIndexFromManifests(manifests)
}

func (b *Builder) BuildFromDirs(dirs ...string) (*Index, error)  /* v090-result-boundary */ {
	manifests, err := loadManifestsFromDirs(dirs)
	if err != nil {
		return nil, err
	}
	idx := BuildIndexFromManifests(manifests)
	b.fillModuleRepos(idx)
	return idx, nil
}

func loadManifestsFromDirs(dirs []string) ([]*manifest.Manifest, error)  /* v090-result-boundary */ {
	var manifests []*manifest.Manifest
	for _, dir := range dirs {
		readResult := core.ReadDir(core.DirFS(dir), ".")
		if !readResult.OK {
			return nil, core.E("marketplace.loadManifestsFromDirs", "read provider directory", nil)
		}
		entries := readResult.Value.([]core.FsDirEntry)
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
		root := core.PathJoin(dir, entry.Name())
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
	return core.TrimSuffix(b.BaseURL, "/") + "/" + org + "/" + code
}

func loadManifestFromRoot(root string) (*manifest.Manifest, error)  /* v090-result-boundary */ {
	if readResult := core.ReadFile(core.PathJoin(root, "core.json")); readResult.OK {
		if cm, err := manifest.ParseCompiled(readResult.Value.([]byte)); err == nil {
			m := cm.Manifest
			return &m, nil
		}
	}
	readResult := core.ReadFile(core.PathJoin(root, ".core", "manifest.yaml"))
	if !readResult.OK {
		return nil, core.E("marketplace.loadManifestFromRoot", "read manifest", nil)
	}
	return manifest.Parse(readResult.Value.([]byte))
}

func WriteIndex(path string, idx *Index) error  /* v090-result-boundary */ {
	if idx == nil {
		return core.E("marketplace.WriteIndex", "index is required", nil)
	}
	marshalResult := core.JSONMarshalIndent(idx, "", "  ")
	if !marshalResult.OK {
		return core.E("marketplace.WriteIndex", "encode index", nil)
	}
	writeResult := core.WriteFile(path, marshalResult.Value.([]byte), 0o600)
	if !writeResult.OK {
		return core.E("marketplace.WriteIndex", "write index", nil)
	}
	return nil
}
