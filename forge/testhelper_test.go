package forge

import (
	json "dappco.re/go/core/scm/internal/ax/jsonx"
	strings "dappco.re/go/core/scm/internal/ax/stringsx"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newMockForgejoServer creates an httptest.Server that mimics the Forgejo API
// endpoints used during client initialisation and common operations.
func newMockForgejoServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := newForgejoMux()
	return httptest.NewServer(mux)
}

// newForgejoMux creates an http.ServeMux with standard Forgejo API responses.
func newForgejoMux() *http.ServeMux {
	mux := http.NewServeMux()

	// The Forgejo SDK calls /api/v1/version during NewClient().
	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]string{"version": "1.21.0"})
	})

	// User info endpoint for GetCurrentUser / GetMyUserInfo.
	mux.HandleFunc("/api/v1/user", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]any{
			"id":         1,
			"login":      "test-user",
			"full_name":  "Test User",
			"email":      "test@example.com",
			"login_name": "test-user",
		})
	})

	// Repos listing (user).
	mux.HandleFunc("/api/v1/user/repos", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, []map[string]any{
			{"id": 1, "name": "repo-a", "full_name": "test-user/repo-a", "owner": map[string]any{"login": "test-user"}},
			{"id": 2, "name": "repo-b", "full_name": "test-user/repo-b", "owner": map[string]any{"login": "test-user"}},
		})
	})

	// Org repos listing + create.
	mux.HandleFunc("/api/v1/orgs/test-org/repos", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			jsonResponse(w, map[string]any{
				"id": 20, "name": "new-repo", "full_name": "test-org/new-repo",
				"owner": map[string]any{"login": "test-org"},
			})
			return
		}
		jsonResponse(w, []map[string]any{
			{"id": 10, "name": "org-repo", "full_name": "test-org/org-repo", "owner": map[string]any{"login": "test-org", "id": 100}},
		})
	})

	// Create org repo (SDK uses /org/ not /orgs/).
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
		if r.Method == http.MethodPatch {
			jsonResponse(w, map[string]any{
				"id": 1, "number": 1, "title": "Issue 1", "state": "open",
				"body": "First issue",
			})
			return
		}
		jsonResponse(w, map[string]any{
			"id": 1, "number": 1, "title": "Issue 1", "state": "open",
			"body": "First issue body",
		})
	})

	// Issue comments.
	mux.HandleFunc("/api/v1/repos/test-org/org-repo/issues/1/comments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			jsonResponse(w, map[string]any{
				"id": 100, "body": "test comment",
				"user": map[string]any{"login": "test-user"},
			})
			return
		}
		jsonResponse(w, []map[string]any{
			{"id": 100, "body": "comment 1", "user": map[string]any{"login": "user1"}, "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"},
			{"id": 101, "body": "comment 2", "user": map[string]any{"login": "user2"}, "created_at": "2026-01-02T00:00:00Z", "updated_at": "2026-01-02T00:00:00Z"},
		})
	})

	// Issue labels.
	mux.HandleFunc("/api/v1/repos/test-org/org-repo/issues/1/labels", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			jsonResponse(w, []map[string]any{
				{"id": 1, "name": "bug", "color": "#ff0000"},
			})
			return
		}
		jsonResponse(w, []map[string]any{
			{"id": 1, "name": "bug", "color": "#ff0000"},
		})
	})

	// Remove issue label.
	mux.HandleFunc("/api/v1/repos/test-org/org-repo/issues/1/labels/1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
	})

	// Pull requests.
	mux.HandleFunc("/api/v1/repos/test-org/org-repo/pulls", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			jsonResponse(w, map[string]any{
				"id": 1, "number": 1, "title": "Test PR", "state": "open",
				"head": map[string]any{"ref": "feature", "label": "feature"},
				"base": map[string]any{"ref": "main", "label": "main"},
			})
			return
		}
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

	// PR merge.
	mux.HandleFunc("/api/v1/repos/test-org/org-repo/pulls/1/merge", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// PR reviews.
	mux.HandleFunc("/api/v1/repos/test-org/org-repo/pulls/1/reviews", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, []map[string]any{
			{"id": 1, "state": "APPROVED", "user": map[string]any{"login": "reviewer1"}},
		})
	})

	// Combined status.
	mux.HandleFunc("/api/v1/repos/test-org/org-repo/commits/main/status", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]any{
			"state": "success",
			"statuses": []map[string]any{
				{"context": "ci/build", "state": "success"},
			},
		})
	})

	// Repo labels.
	mux.HandleFunc("/api/v1/repos/test-org/org-repo/labels", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			jsonResponse(w, map[string]any{
				"id": 5, "name": "new-label", "color": "#00ff00",
			})
			return
		}
		jsonResponse(w, []map[string]any{
			{"id": 1, "name": "bug", "color": "#ff0000"},
			{"id": 2, "name": "feature", "color": "#0000ff"},
		})
	})

	// Webhooks.
	mux.HandleFunc("/api/v1/repos/test-org/org-repo/hooks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			jsonResponse(w, map[string]any{
				"id":     1,
				"type":   "forgejo",
				"active": true,
				"config": map[string]any{"url": "https://example.com/hook"},
			})
			return
		}
		jsonResponse(w, []map[string]any{
			{"id": 1, "type": "forgejo", "active": true, "config": map[string]any{"url": "https://example.com/hook"}},
		})
	})

	// Orgs listing.
	mux.HandleFunc("/api/v1/user/orgs", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, []map[string]any{
			{"id": 100, "login": "test-org", "username": "test-org", "full_name": "Test Organisation"},
		})
	})

	// Single org.
	mux.HandleFunc("/api/v1/orgs/test-org", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]any{
			"id": 100, "login": "test-org", "username": "test-org", "full_name": "Test Organisation",
		})
	})

	// Create org.
	mux.HandleFunc("/api/v1/orgs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			jsonResponse(w, map[string]any{
				"id": 200, "login": "new-org", "username": "new-org", "full_name": "New Organisation",
			})
			return
		}
	})

	// Fork repo.
	mux.HandleFunc("/api/v1/repos/test-org/org-repo/forks", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		jsonResponse(w, map[string]any{
			"id": 30, "name": "org-repo", "full_name": "test-user/org-repo",
			"owner": map[string]any{"login": "test-user"},
		})
	})

	// Migrate repo.
	mux.HandleFunc("/api/v1/repos/migrate", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		jsonResponse(w, map[string]any{
			"id": 40, "name": "migrated-repo", "full_name": "test-user/migrated-repo",
			"owner": map[string]any{"login": "test-user"},
		})
	})

	// Dismiss review.
	mux.HandleFunc("/api/v1/repos/test-org/org-repo/pulls/1/reviews/1/dismissals", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]any{
			"id": 1, "state": "dismissed",
		})
	})

	// Generic fallback — handles PATCH for SetPRDraft and other unmatched routes.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Handle PATCH requests (SetPRDraft).
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

// jsonResponse writes a JSON response with 200 status (unless already set).
func jsonResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

// newTestClient creates a Client backed by the mock server.
func newTestClient(t *testing.T) (*Client, *httptest.Server) {
	t.Helper()
	srv := newMockForgejoServer(t)

	client, err := New(srv.URL, "test-token")
	if err != nil {
		srv.Close()
		t.Fatalf("failed to create test client: %v", err)
	}

	return client, srv
}

// newErrorServer creates a mock server that returns errors for all API calls
// (except /api/v1/version which is needed for client creation).
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
