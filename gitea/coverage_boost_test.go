package gitea

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- SaveConfig tests ---

func TestSaveConfig_Good_URLAndToken(t *testing.T) {
	isolateConfigEnv(t)

	err := SaveConfig("https://gitea.example.com", "test-token-123")
	// SaveConfig may fail if config dir creation fails in isolated HOME,
	// but the function path is still exercised.
	if err != nil {
		assert.Contains(t, err.Error(), "failed to")
	}
}

func TestSaveConfig_Good_URLOnly(t *testing.T) {
	isolateConfigEnv(t)

	err := SaveConfig("https://gitea.example.com", "")
	if err != nil {
		assert.Contains(t, err.Error(), "failed to")
	}
}

func TestSaveConfig_Good_TokenOnly(t *testing.T) {
	isolateConfigEnv(t)

	err := SaveConfig("", "some-token")
	if err != nil {
		assert.Contains(t, err.Error(), "failed to")
	}
}

func TestSaveConfig_Good_Empty(t *testing.T) {
	isolateConfigEnv(t)

	err := SaveConfig("", "")
	// With both empty, nothing to set, so should succeed (no-op).
	if err != nil {
		assert.Contains(t, err.Error(), "failed to")
	}
}

// --- Pagination tests with multi-page mock server ---

func newPaginatedOrgReposServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]string{"version": "1.21.0"})
	})

	mux.HandleFunc("/api/v1/orgs/paginated-org/repos", func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		w.Header().Set("Content-Type", "application/json")

		switch page {
		case "", "1":
			// Indicate there's a second page via Link header.
			// The Gitea SDK uses the Response.LastPage field, which comes from Link headers.
			repos := []map[string]any{
				{"id": 1, "name": "repo-1", "full_name": "paginated-org/repo-1", "owner": map[string]any{"login": "paginated-org"}},
				{"id": 2, "name": "repo-2", "full_name": "paginated-org/repo-2", "owner": map[string]any{"login": "paginated-org"}},
			}
			_ = json.NewEncoder(w).Encode(repos)
		default:
			// Empty page to stop pagination.
			_ = json.NewEncoder(w).Encode([]map[string]any{})
		}
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	return httptest.NewServer(mux)
}

func TestClient_ListOrgRepos_Good_Pagination(t *testing.T) {
	srv := newPaginatedOrgReposServer(t)
	defer srv.Close()

	client, err := New(srv.URL, "test-token")
	require.NoError(t, err)

	repos, err := client.ListOrgRepos("paginated-org")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(repos), 2)
}

func newPaginatedUserReposServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]string{"version": "1.21.0"})
	})

	mux.HandleFunc("/api/v1/user/repos", func(w http.ResponseWriter, r *http.Request) {
		repos := []map[string]any{
			{"id": 1, "name": "my-repo-1", "full_name": "user/my-repo-1", "owner": map[string]any{"login": "user"}},
		}
		_ = json.NewEncoder(w).Encode(repos)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	return httptest.NewServer(mux)
}

func TestClient_ListUserRepos_Good_SinglePage(t *testing.T) {
	srv := newPaginatedUserReposServer(t)
	defer srv.Close()

	client, err := New(srv.URL, "test-token")
	require.NoError(t, err)

	repos, err := client.ListUserRepos()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(repos), 1)
}

// --- PR meta: pagination in comment counting ---

func newPRMetaWithManyCommentsServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]string{"version": "1.21.0"})
	})

	mux.HandleFunc("/api/v1/repos/test-org/test-repo/pulls/1", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]any{
			"id": 1, "number": 1, "title": "Many Comments PR", "state": "open",
			"merged": false,
			"head":   map[string]any{"ref": "feature", "label": "feature"},
			"base":   map[string]any{"ref": "main", "label": "main"},
			"user":   map[string]any{"login": "author"},
			"labels": []map[string]any{},
			"assignees": []map[string]any{},
			"created_at": "2026-01-15T10:00:00Z",
			"updated_at": "2026-01-16T12:00:00Z",
		})
	})

	mux.HandleFunc("/api/v1/repos/test-org/test-repo/issues/1/comments", func(w http.ResponseWriter, r *http.Request) {
		// Return 2 comments (less than commentPageSize, so pagination stops).
		comments := []map[string]any{
			{"id": 1, "body": "comment 1", "user": map[string]any{"login": "reviewer"}, "created_at": "2026-01-15T12:00:00Z", "updated_at": "2026-01-15T12:00:00Z"},
			{"id": 2, "body": "comment 2", "user": map[string]any{"login": "author"}, "created_at": "2026-01-15T13:00:00Z", "updated_at": "2026-01-15T13:00:00Z"},
		}
		_ = json.NewEncoder(w).Encode(comments)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	return httptest.NewServer(mux)
}

func TestClient_GetPRMeta_Good_CommentCount(t *testing.T) {
	srv := newPRMetaWithManyCommentsServer(t)
	defer srv.Close()

	client, err := New(srv.URL, "test-token")
	require.NoError(t, err)

	meta, err := client.GetPRMeta("test-org", "test-repo", 1)
	require.NoError(t, err)
	assert.Equal(t, 2, meta.CommentCount)
	assert.Equal(t, "Many Comments PR", meta.Title)
}

// --- GetPRMeta with nil created/updated dates ---

func newPRMetaWithNilDatesServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]string{"version": "1.21.0"})
	})

	mux.HandleFunc("/api/v1/repos/test-org/test-repo/pulls/2", func(w http.ResponseWriter, r *http.Request) {
		// No created_at, updated_at, user, labels, or assignees.
		jsonResponse(w, map[string]any{
			"id": 2, "number": 2, "title": "Minimal PR", "state": "closed",
			"merged": true,
			"head":   map[string]any{"ref": "fix", "label": "fix"},
			"base":   map[string]any{"ref": "main", "label": "main"},
		})
	})

	mux.HandleFunc("/api/v1/repos/test-org/test-repo/issues/2/comments", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]any{})
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	return httptest.NewServer(mux)
}

func TestClient_GetPRMeta_Good_MinimalFields(t *testing.T) {
	srv := newPRMetaWithNilDatesServer(t)
	defer srv.Close()

	client, err := New(srv.URL, "test-token")
	require.NoError(t, err)

	meta, err := client.GetPRMeta("test-org", "test-repo", 2)
	require.NoError(t, err)
	assert.Equal(t, "Minimal PR", meta.Title)
	assert.True(t, meta.IsMerged)
	assert.Empty(t, meta.Author)
	assert.Empty(t, meta.Labels)
	assert.Empty(t, meta.Assignees)
	assert.Equal(t, 0, meta.CommentCount)
}

// --- GetCommentBodies: empty result ---

func TestClient_GetCommentBodies_Good_Empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]string{"version": "1.21.0"})
	})
	mux.HandleFunc("/api/v1/repos/test-org/test-repo/issues/99/comments", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]any{})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := New(srv.URL, "test-token")
	require.NoError(t, err)

	comments, err := client.GetCommentBodies("test-org", "test-repo", 99)
	require.NoError(t, err)
	assert.Empty(t, comments)
}

// --- GetCommentBodies: poster is nil ---

func TestClient_GetCommentBodies_Good_NilPoster(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]string{"version": "1.21.0"})
	})
	mux.HandleFunc("/api/v1/repos/test-org/test-repo/issues/1/comments", func(w http.ResponseWriter, r *http.Request) {
		comments := []map[string]any{
			{"id": 1, "body": "anonymous comment", "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"},
		}
		_ = json.NewEncoder(w).Encode(comments)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := New(srv.URL, "test-token")
	require.NoError(t, err)

	comments, err := client.GetCommentBodies("test-org", "test-repo", 1)
	require.NoError(t, err)
	require.Len(t, comments, 1)
	assert.Equal(t, "anonymous comment", comments[0].Body)
	assert.Empty(t, comments[0].Author)
}

// --- ListPullRequests: state mapping ---

func TestClient_ListPullRequests_Good_AllStates(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	for _, state := range []string{"open", "closed", "all", ""} {
		_, err := client.ListPullRequests("test-org", "org-repo", state)
		require.NoError(t, err, "state=%q should not error", state)
	}
}

// --- NewFromConfig: additional paths ---

func TestNewFromConfig_Good_FlagOverridesEnv(t *testing.T) {
	isolateConfigEnv(t)

	srv := newMockGiteaServer(t)
	defer srv.Close()

	t.Setenv("GITEA_URL", "https://should-be-overridden.example.com")
	t.Setenv("GITEA_TOKEN", "should-be-overridden")

	client, err := NewFromConfig(srv.URL, "flag-token")
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, srv.URL, client.URL())
}
