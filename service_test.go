// SPDX-License-Identifier: EUPL-1.2

package scm

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"dappco.re/go/core"
	"dappco.re/go/core/io"
	scmgit "dappco.re/go/core/scm/git"
	"dappco.re/go/core/scm/repos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	exec "golang.org/x/sys/execabs"
)

func TestCoreService_OnStartup_RegistersSyncActions_Good(t *testing.T) {
	m := io.NewMockMedium()
	require.NoError(t, m.Write("/tmp/repos.yaml", `
version: 1
org: core
base_path: /tmp/code
repos:
  go-scm:
    type: module
    description: SCM integration
`))

	c := core.New()
	factory := NewCoreService(ServiceOptions{
		Medium:       m,
		RegistryPath: "/tmp/repos.yaml",
	})

	svcAny, err := factory(c)
	require.NoError(t, err)
	svc := svcAny.(*CoreService)

	result := svc.OnStartup(context.Background())
	require.True(t, result.OK)
	assert.True(t, c.Action("repo.sync").Exists())
	assert.True(t, c.Action("repo.sync.all").Exists())
}

func TestCoreService_HandleIPCEvents_Good_PointerWorkspacePushed_Good(t *testing.T) {
	c := core.New()

	called := false
	c.Action("repo.sync", func(ctx context.Context, opts core.Options) core.Result {
		called = true
		assert.Equal(t, "core", opts.String("org"))
		assert.Equal(t, "go-scm", opts.String("repo"))
		assert.Equal(t, "dev", opts.String("branch"))
		return core.Result{OK: true}
	})

	svc := &CoreService{ServiceRuntime: core.NewServiceRuntime(c, ServiceOptions{})}
	result := svc.HandleIPCEvents(c, &WorkspacePushed{
		Org:    "core",
		Repo:   "go-scm",
		Branch: "dev",
		Root:   "/tmp/code",
	})

	require.True(t, result.OK)
	assert.True(t, called)
}

func TestCoreService_HandleIPCEvents_Good_WorkspacePushed_InvalidatesRegistryCache_Good(t *testing.T) {
	m := io.NewMockMedium()
	require.NoError(t, m.Write("/tmp/repos.yaml", `
version: 1
org: core
base_path: /tmp/code
repos:
  go-scm:
    type: module
`))

	c := core.New()
	factory := NewCoreService(ServiceOptions{
		Medium:       m,
		RegistryPath: "/tmp/repos.yaml",
	})

	svcAny, err := factory(c)
	require.NoError(t, err)
	svc := svcAny.(*CoreService)

	// Prime the cache with the initial registry contents.
	regs, err := svc.loadRegistries()
	require.NoError(t, err)
	require.Len(t, regs, 1)

	require.NoError(t, m.Write("/tmp/repos.yaml", `
version: 1
org: core
base_path: /tmp/code
repos:
  go-scm:
    type: module
  go-io:
    type: module
`))

	c.Action("repo.sync.all", func(ctx context.Context, opts core.Options) core.Result {
		regs, err := svc.loadRegistries()
		if err != nil {
			return core.Result{Value: err, OK: false}
		}
		merged := repos.MergeRegistries(regs...)
		return core.Result{Value: len(merged.List()), OK: true}
	})

	result := svc.HandleIPCEvents(c, &WorkspacePushed{
		Org:    "core",
		Repo:   "",
		Branch: "dev",
		Root:   "/tmp/code",
	})

	require.True(t, result.OK)
	count, ok := result.Value.(int)
	require.True(t, ok)
	assert.Equal(t, 2, count)
}

func TestCoreService_ResolveRepo_FallsBackToWorkspaceRoot_Good(t *testing.T) {
	m := io.NewMockMedium()
	require.NoError(t, m.Write("/tmp/empty-repos.yaml", `
version: 1
org: core
base_path: /tmp/code
repos: {}
`))

	c := core.New()
	factory := NewCoreService(ServiceOptions{
		Medium:        m,
		RegistryPath:  "/tmp/empty-repos.yaml",
		WorkspaceRoot: "/tmp/code",
	})

	svcAny, err := factory(c)
	require.NoError(t, err)
	svc := svcAny.(*CoreService)

	repo, reg, path, err := svc.resolveRepo("go-scm", "core", "")
	require.NoError(t, err)
	assert.Nil(t, reg)
	require.NotNil(t, repo)
	assert.Equal(t, "go-scm", repo.Name)
	assert.Equal(t, "/tmp/code/core/go-scm", path)
}

func TestCoreService_LoadRegistries_Good_ScansWorkspaceRoot_WhenNoRegistry_Good(t *testing.T) {
	root := t.TempDir()
	repoDir := filepath.Join(root, "go-scm")
	require.NoError(t, os.MkdirAll(filepath.Join(repoDir, ".git"), 0755))
	t.Setenv("CORE_REPOS", "")
	t.Setenv("HOME", root)

	c := core.New()
	factory := NewCoreService(ServiceOptions{
		Medium:        io.Local,
		WorkspaceRoot: root,
	})

	svcAny, err := factory(c)
	require.NoError(t, err)
	svc := svcAny.(*CoreService)

	regs, err := svc.loadRegistries()
	require.NoError(t, err)
	require.Len(t, regs, 1)

	repo, ok := regs[0].Get("go-scm")
	require.True(t, ok)
	assert.Equal(t, "go-scm", filepath.Base(repo.Path))
}

func TestRepoBranch_Good_PrefersRepoBranch_Good(t *testing.T) {
	repo := &repos.Repo{Branch: "dev"}
	reg := &repos.Registry{
		Defaults: repos.RegistryDefaults{Branch: "main"},
	}

	assert.Equal(t, "dev", repoBranch(repo, reg, ""))
	assert.Equal(t, "dev", repoBranch(repo, reg, "fallback"))
}

func TestRepoBranch_Good_UsesRegistryDefaultBeforeFallback_Good(t *testing.T) {
	repo := &repos.Repo{}
	reg := &repos.Registry{
		Defaults: repos.RegistryDefaults{Branch: "dev"},
	}

	assert.Equal(t, "dev", repoBranch(repo, reg, "fallback"))
}

func TestRepoBranch_Good_UsesFallbackWhenUnset_Good(t *testing.T) {
	assert.Equal(t, "fallback", repoBranch(&repos.Repo{}, &repos.Registry{}, "fallback"))
}

func TestCoreService_ResolveRepo_Good_RespectsRequestedOrg_Good(t *testing.T) {
	m := io.NewMockMedium()
	require.NoError(t, m.Write("/tmp/alpha.yaml", `
version: 1
org: alpha
base_path: /tmp/alpha
repos:
  go-scm:
    type: module
`))
	require.NoError(t, m.Write("/tmp/beta.yaml", `
version: 1
org: beta
base_path: /tmp/beta
repos:
  go-scm:
    type: module
`))

	alpha, err := repos.LoadRegistry(m, "/tmp/alpha.yaml")
	require.NoError(t, err)
	beta, err := repos.LoadRegistry(m, "/tmp/beta.yaml")
	require.NoError(t, err)

	c := core.New()
	factory := NewCoreService(ServiceOptions{
		Medium:        m,
		WorkspaceRoot: "/tmp/code",
	})

	svcAny, err := factory(c)
	require.NoError(t, err)
	svc := svcAny.(*CoreService)
	svc.registries = []*repos.Registry{alpha, beta}

	repo, reg, path, err := svc.resolveRepo("go-scm", "beta", "")
	require.NoError(t, err)
	require.NotNil(t, repo)
	require.NotNil(t, reg)
	assert.Equal(t, "beta", reg.Org)
	assert.Equal(t, "/tmp/beta/go-scm", path)
}

func TestCoreService_HandleRepoSyncAll_Good_UsesRegistryDefaultBranch_Good(t *testing.T) {
	root := t.TempDir()
	remote := filepath.Join(root, "remote.git")
	source := filepath.Join(root, "source")
	clone := filepath.Join(root, "go-scm")
	regPath := filepath.Join(root, "repos.yaml")

	require.NoError(t, os.MkdirAll(remote, 0755))
	runGitCommand(t, remote, "init", "--bare")

	require.NoError(t, os.MkdirAll(source, 0755))
	runGitCommand(t, source, "init")
	runGitCommand(t, source, "checkout", "-b", "dev")
	require.NoError(t, os.WriteFile(filepath.Join(source, "README.md"), []byte("hello\n"), 0644))
	runGitCommand(t, source, "add", "README.md")
	runGitCommand(t, source, "commit", "-m", "initial")
	runGitCommand(t, source, "remote", "add", "origin", remote)
	runGitCommand(t, source, "push", "-u", "origin", "dev")

	require.NoError(t, scmgit.Clone(context.Background(), remote, clone, "dev"))

	require.NoError(t, os.WriteFile(regPath, []byte(`
version: 1
org: core
base_path: `+root+`
defaults:
  branch: dev
repos:
  go-scm:
    type: module
`), 0644))

	c := core.New()
	factory := NewCoreService(ServiceOptions{
		Medium:       io.Local,
		RegistryPath: regPath,
	})

	svcAny, err := factory(c)
	require.NoError(t, err)
	svc := svcAny.(*CoreService)

	result := svc.handleRepoSyncAll(context.Background(), core.NewOptions())
	require.True(t, result.OK)

	value, ok := result.Value.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, 1, value["synced"])
}

func runGitCommand(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmdArgs := append([]string{"-C", dir, "-c", "user.email=test@test.com", "-c", "user.name=test"}, args...)
	cmd := exec.Command("git", cmdArgs...)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v: %s", args, string(out))
}
