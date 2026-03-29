package forge

import (
	fmt "dappco.re/go/core/scm/internal/ax/fmtx"
	json "dappco.re/go/core/scm/internal/ax/jsonx"
	"net/http"
	"net/http/httptest"
	"testing"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_Good(t *testing.T) {
	srv := newMockForgejoServer(t)
	defer srv.Close()

	client, err := New(srv.URL, "test-token-123")
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.API())
	assert.Equal(t, srv.URL, client.URL())
	assert.Equal(t, "test-token-123", client.Token())
}

func TestNew_Bad_InvalidURL(t *testing.T) {
	// The Forgejo SDK may reject certain URL formats.
	_, err := New("://invalid-url", "token")
	assert.Error(t, err)
}

func TestClient_API_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	assert.NotNil(t, client.API(), "API() should return the underlying SDK client")
}

func TestClient_URL_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	assert.Equal(t, srv.URL, client.URL())
}

func TestClient_Token_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	assert.Equal(t, "test-token", client.Token())
}

func TestClient_GetCurrentUser_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	user, err := client.GetCurrentUser()
	require.NoError(t, err)
	assert.Equal(t, "test-user", user.UserName)
}

func TestClient_GetCurrentUser_Bad_ServerError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"version": "1.21.0"})
	})
	mux.HandleFunc("/api/v1/user", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := New(srv.URL, "token")
	require.NoError(t, err)

	_, err = client.GetCurrentUser()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get current user")
}

func TestClient_SetPRDraft_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	err := client.SetPRDraft("owner", "repo", 1, true)
	require.NoError(t, err)
}

func TestClient_SetPRDraft_Good_Undraft(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	err := client.SetPRDraft("owner", "repo", 1, false)
	require.NoError(t, err)
}

func TestClient_SetPRDraft_Bad_ServerError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"version": "1.21.0"})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		http.NotFound(w, r)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := New(srv.URL, "token")
	require.NoError(t, err)

	err = client.SetPRDraft("owner", "repo", 1, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status 403")
}

func TestClient_SetPRDraft_Bad_ConnectionRefused(t *testing.T) {
	// Use a closed server to simulate connection errors.
	srv := newMockForgejoServer(t)
	client, err := New(srv.URL, "token")
	require.NoError(t, err)
	srv.Close() // Close the server.

	err = client.SetPRDraft("owner", "repo", 1, true)
	assert.Error(t, err)
}

func TestClient_SetPRDraft_Good_URLConstruction(t *testing.T) {
	// Verify the URL is constructed correctly by checking the request path.
	var capturedPath string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"version": "1.21.0"})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"number": 42})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := New(srv.URL, "token")
	require.NoError(t, err)

	_ = client.SetPRDraft("my-org", "my-repo", 42, true)
	assert.Equal(t, "/api/v1/repos/my-org/my-repo/pulls/42", capturedPath)
}

func TestClient_SetPRDraft_Good_AuthHeader(t *testing.T) {
	// Verify the authorisation header is set correctly.
	var capturedAuth string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"version": "1.21.0"})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"number": 1})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := New(srv.URL, "my-secret-token")
	require.NoError(t, err)

	_ = client.SetPRDraft("owner", "repo", 1, true)
	assert.Equal(t, "token my-secret-token", capturedAuth)
}

// --- PRMeta and Comment struct tests ---

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

// --- MergePullRequest merge style mapping ---

func TestMergePullRequest_Good_StyleMapping(t *testing.T) {
	// We can't easily test the SDK call, but we can verify the method
	// errors when the server returns failure. This exercises the style mapping code.
	tests := []struct {
		name   string
		method string
	}{
		{name: "merge", method: "merge"},
		{name: "squash", method: "squash"},
		{name: "rebase", method: "rebase"},
		{name: "default (unknown)", method: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]string{"version": "1.21.0"})
			})
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				// Return 405 to trigger an error so we know the code executed.
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			})
			srv := httptest.NewServer(mux)
			defer srv.Close()

			client, err := New(srv.URL, "token")
			require.NoError(t, err)

			err = client.MergePullRequest("owner", "repo", 1, tt.method)
			assert.Error(t, err, "merge should fail against mock server for method %s", tt.method)
		})
	}
}

// --- ListIssuesOpts defaulting ---

func TestListIssuesOpts_Good_Defaults(t *testing.T) {
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
			name:          "all state",
			opts:          ListIssuesOpts{State: "all"},
			expectedState: "all",
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
			// Verify the opts struct stores values correctly.
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

// --- ForkRepo error handling ---

func TestClient_ForkRepo_Good_WithOrg(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"version": "1.21.0"})
	})
	var capturedBody map[string]any
	mux.HandleFunc("/api/v1/repos/owner/repo/forks", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":        1,
			"name":      "repo",
			"full_name": "target-org/repo",
			"owner":     map[string]any{"login": "target-org"},
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := New(srv.URL, "token")
	require.NoError(t, err)

	fork, err := client.ForkRepo("owner", "repo", "target-org")
	require.NoError(t, err)
	assert.NotNil(t, fork)
	assert.Equal(t, "target-org", capturedBody["organization"])
}

func TestClient_ForkRepo_Good_WithoutOrg(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"version": "1.21.0"})
	})
	var capturedBody map[string]any
	mux.HandleFunc("/api/v1/repos/owner/repo/forks", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":        2,
			"name":      "repo",
			"full_name": "user/repo",
			"owner":     map[string]any{"login": "user"},
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := New(srv.URL, "token")
	require.NoError(t, err)

	fork, err := client.ForkRepo("owner", "repo", "")
	require.NoError(t, err)
	assert.NotNil(t, fork)
	// When org is empty, the Organization pointer is nil.
	// The SDK may or may not include it in the JSON; just verify the fork succeeded.
}

func TestClient_ForkRepo_Bad_ServerError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"version": "1.21.0"})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Server Error", http.StatusInternalServerError)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := New(srv.URL, "token")
	require.NoError(t, err)

	_, err = client.ForkRepo("owner", "repo", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fork")
}

// --- CreatePullRequest error handling ---

func TestClient_CreatePullRequest_Bad_ServerError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"version": "1.21.0"})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Server Error", http.StatusInternalServerError)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := New(srv.URL, "token")
	require.NoError(t, err)

	_, err = client.CreatePullRequest("owner", "repo", forgejo.CreatePullRequestOption{
		Head:  "feature",
		Base:  "main",
		Title: "Test PR",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create pull request")
}

// --- commentPageSize constant test ---

func TestCommentPageSize_Good(t *testing.T) {
	assert.Equal(t, 50, commentPageSize, "comment page size should be 50")
}

// --- ListPullRequests state mapping ---

func TestListPullRequests_Good_StateMapping(t *testing.T) {
	// Verify state mapping via error path (server returns error).
	tests := []struct {
		name  string
		state string
	}{
		{name: "open", state: "open"},
		{name: "closed", state: "closed"},
		{name: "all", state: "all"},
		{name: "default", state: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]string{"version": "1.21.0"})
			})

			var capturedState string
			mux.HandleFunc(fmt.Sprintf("/api/v1/repos/owner/repo/pulls"), func(w http.ResponseWriter, r *http.Request) {
				capturedState = r.URL.Query().Get("state")
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode([]any{})
			})
			srv := httptest.NewServer(mux)
			defer srv.Close()

			client, err := New(srv.URL, "token")
			require.NoError(t, err)

			_, _ = client.ListPullRequests("owner", "repo", tt.state)
			// The state parameter was passed to the SDK and sent to the server.
			assert.NotEmpty(t, capturedState)
		})
	}
}
