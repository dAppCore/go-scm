// SPDX-License-Identifier: EUPL-1.2

package gitea

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	sdk "code.gitea.io/sdk/gitea"
	core "dappco.re/go"
)

const (
	sonarGiteaTestApplicationJson         = "application/json"
	sonarGiteaTestConfigYaml              = "config.yaml"
	sonarGiteaTestContentType             = "Content-Type"
	sonarGiteaTestHttpsExampleTestRepoGit = "https://example.test/repo.git"
)

func ax7GiteaClient(t *core.T) *Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/version" {
			w.Header().Set(sonarGiteaTestContentType, sonarGiteaTestApplicationJson)
			_, _ = w.Write([]byte(`{"version":"1.22.0"}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"blocked"}`))
	}))
	t.Cleanup(server.Close)
	client, err := New(server.URL, "token")
	core.RequireNoError(t, err)
	return client
}

func TestGitea_New_Good(t *core.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(sonarGiteaTestContentType, sonarGiteaTestApplicationJson)
		_, _ = w.Write([]byte(`{"version":"1.22.0"}`))
	}))
	t.Cleanup(server.Close)
	client, err := New(server.URL, "token")
	core.AssertNoError(t, err)
	core.AssertEqual(t, server.URL, client.URL())
}

func TestGitea_New_Bad(t *core.T) {
	client, err := New("", "token")
	core.AssertError(t, err)
	core.AssertNil(t, client)
}

func TestGitea_New_Ugly(t *core.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(sonarGiteaTestContentType, sonarGiteaTestApplicationJson)
		_, _ = w.Write([]byte(`{"version":"1.22.0"}`))
	}))
	t.Cleanup(server.Close)
	client, err := New(server.URL, "")
	core.AssertNoError(t, err)
	core.AssertNotNil(t, client.API())
}

func TestGitea_NewFromConfig_Good(t *core.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(sonarGiteaTestContentType, sonarGiteaTestApplicationJson)
		_, _ = w.Write([]byte(`{"version":"1.22.0"}`))
	}))
	t.Cleanup(server.Close)
	client, err := NewFromConfig(server.URL, "token")
	core.AssertNoError(t, err)
	core.AssertEqual(t, server.URL, client.URL())
}

func TestGitea_NewFromConfig_Bad(t *core.T) {
	client, err := NewFromConfig("http://example.test", "")
	core.AssertError(t, err)
	core.AssertNil(t, client)
}

func TestGitea_NewFromConfig_Ugly(t *core.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(sonarGiteaTestContentType, sonarGiteaTestApplicationJson)
		_, _ = w.Write([]byte(`{"version":"1.22.0"}`))
	}))
	t.Cleanup(server.Close)
	t.Setenv("GITEA_URL", server.URL)
	t.Setenv("GITEA_TOKEN", "env-token")
	client, err := NewFromConfig("", "")
	core.AssertNoError(t, err)
	core.AssertEqual(t, server.URL, client.URL())
}

func TestGitea_ResolveConfig_Good(t *core.T) {
	url, token, err := ResolveConfig("http://flag.test", "flag-token")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "http://flag.test", url)
	core.AssertEqual(t, "flag-token", token)
}

func TestGitea_ResolveConfig_Bad(t *core.T) {
	t.Setenv("GITEA_URL", "")
	t.Setenv("GITEA_TOKEN", "")
	url, token, err := ResolveConfig("", "")
	core.AssertNoError(t, err)
	core.AssertEqual(t, DefaultURL, url)
	core.AssertEqual(t, "", token)
}

func TestGitea_ResolveConfig_Ugly(t *core.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("GITEA_URL", "")
	t.Setenv("GITEA_TOKEN", "")
	core.RequireNoError(t, os.MkdirAll(filepath.Join(home, ".core"), 0o755))
	core.RequireNoError(t, os.WriteFile(filepath.Join(home, ".core", sonarGiteaTestConfigYaml), []byte("gitea:\n  url: http://file.test\n  token: file-token\n"), 0o600))
	url, token, err := ResolveConfig("", "")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "http://file.test", url)
	core.AssertEqual(t, "file-token", token)
}

func TestGitea_SaveConfig_Good(t *core.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	err := SaveConfig("http://save.test", "saved-token")
	core.AssertNoError(t, err)
	raw, readErr := os.ReadFile(filepath.Join(home, ".core", sonarGiteaTestConfigYaml))
	core.RequireNoError(t, readErr)
	core.AssertContains(t, string(raw), "saved-token")
}

func TestGitea_SaveConfig_Bad(t *core.T) {
	err := SaveConfig("", "")
	core.AssertError(t, err)
	core.AssertContains(t, err.Error(), "required")
}

func TestGitea_SaveConfig_Ugly(t *core.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	err := SaveConfig("", "token-only")
	core.AssertNoError(t, err)
	raw, readErr := os.ReadFile(filepath.Join(home, ".core", sonarGiteaTestConfigYaml))
	core.RequireNoError(t, readErr)
	core.AssertContains(t, string(raw), "token-only")
}

func TestGitea_Client_API_Good(t *core.T) {
	client := ax7GiteaClient(t)
	api := client.API()
	core.AssertNotNil(t, api)
}

func TestGitea_Client_API_Bad(t *core.T) {
	client := &Client{}
	api := client.API()
	core.AssertNil(t, api)
}

func TestGitea_Client_API_Ugly(t *core.T) {
	var client *Client
	api := client.API()
	core.AssertNil(t, api)
}

func TestGitea_Client_URL_Good(t *core.T) {
	client := ax7GiteaClient(t)
	url := client.URL()
	core.AssertContains(t, url, "127.0.0.1")
}

func TestGitea_Client_URL_Bad(t *core.T) {
	client := &Client{}
	url := client.URL()
	core.AssertEqual(t, "", url)
}

func TestGitea_Client_URL_Ugly(t *core.T) {
	var client *Client
	url := client.URL()
	core.AssertEqual(t, "", url)
}

func TestGitea_Client_ListIssues_Good(t *core.T) {
	client := ax7GiteaClient(t)
	_, err := client.ListIssues("owner", "repo", ListIssuesOpts{State: "all", Limit: 1})
	core.AssertError(t, err)
}

func TestGitea_Client_ListIssues_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.ListIssues("owner", "repo", ListIssuesOpts{}) })
	core.AssertNil(t, client.API())
}

func TestGitea_Client_ListIssues_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.ListIssues("owner", "repo", ListIssuesOpts{}) })
	core.AssertEqual(t, "", client.URL())
}

func TestGitea_Client_ListIssuesIter_Good(t *core.T) {
	client := ax7GiteaClient(t)
	var gotErr error
	for _, err := range client.ListIssuesIter("owner", "repo", ListIssuesOpts{State: "closed", Limit: 1}) {
		gotErr = err
		break
	}
	core.AssertError(t, gotErr)
}

func TestGitea_Client_ListIssuesIter_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() {
		for range client.ListIssuesIter("owner", "repo", ListIssuesOpts{}) {
			break
		}
	})
	core.AssertNil(t, client.API())
}

func TestGitea_Client_ListIssuesIter_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() {
		for range client.ListIssuesIter("owner", "repo", ListIssuesOpts{}) {
			break
		}
	})
	core.AssertEqual(t, "", client.URL())
}

func TestGitea_Client_GetIssue_Good(t *core.T) {
	client := ax7GiteaClient(t)
	_, err := client.GetIssue("owner", "repo", 7)
	core.AssertError(t, err)
}

func TestGitea_Client_GetIssue_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.GetIssue("owner", "repo", 7) })
	core.AssertNil(t, client.API())
}

func TestGitea_Client_GetIssue_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.GetIssue("owner", "repo", 7) })
	core.AssertEqual(t, "", client.URL())
}

func TestGitea_Client_CreateIssue_Good(t *core.T) {
	client := ax7GiteaClient(t)
	_, err := client.CreateIssue("owner", "repo", sdk.CreateIssueOption{Title: "demo"})
	core.AssertError(t, err)
}

func TestGitea_Client_CreateIssue_Bad(t *core.T) {
	client := ax7GiteaClient(t)
	_, err := client.CreateIssue("", "", sdk.CreateIssueOption{})
	core.AssertError(t, err)
}

func TestGitea_Client_CreateIssue_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.CreateIssue("owner", "repo", sdk.CreateIssueOption{}) })
	core.AssertEqual(t, "", client.URL())
}

func TestGitea_Client_EditIssue_Good(t *core.T) {
	client := ax7GiteaClient(t)
	_, err := client.EditIssue("owner", "repo", 7, sdk.EditIssueOption{})
	core.AssertError(t, err)
}

func TestGitea_Client_EditIssue_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.EditIssue("owner", "repo", 7, sdk.EditIssueOption{}) })
	core.AssertNil(t, client.API())
}

func TestGitea_Client_EditIssue_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.EditIssue("owner", "repo", 7, sdk.EditIssueOption{}) })
	core.AssertEqual(t, "", client.URL())
}

func TestGitea_Client_AssignIssue_Good(t *core.T) {
	client := ax7GiteaClient(t)
	err := client.AssignIssue("owner", "repo", 7, []string{"agent"})
	core.AssertError(t, err)
}

func TestGitea_Client_AssignIssue_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _ = client.AssignIssue("owner", "repo", 7, []string{"agent"}) })
	core.AssertNil(t, client.API())
}

func TestGitea_Client_AssignIssue_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _ = client.AssignIssue("owner", "repo", 7, nil) })
	core.AssertEqual(t, "", client.URL())
}

func TestGitea_Client_CreateIssueComment_Good(t *core.T) {
	client := ax7GiteaClient(t)
	err := client.CreateIssueComment("owner", "repo", 7, "body")
	core.AssertError(t, err)
}

func TestGitea_Client_CreateIssueComment_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _ = client.CreateIssueComment("owner", "repo", 7, "body") })
	core.AssertNil(t, client.API())
}

func TestGitea_Client_CreateIssueComment_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _ = client.CreateIssueComment("owner", "repo", 7, "body") })
	core.AssertEqual(t, "", client.URL())
}

func TestGitea_Client_ListIssueComments_Good(t *core.T) {
	client := ax7GiteaClient(t)
	_, err := client.ListIssueComments("owner", "repo", 7)
	core.AssertError(t, err)
}

func TestGitea_Client_ListIssueComments_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.ListIssueComments("owner", "repo", 7) })
	core.AssertNil(t, client.API())
}

func TestGitea_Client_ListIssueComments_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.ListIssueComments("owner", "repo", 7) })
	core.AssertEqual(t, "", client.URL())
}

func TestGitea_Client_ListIssueCommentsIter_Good(t *core.T) {
	client := ax7GiteaClient(t)
	var gotErr error
	for _, err := range client.ListIssueCommentsIter("owner", "repo", 7) {
		gotErr = err
		break
	}
	core.AssertError(t, gotErr)
}

func TestGitea_Client_ListIssueCommentsIter_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() {
		for range client.ListIssueCommentsIter("owner", "repo", 7) {
			break
		}
	})
	core.AssertNil(t, client.API())
}

func TestGitea_Client_ListIssueCommentsIter_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() {
		for range client.ListIssueCommentsIter("owner", "repo", 7) {
			break
		}
	})
	core.AssertEqual(t, "", client.URL())
}

func TestGitea_Client_GetIssueLabels_Good(t *core.T) {
	client := ax7GiteaClient(t)
	_, err := client.GetIssueLabels("owner", "repo", 7)
	core.AssertError(t, err)
}

func TestGitea_Client_GetIssueLabels_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.GetIssueLabels("owner", "repo", 7) })
	core.AssertNil(t, client.API())
}

func TestGitea_Client_GetIssueLabels_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.GetIssueLabels("owner", "repo", 7) })
	core.AssertEqual(t, "", client.URL())
}

func TestGitea_Client_AddIssueLabels_Good(t *core.T) {
	client := ax7GiteaClient(t)
	err := client.AddIssueLabels("owner", "repo", 7, []int64{1})
	core.AssertError(t, err)
}

func TestGitea_Client_AddIssueLabels_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _ = client.AddIssueLabels("owner", "repo", 7, []int64{1}) })
	core.AssertNil(t, client.API())
}

func TestGitea_Client_AddIssueLabels_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _ = client.AddIssueLabels("owner", "repo", 7, nil) })
	core.AssertEqual(t, "", client.URL())
}

func TestGitea_Client_RemoveIssueLabel_Good(t *core.T) {
	client := ax7GiteaClient(t)
	err := client.RemoveIssueLabel("owner", "repo", 7, 1)
	core.AssertError(t, err)
}

func TestGitea_Client_RemoveIssueLabel_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _ = client.RemoveIssueLabel("owner", "repo", 7, 1) })
	core.AssertNil(t, client.API())
}

func TestGitea_Client_RemoveIssueLabel_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _ = client.RemoveIssueLabel("owner", "repo", 7, 1) })
	core.AssertEqual(t, "", client.URL())
}

func TestGitea_Client_CloseIssue_Good(t *core.T) {
	client := ax7GiteaClient(t)
	err := client.CloseIssue("owner", "repo", 7)
	core.AssertError(t, err)
}

func TestGitea_Client_CloseIssue_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _ = client.CloseIssue("owner", "repo", 7) })
	core.AssertNil(t, client.API())
}

func TestGitea_Client_CloseIssue_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _ = client.CloseIssue("owner", "repo", 7) })
	core.AssertEqual(t, "", client.URL())
}

func TestGitea_Client_GetPullRequest_Good(t *core.T) {
	client := ax7GiteaClient(t)
	_, err := client.GetPullRequest("owner", "repo", 7)
	core.AssertError(t, err)
}

func TestGitea_Client_GetPullRequest_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.GetPullRequest("owner", "repo", 7) })
	core.AssertNil(t, client.API())
}

func TestGitea_Client_GetPullRequest_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.GetPullRequest("owner", "repo", 7) })
	core.AssertEqual(t, "", client.URL())
}

func TestGitea_Client_ListPullRequests_Good(t *core.T) {
	client := ax7GiteaClient(t)
	_, err := client.ListPullRequests("owner", "repo", "all")
	core.AssertError(t, err)
}

func TestGitea_Client_ListPullRequests_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.ListPullRequests("owner", "repo", "closed") })
	core.AssertNil(t, client.API())
}

func TestGitea_Client_ListPullRequests_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.ListPullRequests("owner", "repo", "") })
	core.AssertEqual(t, "", client.URL())
}

func TestGitea_Client_ListPullRequestsIter_Good(t *core.T) {
	client := ax7GiteaClient(t)
	var gotErr error
	for _, err := range client.ListPullRequestsIter("owner", "repo", "all") {
		gotErr = err
		break
	}
	core.AssertError(t, gotErr)
}

func TestGitea_Client_ListPullRequestsIter_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() {
		for range client.ListPullRequestsIter("owner", "repo", "closed") {
			break
		}
	})
	core.AssertNil(t, client.API())
}

func TestGitea_Client_ListPullRequestsIter_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() {
		for range client.ListPullRequestsIter("owner", "repo", "") {
			break
		}
	})
	core.AssertEqual(t, "", client.URL())
}

func TestGitea_Client_CreateOrgRepo_Good(t *core.T) {
	client := ax7GiteaClient(t)
	_, err := client.CreateOrgRepo("org", sdk.CreateRepoOption{Name: "repo"})
	core.AssertError(t, err)
}

func TestGitea_Client_CreateOrgRepo_Bad(t *core.T) {
	client := ax7GiteaClient(t)
	_, err := client.CreateOrgRepo("", sdk.CreateRepoOption{})
	core.AssertError(t, err)
}

func TestGitea_Client_CreateOrgRepo_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.CreateOrgRepo("org", sdk.CreateRepoOption{}) })
	core.AssertEqual(t, "", client.URL())
}

func TestGitea_Client_DeleteRepo_Good(t *core.T) {
	client := ax7GiteaClient(t)
	err := client.DeleteRepo("owner", "repo")
	core.AssertError(t, err)
}

func TestGitea_Client_DeleteRepo_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _ = client.DeleteRepo("owner", "repo") })
	core.AssertNil(t, client.API())
}

func TestGitea_Client_DeleteRepo_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _ = client.DeleteRepo("owner", "repo") })
	core.AssertEqual(t, "", client.URL())
}

func TestGitea_Client_GetRepo_Good(t *core.T) {
	client := ax7GiteaClient(t)
	_, err := client.GetRepo("owner", "repo")
	core.AssertError(t, err)
}

func TestGitea_Client_GetRepo_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.GetRepo("owner", "repo") })
	core.AssertNil(t, client.API())
}

func TestGitea_Client_GetRepo_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.GetRepo("owner", "repo") })
	core.AssertEqual(t, "", client.URL())
}

func TestGitea_Client_ListOrgRepos_Good(t *core.T) {
	client := ax7GiteaClient(t)
	_, err := client.ListOrgRepos("org")
	core.AssertError(t, err)
}

func TestGitea_Client_ListOrgRepos_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.ListOrgRepos("org") })
	core.AssertNil(t, client.API())
}

func TestGitea_Client_ListOrgRepos_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.ListOrgRepos("org") })
	core.AssertEqual(t, "", client.URL())
}

func TestGitea_Client_ListOrgReposIter_Good(t *core.T) {
	client := ax7GiteaClient(t)
	var gotErr error
	for _, err := range client.ListOrgReposIter("org") {
		gotErr = err
		break
	}
	core.AssertError(t, gotErr)
}

func TestGitea_Client_ListOrgReposIter_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() {
		for range client.ListOrgReposIter("org") {
			break
		}
	})
	core.AssertNil(t, client.API())
}

func TestGitea_Client_ListOrgReposIter_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() {
		for range client.ListOrgReposIter("org") {
			break
		}
	})
	core.AssertEqual(t, "", client.URL())
}

func TestGitea_Client_ListUserRepos_Good(t *core.T) {
	client := ax7GiteaClient(t)
	_, err := client.ListUserRepos()
	core.AssertError(t, err)
}

func TestGitea_Client_ListUserRepos_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.ListUserRepos() })
	core.AssertNil(t, client.API())
}

func TestGitea_Client_ListUserRepos_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.ListUserRepos() })
	core.AssertEqual(t, "", client.URL())
}

func TestGitea_Client_ListUserReposIter_Good(t *core.T) {
	client := ax7GiteaClient(t)
	var gotErr error
	for _, err := range client.ListUserReposIter() {
		gotErr = err
		break
	}
	core.AssertError(t, gotErr)
}

func TestGitea_Client_ListUserReposIter_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() {
		for range client.ListUserReposIter() {
			break
		}
	})
	core.AssertNil(t, client.API())
}

func TestGitea_Client_ListUserReposIter_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() {
		for range client.ListUserReposIter() {
			break
		}
	})
	core.AssertEqual(t, "", client.URL())
}

func TestGitea_Client_CreateMirror_Good(t *core.T) {
	client := ax7GiteaClient(t)
	_, err := client.CreateMirror("owner", "repo", sonarGiteaTestHttpsExampleTestRepoGit, "token")
	core.AssertError(t, err)
}

func TestGitea_Client_CreateMirror_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.CreateMirror("owner", "repo", sonarGiteaTestHttpsExampleTestRepoGit, "") })
	core.AssertNil(t, client.API())
}

func TestGitea_Client_CreateMirror_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.CreateMirror("owner", "repo", "", "") })
	core.AssertEqual(t, "", client.URL())
}

func TestGitea_Client_CreateMirrorFromService_Good(t *core.T) {
	client := ax7GiteaClient(t)
	_, err := client.CreateMirrorFromService("owner", "repo", sonarGiteaTestHttpsExampleTestRepoGit, sdk.GitServicePlain, "token")
	core.AssertError(t, err)
}

func TestGitea_Client_CreateMirrorFromService_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() {
		_, _ = client.CreateMirrorFromService("owner", "repo", sonarGiteaTestHttpsExampleTestRepoGit, sdk.GitServicePlain, "")
	})
	core.AssertNil(t, client.API())
}

func TestGitea_Client_CreateMirrorFromService_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.CreateMirrorFromService("owner", "repo", "", sdk.GitServicePlain, "") })
	core.AssertEqual(t, "", client.URL())
}

func TestGitea_Client_GetIssueBody_Good(t *core.T) {
	client := ax7GiteaClient(t)
	_, err := client.GetIssueBody("owner", "repo", 7)
	core.AssertError(t, err)
}

func TestGitea_Client_GetIssueBody_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.GetIssueBody("owner", "repo", 7) })
	core.AssertNil(t, client.API())
}

func TestGitea_Client_GetIssueBody_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.GetIssueBody("owner", "repo", 7) })
	core.AssertEqual(t, "", client.URL())
}

func TestGitea_Client_GetCommentBodies_Good(t *core.T) {
	client := ax7GiteaClient(t)
	_, err := client.GetCommentBodies("owner", "repo", 7)
	core.AssertError(t, err)
}

func TestGitea_Client_GetCommentBodies_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.GetCommentBodies("owner", "repo", 7) })
	core.AssertNil(t, client.API())
}

func TestGitea_Client_GetCommentBodies_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.GetCommentBodies("owner", "repo", 7) })
	core.AssertEqual(t, "", client.URL())
}

func TestGitea_Client_GetPRMeta_Good(t *core.T) {
	client := ax7GiteaClient(t)
	_, err := client.GetPRMeta("owner", "repo", 7)
	core.AssertError(t, err)
}

func TestGitea_Client_GetPRMeta_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.GetPRMeta("owner", "repo", 7) })
	core.AssertNil(t, client.API())
}

func TestGitea_Client_GetPRMeta_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.GetPRMeta("owner", "repo", 7) })
	core.AssertEqual(t, "", client.URL())
}
