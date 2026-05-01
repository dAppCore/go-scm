// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	"net/http"
	"net/http/httptest"

	forgejo "codeberg.org/forgejo/go-sdk/forgejo"
	core "dappco.re/go"
)

const (
	sonarForgeTestApplicationJson      = "application/json"
	sonarForgeTestConfigYaml           = "config.yaml"
	sonarForgeTestContentType          = "Content-Type"
	sonarForgeTestUnexpectedHttpStatus = "unexpected HTTP status"
)

func testForgeClient(t *core.T) *Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/version" {
			w.Header().Set(sonarForgeTestContentType, sonarForgeTestApplicationJson)
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

func TestForge_New_Good(t *core.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(sonarForgeTestContentType, sonarForgeTestApplicationJson)
		_, _ = w.Write([]byte(`{"version":"1.22.0"}`))
	}))
	t.Cleanup(server.Close)
	client, err := New(server.URL, "token")
	core.AssertNoError(t, err)
	core.AssertEqual(t, server.URL, client.URL())
}

func TestForge_New_Bad(t *core.T) {
	client, err := New("", "token")
	core.AssertError(t, err)
	core.AssertNil(t, client)
}

func TestForge_New_Ugly(t *core.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(sonarForgeTestContentType, sonarForgeTestApplicationJson)
		_, _ = w.Write([]byte(`{"version":"1.22.0"}`))
	}))
	t.Cleanup(server.Close)
	client, err := New(server.URL, "")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "", client.Token())
}

func TestForge_NewFromConfig_Good(t *core.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(sonarForgeTestContentType, sonarForgeTestApplicationJson)
		_, _ = w.Write([]byte(`{"version":"1.22.0"}`))
	}))
	t.Cleanup(server.Close)
	client, err := NewFromConfig(server.URL, "token")
	core.AssertNoError(t, err)
	core.AssertEqual(t, server.URL, client.URL())
}

func TestForge_NewFromConfig_Bad(t *core.T) {
	client, err := NewFromConfig("http://example.test", "")
	core.AssertError(t, err)
	core.AssertNil(t, client)
}

func TestForge_NewFromConfig_Ugly(t *core.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(sonarForgeTestContentType, sonarForgeTestApplicationJson)
		_, _ = w.Write([]byte(`{"version":"1.22.0"}`))
	}))
	t.Cleanup(server.Close)
	t.Setenv("FORGE_URL", server.URL)
	t.Setenv("FORGE_TOKEN", "env-token")
	client, err := NewFromConfig("", "")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "env-token", client.Token())
}

func TestForge_ResolveConfig_Good(t *core.T) {
	url, token, err := ResolveConfig("http://flag.test", "flag-token")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "http://flag.test", url)
	core.AssertEqual(t, "flag-token", token)
}

func TestForge_ResolveConfig_Bad(t *core.T) {
	t.Setenv("FORGE_URL", "")
	t.Setenv("FORGE_TOKEN", "")
	url, token, err := ResolveConfig("", "")
	core.AssertError(t, err)
	core.AssertEqual(t, "", url)
	core.AssertEqual(t, "", token)
}

func TestForge_ResolveConfig_Ugly(t *core.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("FORGE_URL", "")
	t.Setenv("FORGE_TOKEN", "")
	if r := core.MkdirAll(core.PathJoin(home, ".core"), 0o755); !r.OK {
		core.RequireNoError(t, r.Value.(error))
	}
	if r := core.WriteFile(core.PathJoin(home, ".core", sonarForgeTestConfigYaml), []byte("forge:\n  url: http://file.test\n  token: file-token\n"), 0o600); !r.OK {
		core.RequireNoError(t, r.Value.(error))
	}
	url, token, err := ResolveConfig("", "")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "http://file.test", url)
	core.AssertEqual(t, "file-token", token)
}

func TestForge_SaveConfig_Good(t *core.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	err := SaveConfig("http://save.test", "saved-token")
	core.AssertNoError(t, err)
	r := core.ReadFile(core.PathJoin(home, ".core", sonarForgeTestConfigYaml))
	if !r.OK {
		core.RequireNoError(t, r.Value.(error))
	}
	raw := r.Value.([]byte)
	core.AssertContains(t, string(raw), "saved-token")
}

func TestForge_SaveConfig_Bad(t *core.T) {
	err := SaveConfig("", "")
	core.AssertError(t, err)
	core.AssertContains(t, err.Error(), "required")
}

func TestForge_SaveConfig_Ugly(t *core.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	err := SaveConfig("", "token-only")
	core.AssertNoError(t, err)
	r := core.ReadFile(core.PathJoin(home, ".core", sonarForgeTestConfigYaml))
	if !r.OK {
		core.RequireNoError(t, r.Value.(error))
	}
	raw := r.Value.([]byte)
	core.AssertContains(t, string(raw), "token-only")
}

func TestForge_Client_API_Good(t *core.T) {
	client := testForgeClient(t)
	api := client.API()
	core.AssertNotNil(t, api)
}

func TestForge_Client_API_Bad(t *core.T) {
	client := &Client{}
	api := client.API()
	core.AssertNil(t, api)
}

func TestForge_Client_API_Ugly(t *core.T) {
	var client *Client
	api := client.API()
	core.AssertNil(t, api)
}

func TestForge_Client_URL_Good(t *core.T) {
	client := testForgeClient(t)
	url := client.URL()
	core.AssertContains(t, url, "127.0.0.1")
}

func TestForge_Client_URL_Bad(t *core.T) {
	client := &Client{}
	url := client.URL()
	core.AssertEqual(t, "", url)
}

func TestForge_Client_URL_Ugly(t *core.T) {
	var client *Client
	url := client.URL()
	core.AssertEqual(t, "", url)
}

func TestForge_Client_Token_Good(t *core.T) {
	client := testForgeClient(t)
	token := client.Token()
	core.AssertEqual(t, "token", token)
}

func TestForge_Client_Token_Bad(t *core.T) {
	client := &Client{}
	token := client.Token()
	core.AssertEqual(t, "", token)
}

func TestForge_Client_Token_Ugly(t *core.T) {
	var client *Client
	token := client.Token()
	core.AssertEqual(t, "", token)
}

func TestForge_Error_Error_Good(t *core.T) {
	err := (&httpError{status: http.StatusTeapot}).Error()
	core.AssertEqual(t, sonarForgeTestUnexpectedHttpStatus, err)
	core.AssertEqual(t, http.StatusTeapot, (&httpError{status: http.StatusTeapot}).status)
}

func TestForge_Error_Error_Bad(t *core.T) {
	err := (&httpError{}).Error()
	core.AssertEqual(t, sonarForgeTestUnexpectedHttpStatus, err)
	core.AssertEqual(t, 0, (&httpError{}).status)
}

func TestForge_Error_Error_Ugly(t *core.T) {
	var err *httpError
	got := err.Error()
	core.AssertEqual(t, sonarForgeTestUnexpectedHttpStatus, got)
}

func TestForge_Client_GetCurrentUser_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.GetCurrentUser()
	core.AssertError(t, err)
}

func TestForge_Client_GetCurrentUser_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.GetCurrentUser() })
	core.AssertNil(t, client.API())
}

func TestForge_Client_GetCurrentUser_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.GetCurrentUser() })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_CreatePullRequest_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.CreatePullRequest("owner", "repo", forgejo.CreatePullRequestOption{Title: "demo", Head: "feature", Base: "dev"})
	core.AssertError(t, err)
}

func TestForge_Client_CreatePullRequest_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.CreatePullRequest("owner", "repo", forgejo.CreatePullRequestOption{}) })
	core.AssertNil(t, client.API())
}

func TestForge_Client_CreatePullRequest_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.CreatePullRequest("owner", "repo", forgejo.CreatePullRequestOption{}) })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_ForkRepo_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.ForkRepo("owner", "repo", "org")
	core.AssertError(t, err)
}

func TestForge_Client_ForkRepo_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.ForkRepo("owner", "repo", "") })
	core.AssertNil(t, client.API())
}

func TestForge_Client_ForkRepo_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.ForkRepo("owner", "repo", "") })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_GetIssue_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.GetIssue("owner", "repo", 7)
	core.AssertError(t, err)
}

func TestForge_Client_GetIssue_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.GetIssue("owner", "repo", 7) })
	core.AssertNil(t, client.API())
}

func TestForge_Client_GetIssue_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.GetIssue("owner", "repo", 7) })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_EditIssue_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.EditIssue("owner", "repo", 7, forgejo.EditIssueOption{})
	core.AssertError(t, err)
}

func TestForge_Client_EditIssue_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.EditIssue("owner", "repo", 7, forgejo.EditIssueOption{}) })
	core.AssertNil(t, client.API())
}

func TestForge_Client_EditIssue_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.EditIssue("owner", "repo", 7, forgejo.EditIssueOption{}) })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_CloseIssue_Good(t *core.T) {
	client := testForgeClient(t)
	err := client.CloseIssue("owner", "repo", 7)
	core.AssertError(t, err)
}

func TestForge_Client_CloseIssue_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _ = client.CloseIssue("owner", "repo", 7) })
	core.AssertNil(t, client.API())
}

func TestForge_Client_CloseIssue_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _ = client.CloseIssue("owner", "repo", 7) })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_CreateIssue_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.CreateIssue("owner", "repo", forgejo.CreateIssueOption{Title: "demo"})
	core.AssertError(t, err)
}

func TestForge_Client_CreateIssue_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.CreateIssue("owner", "repo", forgejo.CreateIssueOption{}) })
	core.AssertNil(t, client.API())
}

func TestForge_Client_CreateIssue_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.CreateIssue("owner", "repo", forgejo.CreateIssueOption{}) })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_CreateIssueComment_Good(t *core.T) {
	client := testForgeClient(t)
	err := client.CreateIssueComment("owner", "repo", 7, "body")
	core.AssertError(t, err)
}

func TestForge_Client_CreateIssueComment_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _ = client.CreateIssueComment("owner", "repo", 7, "body") })
	core.AssertNil(t, client.API())
}

func TestForge_Client_CreateIssueComment_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _ = client.CreateIssueComment("owner", "repo", 7, "body") })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_AssignIssue_Good(t *core.T) {
	client := testForgeClient(t)
	err := client.AssignIssue("owner", "repo", 7, []string{"agent"})
	core.AssertError(t, err)
}

func TestForge_Client_AssignIssue_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _ = client.AssignIssue("owner", "repo", 7, []string{"agent"}) })
	core.AssertNil(t, client.API())
}

func TestForge_Client_AssignIssue_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _ = client.AssignIssue("owner", "repo", 7, nil) })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_ListIssueComments_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.ListIssueComments("owner", "repo", 7)
	core.AssertError(t, err)
}

func TestForge_Client_ListIssueComments_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.ListIssueComments("owner", "repo", 7) })
	core.AssertNil(t, client.API())
}

func TestForge_Client_ListIssueComments_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.ListIssueComments("owner", "repo", 7) })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_ListIssueCommentsIter_Good(t *core.T) {
	client := testForgeClient(t)
	var gotErr error
	for _, err := range client.ListIssueCommentsIter("owner", "repo", 7) {
		gotErr = err
		break
	}
	core.AssertError(t, gotErr)
}

func TestForge_Client_ListIssueCommentsIter_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() {
		for range client.ListIssueCommentsIter("owner", "repo", 7) {
			break
		}
	})
	core.AssertNil(t, client.API())
}

func TestForge_Client_ListIssueCommentsIter_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() {
		for range client.ListIssueCommentsIter("owner", "repo", 7) {
			break
		}
	})
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_GetIssueLabels_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.GetIssueLabels("owner", "repo", 7)
	core.AssertError(t, err)
}

func TestForge_Client_GetIssueLabels_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.GetIssueLabels("owner", "repo", 7) })
	core.AssertNil(t, client.API())
}

func TestForge_Client_GetIssueLabels_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.GetIssueLabels("owner", "repo", 7) })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_ListIssues_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.ListIssues("owner", "repo", ListIssuesOpts{State: "all", Limit: 1})
	core.AssertError(t, err)
}

func TestForge_Client_ListIssues_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.ListIssues("owner", "repo", ListIssuesOpts{}) })
	core.AssertNil(t, client.API())
}

func TestForge_Client_ListIssues_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.ListIssues("owner", "repo", ListIssuesOpts{}) })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_ListPullRequests_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.ListPullRequests("owner", "repo", "all")
	core.AssertError(t, err)
}

func TestForge_Client_ListPullRequests_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.ListPullRequests("owner", "repo", "closed") })
	core.AssertNil(t, client.API())
}

func TestForge_Client_ListPullRequests_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.ListPullRequests("owner", "repo", "") })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_ListPullRequestsIter_Good(t *core.T) {
	client := testForgeClient(t)
	var gotErr error
	for _, err := range client.ListPullRequestsIter("owner", "repo", "all") {
		gotErr = err
		break
	}
	core.AssertError(t, gotErr)
}

func TestForge_Client_ListPullRequestsIter_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() {
		for range client.ListPullRequestsIter("owner", "repo", "closed") {
			break
		}
	})
	core.AssertNil(t, client.API())
}

func TestForge_Client_ListPullRequestsIter_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() {
		for range client.ListPullRequestsIter("owner", "repo", "") {
			break
		}
	})
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_GetPullRequest_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.GetPullRequest("owner", "repo", 7)
	core.AssertError(t, err)
}

func TestForge_Client_GetPullRequest_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.GetPullRequest("owner", "repo", 7) })
	core.AssertNil(t, client.API())
}

func TestForge_Client_GetPullRequest_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.GetPullRequest("owner", "repo", 7) })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_CreateOrgRepo_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.CreateOrgRepo("org", forgejo.CreateRepoOption{Name: "repo"})
	core.AssertError(t, err)
}

func TestForge_Client_CreateOrgRepo_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.CreateOrgRepo("org", forgejo.CreateRepoOption{}) })
	core.AssertNil(t, client.API())
}

func TestForge_Client_CreateOrgRepo_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.CreateOrgRepo("org", forgejo.CreateRepoOption{}) })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_DeleteRepo_Good(t *core.T) {
	client := testForgeClient(t)
	err := client.DeleteRepo("owner", "repo")
	core.AssertError(t, err)
}

func TestForge_Client_DeleteRepo_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _ = client.DeleteRepo("owner", "repo") })
	core.AssertNil(t, client.API())
}

func TestForge_Client_DeleteRepo_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _ = client.DeleteRepo("owner", "repo") })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_GetRepo_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.GetRepo("owner", "repo")
	core.AssertError(t, err)
}

func TestForge_Client_GetRepo_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.GetRepo("owner", "repo") })
	core.AssertNil(t, client.API())
}

func TestForge_Client_GetRepo_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.GetRepo("owner", "repo") })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_ListOrgRepos_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.ListOrgRepos("org")
	core.AssertError(t, err)
}

func TestForge_Client_ListOrgRepos_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.ListOrgRepos("org") })
	core.AssertNil(t, client.API())
}

func TestForge_Client_ListOrgRepos_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.ListOrgRepos("org") })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_ListOrgReposIter_Good(t *core.T) {
	client := testForgeClient(t)
	var gotErr error
	for _, err := range client.ListOrgReposIter("org") {
		gotErr = err
		break
	}
	core.AssertError(t, gotErr)
}

func TestForge_Client_ListOrgReposIter_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() {
		for range client.ListOrgReposIter("org") {
			break
		}
	})
	core.AssertNil(t, client.API())
}

func TestForge_Client_ListOrgReposIter_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() {
		for range client.ListOrgReposIter("org") {
			break
		}
	})
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_ListUserRepos_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.ListUserRepos()
	core.AssertError(t, err)
}

func TestForge_Client_ListUserRepos_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.ListUserRepos() })
	core.AssertNil(t, client.API())
}

func TestForge_Client_ListUserRepos_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.ListUserRepos() })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_ListUserReposIter_Good(t *core.T) {
	client := testForgeClient(t)
	var gotErr error
	for _, err := range client.ListUserReposIter() {
		gotErr = err
		break
	}
	core.AssertError(t, gotErr)
}

func TestForge_Client_ListUserReposIter_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() {
		for range client.ListUserReposIter() {
			break
		}
	})
	core.AssertNil(t, client.API())
}

func TestForge_Client_ListUserReposIter_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() {
		for range client.ListUserReposIter() {
			break
		}
	})
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_MigrateRepo_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.MigrateRepo(forgejo.MigrateRepoOption{RepoName: "repo", CloneAddr: "https://example.test/repo.git"})
	core.AssertError(t, err)
}

func TestForge_Client_MigrateRepo_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.MigrateRepo(forgejo.MigrateRepoOption{}) })
	core.AssertNil(t, client.API())
}

func TestForge_Client_MigrateRepo_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.MigrateRepo(forgejo.MigrateRepoOption{}) })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_CreateRepoLabel_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.CreateRepoLabel("owner", "repo", forgejo.CreateLabelOption{Name: "ready", Color: "00ff00"})
	core.AssertError(t, err)
}

func TestForge_Client_CreateRepoLabel_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.CreateRepoLabel("owner", "repo", forgejo.CreateLabelOption{}) })
	core.AssertNil(t, client.API())
}

func TestForge_Client_CreateRepoLabel_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.CreateRepoLabel("owner", "repo", forgejo.CreateLabelOption{}) })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_ListRepoLabels_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.ListRepoLabels("owner", "repo")
	core.AssertError(t, err)
}

func TestForge_Client_ListRepoLabels_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.ListRepoLabels("owner", "repo") })
	core.AssertNil(t, client.API())
}

func TestForge_Client_ListRepoLabels_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.ListRepoLabels("owner", "repo") })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_ListRepoLabelsIter_Good(t *core.T) {
	client := testForgeClient(t)
	var gotErr error
	for _, err := range client.ListRepoLabelsIter("owner", "repo") {
		gotErr = err
		break
	}
	core.AssertError(t, gotErr)
}

func TestForge_Client_ListRepoLabelsIter_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() {
		for range client.ListRepoLabelsIter("owner", "repo") {
			break
		}
	})
	core.AssertNil(t, client.API())
}

func TestForge_Client_ListRepoLabelsIter_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() {
		for range client.ListRepoLabelsIter("owner", "repo") {
			break
		}
	})
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_ListOrgLabels_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.ListOrgLabels("org")
	core.AssertError(t, err)
}

func TestForge_Client_ListOrgLabels_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.ListOrgLabels("org") })
	core.AssertNil(t, client.API())
}

func TestForge_Client_ListOrgLabels_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.ListOrgLabels("org") })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_ListOrgLabelsIter_Good(t *core.T) {
	client := testForgeClient(t)
	var gotErr error
	for _, err := range client.ListOrgLabelsIter("org") {
		gotErr = err
		break
	}
	core.AssertError(t, gotErr)
}

func TestForge_Client_ListOrgLabelsIter_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() {
		for range client.ListOrgLabelsIter("org") {
			break
		}
	})
	core.AssertNil(t, client.API())
}

func TestForge_Client_ListOrgLabelsIter_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() {
		for range client.ListOrgLabelsIter("org") {
			break
		}
	})
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_GetLabelByName_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.GetLabelByName("owner", "repo", "ready")
	core.AssertError(t, err)
}

func TestForge_Client_GetLabelByName_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.GetLabelByName("owner", "repo", "ready") })
	core.AssertNil(t, client.API())
}

func TestForge_Client_GetLabelByName_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.GetLabelByName("owner", "repo", "ready") })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_EnsureLabel_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.EnsureLabel("owner", "repo", "ready", "00ff00")
	core.AssertError(t, err)
}

func TestForge_Client_EnsureLabel_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.EnsureLabel("owner", "repo", "ready", "00ff00") })
	core.AssertNil(t, client.API())
}

func TestForge_Client_EnsureLabel_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.EnsureLabel("owner", "repo", "ready", "00ff00") })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_AddIssueLabels_Good(t *core.T) {
	client := testForgeClient(t)
	err := client.AddIssueLabels("owner", "repo", 7, []int64{1})
	core.AssertError(t, err)
}

func TestForge_Client_AddIssueLabels_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _ = client.AddIssueLabels("owner", "repo", 7, []int64{1}) })
	core.AssertNil(t, client.API())
}

func TestForge_Client_AddIssueLabels_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _ = client.AddIssueLabels("owner", "repo", 7, nil) })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_RemoveIssueLabel_Good(t *core.T) {
	client := testForgeClient(t)
	err := client.RemoveIssueLabel("owner", "repo", 7, 1)
	core.AssertError(t, err)
}

func TestForge_Client_RemoveIssueLabel_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _ = client.RemoveIssueLabel("owner", "repo", 7, 1) })
	core.AssertNil(t, client.API())
}

func TestForge_Client_RemoveIssueLabel_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _ = client.RemoveIssueLabel("owner", "repo", 7, 1) })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_MergePullRequest_Good(t *core.T) {
	client := testForgeClient(t)
	err := client.MergePullRequest("owner", "repo", 7, "squash")
	core.AssertNoError(t, err)
}

func TestForge_Client_MergePullRequest_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _ = client.MergePullRequest("owner", "repo", 7, "merge") })
	core.AssertNil(t, client.API())
}

func TestForge_Client_MergePullRequest_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _ = client.MergePullRequest("owner", "repo", 7, "rebase") })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_SetPRDraft_Good(t *core.T) {
	client := testForgeClient(t)
	err := client.SetPRDraft("owner", "repo", 7, false)
	core.AssertError(t, err)
}

func TestForge_Client_SetPRDraft_Bad(t *core.T) {
	client := &Client{url: "://bad"}
	err := client.SetPRDraft("owner", "repo", 7, false)
	core.AssertError(t, err)
}

func TestForge_Client_SetPRDraft_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _ = client.SetPRDraft("owner", "repo", 7, false) })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_ListPRReviews_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.ListPRReviews("owner", "repo", 7)
	core.AssertError(t, err)
}

func TestForge_Client_ListPRReviews_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.ListPRReviews("owner", "repo", 7) })
	core.AssertNil(t, client.API())
}

func TestForge_Client_ListPRReviews_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.ListPRReviews("owner", "repo", 7) })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_ListPRReviewsIter_Good(t *core.T) {
	client := testForgeClient(t)
	var gotErr error
	for _, err := range client.ListPRReviewsIter("owner", "repo", 7) {
		gotErr = err
		break
	}
	core.AssertError(t, gotErr)
}

func TestForge_Client_ListPRReviewsIter_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() {
		for range client.ListPRReviewsIter("owner", "repo", 7) {
			break
		}
	})
	core.AssertNil(t, client.API())
}

func TestForge_Client_ListPRReviewsIter_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() {
		for range client.ListPRReviewsIter("owner", "repo", 7) {
			break
		}
	})
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_GetCombinedStatus_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.GetCombinedStatus("owner", "repo", "sha")
	core.AssertError(t, err)
}

func TestForge_Client_GetCombinedStatus_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.GetCombinedStatus("owner", "repo", "sha") })
	core.AssertNil(t, client.API())
}

func TestForge_Client_GetCombinedStatus_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.GetCombinedStatus("owner", "repo", "sha") })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_DismissReview_Good(t *core.T) {
	client := testForgeClient(t)
	err := client.DismissReview("owner", "repo", 7, 9, "stale")
	core.AssertError(t, err)
}

func TestForge_Client_DismissReview_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _ = client.DismissReview("owner", "repo", 7, 9, "stale") })
	core.AssertNil(t, client.API())
}

func TestForge_Client_DismissReview_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _ = client.DismissReview("owner", "repo", 7, 9, "stale") })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_UndismissReview_Good(t *core.T) {
	client := testForgeClient(t)
	err := client.UndismissReview("owner", "repo", 7, 9)
	core.AssertError(t, err)
}

func TestForge_Client_UndismissReview_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _ = client.UndismissReview("owner", "repo", 7, 9) })
	core.AssertNil(t, client.API())
}

func TestForge_Client_UndismissReview_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _ = client.UndismissReview("owner", "repo", 7, 9) })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_GetIssueBody_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.GetIssueBody("owner", "repo", 7)
	core.AssertError(t, err)
}

func TestForge_Client_GetIssueBody_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.GetIssueBody("owner", "repo", 7) })
	core.AssertNil(t, client.API())
}

func TestForge_Client_GetIssueBody_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.GetIssueBody("owner", "repo", 7) })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_GetCommentBodies_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.GetCommentBodies("owner", "repo", 7)
	core.AssertError(t, err)
}

func TestForge_Client_GetCommentBodies_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.GetCommentBodies("owner", "repo", 7) })
	core.AssertNil(t, client.API())
}

func TestForge_Client_GetCommentBodies_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.GetCommentBodies("owner", "repo", 7) })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_GetPRMeta_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.GetPRMeta("owner", "repo", 7)
	core.AssertError(t, err)
}

func TestForge_Client_GetPRMeta_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.GetPRMeta("owner", "repo", 7) })
	core.AssertNil(t, client.API())
}

func TestForge_Client_GetPRMeta_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.GetPRMeta("owner", "repo", 7) })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_CreateOrg_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.CreateOrg(forgejo.CreateOrgOption{Name: "org"})
	core.AssertError(t, err)
}

func TestForge_Client_CreateOrg_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.CreateOrg(forgejo.CreateOrgOption{}) })
	core.AssertNil(t, client.API())
}

func TestForge_Client_CreateOrg_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.CreateOrg(forgejo.CreateOrgOption{}) })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_GetOrg_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.GetOrg("org")
	core.AssertError(t, err)
}

func TestForge_Client_GetOrg_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.GetOrg("org") })
	core.AssertNil(t, client.API())
}

func TestForge_Client_GetOrg_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.GetOrg("org") })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_ListMyOrgs_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.ListMyOrgs()
	core.AssertError(t, err)
}

func TestForge_Client_ListMyOrgs_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.ListMyOrgs() })
	core.AssertNil(t, client.API())
}

func TestForge_Client_ListMyOrgs_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.ListMyOrgs() })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_ListMyOrgsIter_Good(t *core.T) {
	client := testForgeClient(t)
	var gotErr error
	for _, err := range client.ListMyOrgsIter() {
		gotErr = err
		break
	}
	core.AssertError(t, gotErr)
}

func TestForge_Client_ListMyOrgsIter_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() {
		for range client.ListMyOrgsIter() {
			break
		}
	})
	core.AssertNil(t, client.API())
}

func TestForge_Client_ListMyOrgsIter_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() {
		for range client.ListMyOrgsIter() {
			break
		}
	})
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_CreateRepoWebhook_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.CreateRepoWebhook("owner", "repo", forgejo.CreateHookOption{Type: "gitea"})
	core.AssertError(t, err)
}

func TestForge_Client_CreateRepoWebhook_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.CreateRepoWebhook("owner", "repo", forgejo.CreateHookOption{}) })
	core.AssertNil(t, client.API())
}

func TestForge_Client_CreateRepoWebhook_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.CreateRepoWebhook("owner", "repo", forgejo.CreateHookOption{}) })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_ListRepoWebhooks_Good(t *core.T) {
	client := testForgeClient(t)
	_, err := client.ListRepoWebhooks("owner", "repo")
	core.AssertError(t, err)
}

func TestForge_Client_ListRepoWebhooks_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() { _, _ = client.ListRepoWebhooks("owner", "repo") })
	core.AssertNil(t, client.API())
}

func TestForge_Client_ListRepoWebhooks_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() { _, _ = client.ListRepoWebhooks("owner", "repo") })
	core.AssertEqual(t, "", client.URL())
}

func TestForge_Client_ListRepoWebhooksIter_Good(t *core.T) {
	client := testForgeClient(t)
	var gotErr error
	for _, err := range client.ListRepoWebhooksIter("owner", "repo") {
		gotErr = err
		break
	}
	core.AssertError(t, gotErr)
}

func TestForge_Client_ListRepoWebhooksIter_Bad(t *core.T) {
	client := &Client{}
	core.AssertPanics(t, func() {
		for range client.ListRepoWebhooksIter("owner", "repo") {
			break
		}
	})
	core.AssertNil(t, client.API())
}

func TestForge_Client_ListRepoWebhooksIter_Ugly(t *core.T) {
	var client *Client
	core.AssertPanics(t, func() {
		for range client.ListRepoWebhooksIter("owner", "repo") {
			break
		}
	})
	core.AssertEqual(t, "", client.URL())
}
