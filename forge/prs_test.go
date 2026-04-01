// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	json "dappco.re/go/core/scm/internal/ax/jsonx"
	strings "dappco.re/go/core/scm/internal/ax/stringsx"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_MergePullRequest_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	err := client.MergePullRequest("test-org", "org-repo", 1, "merge")
	require.NoError(t, err)
}

func TestClient_MergePullRequest_Good_Squash_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	err := client.MergePullRequest("test-org", "org-repo", 1, "squash")
	require.NoError(t, err)
}

func TestClient_MergePullRequest_Good_Rebase_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	err := client.MergePullRequest("test-org", "org-repo", 1, "rebase")
	require.NoError(t, err)
}

func TestClient_MergePullRequest_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	err := client.MergePullRequest("test-org", "org-repo", 1, "merge")
	assert.Error(t, err)
	// The error may be "failed to merge" or "merge returned false" depending on
	// how the error server responds.
	assert.True(t,
		strings.Contains(err.Error(), "failed to merge") ||
			strings.Contains(err.Error(), "merge returned false"),
		"unexpected error: %s", err.Error())
}

func TestClient_ListPRReviews_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	reviews, err := client.ListPRReviews("test-org", "org-repo", 1)
	require.NoError(t, err)
	require.Len(t, reviews, 1)
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
		case "3":
			jsonResponse(w, []map[string]any{})
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
		states = append(states, review.State)
	}

	require.Equal(t, []string{"APPROVED", "REQUEST_CHANGES"}, states)
}

func TestClient_ListPRReviews_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.ListPRReviews("test-org", "org-repo", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list reviews")
}

func TestClient_ListPRReviewsIter_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	var got bool
	for _, err := range client.ListPRReviewsIter("test-org", "org-repo", 1) {
		assert.Error(t, err)
		got = true
	}

	assert.True(t, got)
}

func TestClient_GetCombinedStatus_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	status, err := client.GetCombinedStatus("test-org", "org-repo", "main")
	require.NoError(t, err)
	assert.NotNil(t, status)
}

func TestClient_GetCombinedStatus_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.GetCombinedStatus("test-org", "org-repo", "main")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get combined status")
}

func TestClient_DismissReview_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	err := client.DismissReview("test-org", "org-repo", 1, 1, "outdated review")
	require.NoError(t, err)
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
	assert.NoError(t, err)
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
