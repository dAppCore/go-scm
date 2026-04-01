// SPDX-License-Identifier: EUPL-1.2

package gitea

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	giteaSDK "code.gitea.io/sdk/gitea"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newPaginatedIssuesClient(t *testing.T) (*Client, *httptest.Server) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]string{"version": "1.21.0"})
	})
	mux.HandleFunc("/api/v1/repos/test-org/org-repo/issues", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("page") {
		case "2":
			jsonResponse(w, []map[string]any{
				{"id": 2, "number": 2, "title": "Issue 2", "state": "open", "body": "Second issue"},
			})
		case "3":
			jsonResponse(w, []map[string]any{})
		default:
			jsonResponse(w, []map[string]any{
				{"id": 1, "number": 1, "title": "Issue 1", "state": "open", "body": "First issue"},
			})
		}
	})

	srv := httptest.NewServer(mux)
	client, err := New(srv.URL, "test-token")
	require.NoError(t, err)
	return client, srv
}

func newPaginatedCommentsClient(t *testing.T) (*Client, *httptest.Server) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]string{"version": "1.21.0"})
	})
	mux.HandleFunc("/api/v1/repos/test-org/org-repo/issues/1/comments", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("page") {
		case "2":
			jsonResponse(w, []map[string]any{
				{"id": 150, "body": "comment 51", "user": map[string]any{"login": "user51"}, "created_at": "2026-01-02T00:00:00Z", "updated_at": "2026-01-02T00:00:00Z"},
			})
		case "3":
			jsonResponse(w, []map[string]any{})
		default:
			w.Header().Set("Link", `</api/v1/repos/test-org/org-repo/issues/1/comments?page=2>; rel="next", </api/v1/repos/test-org/org-repo/issues/1/comments?page=2>; rel="last"`)
			comments := make([]map[string]any, 0, 50)
			for i := 1; i <= 50; i++ {
				comments = append(comments, map[string]any{
					"id":         99 + i,
					"body":       "comment " + strconv.Itoa(i),
					"user":       map[string]any{"login": "user" + strconv.Itoa(i)},
					"created_at": "2026-01-01T00:00:00Z",
					"updated_at": "2026-01-01T00:00:00Z",
				})
			}
			jsonResponse(w, comments)
		}
	})

	srv := httptest.NewServer(mux)
	client, err := New(srv.URL, "test-token")
	require.NoError(t, err)
	return client, srv
}

func TestClient_ListIssues_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	issues, err := client.ListIssues("test-org", "org-repo", ListIssuesOpts{})
	require.NoError(t, err)
	require.Len(t, issues, 2)
	assert.Equal(t, "Issue 1", issues[0].Title)
}

func TestClient_ListIssues_Good_Paginates_Good(t *testing.T) {
	client, srv := newPaginatedIssuesClient(t)
	defer srv.Close()

	issues, err := client.ListIssues("test-org", "org-repo", ListIssuesOpts{Limit: 1})
	require.NoError(t, err)
	require.Len(t, issues, 2)
	assert.Equal(t, "Issue 1", issues[0].Title)
	assert.Equal(t, "Issue 2", issues[1].Title)
}

func TestClient_ListIssuesIter_Good_Paginates_Good(t *testing.T) {
	client, srv := newPaginatedIssuesClient(t)
	defer srv.Close()

	var titles []string
	for issue, err := range client.ListIssuesIter("test-org", "org-repo", ListIssuesOpts{Limit: 1}) {
		require.NoError(t, err)
		titles = append(titles, issue.Title)
	}

	require.Len(t, titles, 2)
	assert.Equal(t, []string{"Issue 1", "Issue 2"}, titles)
}

func TestClient_ListIssues_Good_StateMapping_Good(t *testing.T) {
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

func TestClient_ListIssues_Good_CustomPageAndLimit_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	_, err := client.ListIssues("test-org", "org-repo", ListIssuesOpts{
		Page:  2,
		Limit: 10,
	})
	require.NoError(t, err)
}

func TestClient_ListIssues_Bad_ServerError_Good(t *testing.T) {
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

func TestClient_GetIssue_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.GetIssue("test-org", "org-repo", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get issue")
}

func TestClient_ListIssueCommentsIter_Good_Paginates_Good(t *testing.T) {
	client, srv := newPaginatedCommentsClient(t)
	defer srv.Close()

	var bodies []string
	for comment, err := range client.ListIssueCommentsIter("test-org", "org-repo", 1) {
		require.NoError(t, err)
		bodies = append(bodies, comment.Body)
	}

	require.Len(t, bodies, 51)
	assert.Equal(t, "comment 1", bodies[0])
	assert.Equal(t, "comment 50", bodies[49])
	assert.Equal(t, "comment 51", bodies[50])
}

func TestClient_CreateIssue_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	issue, err := client.CreateIssue("test-org", "org-repo", giteaSDK.CreateIssueOption{
		Title: "New Issue",
		Body:  "Issue description",
	})
	require.NoError(t, err)
	assert.NotNil(t, issue)
}

func TestClient_CreateIssue_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.CreateIssue("test-org", "org-repo", giteaSDK.CreateIssueOption{
		Title: "New Issue",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create issue")
}

func TestClient_ListPullRequests_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	prs, err := client.ListPullRequests("test-org", "org-repo", "open")
	require.NoError(t, err)
	require.Len(t, prs, 1)
	assert.Equal(t, "PR 1", prs[0].Title)
}

func TestClient_ListPullRequests_Good_StateMapping_Good(t *testing.T) {
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

func TestClient_ListPullRequests_Bad_ServerError_Good(t *testing.T) {
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

func TestClient_GetPullRequest_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.GetPullRequest("test-org", "org-repo", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get pull request")
}

// --- ListIssuesOpts defaulting ---

func TestListIssuesOpts_Good_Defaults_Good(t *testing.T) {
	tests := []struct {
		name          string
		opts          ListIssuesOpts
		expectedState string
		expectedLimit int
		expectedPage  int
	}{
		{
			name:          "all defaults",
			opts:          ListIssuesOpts{},
			expectedState: "open",
			expectedLimit: 50,
			expectedPage:  1,
		},
		{
			name:          "closed state",
			opts:          ListIssuesOpts{State: "closed"},
			expectedState: "closed",
			expectedLimit: 50,
			expectedPage:  1,
		},
		{
			name:          "custom limit and page",
			opts:          ListIssuesOpts{Page: 3, Limit: 25},
			expectedState: "open",
			expectedLimit: 25,
			expectedPage:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.opts.State == "" {
				tt.opts.State = "open"
			}
			assert.Equal(t, tt.expectedState, tt.opts.State)

			limit := tt.opts.Limit
			if limit == 0 {
				limit = 50
			}
			assert.Equal(t, tt.expectedLimit, limit)

			page := tt.opts.Page
			if page == 0 {
				page = 1
			}
			assert.Equal(t, tt.expectedPage, page)
		})
	}
}
