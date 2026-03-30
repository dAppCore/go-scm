// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	"testing"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"

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

func TestClient_ListOrgRepos_Bad_ServerError(t *testing.T) {
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

func TestClient_ListUserRepos_Bad_ServerError(t *testing.T) {
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

func TestClient_GetRepo_Bad_ServerError(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.GetRepo("test-org", "org-repo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get repo")
}

func TestClient_CreateOrgRepo_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	repo, err := client.CreateOrgRepo("test-org", forgejo.CreateRepoOption{
		Name:        "new-repo",
		Description: "A new repository",
	})
	require.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestClient_CreateOrgRepo_Bad_ServerError(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.CreateOrgRepo("test-org", forgejo.CreateRepoOption{
		Name: "new-repo",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create org repo")
}

func TestClient_DeleteRepo_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	err := client.DeleteRepo("test-org", "org-repo")
	require.NoError(t, err)
}

func TestClient_DeleteRepo_Bad_ServerError(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	err := client.DeleteRepo("test-org", "org-repo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete repo")
}

func TestClient_MigrateRepo_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	repo, err := client.MigrateRepo(forgejo.MigrateRepoOption{
		RepoName:  "migrated-repo",
		RepoOwner: "test-user",
		CloneAddr: "https://github.com/example/repo.git",
	})
	require.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestClient_MigrateRepo_Bad_ServerError(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.MigrateRepo(forgejo.MigrateRepoOption{
		RepoName:  "migrated-repo",
		RepoOwner: "test-user",
		CloneAddr: "https://github.com/example/repo.git",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to migrate repo")
}
