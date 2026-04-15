// SPDX-License-Identifier: EUPL-1.2

package scm

import (
	"testing"

	"dappco.re/go/core/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadRegistry_Good_RootWrapper_Good(t *testing.T) {
	m := io.NewMockMedium()
	require.NoError(t, m.Write("/tmp/repos.yaml", `
version: 1
org: core
base_path: /tmp/code
repos:
  go-scm:
    type: module
`))

	reg, err := LoadRegistry(m, "/tmp/repos.yaml")
	require.NoError(t, err)
	require.NotNil(t, reg)
	assert.Equal(t, "core", reg.Org)

	repo, ok := reg.Get("go-scm")
	require.True(t, ok)
	assert.Equal(t, "/tmp/code/go-scm", repo.Path)
}

func TestMergeRegistries_Good_RootWrapper_Good(t *testing.T) {
	first := NewRegistry()
	first.Repos["alpha"] = &Repo{Name: "alpha", Path: "/tmp/alpha"}

	second := NewRegistry()
	second.Repos["alpha"] = &Repo{Name: "alpha", Path: "/tmp/override"}
	second.Repos["beta"] = &Repo{Name: "beta", Path: "/tmp/beta"}

	merged := MergeRegistries(first, second)
	require.NotNil(t, merged)

	repos := merged.List()
	require.Len(t, repos, 2)
	assert.Equal(t, "alpha", repos[0].Name)
	assert.Equal(t, "/tmp/alpha", repos[0].Path)
	assert.Equal(t, "beta", repos[1].Name)
}
