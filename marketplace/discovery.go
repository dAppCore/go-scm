// SPDX-License-Identifier: EUPL-1.2

package marketplace

import (
	"path/filepath"
	"sort"

	"dappco.re/go/scm/internal/ax/osx"
	"dappco.re/go/scm/manifest"
	"gopkg.in/yaml.v3"
)

type DiscoveredProvider struct {
	Dir      string
	Manifest *manifest.Manifest
}

type ProviderRegistryEntry struct {
	Installed string `yaml:"installed" json:"installed"`
	Version   string `yaml:"version" json:"version"`
	Source    string `yaml:"source" json:"source"`
	AutoStart bool   `yaml:"auto_start" json:"auto_start"`
}

type ProviderRegistryFile struct {
	Version   int                              `yaml:"version" json:"version"`
	Providers map[string]ProviderRegistryEntry `yaml:"providers" json:"providers"`
}

func (r *ProviderRegistryFile) Add(code string, entry ProviderRegistryEntry) {
	if r == nil {
		return
	}
	if r.Providers == nil {
		r.Providers = map[string]ProviderRegistryEntry{}
	}
	r.Providers[code] = entry
}

func (r *ProviderRegistryFile) Get(code string) (ProviderRegistryEntry, bool) {
	if r == nil || r.Providers == nil {
		return ProviderRegistryEntry{}, false
	}
	entry, ok := r.Providers[code]
	return entry, ok
}

func (r *ProviderRegistryFile) List() []string {
	if r == nil {
		return nil
	}
	out := make([]string, 0, len(r.Providers))
	for code := range r.Providers {
		out = append(out, code)
	}
	sort.Strings(out)
	return out
}

func (r *ProviderRegistryFile) AutoStartProviders() []string {
	if r == nil {
		return nil
	}
	var out []string
	for code, entry := range r.Providers {
		if entry.AutoStart {
			out = append(out, code)
		}
	}
	sort.Strings(out)
	return out
}

func (r *ProviderRegistryFile) Remove(code string) {
	if r == nil || r.Providers == nil {
		return
	}
	delete(r.Providers, code)
}

func DiscoverProviders(dir string) ([]DiscoveredProvider, error) {
	entries, err := osx.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []DiscoveredProvider
	for _, entry := range entries {
		if entry == nil || !entry.IsDir() {
			continue
		}
		root := filepath.Join(dir, entry.Name())
		raw, err := osx.ReadFile(filepath.Join(root, ".core", "manifest.yaml"))
		if err != nil {
			continue
		}
		m, err := manifest.Parse(raw)
		if err != nil || m == nil || !m.IsProvider() {
			continue
		}
		out = append(out, DiscoveredProvider{Dir: root, Manifest: m})
	}
	return out, nil
}

func LoadProviderRegistry(path string) (*ProviderRegistryFile, error) {
	raw, err := osx.ReadFile(path)
	if err != nil {
		return &ProviderRegistryFile{Version: 1, Providers: map[string]ProviderRegistryEntry{}}, nil
	}
	var reg ProviderRegistryFile
	if err := yaml.Unmarshal(raw, &reg); err != nil {
		return nil, err
	}
	if reg.Version == 0 {
		reg.Version = 1
	}
	if reg.Providers == nil {
		reg.Providers = map[string]ProviderRegistryEntry{}
	}
	return &reg, nil
}

func SaveProviderRegistry(path string, reg *ProviderRegistryFile) error {
	if reg == nil {
		reg = &ProviderRegistryFile{Version: 1, Providers: map[string]ProviderRegistryEntry{}}
	}
	raw, err := yaml.Marshal(reg)
	if err != nil {
		return err
	}
	return osx.WriteFile(path, raw, 0o600)
}
