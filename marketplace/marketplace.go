// SPDX-License-Identifier: EUPL-1.2

package marketplace

import (
	"encoding/json"
	"sort"
	"strings"

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
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, err
	}
	return &idx, nil
}

func (idx *Index) Find(code string) (Module, bool) {
	if idx == nil {
		return Module{}, false
	}
	for _, mod := range idx.Modules {
		if strings.EqualFold(mod.Code, code) {
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
		if strings.EqualFold(mod.Category, category) {
			out = append(out, mod)
		}
	}
	return out
}

func (idx *Index) Search(query string) []Module {
	if idx == nil {
		return nil
	}
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return append([]Module(nil), idx.Modules...)
	}
	var out []Module
	for _, mod := range idx.Modules {
		if strings.Contains(strings.ToLower(mod.Code), q) || strings.Contains(strings.ToLower(mod.Name), q) || strings.Contains(strings.ToLower(mod.Category), q) {
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
			Version:  strings.TrimSpace(m.Version),
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
	return strings.TrimSpace(layout)
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
