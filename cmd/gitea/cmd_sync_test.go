// SPDX-Licence-Identifier: EUPL-1.2

package gitea

import (
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildRepoList_Good(t *testing.T) {
	basePath := filepath.Join(t.TempDir(), "repos")

	repos, err := buildRepoList(nil, []string{"host-uk/core"}, basePath)
	require.NoError(t, err)
	require.Len(t, repos, 1)
	assert.Equal(t, "core", repos[0].name)
	assert.Equal(t, filepath.Join(basePath, "core"), repos[0].localPath)
}

func TestBuildRepoList_Bad_PathTraversal(t *testing.T) {
	basePath := filepath.Join(t.TempDir(), "repos")

	_, err := buildRepoList(nil, []string{"../escape"}, basePath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid repo argument")
}

func TestBuildRepoList_Good_OwnerRepo(t *testing.T) {
	basePath := filepath.Join(t.TempDir(), "repos")

	repos, err := buildRepoList(nil, []string{"Host-UK/core"}, basePath)
	require.NoError(t, err)
	require.Len(t, repos, 1)
	assert.Equal(t, "core", repos[0].name)
	assert.Equal(t, filepath.Join(basePath, "core"), repos[0].localPath)
}

func TestBuildRepoList_Bad_PathTraversal_OwnerRepo(t *testing.T) {
	basePath := filepath.Join(t.TempDir(), "repos")

	_, err := buildRepoList(nil, []string{"host-uk/../escape"}, basePath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid repo argument")
}

func TestBuildRepoList_Bad_PathTraversal_OwnerRepoEncoded(t *testing.T) {
	basePath := filepath.Join(t.TempDir(), "repos")

	_, err := buildRepoList(nil, []string{"host-uk%2F..%2Fescape"}, basePath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid repo argument")
}
