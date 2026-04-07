// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetPRMeta_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	meta, err := client.GetPRMeta("test-org", "org-repo", 1)
	require.NoError(t, err)
	assert.Equal(t, "PR 1", meta.Title)
	assert.Equal(t, "open", meta.State)
	assert.Equal(t, "feature", meta.Branch)
	assert.Equal(t, "main", meta.BaseBranch)
	assert.Equal(t, "author", meta.Author)
	assert.Contains(t, meta.Labels, "enhancement")
	assert.Contains(t, meta.Assignees, "dev1")
	assert.False(t, meta.IsMerged)
}

func TestClient_GetPRMeta_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.GetPRMeta("test-org", "org-repo", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get PR metadata")
}

func TestClient_GetCommentBodies_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	comments, err := client.GetCommentBodies("test-org", "org-repo", 1)
	require.NoError(t, err)
	require.Len(t, comments, 2)
	assert.Equal(t, "comment 1", comments[0].Body)
	assert.Equal(t, "user1", comments[0].Author)
	assert.Equal(t, "comment 2", comments[1].Body)
	assert.Equal(t, "user2", comments[1].Author)
}

func TestClient_GetCommentBodies_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.GetCommentBodies("test-org", "org-repo", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get PR comments")
}

func TestClient_GetIssueBody_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	body, err := client.GetIssueBody("test-org", "org-repo", 1)
	require.NoError(t, err)
	assert.Equal(t, "First issue body", body)
}

func TestClient_GetIssueBody_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.GetIssueBody("test-org", "org-repo", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get issue body")
}
