// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	"testing"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_ListIssues_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	issues, err := client.ListIssues("test-org", "org-repo", ListIssuesOpts{})
	require.NoError(t, err)
	require.Len(t, issues, 2)
	assert.Equal(t, "Issue 1", issues[0].Title)
}

func TestClient_ListIssues_Good_StateMapping(t *testing.T) {
	tests := []struct {
		name  string
		state string
	}{
		{name: "open", state: "open"},
		{name: "closed", state: "closed"},
		{name: "all", state: "all"},
		{name: "default (empty)", state: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, srv := newTestClient(t)
			defer srv.Close()

			_, err := client.ListIssues("test-org", "org-repo", ListIssuesOpts{State: tt.state})
			require.NoError(t, err)
		})
	}
}

func TestClient_ListIssues_Good_CustomPageAndLimit(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	_, err := client.ListIssues("test-org", "org-repo", ListIssuesOpts{
		Page:  2,
		Limit: 10,
	})
	require.NoError(t, err)
}

func TestClient_ListIssues_Bad_ServerError(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.ListIssues("test-org", "org-repo", ListIssuesOpts{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list issues")
}

func TestClient_GetIssue_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	issue, err := client.GetIssue("test-org", "org-repo", 1)
	require.NoError(t, err)
	assert.Equal(t, "Issue 1", issue.Title)
}

func TestClient_GetIssue_Bad_ServerError(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.GetIssue("test-org", "org-repo", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get issue")
}

func TestClient_CreateIssue_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	issue, err := client.CreateIssue("test-org", "org-repo", forgejo.CreateIssueOption{
		Title: "New Issue",
		Body:  "Issue description",
	})
	require.NoError(t, err)
	assert.NotNil(t, issue)
}

func TestClient_CreateIssue_Bad_ServerError(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.CreateIssue("test-org", "org-repo", forgejo.CreateIssueOption{
		Title: "New Issue",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create issue")
}

func TestClient_EditIssue_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	issue, err := client.EditIssue("test-org", "org-repo", 1, forgejo.EditIssueOption{
		Title: "Updated Title",
	})
	require.NoError(t, err)
	assert.NotNil(t, issue)
}

func TestClient_EditIssue_Bad_ServerError(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.EditIssue("test-org", "org-repo", 1, forgejo.EditIssueOption{
		Title: "Updated Title",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to edit issue")
}

func TestClient_AssignIssue_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	err := client.AssignIssue("test-org", "org-repo", 1, []string{"dev1", "dev2"})
	require.NoError(t, err)
}

func TestClient_AssignIssue_Bad_ServerError(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	err := client.AssignIssue("test-org", "org-repo", 1, []string{"dev1"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to assign issue")
}

func TestClient_ListPullRequests_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	prs, err := client.ListPullRequests("test-org", "org-repo", "open")
	require.NoError(t, err)
	require.Len(t, prs, 1)
	assert.Equal(t, "PR 1", prs[0].Title)
}

func TestClient_ListPullRequests_Good_StateMapping(t *testing.T) {
	tests := []struct {
		name  string
		state string
	}{
		{name: "open", state: "open"},
		{name: "closed", state: "closed"},
		{name: "all", state: "all"},
		{name: "default (empty)", state: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, srv := newTestClient(t)
			defer srv.Close()

			_, err := client.ListPullRequests("test-org", "org-repo", tt.state)
			require.NoError(t, err)
		})
	}
}

func TestClient_ListPullRequests_Bad_ServerError(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.ListPullRequests("test-org", "org-repo", "open")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list pull requests")
}

func TestClient_GetPullRequest_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	pr, err := client.GetPullRequest("test-org", "org-repo", 1)
	require.NoError(t, err)
	assert.Equal(t, "PR 1", pr.Title)
}

func TestClient_GetPullRequest_Bad_ServerError(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.GetPullRequest("test-org", "org-repo", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get pull request")
}

func TestClient_CreateIssueComment_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	err := client.CreateIssueComment("test-org", "org-repo", 1, "LGTM")
	require.NoError(t, err)
}

func TestClient_CreateIssueComment_Bad_ServerError(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	err := client.CreateIssueComment("test-org", "org-repo", 1, "LGTM")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create comment")
}

func TestClient_ListIssueComments_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	comments, err := client.ListIssueComments("test-org", "org-repo", 1)
	require.NoError(t, err)
	require.Len(t, comments, 2)
	assert.Equal(t, "comment 1", comments[0].Body)
}

func TestClient_ListIssueComments_Bad_ServerError(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.ListIssueComments("test-org", "org-repo", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list comments")
}

func TestClient_CloseIssue_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	err := client.CloseIssue("test-org", "org-repo", 1)
	require.NoError(t, err)
}

func TestClient_CloseIssue_Bad_ServerError(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	err := client.CloseIssue("test-org", "org-repo", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to close issue")
}
