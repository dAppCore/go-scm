// SPDX-License-Identifier: EUPL-1.2

package plugin

import (
	// Note: AX-6 — Registry listing must be deterministic across map iteration (no core sort primitive).
	"sort"

	core "dappco.re/go/core"
	coreio "dappco.re/go/io"
	"dappco.re/go/scm/internal/ax/jsonx"
)

type Registry struct {
	medium   coreio.Medium
	basePath string
	plugins  map[string]*PluginConfig
}

func NewRegistry(m coreio.Medium, basePath string) *Registry {
	return &Registry{medium: m, basePath: basePath, plugins: map[string]*PluginConfig{}}
}

func (r *Registry) Add(cfg *PluginConfig) error {
	if r == nil || cfg == nil {
		return core.E("plugin.Registry.Add", "config is required", nil)
	}
	if r.plugins == nil {
		r.plugins = map[string]*PluginConfig{}
	}
	r.plugins[cfg.Name] = cfg
	return nil
}

func (r *Registry) Get(name string) (*PluginConfig, bool) {
	if r == nil {
		return nil, false
	}
	cfg, ok := r.plugins[name]
	return cfg, ok
}

func (r *Registry) List() []*PluginConfig {
	if r == nil {
		return nil
	}
	names := make([]string, 0, len(r.plugins))
	for name := range r.plugins {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]*PluginConfig, 0, len(names))
	for _, name := range names {
		out = append(out, r.plugins[name])
	}
	return out
}

func (r *Registry) Load() error {
	if r == nil || r.medium == nil {
		return nil
	}
	raw, err := r.medium.Read(r.basePath + "/registry.json")
	if err != nil {
		return nil
	}
	var data struct {
		Plugins map[string]*PluginConfig `json:"plugins"`
	}
	if err := jsonx.Unmarshal([]byte(raw), &data); err != nil {
		return err
	}
	r.plugins = data.Plugins
	if r.plugins == nil {
		r.plugins = map[string]*PluginConfig{}
	}
	return nil
}

func (r *Registry) Remove(name string) error {
	if r == nil {
		return core.E("plugin.Registry.Remove", "registry is required", nil)
	}
	delete(r.plugins, name)
	return nil
}

func (r *Registry) Save() error {
	if r == nil || r.medium == nil {
		return nil
	}
	raw, err := jsonx.MarshalIndent(map[string]any{"plugins": r.plugins}, "", "  ")
	if err != nil {
		return err
	}
	return r.medium.Write(r.basePath+"/registry.json", string(raw))
}
