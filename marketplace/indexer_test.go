// SPDX-License-Identifier: EUPL-1.2

package marketplace

import (
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
	"testing"

	"dappco.re/go/core/io"
	json "dappco.re/go/core/scm/internal/ax/jsonx"
	"dappco.re/go/core/scm/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildIndex_Good_CategoriesAndRepoURLs_Good(t *testing.T) {
	medium := io.NewMockMedium()

	require.NoError(t, medium.Write("/repos/a/.core/manifest.yaml", `
code: a
name: Alpha
version: 1.0.0
sign: key-a
`))
	require.NoError(t, medium.Write("/repos/b/.core/manifest.yaml", `
code: b
name: Beta
version: 1.0.0
sign: key-b
`))
	require.NoError(t, medium.Write("/repos/c/.core/manifest.yaml", `
code: c
name: Gamma
version: 1.0.0
sign: key-c
`))

	idx, err := BuildIndex(medium, []string{"/repos/a", "/repos/b", "/repos/c"}, IndexOptions{
		ForgeURL: "https://forge.example.com",
		Org:      "core",
		CategoryFn: func(code string) string {
			switch code {
			case "a", "b":
				return "tools"
			default:
				return "products"
			}
		},
	})
	require.NoError(t, err)

	require.Len(t, idx.Modules, 3)
	assert.Equal(t, "a", idx.Modules[0].Code)
	assert.Equal(t, "https://forge.example.com/core/a.git", idx.Modules[0].Repo)
	assert.Equal(t, "tools", idx.Modules[0].Category)
	assert.Equal(t, "key-a", idx.Modules[0].SignKey)
	assert.Equal(t, []string{"products", "tools"}, idx.Categories)
}

func TestBuildIndex_Good_SkipsMissingManifest_Good(t *testing.T) {
	medium := io.NewMockMedium()

	require.NoError(t, medium.Write("/repos/one/.core/manifest.yaml", `
code: one
name: One
version: 1.0.0
sign: key-one
`))

	idx, err := BuildIndex(medium, []string{"/repos/one", "/repos/missing"}, IndexOptions{})
	require.NoError(t, err)

	require.Len(t, idx.Modules, 1)
	assert.Equal(t, "one", idx.Modules[0].Code)
	assert.Empty(t, idx.Categories)
}

func TestBuildIndex_Good_PrefersCompiledManifest_Good(t *testing.T) {
	medium := io.NewMockMedium()

	repoDir := "/repos/compiled"
	require.NoError(t, medium.EnsureDir(filepath.Join(repoDir, ".core")))

	cm := &manifest.CompiledManifest{
		Manifest: manifest.Manifest{
			Code:    "compiled",
			Name:    "Compiled Module",
			Version: "2.0.0",
			Sign:    "key-compiled",
		},
		Commit: "abc123",
	}
	raw, err := json.Marshal(cm)
	require.NoError(t, err)
	require.NoError(t, medium.Write(filepath.Join(repoDir, "core.json"), string(raw)))
	require.NoError(t, medium.Write(filepath.Join(repoDir, ".core", "manifest.yaml"), `
code: source
name: Source Module
version: 1.0.0
sign: key-source
`))

	idx, err := BuildIndex(medium, []string{repoDir}, IndexOptions{})
	require.NoError(t, err)

	require.Len(t, idx.Modules, 1)
	assert.Equal(t, "compiled", idx.Modules[0].Code)
	assert.Equal(t, "key-compiled", idx.Modules[0].SignKey)
}

func TestBuildIndex_Good_PrefersSignKeyField_Good(t *testing.T) {
	medium := io.NewMockMedium()

	require.NoError(t, medium.Write("/repos/signed/.core/manifest.yaml", `
code: signed
name: Signed Module
version: 1.0.0
sign: signature-fallback
sign_key: public-key-preferred
`))

	idx, err := BuildIndex(medium, []string{"/repos/signed"}, IndexOptions{})
	require.NoError(t, err)

	require.Len(t, idx.Modules, 1)
	assert.Equal(t, "public-key-preferred", idx.Modules[0].SignKey)
}
