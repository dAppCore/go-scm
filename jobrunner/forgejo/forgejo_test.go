// SPDX-License-Identifier: EUPL-1.2

package forgejo

import (
	"context"
	"net/http"
	"net/http/httptest"

	core "dappco.re/go"
	coreforge "dappco.re/go/scm/forge"
	"dappco.re/go/scm/jobrunner"
)

func ax7ForgeClient(t *core.T) *coreforge.Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/version" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"version":"1.22.0"}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"blocked"}`))
	}))
	t.Cleanup(server.Close)
	client, err := coreforge.New(server.URL, "token")
	core.RequireNoError(t, err)
	return client
}

func TestForgejo_New_Good(t *core.T) {
	source := New(Config{Repos: []string{"core/go-scm"}}, nil)
	core.AssertEqual(t, []string{"core/go-scm"}, source.repos)
	core.AssertNil(t, source.forge)
}

func TestForgejo_New_Bad(t *core.T) {
	source := New(Config{}, nil)
	core.AssertEmpty(t, source.repos)
	core.AssertNil(t, source.forge)
}

func TestForgejo_New_Ugly(t *core.T) {
	repos := []string{"core/go-scm"}
	source := New(Config{Repos: repos}, nil)
	repos[0] = "mutated/repo"
	core.AssertEqual(t, "core/go-scm", source.repos[0])
}

func TestForgejo_ForgejoSource_Name_Good(t *core.T) {
	source := New(Config{}, nil)
	got := source.Name()
	core.AssertEqual(t, "forgejo", got)
}

func TestForgejo_ForgejoSource_Name_Bad(t *core.T) {
	source := &ForgejoSource{}
	got := source.Name()
	core.AssertEqual(t, "forgejo", got)
}

func TestForgejo_ForgejoSource_Name_Ugly(t *core.T) {
	var source *ForgejoSource
	got := source.Name()
	core.AssertEqual(t, "forgejo", got)
}

func TestForgejo_ForgejoSource_Poll_Good(t *core.T) {
	source := New(Config{Repos: []string{"core/go-scm"}}, nil)
	signals, err := source.Poll(context.Background())
	core.AssertNoError(t, err)
	core.AssertNil(t, signals)
}

func TestForgejo_ForgejoSource_Poll_Bad(t *core.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	source := New(Config{Repos: []string{"core/go-scm"}}, nil)
	signals, err := source.Poll(ctx)
	core.AssertError(t, err)
	core.AssertNil(t, signals)
}

func TestForgejo_ForgejoSource_Poll_Ugly(t *core.T) {
	source := New(Config{Repos: []string{"broken-ref"}}, ax7ForgeClient(t))
	signals, err := source.Poll(context.Background())
	core.AssertError(t, err)
	core.AssertNil(t, signals)
}

func TestForgejo_ForgejoSource_Report_Good(t *core.T) {
	source := New(Config{}, nil)
	err := source.Report(context.Background(), &jobrunner.ActionResult{Action: "noop"})
	core.AssertNoError(t, err)
	core.AssertEmpty(t, source.repos)
}

func TestForgejo_ForgejoSource_Report_Bad(t *core.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	source := New(Config{}, nil)
	err := source.Report(ctx, &jobrunner.ActionResult{Action: "noop"})
	core.AssertError(t, err)
}

func TestForgejo_ForgejoSource_Report_Ugly(t *core.T) {
	source := New(Config{}, ax7ForgeClient(t))
	err := source.Report(context.Background(), &jobrunner.ActionResult{Action: "report", RepoOwner: "owner", RepoName: "repo", EpicNumber: 7})
	core.AssertError(t, err)
}
