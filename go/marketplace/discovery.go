// SPDX-License-Identifier: EUPL-1.2

package marketplace

import (
	"sort"

	core "dappco.re/go"
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
	absResult := core.PathAbs(dir)
	if absResult.OK {
		dir = absResult.Value.(string)
	}
	readDirResult := core.ReadDir(core.DirFS(dir), ".")
	if !readDirResult.OK {
		return nil, core.E("marketplace.DiscoverProviders", "read provider directory", nil)
	}
	entries := readDirResult.Value.([]core.FsDirEntry)
	var out []DiscoveredProvider
	for _, entry := range entries {
		if entry == nil || !entry.IsDir() {
			continue
		}
		root := core.PathJoin(dir, entry.Name())
		readResult := core.ReadFile(core.PathJoin(root, ".core", "manifest.yaml"))
		if !readResult.OK {
			continue
		}
		m, err := manifest.Parse(readResult.Value.([]byte))
		if err != nil || m == nil || !m.IsProvider() {
			continue
		}
		out = append(out, DiscoveredProvider{Dir: root, Manifest: m})
	}
	return out, nil
}

func LoadProviderRegistry(path string) (*ProviderRegistryFile, error) {
	readResult := core.ReadFile(path)
	if !readResult.OK {
		return &ProviderRegistryFile{Version: 1, Providers: map[string]ProviderRegistryEntry{}}, nil
	}
	var reg ProviderRegistryFile
	if err := yaml.Unmarshal(readResult.Value.([]byte), &reg); err != nil {
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
	writeResult := core.WriteFile(path, raw, 0o600)
	if !writeResult.OK {
		return core.E("marketplace.SaveProviderRegistry", "write provider registry", nil)
	}
	return nil
}
