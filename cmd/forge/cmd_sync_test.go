// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSyncRepoList_Good(t *testing.T) {
	basePath := filepath.Join(t.TempDir(), "repos")

	repos, err := buildSyncRepoList(nil, []string{"host-uk/core"}, basePath)
	require.NoError(t, err)
	require.Len(t, repos, 1)
	assert.Equal(t, "core", repos[0].name)
	assert.Equal(t, filepath.Join(basePath, "core"), repos[0].localPath)
}

func TestBuildSyncRepoList_Bad_PathTraversal_Good(t *testing.T) {
	basePath := filepath.Join(t.TempDir(), "repos")

	_, err := buildSyncRepoList(nil, []string{"../escape"}, basePath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid repo argument")
}

func TestBuildSyncRepoList_Good_OwnerRepo_Good(t *testing.T) {
	basePath := filepath.Join(t.TempDir(), "repos")

	repos, err := buildSyncRepoList(nil, []string{"Host-UK/core"}, basePath)
	require.NoError(t, err)
	require.Len(t, repos, 1)
	assert.Equal(t, "core", repos[0].name)
	assert.Equal(t, filepath.Join(basePath, "core"), repos[0].localPath)
}

func TestBuildSyncRepoList_Bad_PathTraversal_OwnerRepo_Good(t *testing.T) {
	basePath := filepath.Join(t.TempDir(), "repos")

	_, err := buildSyncRepoList(nil, []string{"host-uk/../escape"}, basePath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid repo argument")
}

func TestBuildSyncRepoList_Bad_PathTraversal_OwnerRepoEncoded_Good(t *testing.T) {
	basePath := filepath.Join(t.TempDir(), "repos")

	_, err := buildSyncRepoList(nil, []string{"host-uk%2F..%2Fescape"}, basePath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid repo argument")
}
