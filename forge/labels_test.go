// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	"net/http"
	"net/http/httptest"
	"testing"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func forgejoCreateLabel(name, color string) forgejo.CreateLabelOption {
	return forgejo.CreateLabelOption{Name: name, Color: color}
}

func TestClient_ListRepoLabels_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	labels, err := client.ListRepoLabels("test-org", "org-repo")
	require.NoError(t, err)
	require.Len(t, labels, 2)
	assert.Equal(t, "bug", labels[0].Name)
	assert.Equal(t, "feature", labels[1].Name)
}

func TestClient_ListRepoLabels_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.ListRepoLabels("test-org", "org-repo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list repo labels")
}

func TestClient_ListRepoLabelsIter_Good_Paginates_Good(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]string{"version": "1.21.0"})
	})
	mux.HandleFunc("/api/v1/repos/test-org/org-repo/labels", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("page") {
		case "2":
			jsonResponse(w, []map[string]any{
				{"id": 3, "name": "documentation", "color": "#00aa00"},
			})
		default:
			w.Header().Set("Link", "<http://"+r.Host+"/api/v1/repos/test-org/org-repo/labels?page=2>; rel=\"next\", <http://"+r.Host+"/api/v1/repos/test-org/org-repo/labels?page=2>; rel=\"last\"")
			jsonResponse(w, []map[string]any{
				{"id": 1, "name": "bug", "color": "#ff0000"},
				{"id": 2, "name": "feature", "color": "#0000ff"},
			})
		}
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := New(srv.URL, "test-token")
	require.NoError(t, err)

	var names []string
	for label, err := range client.ListRepoLabelsIter("test-org", "org-repo") {
		require.NoError(t, err)
		names = append(names, label.Name)
	}

	require.Len(t, names, 3)
	assert.Equal(t, []string{"bug", "feature", "documentation"}, names)
}

func TestClient_ListRepoLabelsIter_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	for _, err := range client.ListRepoLabelsIter("test-org", "org-repo") {
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list repo labels")
		break
	}
}

func TestClient_CreateRepoLabel_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	label, err := client.CreateRepoLabel("test-org", "org-repo", forgejoCreateLabel("new-label", "#00ff00"))
	require.NoError(t, err)
	assert.NotNil(t, label)
}

func TestClient_CreateRepoLabel_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.CreateRepoLabel("test-org", "org-repo", forgejoCreateLabel("label", "#000"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create repo label")
}

func TestClient_GetLabelByName_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	label, err := client.GetLabelByName("test-org", "org-repo", "bug")
	require.NoError(t, err)
	assert.Equal(t, "bug", label.Name)
}

func TestClient_GetLabelByName_Good_CaseInsensitive_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	label, err := client.GetLabelByName("test-org", "org-repo", "BUG")
	require.NoError(t, err)
	assert.Equal(t, "bug", label.Name)
}

func TestClient_GetLabelByName_Bad_NotFound_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	_, err := client.GetLabelByName("test-org", "org-repo", "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "label nonexistent not found")
}

func TestClient_EnsureLabel_Good_Exists_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	// "bug" already exists in mock server.
	label, err := client.EnsureLabel("test-org", "org-repo", "bug", "#ff0000")
	require.NoError(t, err)
	assert.Equal(t, "bug", label.Name)
}

func TestClient_EnsureLabel_Good_Creates_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	// "urgent" does not exist, so it should be created.
	label, err := client.EnsureLabel("test-org", "org-repo", "urgent", "#ff9900")
	require.NoError(t, err)
	assert.NotNil(t, label)
}

func TestClient_ListOrgLabels_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	labels, err := client.ListOrgLabels("test-org")
	require.NoError(t, err)
	require.Len(t, labels, 3)
	assert.Equal(t, "bug", labels[0].Name)
	assert.Equal(t, "feature", labels[1].Name)
	assert.Equal(t, "documentation", labels[2].Name)
}

func TestClient_ListOrgLabelsIter_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	var names []string
	for label, err := range client.ListOrgLabelsIter("test-org") {
		require.NoError(t, err)
		names = append(names, label.Name)
	}

	require.Len(t, names, 3)
	assert.Equal(t, []string{"bug", "feature", "documentation"}, names)
}

func TestClient_ListOrgLabelsIter_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	for _, err := range client.ListOrgLabelsIter("test-org") {
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list org repos")
		break
	}
}

func TestClient_ListOrgLabels_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.ListOrgLabels("test-org")
	assert.Error(t, err)
}

func TestClient_AddIssueLabels_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	err := client.AddIssueLabels("test-org", "org-repo", 1, []int64{1, 2})
	require.NoError(t, err)
}

func TestClient_AddIssueLabels_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	err := client.AddIssueLabels("test-org", "org-repo", 1, []int64{1})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add labels")
}

func TestClient_RemoveIssueLabel_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	err := client.RemoveIssueLabel("test-org", "org-repo", 1, 1)
	require.NoError(t, err)
}

func TestClient_RemoveIssueLabel_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	err := client.RemoveIssueLabel("test-org", "org-repo", 1, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove label")
}
