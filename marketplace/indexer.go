// SPDX-License-Identifier: EUPL-1.2

package marketplace

import (
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
	"fmt"
	"sort"
	"strings"

	"dappco.re/go/core/io"
	"dappco.re/go/core/scm/manifest"
)

const defaultForgeURL = "https://forge.lthn.ai"

// IndexOptions controls how BuildIndex populates marketplace metadata.
// Usage: IndexOptions{...}
type IndexOptions struct {
	// Org is the default organisation used when constructing repo URLs.
	Org string

	// ForgeURL is the base URL used when constructing repo URLs.
	// If empty, the default Forge URL is used.
	ForgeURL string

	// CategoryFn assigns a category to a module code.
	CategoryFn func(code string) string
}

// BuildIndex reads .core/manifest.yaml from each repository root and produces
// a marketplace index. Repositories without a manifest are skipped silently.
// Categories are deduplicated and sorted.
//
// Example:
//
//	idx, err := marketplace.BuildIndex(
//	  io.Local,
//	  []string{"/tmp/core-scm", "/tmp/core-ui"},
//	  marketplace.IndexOptions{Org: "core", ForgeURL: "https://forge.lthn.ai"},
//	)
func BuildIndex(medium io.Medium, repoPaths []string, opts IndexOptions) (*Index, error) {
	if opts.ForgeURL == "" {
		opts.ForgeURL = defaultForgeURL
	}

	idx := &Index{
		Version: IndexVersion,
	}

	seen := make(map[string]bool)
	categories := make(map[string]bool)

	for _, repoPath := range repoPaths {
		m, err := loadIndexManifest(medium, repoPath)
		if err != nil || m == nil || m.Code == "" {
			continue
		}
		if seen[m.Code] {
			continue
		}
		seen[m.Code] = true

		module := Module{
			Code:    m.Code,
			Name:    m.Name,
			SignKey: manifestSignKey(m),
		}
		if opts.Org != "" {
			baseURL := strings.TrimSuffix(opts.ForgeURL, "/")
			module.Repo = fmt.Sprintf("%s/%s/%s.git", baseURL, opts.Org, m.Code)
		}
		if opts.CategoryFn != nil {
			module.Category = opts.CategoryFn(m.Code)
		}
		if module.Category != "" {
			categories[module.Category] = true
		}

		idx.Modules = append(idx.Modules, module)
	}

	sort.Slice(idx.Modules, func(i, j int) bool {
		return idx.Modules[i].Code < idx.Modules[j].Code
	})

	for category := range categories {
		idx.Categories = append(idx.Categories, category)
	}
	sort.Strings(idx.Categories)

	return idx, nil
}

func manifestSignKey(m *manifest.Manifest) string {
	if m == nil {
		return ""
	}
	if key := strings.TrimSpace(m.SignKey); key != "" {
		return key
	}
	return strings.TrimSpace(m.Sign)
}

func loadIndexManifest(medium io.Medium, repoPath string) (*manifest.Manifest, error) {
	coreJSON := filepath.Join(repoPath, "core.json")
	if raw, err := medium.Read(coreJSON); err == nil {
		cm, parseErr := manifest.ParseCompiled([]byte(raw))
		if parseErr != nil {
			return nil, parseErr
		}
		return &cm.Manifest, nil
	}

	return manifest.Load(medium, repoPath)
}
