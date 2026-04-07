// SPDX-License-Identifier: EUPL-1.2

package marketplace

import (
	json "dappco.re/go/core/scm/internal/ax/jsonx"
	strings "dappco.re/go/core/scm/internal/ax/stringsx"

	"dappco.re/go/core/io"
	coreerr "dappco.re/go/core/log"
)

// Module is a marketplace entry pointing to a module's Git repo.
type Module struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Repo     string `json:"repo"`
	SignKey  string `json:"sign_key"`
	Category string `json:"category"`
}

// Index is the root marketplace catalog.
type Index struct {
	Version    int      `json:"version"`
	Modules    []Module `json:"modules"`
	Categories []string `json:"categories"`
}

// ParseIndex decodes a marketplace index.json.
// Usage: ParseIndex(...)
func ParseIndex(data []byte) (*Index, error) {
	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, coreerr.E("marketplace.ParseIndex", "unmarshal failed", err)
	}
	return &idx, nil
}

// LoadIndex reads and parses a marketplace index.json from the given path.
// Usage: LoadIndex(...)
func LoadIndex(m io.Medium, path string) (*Index, error) {
	data, err := m.Read(path)
	if err != nil {
		return nil, coreerr.E("marketplace.LoadIndex", "read failed", err)
	}
	return ParseIndex([]byte(data))
}

// Search returns modules matching the query in code, name, or category.
// Usage: Search(...)
func (idx *Index) Search(query string) []Module {
	q := strings.ToLower(query)
	var results []Module
	for _, m := range idx.Modules {
		if strings.Contains(strings.ToLower(m.Code), q) ||
			strings.Contains(strings.ToLower(m.Name), q) ||
			strings.Contains(strings.ToLower(m.Category), q) {
			results = append(results, m)
		}
	}
	return results
}

// ByCategory returns all modules in the given category.
// Usage: ByCategory(...)
func (idx *Index) ByCategory(category string) []Module {
	var results []Module
	for _, m := range idx.Modules {
		if m.Category == category {
			results = append(results, m)
		}
	}
	return results
}

// Find returns the module with the given code, or false if not found.
// Usage: Find(...)
func (idx *Index) Find(code string) (Module, bool) {
	for _, m := range idx.Modules {
		if m.Code == code {
			return m, true
		}
	}
	return Module{}, false
}
