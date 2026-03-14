package marketplace

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"forge.lthn.ai/core/go-scm/manifest"
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
func DiscoverProviders(dir string) ([]DiscoveredProvider, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No providers directory — not an error.
		}
		return nil, fmt.Errorf("marketplace.DiscoverProviders: %w", err)
	}

	var providers []DiscoveredProvider
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		providerDir := filepath.Join(dir, e.Name())
		manifestPath := filepath.Join(providerDir, ".core", "manifest.yaml")

		data, err := os.ReadFile(manifestPath)
		if err != nil {
			log.Printf("marketplace: skipping %s: %v", e.Name(), err)
			continue
		}

		m, err := manifest.Parse(data)
		if err != nil {
			log.Printf("marketplace: skipping %s: invalid manifest: %v", e.Name(), err)
			continue
		}

		if !m.IsProvider() {
			log.Printf("marketplace: skipping %s: not a provider (missing namespace or binary)", e.Name())
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
func LoadProviderRegistry(path string) (*ProviderRegistryFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ProviderRegistryFile{
				Version:   1,
				Providers: make(map[string]ProviderRegistryEntry),
			}, nil
		}
		return nil, fmt.Errorf("marketplace.LoadProviderRegistry: %w", err)
	}

	var reg ProviderRegistryFile
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("marketplace.LoadProviderRegistry: %w", err)
	}

	if reg.Providers == nil {
		reg.Providers = make(map[string]ProviderRegistryEntry)
	}

	return &reg, nil
}

// SaveProviderRegistry writes the registry to the given path.
func SaveProviderRegistry(path string, reg *ProviderRegistryFile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("marketplace.SaveProviderRegistry: %w", err)
	}

	data, err := yaml.Marshal(reg)
	if err != nil {
		return fmt.Errorf("marketplace.SaveProviderRegistry: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// Add adds or updates a provider entry in the registry.
func (r *ProviderRegistryFile) Add(code string, entry ProviderRegistryEntry) {
	if r.Providers == nil {
		r.Providers = make(map[string]ProviderRegistryEntry)
	}
	r.Providers[code] = entry
}

// Remove removes a provider entry from the registry.
func (r *ProviderRegistryFile) Remove(code string) {
	delete(r.Providers, code)
}

// Get returns a provider entry and true if found, or zero value and false.
func (r *ProviderRegistryFile) Get(code string) (ProviderRegistryEntry, bool) {
	entry, ok := r.Providers[code]
	return entry, ok
}

// List returns all provider codes in the registry.
func (r *ProviderRegistryFile) List() []string {
	codes := make([]string, 0, len(r.Providers))
	for code := range r.Providers {
		codes = append(codes, code)
	}
	return codes
}

// AutoStartProviders returns codes of providers with auto_start enabled.
func (r *ProviderRegistryFile) AutoStartProviders() []string {
	var codes []string
	for code, entry := range r.Providers {
		if entry.AutoStart {
			codes = append(codes, code)
		}
	}
	return codes
}
