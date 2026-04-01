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

func TestClient_ListOrgRepos_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	repos, err := client.ListOrgRepos("test-org")
	require.NoError(t, err)
	require.Len(t, repos, 1)
	assert.Equal(t, "org-repo", repos[0].Name)
}

func TestClient_ListOrgRepos_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.ListOrgRepos("test-org")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list org repos")
}

func TestClient_ListUserRepos_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	repos, err := client.ListUserRepos()
	require.NoError(t, err)
	require.Len(t, repos, 2)
	assert.Equal(t, "repo-a", repos[0].Name)
	assert.Equal(t, "repo-b", repos[1].Name)
}

func TestClient_ListUserRepos_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.ListUserRepos()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list user repos")
}

func TestClient_GetRepo_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	repo, err := client.GetRepo("test-org", "org-repo")
	require.NoError(t, err)
	assert.Equal(t, "org-repo", repo.Name)
}

func TestClient_GetRepo_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.GetRepo("test-org", "org-repo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get repo")
}

func TestClient_CreateMirror_Good_WithAuth_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	// The Gitea SDK requires an auth token when Service is GitServiceGithub.
	repo, err := client.CreateMirror("test-org", "private-mirror", "https://github.com/example/private.git", "ghp_token123")
	require.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestClient_CreateMirror_Bad_NoAuthToken_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	// GitHub mirrors require an auth token.
	_, err := client.CreateMirror("test-org", "mirrored", "https://github.com/example/repo.git", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create mirror")
}

func TestClient_CreateMirror_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.CreateMirror("test-org", "mirrored", "https://github.com/example/repo.git", "ghp_token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create mirror")
}

func TestClient_CreateMirrorFromService_Good_Gitea_Good(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]string{"version": "1.21.0"})
	})
	mux.HandleFunc("/api/v1/repos/migrate", func(w http.ResponseWriter, r *http.Request) {
		var opts map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&opts))
		assert.Equal(t, "gitea", opts["service"])
		assert.Equal(t, true, opts["mirror"])
		assert.Equal(t, "https://forge.example.org/core/go-scm.git", opts["clone_addr"])
		assert.Equal(t, "secret-token", opts["auth_token"])
		w.WriteHeader(http.StatusCreated)
		jsonResponse(w, map[string]any{
			"id": 40, "name": "public-mirror", "full_name": "test-org/public-mirror",
			"owner":  map[string]any{"login": "test-org"},
			"mirror": true,
		})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := New(srv.URL, "test-token")
	require.NoError(t, err)

	repo, err := client.CreateMirrorFromService("test-org", "public-mirror", "https://forge.example.org/core/go-scm.git", giteaSDK.GitServiceGitea, "secret-token")
	require.NoError(t, err)
	require.NotNil(t, repo)
	assert.Equal(t, "public-mirror", repo.Name)
}

func TestClient_DeleteRepo_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	err := client.DeleteRepo("test-org", "org-repo")
	require.NoError(t, err)
}

func TestClient_DeleteRepo_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	err := client.DeleteRepo("test-org", "org-repo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete repo")
}

func TestClient_CreateOrgRepo_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	repo, err := client.CreateOrgRepo("test-org", giteaSDK.CreateRepoOption{
		Name:        "new-repo",
		Description: "A new repository",
	})
	require.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestClient_CreateOrgRepo_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.CreateOrgRepo("test-org", giteaSDK.CreateRepoOption{
		Name: "new-repo",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create org repo")
}
