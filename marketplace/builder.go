// SPDX-License-Identifier: EUPL-1.2

package marketplace

import (
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
	json "dappco.re/go/core/scm/internal/ax/jsonx"
	os "dappco.re/go/core/scm/internal/ax/osx"
	"sort"
	"strings"

	core "dappco.re/go/core"

	coreio "dappco.re/go/core/io"
	coreerr "dappco.re/go/core/log"
	"dappco.re/go/core/scm/manifest"
)

// IndexVersion is the current marketplace index format version.
const IndexVersion = 1

// Builder constructs a marketplace Index by crawling directories for
// core.json (compiled manifests) or .core/manifest.yaml files.
type Builder struct {
	// Medium is the filesystem abstraction used for manifest reads.
	// If nil, io.Local is used.
	Medium coreio.Medium

	// BaseURL is the prefix for constructing repository URLs, e.g.
	// "https://forge.lthn.ai". When set, module Repo is derived as
	// BaseURL + "/" + org + "/" + code + ".git".
	BaseURL string

	// Org is the default organisation used when constructing Repo URLs.
	Org string
}

// BuildFromDirs scans each directory, then its immediate subdirectories, for
// core.json (preferred) or .core/manifest.yaml. Each valid manifest is added
// to the resulting Index as a Module.
// Usage: BuildFromDirs(...)
func (b *Builder) BuildFromDirs(dirs ...string) (*Index, error) {
	var modules []Module
	seen := make(map[string]bool)

	for _, dir := range dirs {
		if err := b.loadInto(&modules, seen, dir); err != nil {
			core.Warn(core.Sprintf("marketplace: skipping %s: %v", filepath.Base(dir), err))
		}

		entries, err := os.ReadDir(dir)
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

			child := filepath.Join(dir, e.Name())
			if err := b.loadInto(&modules, seen, child); err != nil {
				core.Warn(core.Sprintf("marketplace: skipping %s: %v", e.Name(), err))
			}
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

func (b *Builder) loadInto(modules *[]Module, seen map[string]bool, dir string) error {
	m, err := b.loadFromDir(dir)
	if err != nil {
		return err
	}
	if m == nil {
		return nil
	}
	if seen[m.Code] {
		return nil
	}
	seen[m.Code] = true

	mod := Module{
		Code:    m.Code,
		Name:    m.Name,
		Repo:    b.repoURL(m.Code),
		Version: moduleVersionForRepo(dir, m.Version),
		SignKey: manifestSignKey(m),
	}
	*modules = append(*modules, mod)
	return nil
}

// BuildFromManifests constructs an Index from pre-loaded manifests.
// This is useful when manifests have already been collected (e.g. from
// a Forge API crawl).
// Usage: BuildFromManifests(...)
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
			Code:    m.Code,
			Name:    m.Name,
			Version: m.Version,
			SignKey: manifestSignKey(m),
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

// WriteIndex serialises an Index to JSON and writes it to the given path
// using the provided filesystem medium.
// Usage: WriteIndex(...)
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
	medium := b.Medium
	if medium == nil {
		medium = coreio.Local
	}

	// Prefer compiled manifest (core.json).
	coreJSON := filepath.Join(dir, "core.json")
	if raw, err := medium.Read(coreJSON); err == nil {
		cm, err := manifest.ParseCompiled([]byte(raw))
		if err != nil {
			return nil, coreerr.E("marketplace.Builder.loadFromDir", "parse core.json", err)
		}
		return &cm.Manifest, nil
	}

	// Fall back to source manifest.
	manifestYAML := filepath.Join(dir, ".core", "manifest.yaml")
	raw, err := medium.Read(manifestYAML)
	if err != nil {
		return nil, nil // No manifest — skip silently.
	}

	m, err := manifest.Parse([]byte(raw))
	if err != nil {
		return nil, coreerr.E("marketplace.Builder.loadFromDir", "parse manifest.yaml", err)
	}
	return m, nil
}

// repoURL constructs a module repository URL from the builder config.
func (b *Builder) repoURL(code string) string {
	if b.Org == "" {
		return ""
	}
	baseURL := b.BaseURL
	if baseURL == "" {
		baseURL = defaultForgeURL
	}
	baseURL = strings.TrimSuffix(baseURL, "/")
	return baseURL + "/" + b.Org + "/" + code + ".git"
}
