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

func TestClient_MergePullRequest_Good_Squash(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	err := client.MergePullRequest("test-org", "org-repo", 1, "squash")
	require.NoError(t, err)
}

func TestClient_MergePullRequest_Good_Rebase(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	err := client.MergePullRequest("test-org", "org-repo", 1, "rebase")
	require.NoError(t, err)
}

func TestClient_MergePullRequest_Bad_ServerError(t *testing.T) {
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

func TestClient_ListPRReviews_Bad_ServerError(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.ListPRReviews("test-org", "org-repo", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list reviews")
}

func TestClient_GetCombinedStatus_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	status, err := client.GetCombinedStatus("test-org", "org-repo", "main")
	require.NoError(t, err)
	assert.NotNil(t, status)
}

func TestClient_GetCombinedStatus_Bad_ServerError(t *testing.T) {
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

func TestClient_DismissReview_Bad_ServerError(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	err := client.DismissReview("test-org", "org-repo", 1, 1, "outdated")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to dismiss review")
}

func TestClient_SetPRDraft_Good_Request(t *testing.T) {
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

func TestClient_SetPRDraft_Bad_PathTraversalOwner(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	err := client.SetPRDraft("../owner", "org-repo", 3, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid owner")
}

func TestClient_SetPRDraft_Bad_PathTraversalRepo(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	err := client.SetPRDraft("test-org", "..", 3, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid repo")
}
