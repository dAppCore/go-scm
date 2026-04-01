// SPDX-License-Identifier: EUPL-1.2

package marketplace

import (
	"testing"

	"dappco.re/go/core/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildIndex_Good_CategoriesAndRepoURLs_Good(t *testing.T) {
	medium := io.NewMockMedium()

	require.NoError(t, medium.Write("/repos/a/.core/manifest.yaml", `
code: a
name: Alpha
version: 1.0.0
`))
	require.NoError(t, medium.Write("/repos/b/.core/manifest.yaml", `
code: b
name: Beta
version: 1.0.0
`))
	require.NoError(t, medium.Write("/repos/c/.core/manifest.yaml", `
code: c
name: Gamma
version: 1.0.0
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
	assert.Equal(t, "https://forge.example.com/core/a", idx.Modules[0].Repo)
	assert.Equal(t, "tools", idx.Modules[0].Category)
	assert.Equal(t, []string{"products", "tools"}, idx.Categories)
}

func TestBuildIndex_Good_SkipsMissingManifest_Good(t *testing.T) {
	medium := io.NewMockMedium()

	require.NoError(t, medium.Write("/repos/one/.core/manifest.yaml", `
code: one
name: One
version: 1.0.0
`))

	idx, err := BuildIndex(medium, []string{"/repos/one", "/repos/missing"}, IndexOptions{})
	require.NoError(t, err)

	require.Len(t, idx.Modules, 1)
	assert.Equal(t, "one", idx.Modules[0].Code)
	assert.Empty(t, idx.Categories)
}
