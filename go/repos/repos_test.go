// SPDX-License-Identifier: EUPL-1.2

package repos

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	core "dappco.re/go"
	coreio "dappco.re/go/io"
)

func ax7ReposRegistry() *Registry {
	return &Registry{
		Version:  1,
		BasePath: "/work",
		Repos: map[string]*Repo{
			"api":  {Type: "service", DependsOn: []string{"core"}},
			"core": {Type: "library"},
		},
	}
}

func ax7ReposGitCommand(t *core.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(out))
	}
}

func ax7ReposGitRepo(t *core.T) string {
	root := t.TempDir()
	remote := filepath.Join(root, "remote.git")
	work := filepath.Join(root, "work")
	ax7ReposGitCommand(t, root, "init", "--bare", remote)
	ax7ReposGitCommand(t, root, "clone", remote, work)
	ax7ReposGitCommand(t, work, "config", "user.name", "AX7")
	ax7ReposGitCommand(t, work, "config", "user.email", "ax7@example.test")
	ax7ReposGitCommand(t, work, "checkout", "-b", "dev")
	core.RequireNoError(t, os.WriteFile(filepath.Join(work, "README.md"), []byte("ready\n"), 0o600))
	ax7ReposGitCommand(t, work, "add", "README.md")
	ax7ReposGitCommand(t, work, "commit", "-m", "initial")
	ax7ReposGitCommand(t, work, "push", "-u", "origin", "dev")
	return work
}

func TestRepos_Repo_Exists_Good(t *core.T) {
	repo := &Repo{Path: t.TempDir()}
	got := repo.Exists()
	core.AssertTrue(t, got)
}

func TestRepos_Repo_Exists_Bad(t *core.T) {
	repo := &Repo{Path: filepath.Join(t.TempDir(), "missing")}
	got := repo.Exists()
	core.AssertFalse(t, got)
}

func TestRepos_Repo_Exists_Ugly(t *core.T) {
	var repo *Repo
	got := repo.Exists()
	core.AssertFalse(t, got)
}

func TestRepos_Repo_IsGitRepo_Good(t *core.T) {
	dir := t.TempDir()
	core.RequireNoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0o755))
	repo := &Repo{Path: dir}
	got := repo.IsGitRepo()
	core.AssertTrue(t, got)
}

func TestRepos_Repo_IsGitRepo_Bad(t *core.T) {
	repo := &Repo{Path: t.TempDir()}
	got := repo.IsGitRepo()
	core.AssertFalse(t, got)
}

func TestRepos_Repo_IsGitRepo_Ugly(t *core.T) {
	var repo *Repo
	got := repo.IsGitRepo()
	core.AssertFalse(t, got)
}

func TestRepos_Registry_List_Good(t *core.T) {
	got := ax7ReposRegistry().List()
	core.AssertLen(t, got, 2)
	core.AssertEqual(t, "api", got[0].Name)
	core.AssertEqual(t, "/work/api", got[0].Path)
}

func TestRepos_Registry_List_Bad(t *core.T) {
	registry := &Registry{}
	got := registry.List()
	core.AssertNil(t, got)
}

func TestRepos_Registry_List_Ugly(t *core.T) {
	var registry *Registry
	got := registry.List()
	core.AssertNil(t, got)
}

func TestRepos_Registry_Get_Good(t *core.T) {
	repo, ok := ax7ReposRegistry().Get("api")
	core.AssertTrue(t, ok)
	core.AssertEqual(t, "api", repo.Name)
	core.AssertEqual(t, "/work/api", repo.Path)
}

func TestRepos_Registry_Get_Bad(t *core.T) {
	_, ok := ax7ReposRegistry().Get("missing")
	core.AssertFalse(
		t, ok,
	)
}

func TestRepos_Registry_Get_Ugly(t *core.T) {
	var registry *Registry
	_, ok := registry.Get("api")
	core.AssertFalse(t, ok)
}

func TestRepos_Registry_ByType_Good(t *core.T) {
	got := ax7ReposRegistry().ByType("SERVICE")
	core.AssertLen(t, got, 1)
	core.AssertEqual(t, "api", got[0].Name)
}

func TestRepos_Registry_ByType_Bad(t *core.T) {
	got := ax7ReposRegistry().ByType("missing")
	core.AssertEmpty(
		t, got,
	)
}

func TestRepos_Registry_ByType_Ugly(t *core.T) {
	var registry *Registry
	got := registry.ByType("service")
	core.AssertEmpty(t, got)
}

func TestRepos_Registry_TopologicalOrder_Good(t *core.T) {
	got, err := ax7ReposRegistry().TopologicalOrder()
	core.AssertNoError(t, err)
	core.AssertEqual(t, "core", got[0].Name)
	core.AssertEqual(t, "api", got[1].Name)
}

func TestRepos_Registry_TopologicalOrder_Bad(t *core.T) {
	registry := &Registry{Repos: map[string]*Repo{"a": {DependsOn: []string{"b"}}, "b": {DependsOn: []string{"a"}}}}
	_, err := registry.TopologicalOrder()
	core.AssertError(t, err)
}

func TestRepos_Registry_TopologicalOrder_Ugly(t *core.T) {
	registry := &Registry{}
	got, err := registry.TopologicalOrder()
	core.AssertNoError(t, err)
	core.AssertNil(t, got)
}

func TestRepos_LoadRegistry_Good(t *core.T) {
	medium := coreio.NewMockMedium()
	core.RequireNoError(t, medium.Write("repos.yaml", "version: 1\nbase_path: /work\nrepos:\n  core:\n    type: library\n"))
	registry, err := LoadRegistry(medium, "repos.yaml")
	core.AssertNoError(t, err)
	repo, ok := registry.Get("core")
	core.AssertTrue(t, ok)
	core.AssertEqual(t, "/work/core", repo.Path)
}

func TestRepos_LoadRegistry_Bad(t *core.T) {
	_, err := LoadRegistry(nil, "repos.yaml")
	core.AssertError(
		t, err,
	)
}

func TestRepos_LoadRegistry_Ugly(t *core.T) {
	medium := coreio.NewMockMedium()
	core.RequireNoError(t, medium.Write("repos.yaml", "repos: ["))
	_, err := LoadRegistry(medium, "repos.yaml")
	core.AssertError(t, err)
}

func TestRepos_FindRegistry_Good(t *core.T) {
	medium := coreio.NewMockMedium()
	core.RequireNoError(t, medium.Write("repos.yaml", "version: 1\nrepos: {}\n"))
	path, err := FindRegistry(medium)
	core.AssertNoError(t, err)
	core.AssertEqual(t, "repos.yaml", path)
}

func TestRepos_FindRegistry_Bad(t *core.T) {
	_, err := FindRegistry(nil)
	core.AssertError(
		t, err,
	)
}

func TestRepos_FindRegistry_Ugly(t *core.T) {
	_, err := FindRegistry(coreio.NewMockMedium())
	core.AssertError(
		t, err,
	)
}

func TestRepos_ScanDirectory_Good(t *core.T) {
	medium := coreio.NewMockMedium()
	core.RequireNoError(t, medium.EnsureDir("Code/core"))
	registry, err := ScanDirectory(medium, "Code")
	core.AssertNoError(t, err)
	repo, ok := registry.Get("core")
	core.AssertTrue(t, ok)
	core.AssertEqual(t, "Code/core", repo.Path)
}

func TestRepos_ScanDirectory_Bad(t *core.T) {
	_, err := ScanDirectory(nil, "Code")
	core.AssertError(
		t, err,
	)
}

func TestRepos_ScanDirectory_Ugly(t *core.T) {
	registry, err := ScanDirectory(coreio.NewMockMedium(), "Code")
	core.AssertNoError(t, err)
	core.AssertEmpty(t, registry.Repos)
}

func TestRepos_Registry_Save_Good(t *core.T) {
	medium := coreio.NewMockMedium()
	registry := ax7ReposRegistry()
	registry.medium = medium
	err := registry.Save("repos.yaml")
	core.AssertNoError(t, err)
	raw, readErr := medium.Read("repos.yaml")
	core.RequireNoError(t, readErr)
	core.AssertContains(t, raw, "repos:")
}

func TestRepos_Registry_Save_Bad(t *core.T) {
	var registry *Registry
	err := registry.Save("repos.yaml")
	core.AssertError(t, err)
}

func TestRepos_Registry_Save_Ugly(t *core.T) {
	registry := ax7ReposRegistry()
	path := filepath.Join(t.TempDir(), "repos.yaml")
	err := registry.Save(path)
	core.AssertNoError(t, err)
}

func TestRepos_Registry_SyncRepo_Good(t *core.T) {
	repoPath := ax7ReposGitRepo(t)
	registry := &Registry{Repos: map[string]*Repo{"demo": {Path: repoPath}}}
	err := registry.SyncRepo(context.Background(), "demo", "origin", "dev")
	core.AssertNoError(t, err)
}

func TestRepos_Registry_SyncRepo_Bad(t *core.T) {
	err := ax7ReposRegistry().SyncRepo(context.Background(), "missing", "origin", "dev")
	core.AssertError(
		t, err,
	)
}

func TestRepos_Registry_SyncRepo_Ugly(t *core.T) {
	var registry *Registry
	err := registry.SyncRepo(context.Background(), "demo", "origin", "dev")
	core.AssertError(t, err)
}

func TestRepos_Registry_SyncAll_Good(t *core.T) {
	repoPath := ax7ReposGitRepo(t)
	registry := &Registry{Repos: map[string]*Repo{"demo": {Path: repoPath}}}
	got := registry.SyncAll(context.Background(), "origin", "dev")
	core.AssertLen(t, got, 1)
	core.AssertTrue(t, got[0].Success)
}

func TestRepos_Registry_SyncAll_Bad(t *core.T) {
	got := (&Registry{}).SyncAll(context.Background(), "origin", "dev")
	core.AssertEmpty(
		t, got,
	)
}

func TestRepos_Registry_SyncAll_Ugly(t *core.T) {
	var registry *Registry
	got := registry.SyncAll(context.Background(), "origin", "dev")
	core.AssertNil(t, got)
}

func TestRepos_NewGitState_Good(t *core.T) {
	state := NewGitState()
	core.AssertEqual(t, 1, state.Version)
	core.AssertNotNil(t, state.Repos)
	core.AssertNotNil(t, state.Agents)
}

func TestRepos_NewGitState_Bad(t *core.T) {
	state := NewGitState()
	core.AssertEmpty(
		t, state.Repos,
	)
}

func TestRepos_NewGitState_Ugly(t *core.T) {
	state := NewGitState()
	state.ensure()
	core.AssertEqual(t, 1, state.Version)
}

func TestRepos_GitState_TouchPull_Good(t *core.T) {
	state := NewGitState()
	state.TouchPull("core")
	core.AssertFalse(t, state.Repos["core"].LastPull.IsZero())
}

func TestRepos_GitState_TouchPull_Bad(t *core.T) {
	state := &GitState{}
	state.TouchPull("")
	core.AssertFalse(t, state.Repos[""].LastPull.IsZero())
}

func TestRepos_GitState_TouchPull_Ugly(t *core.T) {
	var state *GitState
	core.AssertPanics(
		t, func() { state.TouchPull("core") },
	)
}

func TestRepos_GitState_TouchPush_Good(t *core.T) {
	state := NewGitState()
	state.TouchPush("core")
	core.AssertFalse(t, state.Repos["core"].LastPush.IsZero())
}

func TestRepos_GitState_TouchPush_Bad(t *core.T) {
	state := &GitState{}
	state.TouchPush("")
	core.AssertFalse(t, state.Repos[""].LastPush.IsZero())
}

func TestRepos_GitState_TouchPush_Ugly(t *core.T) {
	var state *GitState
	core.AssertPanics(
		t, func() { state.TouchPush("core") },
	)
}

func TestRepos_GitState_UpdateRepo_Good(t *core.T) {
	state := NewGitState()
	state.UpdateRepo("core", "dev", "origin", 1, 2)
	core.AssertEqual(t, "dev", state.Repos["core"].Branch)
	core.AssertEqual(t, 2, state.Repos["core"].Behind)
}

func TestRepos_GitState_UpdateRepo_Bad(t *core.T) {
	state := NewGitState()
	state.UpdateRepo("", "", "", 0, 0)
	core.AssertNotNil(t, state.Repos[""])
}

func TestRepos_GitState_UpdateRepo_Ugly(t *core.T) {
	var state *GitState
	core.AssertPanics(
		t, func() { state.UpdateRepo("core", "dev", "origin", 0, 0) },
	)
}

func TestRepos_GitState_Heartbeat_Good(t *core.T) {
	state := NewGitState()
	state.Heartbeat("codex", []string{"core"})
	core.AssertFalse(t, state.Agents["codex"].LastSeen.IsZero())
	core.AssertEqual(t, []string{"core"}, state.Agents["codex"].Active)
}

func TestRepos_GitState_Heartbeat_Bad(t *core.T) {
	state := NewGitState()
	state.Heartbeat("", nil)
	core.AssertNotNil(t, state.Agents[""])
}

func TestRepos_GitState_Heartbeat_Ugly(t *core.T) {
	var state *GitState
	core.AssertPanics(
		t, func() { state.Heartbeat("codex", nil) },
	)
}

func TestRepos_GitState_StaleAgents_Good(t *core.T) {
	state := NewGitState()
	state.Agents["old"] = &AgentState{LastSeen: time.Now().Add(-time.Hour)}
	got := state.StaleAgents(time.Minute)
	core.AssertEqual(t, []string{"old"}, got)
}

func TestRepos_GitState_StaleAgents_Bad(t *core.T) {
	state := NewGitState()
	state.Heartbeat("fresh", nil)
	got := state.StaleAgents(time.Hour)
	core.AssertEmpty(t, got)
}

func TestRepos_GitState_StaleAgents_Ugly(t *core.T) {
	var state *GitState
	got := state.StaleAgents(time.Minute)
	core.AssertNil(t, got)
}

func TestRepos_GitState_ActiveAgentsFor_Good(t *core.T) {
	state := NewGitState()
	state.Heartbeat("codex", []string{"core"})
	got := state.ActiveAgentsFor("core", time.Hour)
	core.AssertEqual(t, []string{"codex"}, got)
}

func TestRepos_GitState_ActiveAgentsFor_Bad(t *core.T) {
	state := NewGitState()
	state.Heartbeat("codex", []string{"other"})
	got := state.ActiveAgentsFor("core", time.Hour)
	core.AssertEmpty(t, got)
}

func TestRepos_GitState_ActiveAgentsFor_Ugly(t *core.T) {
	var state *GitState
	got := state.ActiveAgentsFor("core", time.Hour)
	core.AssertNil(t, got)
}

func TestRepos_GitState_NeedsPull_Good(t *core.T) {
	state := NewGitState()
	state.Repos["core"] = &RepoGitState{LastPull: time.Now().Add(-time.Hour)}
	got := state.NeedsPull("core", time.Minute)
	core.AssertTrue(t, got)
}

func TestRepos_GitState_NeedsPull_Bad(t *core.T) {
	state := NewGitState()
	state.TouchPull("core")
	got := state.NeedsPull("core", time.Hour)
	core.AssertFalse(t, got)
}

func TestRepos_GitState_NeedsPull_Ugly(t *core.T) {
	var state *GitState
	got := state.NeedsPull("core", time.Hour)
	core.AssertTrue(t, got)
}

func TestRepos_LoadGitState_Good(t *core.T) {
	medium := coreio.NewMockMedium()
	core.RequireNoError(t, medium.Write("root/.core/git.yaml", "version: 1\nrepos:\n  core:\n    branch: dev\n"))
	state, err := LoadGitState(medium, "root")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "dev", state.Repos["core"].Branch)
}

func TestRepos_LoadGitState_Bad(t *core.T) {
	_, err := LoadGitState(nil, "root")
	core.AssertError(
		t, err,
	)
}

func TestRepos_LoadGitState_Ugly(t *core.T) {
	state, err := LoadGitState(coreio.NewMockMedium(), "root")
	core.AssertNoError(t, err)
	core.AssertEqual(t, 1, state.Version)
}

func TestRepos_SaveGitState_Good(t *core.T) {
	medium := coreio.NewMockMedium()
	state := NewGitState()
	state.UpdateRepo("core", "dev", "origin", 0, 0)
	err := SaveGitState(medium, "root", state)
	core.AssertNoError(t, err)
	raw, readErr := medium.Read("root/.core/git.yaml")
	core.RequireNoError(t, readErr)
	core.AssertContains(t, raw, "core")
}

func TestRepos_SaveGitState_Bad(t *core.T) {
	err := SaveGitState(nil, "root", NewGitState())
	core.AssertError(
		t, err,
	)
}

func TestRepos_SaveGitState_Ugly(t *core.T) {
	err := SaveGitState(coreio.NewMockMedium(), "root", nil)
	core.AssertNoError(
		t, err,
	)
}

func TestRepos_DefaultWorkConfig_Good(t *core.T) {
	cfg := DefaultWorkConfig()
	core.AssertEqual(t, 1, cfg.Version)
	core.AssertTrue(t, cfg.Sync.AutoPull)
}

func TestRepos_DefaultWorkConfig_Bad(t *core.T) {
	cfg := DefaultWorkConfig()
	core.AssertEqual(
		t, time.Minute, cfg.Sync.Interval,
	)
}

func TestRepos_DefaultWorkConfig_Ugly(t *core.T) {
	cfg := DefaultWorkConfig()
	core.AssertTrue(
		t, cfg.Agents.WarnOnOverlap,
	)
}

func TestRepos_WorkConfig_HasTrigger_Good(t *core.T) {
	cfg := &WorkConfig{Triggers: []string{"push"}}
	core.AssertTrue(
		t, cfg.HasTrigger("PUSH"),
	)
}

func TestRepos_WorkConfig_HasTrigger_Bad(t *core.T) {
	cfg := &WorkConfig{Triggers: []string{"push"}}
	core.AssertFalse(
		t, cfg.HasTrigger("pull"),
	)
}

func TestRepos_WorkConfig_HasTrigger_Ugly(t *core.T) {
	var cfg *WorkConfig
	core.AssertFalse(
		t, cfg.HasTrigger("push"),
	)
}

func TestRepos_LoadWorkConfig_Good(t *core.T) {
	medium := coreio.NewMockMedium()
	core.RequireNoError(t, medium.Write("root/.core/work.yaml", "version: 2\ntriggers: [push]\n"))
	cfg, err := LoadWorkConfig(medium, "root")
	core.AssertNoError(t, err)
	core.AssertEqual(t, 2, cfg.Version)
	core.AssertTrue(t, cfg.HasTrigger("push"))
}

func TestRepos_LoadWorkConfig_Bad(t *core.T) {
	_, err := LoadWorkConfig(nil, "root")
	core.AssertError(
		t, err,
	)
}

func TestRepos_LoadWorkConfig_Ugly(t *core.T) {
	cfg, err := LoadWorkConfig(coreio.NewMockMedium(), "root")
	core.AssertNoError(t, err)
	core.AssertEqual(t, 1, cfg.Version)
}

func TestRepos_SaveWorkConfig_Good(t *core.T) {
	medium := coreio.NewMockMedium()
	err := SaveWorkConfig(medium, "root", &WorkConfig{Version: 2})
	core.AssertNoError(t, err)
	raw, readErr := medium.Read("root/.core/work.yaml")
	core.RequireNoError(t, readErr)
	core.AssertContains(t, raw, "version: 2")
}

func TestRepos_SaveWorkConfig_Bad(t *core.T) {
	err := SaveWorkConfig(nil, "root", DefaultWorkConfig())
	core.AssertError(
		t, err,
	)
}

func TestRepos_SaveWorkConfig_Ugly(t *core.T) {
	err := SaveWorkConfig(coreio.NewMockMedium(), "root", nil)
	core.AssertNoError(
		t, err,
	)
}

func TestRepos_DefaultKBConfig_Good(t *core.T) {
	cfg := DefaultKBConfig()
	core.AssertEqual(t, 1, cfg.Version)
	core.AssertEqual(t, ".core/wiki", cfg.Wiki.Dir)
}

func TestRepos_DefaultKBConfig_Bad(t *core.T) {
	cfg := DefaultKBConfig()
	core.AssertFalse(
		t, cfg.Wiki.Enabled,
	)
}

func TestRepos_DefaultKBConfig_Ugly(t *core.T) {
	cfg := DefaultKBConfig()
	core.AssertEqual(
		t, "", cfg.Wiki.Remote,
	)
}

func TestRepos_KBConfig_WikiRepoURL_Good(t *core.T) {
	cfg := &KBConfig{Wiki: WikiConfig{Remote: "https://forge.example/core"}}
	got := cfg.WikiRepoURL("go-scm")
	core.AssertEqual(t, "https://forge.example/core/go-scm.wiki.git", got)
}

func TestRepos_KBConfig_WikiRepoURL_Bad(t *core.T) {
	cfg := &KBConfig{}
	got := cfg.WikiRepoURL("go-scm")
	core.AssertEqual(t, "", got)
}

func TestRepos_KBConfig_WikiRepoURL_Ugly(t *core.T) {
	var cfg *KBConfig
	got := cfg.WikiRepoURL("go-scm")
	core.AssertEqual(t, "", got)
}

func TestRepos_KBConfig_WikiLocalPath_Good(t *core.T) {
	cfg := &KBConfig{Wiki: WikiConfig{Dir: "wiki"}}
	got := cfg.WikiLocalPath("/root", "go-scm")
	core.AssertEqual(t, filepath.Join("/root", "wiki", "go-scm"), got)
}

func TestRepos_KBConfig_WikiLocalPath_Bad(t *core.T) {
	cfg := &KBConfig{}
	got := cfg.WikiLocalPath("/root", "go-scm")
	core.AssertEqual(t, filepath.Join("/root", ".core/wiki", "go-scm"), got)
}

func TestRepos_KBConfig_WikiLocalPath_Ugly(t *core.T) {
	var cfg *KBConfig
	got := cfg.WikiLocalPath("", "")
	core.AssertEqual(t, filepath.Join(".core/wiki"), got)
}

func TestRepos_LoadKBConfig_Good(t *core.T) {
	medium := coreio.NewMockMedium()
	core.RequireNoError(t, medium.Write("root/.core/kb.yaml", "version: 2\nwiki:\n  dir: wiki\n"))
	cfg, err := LoadKBConfig(medium, "root")
	core.AssertNoError(t, err)
	core.AssertEqual(t, 2, cfg.Version)
	core.AssertEqual(t, "wiki", cfg.Wiki.Dir)
}

func TestRepos_LoadKBConfig_Bad(t *core.T) {
	_, err := LoadKBConfig(nil, "root")
	core.AssertError(
		t, err,
	)
}

func TestRepos_LoadKBConfig_Ugly(t *core.T) {
	cfg, err := LoadKBConfig(coreio.NewMockMedium(), "root")
	core.AssertNoError(t, err)
	core.AssertEqual(t, 1, cfg.Version)
}

func TestRepos_SaveKBConfig_Good(t *core.T) {
	medium := coreio.NewMockMedium()
	err := SaveKBConfig(medium, "root", &KBConfig{Version: 2})
	core.AssertNoError(t, err)
	raw, readErr := medium.Read("root/.core/kb.yaml")
	core.RequireNoError(t, readErr)
	core.AssertContains(t, raw, "version: 2")
}

func TestRepos_SaveKBConfig_Bad(t *core.T) {
	err := SaveKBConfig(nil, "root", DefaultKBConfig())
	core.AssertError(
		t, err,
	)
}

func TestRepos_SaveKBConfig_Ugly(t *core.T) {
	err := SaveKBConfig(coreio.NewMockMedium(), "root", nil)
	core.AssertNoError(
		t, err,
	)
}

func TestRepos_NewService_Good(t *core.T) {
	result := NewService(ServiceOptions{Root: t.TempDir()})(core.New())
	core.AssertTrue(t, result.OK)
	core.AssertNotNil(t, result.Value)
}

func TestRepos_NewService_Bad(t *core.T) {
	result := NewService(ServiceOptions{})(nil)
	core.AssertTrue(t, result.OK)
	core.AssertNotNil(t, result.Value)
}

func TestRepos_NewService_Ugly(t *core.T) {
	result := NewService(ServiceOptions{Remote: "upstream", Branch: "main"})(core.New())
	service := result.Value.(*Service)
	core.AssertEqual(t, "upstream", service.Options().Remote)
	core.AssertEqual(t, "main", service.Options().Branch)
}

func TestRepos_Service_OnStartup_Good(t *core.T) {
	root := t.TempDir()
	core.RequireNoError(t, os.MkdirAll(filepath.Join(root, ".core"), 0o755))
	core.RequireNoError(t, os.WriteFile(filepath.Join(root, ".core", "repos.yaml"), []byte("version: 1\nrepos: {}\n"), 0o600))
	c := core.New(core.WithService(NewService(ServiceOptions{Root: root})))
	result := c.ServiceStartup(context.Background(), nil)
	core.AssertTrue(t, result.OK)
	core.AssertTrue(t, c.Action("repo.sync").Exists())
}

func TestRepos_Service_OnStartup_Bad(t *core.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	service := &Service{}
	result := service.OnStartup(ctx)
	core.AssertFalse(t, result.OK)
	core.AssertErrorIs(t, result.Value.(error), context.Canceled)
}

func TestRepos_Service_OnStartup_Ugly(t *core.T) {
	var service *Service
	result := service.OnStartup(context.Background())
	core.AssertTrue(t, result.OK)
}

func TestRepos_Service_HandleIPCEvents_Good(t *core.T) {
	service := &Service{}
	result := service.HandleIPCEvents(core.New(), "ignored")
	core.AssertTrue(t, result.OK)
}

func TestRepos_Service_HandleIPCEvents_Bad(t *core.T) {
	service := &Service{}
	result := service.HandleIPCEvents(core.New(), (*WorkspacePushed)(nil))
	core.AssertTrue(t, result.OK)
}

func TestRepos_Service_HandleIPCEvents_Ugly(t *core.T) {
	var service *Service
	result := service.HandleIPCEvents(core.New(), WorkspacePushed{Repo: "demo"})
	core.AssertTrue(t, result.OK)
}
