package forge

import (
	"strings"
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
