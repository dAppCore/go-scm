# Manifest → core.json Pipeline Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a build step that compiles `.core/manifest.yaml` into a `core.json` distribution artifact at the repo root, and a catalogue generator that indexes `core.json` files across repos into a marketplace index.

**Architecture:** `manifest.Compile()` reads `.core/manifest.yaml`, injects version/commit metadata, and writes `core.json` at the distribution root. `marketplace.BuildIndex()` crawls a repos registry (or forge org), extracts manifests, and produces an `index.json` catalogue. Both use go-io Medium for filesystem abstraction. Tests use `io.NewMockMedium()`.

**Tech Stack:** Go, go-io Medium, go-scm manifest/marketplace packages, testify

---

## File Structure

| File | Action | Purpose |
|------|--------|---------|
| `manifest/compile.go` | Create | `Compile()` — manifest.yaml → core.json with build metadata |
| `manifest/compile_test.go` | Create | Tests for compilation, metadata injection, signing |
| `marketplace/indexer.go` | Create | `BuildIndex()` — crawl repos, extract manifests, build catalogue |
| `marketplace/indexer_test.go` | Create | Tests for indexing, dedup, category extraction |

---

## Task 1: Manifest Compilation (manifest.yaml → core.json)

**Files:**
- Create: `manifest/compile.go`
- Create: `manifest/compile_test.go`

The `Compile` function reads `.core/manifest.yaml`, injects build metadata (version, commit, build time), and writes `core.json` at the target root. The output is JSON (not YAML) so consumers don't need a YAML parser.

- [ ] **Step 1: Write the failing test for Compile**

```go
// compile_test.go
package manifest

import (
	"encoding/json"
	"testing"

	io "dappco.re/go/core/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompile_Good(t *testing.T) {
	medium := io.NewMockMedium()

	// Write a manifest.yaml
	manifest := `code: core/api
name: Core API
version: 0.1.0
namespace: api
binary: ./bin/core-api
licence: EUPL-1.2
`
	medium.WriteString("/project/.core/manifest.yaml", manifest)

	// Compile with build metadata
	err := Compile(medium, "/project", CompileOptions{
		Version: "1.2.3",
		Commit:  "abc1234",
	})
	require.NoError(t, err)

	// Read core.json
	data, err := medium.Read("/project/core.json")
	require.NoError(t, err)

	var result CompiledManifest
	require.NoError(t, json.Unmarshal([]byte(data), &result))

	assert.Equal(t, "core/api", result.Code)
	assert.Equal(t, "Core API", result.Name)
	assert.Equal(t, "1.2.3", result.Version)
	assert.Equal(t, "abc1234", result.Commit)
	assert.Equal(t, "api", result.Namespace)
	assert.NotEmpty(t, result.BuiltAt)
}

func TestCompile_Good_PreservesSign(t *testing.T) {
	medium := io.NewMockMedium()

	manifest := `code: core/api
name: Core API
version: 0.1.0
sign: "dGVzdHNpZw=="
`
	medium.WriteString("/project/.core/manifest.yaml", manifest)

	err := Compile(medium, "/project", CompileOptions{})
	require.NoError(t, err)

	data, err := medium.Read("/project/core.json")
	require.NoError(t, err)

	var result CompiledManifest
	require.NoError(t, json.Unmarshal([]byte(data), &result))

	assert.Equal(t, "dGVzdHNpZw==", result.Sign)
}

func TestCompile_Bad_NoManifest(t *testing.T) {
	medium := io.NewMockMedium()

	err := Compile(medium, "/project", CompileOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "manifest.Compile")
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v -run TestCompile ./manifest/`
Expected: FAIL — `Compile` undefined

- [ ] **Step 3: Write minimal implementation**

```go
// compile.go
package manifest

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	io "dappco.re/go/core/io"
)

// CompiledManifest is the core.json distribution format.
// Embeds the full Manifest plus build metadata.
type CompiledManifest struct {
	Manifest

	// Build metadata — injected at compile time, not in source manifest.
	Commit  string `json:"commit,omitempty"`
	BuiltAt string `json:"built_at,omitempty"`
}

// CompileOptions controls what metadata is injected during compilation.
type CompileOptions struct {
	Version string // Override version (e.g. from git tag)
	Commit  string // Git commit hash
	Output  string // Output path (default: "core.json" at root)
}

// Compile reads .core/manifest.yaml, injects build metadata, and writes
// core.json at the distribution root.
func Compile(medium io.Medium, root string, opts CompileOptions) error {
	m, err := Load(medium, root)
	if err != nil {
		return fmt.Errorf("manifest.Compile: %w", err)
	}

	compiled := CompiledManifest{
		Manifest: *m,
		Commit:   opts.Commit,
		BuiltAt:  time.Now().UTC().Format(time.RFC3339),
	}

	// Override version if provided (e.g. from git tag)
	if opts.Version != "" {
		compiled.Version = opts.Version
	}

	data, err := json.MarshalIndent(compiled, "", "  ")
	if err != nil {
		return fmt.Errorf("manifest.Compile: marshal: %w", err)
	}

	outPath := opts.Output
	if outPath == "" {
		outPath = filepath.Join(root, "core.json")
	}

	if err := medium.Write(outPath, string(data)); err != nil {
		return fmt.Errorf("manifest.Compile: write: %w", err)
	}

	return nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v -run TestCompile ./manifest/`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add manifest/compile.go manifest/compile_test.go
git commit -m "feat(manifest): compile .core/manifest.yaml to core.json

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

## Task 2: Marketplace Index Builder

**Files:**
- Create: `marketplace/indexer.go`
- Create: `marketplace/indexer_test.go`

The `BuildIndex` function takes a list of directory paths (repos), loads each `.core/manifest.yaml`, extracts Module entries, deduplicates categories, and produces an `Index`.

- [ ] **Step 1: Write the failing test for BuildIndex**

```go
// indexer_test.go
package marketplace

import (
	"testing"

	io "dappco.re/go/core/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildIndex_Good(t *testing.T) {
	medium := io.NewMockMedium()

	// Two repos with manifests
	medium.WriteString("/repos/core-api/.core/manifest.yaml", `
code: core/api
name: Core API
version: 0.1.0
namespace: api
binary: ./bin/api
`)
	medium.WriteString("/repos/core-bio/.core/manifest.yaml", `
code: core/bio
name: Bio
version: 0.2.0
namespace: bio
binary: ./bin/bio
`)

	idx, err := BuildIndex(medium, []string{"/repos/core-api", "/repos/core-bio"}, IndexOptions{
		Org: "core",
	})
	require.NoError(t, err)

	assert.Equal(t, 1, idx.Version)
	assert.Len(t, idx.Modules, 2)
	assert.Equal(t, "core/api", idx.Modules[0].Code)
	assert.Equal(t, "core/bio", idx.Modules[1].Code)
}

func TestBuildIndex_Good_SkipsMissingManifest(t *testing.T) {
	medium := io.NewMockMedium()

	// Only one repo has a manifest
	medium.WriteString("/repos/core-api/.core/manifest.yaml", `
code: core/api
name: Core API
version: 0.1.0
`)

	idx, err := BuildIndex(medium, []string{"/repos/core-api", "/repos/no-manifest"}, IndexOptions{})
	require.NoError(t, err)

	assert.Len(t, idx.Modules, 1)
}

func TestBuildIndex_Good_ExtractsCategories(t *testing.T) {
	medium := io.NewMockMedium()

	medium.WriteString("/repos/a/.core/manifest.yaml", `
code: a
name: A
`)
	medium.WriteString("/repos/b/.core/manifest.yaml", `
code: b
name: B
`)

	idx, err := BuildIndex(medium, []string{"/repos/a", "/repos/b"}, IndexOptions{
		CategoryFn: func(code string) string {
			if code == "a" {
				return "tools"
			}
			return "products"
		},
	})
	require.NoError(t, err)

	assert.Contains(t, idx.Categories, "tools")
	assert.Contains(t, idx.Categories, "products")
}

func TestBuildIndex_Bad_EmptyList(t *testing.T) {
	medium := io.NewMockMedium()

	idx, err := BuildIndex(medium, []string{}, IndexOptions{})
	require.NoError(t, err)
	assert.Len(t, idx.Modules, 0)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v -run TestBuildIndex ./marketplace/`
Expected: FAIL — `BuildIndex` undefined

- [ ] **Step 3: Write minimal implementation**

```go
// indexer.go
package marketplace

import (
	"fmt"
	"sort"

	io "dappco.re/go/core/io"
	"dappco.re/go/core/scm/manifest"
)

// IndexOptions controls how the index is built.
type IndexOptions struct {
	Org        string                 // Git org for repo URLs (e.g. "core")
	ForgeURL   string                 // Forge base URL (default: "https://forge.lthn.ai")
	CategoryFn func(code string) string // Optional function to assign category by code
}

// BuildIndex reads .core/manifest.yaml from each repo path and produces
// a marketplace Index. Repos without a manifest are silently skipped.
func BuildIndex(medium io.Medium, repoPaths []string, opts IndexOptions) (*Index, error) {
	if opts.ForgeURL == "" {
		opts.ForgeURL = "https://forge.lthn.ai"
	}

	idx := &Index{Version: 1}
	seen := make(map[string]bool)
	catSet := make(map[string]bool)

	for _, repoPath := range repoPaths {
		m, err := manifest.Load(medium, repoPath)
		if err != nil {
			continue // Skip repos without manifest
		}

		if m.Code == "" {
			continue
		}

		if seen[m.Code] {
			continue // Deduplicate
		}
		seen[m.Code] = true

		module := Module{
			Code:    m.Code,
			Name:    m.Name,
			SignKey: m.Sign,
		}

		// Build repo URL
		if opts.Org != "" {
			module.Repo = fmt.Sprintf("%s/%s/%s.git", opts.ForgeURL, opts.Org, m.Code)
		}

		// Assign category
		if opts.CategoryFn != nil {
			module.Category = opts.CategoryFn(m.Code)
		}
		if module.Category != "" {
			catSet[module.Category] = true
		}

		idx.Modules = append(idx.Modules, module)
	}

	// Sort categories
	for cat := range catSet {
		idx.Categories = append(idx.Categories, cat)
	}
	sort.Strings(idx.Categories)

	return idx, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v -run TestBuildIndex ./marketplace/`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add marketplace/indexer.go marketplace/indexer_test.go
git commit -m "feat(marketplace): index builder — crawl repos, build catalogue

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

## Summary

**Total: 2 tasks, ~10 steps**

After completion:
- `manifest.Compile()` produces `core.json` at distribution root
- `marketplace.BuildIndex()` crawls repo paths and produces `index.json`
- Both are testable via mock Medium (no filesystem)
- Ready for integration into `core build` and `core scm index` CLI commands
