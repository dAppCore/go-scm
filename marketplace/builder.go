package marketplace

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sort"

	coreerr "forge.lthn.ai/core/go-log"
	coreio "forge.lthn.ai/core/go-io"
	"forge.lthn.ai/core/go-scm/manifest"
)

// IndexVersion is the current marketplace index format version.
const IndexVersion = 1

// Builder constructs a marketplace Index by crawling directories for
// core.json (compiled manifests) or .core/manifest.yaml files.
type Builder struct {
	// BaseURL is the prefix for constructing repository URLs, e.g.
	// "https://forge.lthn.ai". When set, module Repo is derived as
	// BaseURL + "/" + org + "/" + code.
	BaseURL string

	// Org is the default organisation used when constructing Repo URLs.
	Org string

	// Medium is the filesystem abstraction used for reading and writing.
	// Falls back to coreio.Local if nil.
	Medium coreio.Medium
}

// medium returns the builder's Medium or falls back to coreio.Local.
func (b *Builder) medium() coreio.Medium {
	if b.Medium != nil {
		return b.Medium
	}
	return coreio.Local
}

// BuildFromDirs scans each directory for subdirectories containing either
// core.json (preferred) or .core/manifest.yaml. Each valid manifest is
// added to the resulting Index as a Module.
func (b *Builder) BuildFromDirs(dirs ...string) (*Index, error) {
	var modules []Module
	seen := make(map[string]bool)

	m := b.medium()
	for _, dir := range dirs {
		entries, err := m.List(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, coreerr.E("marketplace.Builder.BuildFromDirs", "read "+dir, err)
		}

		for _, e := range entries {
			if !e.IsDir() {
				continue
			}

			m, err := b.loadFromDir(filepath.Join(dir, e.Name()))
			if err != nil {
				log.Printf("marketplace: skipping %s: %v", e.Name(), err)
				continue
			}
			if m == nil {
				continue
			}
			if seen[m.Code] {
				continue
			}
			seen[m.Code] = true

			mod := Module{
				Code: m.Code,
				Name: m.Name,
				Repo: b.repoURL(m.Code),
			}
			modules = append(modules, mod)
		}
	}

	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Code < modules[j].Code
	})

	return &Index{
		Version: IndexVersion,
		Modules: modules,
	}, nil
}

// BuildFromManifests constructs an Index from pre-loaded manifests.
// This is useful when manifests have already been collected (e.g. from
// a Forge API crawl).
func BuildFromManifests(manifests []*manifest.Manifest) *Index {
	var modules []Module
	seen := make(map[string]bool)

	for _, m := range manifests {
		if m == nil || m.Code == "" {
			continue
		}
		if seen[m.Code] {
			continue
		}
		seen[m.Code] = true

		modules = append(modules, Module{
			Code: m.Code,
			Name: m.Name,
		})
	}

	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Code < modules[j].Code
	})

	return &Index{
		Version: IndexVersion,
		Modules: modules,
	}
}

// WriteIndex serialises an Index to JSON and writes it to the given path.
func WriteIndex(m coreio.Medium, path string, idx *Index) error {
	if err := m.EnsureDir(filepath.Dir(path)); err != nil {
		return coreerr.E("marketplace.WriteIndex", "mkdir failed", err)
	}
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return coreerr.E("marketplace.WriteIndex", "marshal failed", err)
	}
	return m.Write(path, string(data))
}

// loadFromDir tries core.json first, then falls back to .core/manifest.yaml.
func (b *Builder) loadFromDir(dir string) (*manifest.Manifest, error) {
	m := b.medium()

	// Prefer compiled manifest (core.json).
	coreJSON := filepath.Join(dir, "core.json")
	if raw, err := m.Read(coreJSON); err == nil {
		cm, err := manifest.ParseCompiled([]byte(raw))
		if err != nil {
			return nil, coreerr.E("marketplace.Builder.loadFromDir", "parse core.json", err)
		}
		return &cm.Manifest, nil
	}

	// Fall back to source manifest.
	manifestYAML := filepath.Join(dir, ".core", "manifest.yaml")
	raw, err := m.Read(manifestYAML)
	if err != nil {
		return nil, nil // No manifest — skip silently.
	}

	mf, err := manifest.Parse([]byte(raw))
	if err != nil {
		return nil, coreerr.E("marketplace.Builder.loadFromDir", "parse manifest.yaml", err)
	}
	return mf, nil
}

// repoURL constructs a module repository URL from the builder config.
func (b *Builder) repoURL(code string) string {
	if b.BaseURL == "" || b.Org == "" {
		return ""
	}
	return b.BaseURL + "/" + b.Org + "/" + code
}
