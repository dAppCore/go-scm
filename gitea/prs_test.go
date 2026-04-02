// SPDX-License-Identifier: EUPL-1.2

package gitea

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	giteaSDK "code.gitea.io/sdk/gitea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_MergePullRequest_Good(t *testing.T) {
	var method, path string

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]string{"version": "1.21.0"})
	})
	mux.HandleFunc("/api/v1/repos/test-org/org-repo/pulls/1/merge", func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := New(srv.URL, "test-token")
	require.NoError(t, err)

	err = client.MergePullRequest("test-org", "org-repo", 1, "merge")
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, method)
	assert.Equal(t, "/api/v1/repos/test-org/org-repo/pulls/1/merge", path)
}

func TestClient_MergePullRequest_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	err := client.MergePullRequest("test-org", "org-repo", 1, "merge")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to merge pull request")
}

func TestClient_ListPRReviews_Good(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]string{"version": "1.21.0"})
	})
	mux.HandleFunc("/api/v1/repos/test-org/org-repo/pulls/1/reviews", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, []map[string]any{
			{"id": 1, "state": "APPROVED", "user": map[string]any{"login": "reviewer1"}},
		})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := New(srv.URL, "test-token")
	require.NoError(t, err)

	reviews, err := client.ListPRReviews("test-org", "org-repo", 1)
	require.NoError(t, err)
	require.Len(t, reviews, 1)
	assert.Equal(t, giteaSDK.ReviewStateApproved, reviews[0].State)
}

func TestClient_ListPRReviewsIter_Good_Paginates_Good(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]string{"version": "1.21.0"})
	})
	mux.HandleFunc("/api/v1/repos/test-org/org-repo/pulls/1/reviews", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("page") {
		case "2":
			jsonResponse(w, []map[string]any{
				{"id": 2, "state": "REQUEST_CHANGES", "user": map[string]any{"login": "reviewer2"}},
			})
		default:
			w.Header().Set("Link", "<http://"+r.Host+"/api/v1/repos/test-org/org-repo/pulls/1/reviews?page=2>; rel=\"next\", <http://"+r.Host+"/api/v1/repos/test-org/org-repo/pulls/1/reviews?page=2>; rel=\"last\"")
			jsonResponse(w, []map[string]any{
				{"id": 1, "state": "APPROVED", "user": map[string]any{"login": "reviewer1"}},
			})
		}
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := New(srv.URL, "test-token")
	require.NoError(t, err)

	var states []string
	for review, err := range client.ListPRReviewsIter("test-org", "org-repo", 1) {
		require.NoError(t, err)
		states = append(states, string(review.State))
	}

	assert.Equal(t, []string{"APPROVED", "REQUEST_CHANGES"}, states)
}

func TestClient_ListPRReviews_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.ListPRReviews("test-org", "org-repo", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list reviews")
}

func TestClient_GetCombinedStatus_Good(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]string{"version": "1.21.0"})
	})
	mux.HandleFunc("/api/v1/repos/test-org/org-repo/commits/main/status", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]any{
			"state":       "success",
			"sha":         "abc123",
			"total_count": 1,
			"statuses": []map[string]any{
				{"state": "success", "context": "ci/test"},
			},
		})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := New(srv.URL, "test-token")
	require.NoError(t, err)

	status, err := client.GetCombinedStatus("test-org", "org-repo", "main")
	require.NoError(t, err)
	assert.Equal(t, giteaSDK.StatusSuccess, status.State)
	assert.Equal(t, "abc123", status.SHA)
}

func TestClient_GetCombinedStatus_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.GetCombinedStatus("test-org", "org-repo", "main")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get combined status")
}

func TestClient_DismissReview_Good(t *testing.T) {
	var method, path string
	var payload map[string]any

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]string{"version": "1.21.0"})
	})
	mux.HandleFunc("/api/v1/repos/test-org/org-repo/pulls/1/reviews/1/dismissals", func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := New(srv.URL, "test-token")
	require.NoError(t, err)

	err = client.DismissReview("test-org", "org-repo", 1, 1, "outdated review")
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, method)
	assert.Equal(t, "/api/v1/repos/test-org/org-repo/pulls/1/reviews/1/dismissals", path)
	assert.Equal(t, "outdated review", payload["message"])
}

func TestClient_DismissReview_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	err := client.DismissReview("test-org", "org-repo", 1, 1, "outdated")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to dismiss review")
}

func TestClient_SetPRDraft_Good_Request_Good(t *testing.T) {
	var method, path string
	var payload map[string]any

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]string{"version": "1.21.0"})
	})
	mux.HandleFunc("/api/v1/repos/test-org/org-repo/pulls/3", func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		jsonResponse(w, map[string]any{"number": 3})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := New(srv.URL, "test-token")
	require.NoError(t, err)

	err = client.SetPRDraft("test-org", "org-repo", 3, false)
	require.NoError(t, err)
	assert.Equal(t, http.MethodPatch, method)
	assert.Equal(t, "/api/v1/repos/test-org/org-repo/pulls/3", path)
	assert.Equal(t, false, payload["draft"])
}

func TestClient_SetPRDraft_Bad_PathTraversalOwner_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	err := client.SetPRDraft("../owner", "org-repo", 3, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid owner")
}

func TestClient_SetPRDraft_Bad_PathTraversalRepo_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	err := client.SetPRDraft("test-org", "..", 3, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid repo")
}
