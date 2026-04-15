// SPDX-License-Identifier: EUPL-1.2

package marketplace

import (
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
	os "dappco.re/go/core/scm/internal/ax/osx"
	"sort"

	core "dappco.re/go/core"
	coreio "dappco.re/go/core/io"
	coreerr "dappco.re/go/core/log"
	"dappco.re/go/core/scm/manifest"
	"gopkg.in/yaml.v3"
)

// DiscoveredProvider represents a runtime provider found on disk.
type DiscoveredProvider struct {
	// Dir is the absolute path to the provider directory.
	Dir string

	// Manifest is the parsed manifest from the provider directory.
	Manifest *manifest.Manifest
}

// DiscoverProviders scans the given directory for runtime provider manifests.
// Each subdirectory is checked for a .core/manifest.yaml file. Directories
// without a valid manifest are skipped with a log warning.
// Only manifests with provider fields (namespace + binary) are returned.
// Usage: DiscoverProviders(...)
func DiscoverProviders(dir string) ([]DiscoveredProvider, error) {
	return DiscoverProvidersWithMedium(coreio.Local, dir)
}

// DiscoverProvidersWithMedium scans the given directory for runtime provider
// manifests using the supplied filesystem medium.
// Usage: DiscoverProvidersWithMedium(...)
func DiscoverProvidersWithMedium(medium coreio.Medium, dir string) ([]DiscoveredProvider, error) {
	entries, err := medium.List(dir)
	if err != nil {
		if !medium.Exists(dir) {
			return nil, nil // No providers directory — not an error.
		}
		return nil, coreerr.E("marketplace.DiscoverProviders", "read directory", err)
	}

	var providers []DiscoveredProvider
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		providerDir := filepath.Join(dir, e.Name())
		manifestPath := filepath.Join(providerDir, ".core", "manifest.yaml")

		raw, err := medium.Read(manifestPath)
		if err != nil {
			core.Warn(core.Sprintf("marketplace: skipping %s: %v", e.Name(), err))
			continue
		}

		m, err := manifest.Parse([]byte(raw))
		if err != nil {
			core.Warn(core.Sprintf("marketplace: skipping %s: invalid manifest: %v", e.Name(), err))
			continue
		}

		if !m.IsProvider() {
			core.Warn(core.Sprintf("marketplace: skipping %s: not a provider (missing namespace or binary)", e.Name()))
			continue
		}

		providers = append(providers, DiscoveredProvider{
			Dir:      providerDir,
			Manifest: m,
		})
	}

	return providers, nil
}

// ProviderRegistryEntry records metadata about an installed provider.
type ProviderRegistryEntry struct {
	Installed string `yaml:"installed" json:"installed"`
	Version   string `yaml:"version" json:"version"`
	Source    string `yaml:"source" json:"source"`
	AutoStart bool   `yaml:"auto_start" json:"auto_start"`
}

// ProviderRegistryFile represents the registry.yaml file tracking installed providers.
type ProviderRegistryFile struct {
	Version   int                              `yaml:"version" json:"version"`
	Providers map[string]ProviderRegistryEntry `yaml:"providers" json:"providers"`
}

// LoadProviderRegistry reads a registry.yaml file from the given path.
// Returns an empty registry if the file does not exist.
// Usage: LoadProviderRegistry(...)
func LoadProviderRegistry(path string) (*ProviderRegistryFile, error) {
	return LoadProviderRegistryWithMedium(coreio.Local, path)
}

// LoadProviderRegistryWithMedium reads a registry.yaml file using the supplied
// filesystem medium.
// Usage: LoadProviderRegistryWithMedium(...)
func LoadProviderRegistryWithMedium(medium coreio.Medium, path string) (*ProviderRegistryFile, error) {
	raw, err := medium.Read(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ProviderRegistryFile{
				Version:   1,
				Providers: make(map[string]ProviderRegistryEntry),
			}, nil
		}
		return nil, coreerr.E("marketplace.LoadProviderRegistry", "read failed", err)
	}

	var reg ProviderRegistryFile
	if err := yaml.Unmarshal([]byte(raw), &reg); err != nil {
		return nil, coreerr.E("marketplace.LoadProviderRegistry", "parse failed", err)
	}

	if reg.Providers == nil {
		reg.Providers = make(map[string]ProviderRegistryEntry)
	}

	return &reg, nil
}

// SaveProviderRegistry writes the registry to the given path.
// Usage: SaveProviderRegistry(...)
func SaveProviderRegistry(path string, reg *ProviderRegistryFile) error {
	return SaveProviderRegistryWithMedium(coreio.Local, path, reg)
}

// SaveProviderRegistryWithMedium writes the registry using the supplied
// filesystem medium.
// Usage: SaveProviderRegistryWithMedium(...)
func SaveProviderRegistryWithMedium(medium coreio.Medium, path string, reg *ProviderRegistryFile) error {
	if err := medium.EnsureDir(filepath.Dir(path)); err != nil {
		return coreerr.E("marketplace.SaveProviderRegistry", "ensure directory", err)
	}

	data, err := yaml.Marshal(reg)
	if err != nil {
		return coreerr.E("marketplace.SaveProviderRegistry", "marshal failed", err)
	}

	return medium.Write(path, string(data))
}

// Add adds or updates a provider entry in the registry.
// Usage: Add(...)
func (r *ProviderRegistryFile) Add(code string, entry ProviderRegistryEntry) {
	if r.Providers == nil {
		r.Providers = make(map[string]ProviderRegistryEntry)
	}
	r.Providers[code] = entry
}

// Remove removes a provider entry from the registry.
// Usage: Remove(...)
func (r *ProviderRegistryFile) Remove(code string) {
	delete(r.Providers, code)
}

// Get returns a provider entry and true if found, or zero value and false.
// Usage: Get(...)
func (r *ProviderRegistryFile) Get(code string) (ProviderRegistryEntry, bool) {
	entry, ok := r.Providers[code]
	return entry, ok
}

// List returns all provider codes in the registry.
// Usage: List(...)
func (r *ProviderRegistryFile) List() []string {
	codes := make([]string, 0, len(r.Providers))
	for code := range r.Providers {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	return codes
}

// AutoStartProviders returns codes of providers with auto_start enabled.
// Usage: AutoStartProviders(...)
func (r *ProviderRegistryFile) AutoStartProviders() []string {
	var codes []string
	for code, entry := range r.Providers {
		if entry.AutoStart {
			codes = append(codes, code)
		}
	}
	sort.Strings(codes)
	return codes
}
