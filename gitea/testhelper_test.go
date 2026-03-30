// SPDX-Licence-Identifier: EUPL-1.2

package gitea

import (
	json "dappco.re/go/core/scm/internal/ax/jsonx"
	strings "dappco.re/go/core/scm/internal/ax/stringsx"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newMockGiteaServer creates an httptest.Server that mimics the Gitea API
// endpoints used during client initialisation and common operations.
func newMockGiteaServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := newGiteaMux()
	return httptest.NewServer(mux)
}

// newGiteaMux creates an http.ServeMux with standard Gitea API responses.
func newGiteaMux() *http.ServeMux {
	mux := http.NewServeMux()

	// The Gitea SDK calls /api/v1/version during NewClient().
	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]string{"version": "1.21.0"})
	})

	// User repos listing.
	mux.HandleFunc("/api/v1/user/repos", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, []map[string]any{
			{"id": 1, "name": "repo-a", "full_name": "test-user/repo-a", "owner": map[string]any{"login": "test-user"}},
			{"id": 2, "name": "repo-b", "full_name": "test-user/repo-b", "owner": map[string]any{"login": "test-user"}},
		})
	})

	// Org repos listing.
	mux.HandleFunc("/api/v1/orgs/test-org/repos", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, []map[string]any{
			{"id": 10, "name": "org-repo", "full_name": "test-org/org-repo", "owner": map[string]any{"login": "test-org", "id": 100}},
		})
	})

	// Create org repo (Gitea SDK uses /org/ not /orgs/).
	mux.HandleFunc("/api/v1/org/test-org/repos", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			jsonResponse(w, map[string]any{
				"id": 20, "name": "new-repo", "full_name": "test-org/new-repo",
				"owner": map[string]any{"login": "test-org"},
			})
			return
		}
		jsonResponse(w, []map[string]any{})
	})

	// Get/delete single repo.
	mux.HandleFunc("/api/v1/repos/test-org/org-repo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		jsonResponse(w, map[string]any{
			"id": 10, "name": "org-repo", "full_name": "test-org/org-repo",
			"owner": map[string]any{"login": "test-org"},
		})
	})

	// Issues.
	mux.HandleFunc("/api/v1/repos/test-org/org-repo/issues", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			jsonResponse(w, map[string]any{
				"id": 1, "number": 1, "title": "Test Issue", "state": "open",
				"body": "Issue body text",
			})
			return
		}
		jsonResponse(w, []map[string]any{
			{"id": 1, "number": 1, "title": "Issue 1", "state": "open", "body": "First issue"},
			{"id": 2, "number": 2, "title": "Issue 2", "state": "closed", "body": "Second issue"},
		})
	})

	// Single issue.
	mux.HandleFunc("/api/v1/repos/test-org/org-repo/issues/1", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]any{
			"id": 1, "number": 1, "title": "Issue 1", "state": "open",
			"body": "First issue body",
		})
	})

	// Issue comments.
	mux.HandleFunc("/api/v1/repos/test-org/org-repo/issues/1/comments", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, []map[string]any{
			{"id": 100, "body": "comment 1", "user": map[string]any{"login": "user1"}, "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"},
			{"id": 101, "body": "comment 2", "user": map[string]any{"login": "user2"}, "created_at": "2026-01-02T00:00:00Z", "updated_at": "2026-01-02T00:00:00Z"},
		})
	})

	// Pull requests.
	mux.HandleFunc("/api/v1/repos/test-org/org-repo/pulls", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, []map[string]any{
			{
				"id": 1, "number": 1, "title": "PR 1", "state": "open",
				"head": map[string]any{"ref": "feature", "label": "feature"},
				"base": map[string]any{"ref": "main", "label": "main"},
			},
		})
	})

	// Single pull request.
	mux.HandleFunc("/api/v1/repos/test-org/org-repo/pulls/1", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]any{
			"id": 1, "number": 1, "title": "PR 1", "state": "open",
			"merged": false,
			"head":   map[string]any{"ref": "feature", "label": "feature"},
			"base":   map[string]any{"ref": "main", "label": "main"},
			"user":   map[string]any{"login": "author"},
			"labels": []map[string]any{{"name": "enhancement"}},
			"assignees": []map[string]any{
				{"login": "dev1"},
			},
			"created_at": "2026-01-15T10:00:00Z",
			"updated_at": "2026-01-16T12:00:00Z",
		})
	})

	// Migrate repo.
	mux.HandleFunc("/api/v1/repos/migrate", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		jsonResponse(w, map[string]any{
			"id": 40, "name": "mirrored-repo", "full_name": "test-org/mirrored-repo",
			"owner":  map[string]any{"login": "test-org"},
			"mirror": true,
		})
	})

	// Fallback for PATCH requests and unmatched routes.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch && strings.Contains(r.URL.Path, "/pulls/") {
			jsonResponse(w, map[string]any{
				"number": 1, "title": "test PR", "state": "open",
			})
			return
		}
		http.NotFound(w, r)
	})

	return mux
}

// jsonResponse writes a JSON response.
func jsonResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

// newTestClient creates a Client backed by the mock server.
func newTestClient(t *testing.T) (*Client, *httptest.Server) {
	t.Helper()
	srv := newMockGiteaServer(t)

	client, err := New(srv.URL, "test-token")
	if err != nil {
		srv.Close()
		t.Fatalf("failed to create test client: %v", err)
	}

	return client, srv
}

// newErrorServer creates a mock server that returns errors for all API calls.
func newErrorServer(t *testing.T) (*Client, *httptest.Server) {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]string{"version": "1.21.0"})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	})

	srv := httptest.NewServer(mux)
	client, err := New(srv.URL, "token")
	if err != nil {
		srv.Close()
		t.Fatalf("failed to create error server client: %v", err)
	}

	return client, srv
}
