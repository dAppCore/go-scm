// SPDX-License-Identifier: EUPL-1.2

package syncutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRepoName_Good(t *testing.T) {
	name, err := ParseRepoName("core")
	require.NoError(t, err)
	assert.Equal(t, "core", name)
}

func TestParseRepoName_Good_OwnerRepo(t *testing.T) {
	name, err := ParseRepoName("host-uk/core")
	require.NoError(t, err)
	assert.Equal(t, "core", name)
}

func TestParseRepoName_Bad_PathTraversal(t *testing.T) {
	_, err := ParseRepoName("../escape")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "syncutil.ParseRepoName")
}

func TestParseRepoName_Bad_PathTraversalEncoded(t *testing.T) {
	_, err := ParseRepoName("host-uk%2F..%2Fescape")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "syncutil.ParseRepoName")
}
