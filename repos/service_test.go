// SPDX-License-Identifier: EUPL-1.2

package repos

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	core "dappco.re/go/core"
	"gopkg.in/yaml.v3"
)

func TestServiceRegistersRepoSyncActions(t *testing.T) {
	root := t.TempDir()
	repoPath := filepath.Join(root, "repo1")
	remotePath := filepath.Join(root, "remote.git")
	filePath := filepath.Join(repoPath, "state.txt")

	runGitCmd(t, root, "git", "init", "--bare", remotePath)
	if err := os.MkdirAll(filepath.Dir(repoPath), 0o755); err != nil {
		t.Fatalf("mkdir workspace parent: %v", err)
	}
	runGitCmd(t, root, "git", "clone", remotePath, repoPath)
	runGitCmd(t, repoPath, "git", "-C", repoPath, "config", "user.name", "Test User")
	runGitCmd(t, repoPath, "git", "-C", repoPath, "config", "user.email", "test@example.com")
	runGitCmd(t, repoPath, "git", "-C", repoPath, "checkout", "-b", "dev")

	if err := os.WriteFile(filePath, []byte("remote state\n"), 0o600); err != nil {
		t.Fatalf("write seed file: %v", err)
	}
	runGitCmd(t, repoPath, "git", "-C", repoPath, "add", "state.txt")
	runGitCmd(t, repoPath, "git", "-C", repoPath, "commit", "-m", "initial")
	runGitCmd(t, repoPath, "git", "-C", repoPath, "push", "-u", "origin", "dev")

	registry := &Registry{
		Version:  1,
		BasePath: root,
		Repos: map[string]*Repo{
			"repo1": {Path: repoPath},
		},
	}
	if err := writeRegistry(root, registry); err != nil {
		t.Fatalf("write registry: %v", err)
	}

	svc := NewService(ServiceOptions{Root: root, Branch: "dev", Remote: "origin"})
	c := core.New(core.WithService(svc))
	if r := c.ServiceStartup(context.Background(), nil); !r.OK {
		t.Fatalf("service startup failed: %v", r.Value)
	}

	if !c.Action("repo.sync").Exists() {
		t.Fatalf("repo.sync action was not registered")
	}
	if !c.Action("repo.sync.all").Exists() {
		t.Fatalf("repo.sync.all action was not registered")
	}

	if err := os.WriteFile(filePath, []byte("local changes\n"), 0o600); err != nil {
		t.Fatalf("write local change: %v", err)
	}
	runGitCmd(t, repoPath, "git", "-C", repoPath, "commit", "-am", "local change")

	syncResult := c.Action("repo.sync").Run(context.Background(), core.NewOptions(core.Option{Key: "repo", Value: "repo1"}))
	if !syncResult.OK {
		t.Fatalf("repo.sync failed: %v", syncResult.Value)
	}

	raw, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read synced file: %v", err)
	}
	if got := string(raw); got != "remote state\n" {
		t.Fatalf("unexpected synced file contents: %q", got)
	}

	if err := os.WriteFile(filePath, []byte("local changes again\n"), 0o600); err != nil {
		t.Fatalf("write local change for ipc: %v", err)
	}
	runGitCmd(t, repoPath, "git", "-C", repoPath, "commit", "-am", "local change again")

	if r := c.ACTION(WorkspacePushed{Repo: "repo1"}); !r.OK {
		t.Fatalf("workspace push broadcast failed: %v", r.Value)
	}

	raw, err = os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read ipc-synced file: %v", err)
	}
	if got := string(raw); got != "remote state\n" {
		t.Fatalf("unexpected ipc synced file contents: %q", got)
	}
}

func TestServiceLoadsProjectAndHomeRegistries(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)

	projectRepo := filepath.Join(root, "project-repo")
	homeRepo := filepath.Join(home, "home-repo")

	projectRegistry := &Registry{
		Version:  1,
		BasePath: root,
		Repos: map[string]*Repo{
			"project": {Path: projectRepo},
		},
	}
	homeRegistry := &Registry{
		Version:  1,
		BasePath: home,
		Repos: map[string]*Repo{
			"shared": {Path: homeRepo},
		},
	}

	if err := writeRegistry(root, projectRegistry); err != nil {
		t.Fatalf("write project registry: %v", err)
	}
	if err := writeRegistry(home, homeRegistry); err != nil {
		t.Fatalf("write home registry: %v", err)
	}

	c := core.New(core.WithService(NewService(ServiceOptions{Root: root})))
	if r := c.ServiceStartup(context.Background(), nil); !r.OK {
		t.Fatalf("service startup failed: %v", r.Value)
	}
	svcRes := c.Service("repos")
	if !svcRes.OK {
		t.Fatalf("repos service was not registered")
	}
	svc, ok := svcRes.Value.(*Service)
	if !ok {
		t.Fatalf("unexpected service type: %T", svcRes.Value)
	}
	reg, err := svc.loadRegistryAt(root)
	if err != nil {
		t.Fatalf("load merged registry: %v", err)
	}

	if _, ok := reg.Get("project"); !ok {
		t.Fatalf("expected project registry entry to be loaded")
	}
	if _, ok := reg.Get("shared"); !ok {
		t.Fatalf("expected home registry entry to be loaded")
	}
	if got := reg.Repos["project"].Path; got != projectRepo {
		t.Fatalf("unexpected project repo path: %q", got)
	}
	if got := reg.Repos["shared"].Path; got != homeRepo {
		t.Fatalf("unexpected home repo path: %q", got)
	}
}

func TestServiceSyncRepoFallsBackToWorkspacePath(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CORE_REPOS", "")
	org := "core"
	repoName := "repo1"
	repoPath := filepath.Join(root, org, repoName)
	remotePath := filepath.Join(root, "remote.git")
	filePath := filepath.Join(repoPath, "state.txt")

	runGitCmd(t, root, "git", "init", "--bare", remotePath)
	runGitCmd(t, root, "git", "clone", remotePath, repoPath)
	runGitCmd(t, repoPath, "git", "-C", repoPath, "config", "user.name", "Test User")
	runGitCmd(t, repoPath, "git", "-C", repoPath, "config", "user.email", "test@example.com")
	runGitCmd(t, repoPath, "git", "-C", repoPath, "checkout", "-b", "dev")

	if err := os.WriteFile(filePath, []byte("remote state\n"), 0o600); err != nil {
		t.Fatalf("write seed file: %v", err)
	}
	runGitCmd(t, repoPath, "git", "-C", repoPath, "add", "state.txt")
	runGitCmd(t, repoPath, "git", "-C", repoPath, "commit", "-m", "initial")
	runGitCmd(t, repoPath, "git", "-C", repoPath, "push", "-u", "origin", "dev")

	svc := &Service{ServiceRuntime: core.NewServiceRuntime(core.New(), ServiceOptions{
		Root:         root,
		RegistryPath: filepath.Join(root, ".core", "missing.yaml"),
	})}

	if err := os.WriteFile(filePath, []byte("local changes\n"), 0o600); err != nil {
		t.Fatalf("write local change: %v", err)
	}
	runGitCmd(t, repoPath, "git", "-C", repoPath, "commit", "-am", "local change")

	result, err := svc.syncRepo(context.Background(), core.NewOptions(
		core.Option{Key: "root", Value: root},
		core.Option{Key: "org", Value: org},
		core.Option{Key: "repo", Value: repoName},
		core.Option{Key: "remote", Value: "origin"},
		core.Option{Key: "branch", Value: "dev"},
	))
	if err != nil {
		t.Fatalf("sync repo fallback: %v", err)
	}
	if result == nil || !result.Success {
		t.Fatalf("unexpected sync result: %#v", result)
	}

	raw, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read synced file: %v", err)
	}
	if got := string(raw); got != "remote state\n" {
		t.Fatalf("unexpected synced file contents: %q", got)
	}
}

func TestServiceSyncRepoFallsBackWhenRegistryMissesRepo(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	org := "core"
	repoName := "repo1"
	repoPath := filepath.Join(root, org, repoName)
	remotePath := filepath.Join(root, "remote.git")
	filePath := filepath.Join(repoPath, "state.txt")

	runGitCmd(t, root, "git", "init", "--bare", remotePath)
	runGitCmd(t, root, "git", "clone", remotePath, repoPath)
	runGitCmd(t, repoPath, "git", "-C", repoPath, "config", "user.name", "Test User")
	runGitCmd(t, repoPath, "git", "-C", repoPath, "config", "user.email", "test@example.com")
	runGitCmd(t, repoPath, "git", "-C", repoPath, "checkout", "-b", "dev")

	if err := os.WriteFile(filePath, []byte("remote state\n"), 0o600); err != nil {
		t.Fatalf("write seed file: %v", err)
	}
	runGitCmd(t, repoPath, "git", "-C", repoPath, "add", "state.txt")
	runGitCmd(t, repoPath, "git", "-C", repoPath, "commit", "-m", "initial")
	runGitCmd(t, repoPath, "git", "-C", repoPath, "push", "-u", "origin", "dev")

	registry := &Registry{
		Version:  1,
		BasePath: root,
		Repos: map[string]*Repo{
			"other": {Path: filepath.Join(root, org, "other")},
		},
	}
	if err := writeRegistry(root, registry); err != nil {
		t.Fatalf("write registry: %v", err)
	}

	svc := &Service{ServiceRuntime: core.NewServiceRuntime(core.New(), ServiceOptions{
		Root:   root,
		Branch: "dev",
		Remote: "origin",
	})}

	if err := os.WriteFile(filePath, []byte("local changes\n"), 0o600); err != nil {
		t.Fatalf("write local change: %v", err)
	}
	runGitCmd(t, repoPath, "git", "-C", repoPath, "commit", "-am", "local change")

	result, err := svc.syncRepo(context.Background(), core.NewOptions(
		core.Option{Key: "root", Value: root},
		core.Option{Key: "org", Value: org},
		core.Option{Key: "repo", Value: repoName},
		core.Option{Key: "remote", Value: "origin"},
		core.Option{Key: "branch", Value: "dev"},
	))
	if err != nil {
		t.Fatalf("sync repo fallback: %v", err)
	}
	if result == nil || !result.Success {
		t.Fatalf("unexpected sync result: %#v", result)
	}
	if got := result.Path; got != repoPath {
		t.Fatalf("unexpected fallback path: %q", got)
	}

	raw, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read synced file: %v", err)
	}
	if got := string(raw); got != "remote state\n" {
		t.Fatalf("unexpected synced file contents: %q", got)
	}
}

func writeRegistry(root string, reg *Registry) error {
	raw, err := yaml.Marshal(reg)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(root, ".core"), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, ".core", "repos.yaml"), raw, 0o600)
}

func runGitCmd(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, string(out))
	}
}
