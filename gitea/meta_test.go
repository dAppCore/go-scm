// SPDX-Licence-Identifier: EUPL-1.2

package gitea

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

func TestClient_GetPRMeta_Bad_ServerError(t *testing.T) {
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

func TestClient_GetCommentBodies_Bad_ServerError(t *testing.T) {
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

func TestClient_GetIssueBody_Bad_ServerError(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.GetIssueBody("test-org", "org-repo", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get issue body")
}

// --- PRMeta struct tests ---

func TestPRMeta_Good_Fields(t *testing.T) {
	meta := &PRMeta{
		Number:       42,
		Title:        "Test PR",
		State:        "open",
		Author:       "testuser",
		Branch:       "feature/test",
		BaseBranch:   "main",
		Labels:       []string{"bug", "urgent"},
		Assignees:    []string{"dev1", "dev2"},
		IsMerged:     false,
		CommentCount: 5,
	}

	assert.Equal(t, int64(42), meta.Number)
	assert.Equal(t, "Test PR", meta.Title)
	assert.Equal(t, "open", meta.State)
	assert.Equal(t, "testuser", meta.Author)
	assert.Equal(t, "feature/test", meta.Branch)
	assert.Equal(t, "main", meta.BaseBranch)
	assert.Equal(t, []string{"bug", "urgent"}, meta.Labels)
	assert.Equal(t, []string{"dev1", "dev2"}, meta.Assignees)
	assert.False(t, meta.IsMerged)
	assert.Equal(t, 5, meta.CommentCount)
}

func TestComment_Good_Fields(t *testing.T) {
	comment := Comment{
		ID:     123,
		Author: "reviewer",
		Body:   "LGTM",
	}

	assert.Equal(t, int64(123), comment.ID)
	assert.Equal(t, "reviewer", comment.Author)
	assert.Equal(t, "LGTM", comment.Body)
}

func TestCommentPageSize_Good(t *testing.T) {
	assert.Equal(t, 50, commentPageSize, "comment page size should be 50")
}
