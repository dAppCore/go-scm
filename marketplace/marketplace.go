// SPDX-License-Identifier: EUPL-1.2

package marketplace

import (
	// Note: AX-6 — Marketplace modules and categories must be ordered deterministically (no core sort primitive).
	"sort"

	core "dappco.re/go"
	"dappco.re/go/scm/manifest"
)

type Module struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Version  string `json:"version,omitempty"`
	Repo     string `json:"repo"`
	Sign     string `json:"sign,omitempty"`
	SignKey  string `json:"sign_key"`
	Category string `json:"category"`
}

type Index struct {
	Version    int      `json:"version"`
	Modules    []Module `json:"modules"`
	Categories []string `json:"categories"`
}

func ParseIndex(data []byte) (*Index, error) {
	var idx Index
	r := core.JSONUnmarshal(data, &idx)
	if !r.OK {
		return nil, resultError("marketplace.ParseIndex", "unmarshal index", r)
	}
	return &idx, nil
}

func (idx *Index) Find(code string) (Module, bool) {
	if idx == nil {
		return Module{}, false
	}
	for _, mod := range idx.Modules {
		if core.Lower(mod.Code) == core.Lower(code) {
			return mod, true
		}
	}
	return Module{}, false
}

func (idx *Index) ByCategory(category string) []Module {
	if idx == nil {
		return nil
	}
	var out []Module
	for _, mod := range idx.Modules {
		if core.Lower(mod.Category) == core.Lower(category) {
			out = append(out, mod)
		}
	}
	return out
}

func (idx *Index) Search(query string) []Module {
	if idx == nil {
		return nil
	}
	q := core.Lower(core.Trim(query))
	if q == "" {
		return append([]Module(nil), idx.Modules...)
	}
	var out []Module
	for _, mod := range idx.Modules {
		if core.Contains(core.Lower(mod.Code), q) || core.Contains(core.Lower(mod.Name), q) || core.Contains(core.Lower(mod.Category), q) {
			out = append(out, mod)
		}
	}
	return out
}

func BuildIndexFromManifests(manifests []*manifest.Manifest) *Index {
	idx := &Index{Version: 1}
	if len(manifests) == 0 {
		return idx
	}
	cats := map[string]struct{}{}
	for _, m := range manifests {
		if m == nil {
			continue
		}
		mod := Module{
			Code:     m.Code,
			Name:     m.Name,
			Version:  core.Trim(m.Version),
			Repo:     "",
			Sign:     m.Sign,
			SignKey:  m.SignKey,
			Category: firstCategory(m.Modules, m.Layout),
		}
		if mod.Version == "" {
			mod.Version = "latest"
		}
		idx.Modules = append(idx.Modules, mod)
		if mod.Category != "" {
			cats[mod.Category] = struct{}{}
		}
	}
	idx.Categories = sortedKeys(cats)
	sort.SliceStable(idx.Modules, func(i, j int) bool { return idx.Modules[i].Code < idx.Modules[j].Code })
	return idx
}

func firstCategory(modules []string, layout string) string {
	if len(modules) > 0 {
		return modules[0]
	}
	return core.Trim(layout)
}

func sortedKeys(m map[string]struct{}) []string {
	if len(m) == 0 {
		return nil
	}
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func resultError(op, msg string, r core.Result) error {
	if err, ok := r.Value.(error); ok {
		return core.E(op, msg, err)
	}
	return core.E(op, msg, nil)
}
