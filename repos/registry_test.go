// SPDX-License-Identifier: EUPL-1.2

package repos

import (
	"os"
	"path/filepath"
	"testing"

	"dappco.re/go/core/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── LoadRegistry ───────────────────────────────────────────────────

func TestLoadRegistry_Good(t *testing.T) {
	m := io.NewMockMedium()
	yaml := `
version: 1
org: host-uk
base_path: /tmp/repos
repos:
  core:
    type: foundation
    description: Core package
`
	_ = m.Write("/tmp/repos.yaml", yaml)

	reg, err := LoadRegistry(m, "/tmp/repos.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, reg)
	assert.Equal(t, "host-uk", reg.Org)
	assert.Equal(t, "/tmp/repos", reg.BasePath)
	assert.Equal(t, m, reg.medium)

	repo, ok := reg.Get("core")
	assert.True(t, ok)
	assert.Equal(t, "core", repo.Name)
	assert.Equal(t, "/tmp/repos/core", repo.Path)
	assert.Equal(t, reg, repo.registry)
}

func TestLoadRegistry_Good_WithDefaults_Good(t *testing.T) {
	m := io.NewMockMedium()
	yaml := `
version: 1
org: host-uk
base_path: /tmp/repos
defaults:
  ci: github-actions
  license: EUPL-1.2
  branch: main
repos:
  core-php:
    type: foundation
    description: Foundation
  core-admin:
    type: module
    description: Admin panel
`
	_ = m.Write("/tmp/repos.yaml", yaml)

	reg, err := LoadRegistry(m, "/tmp/repos.yaml")
	require.NoError(t, err)

	php, ok := reg.Get("core-php")
	require.True(t, ok)
	assert.Equal(t, "github-actions", php.CI)

	admin, ok := reg.Get("core-admin")
	require.True(t, ok)
	assert.Equal(t, "github-actions", admin.CI)
}

func TestLoadRegistry_Good_WithDefaults_Licence_Good(t *testing.T) {
	m := io.NewMockMedium()
	yaml := `
version: 1
org: host-uk
base_path: /tmp/repos
defaults:
  ci: github-actions
  licence: EUPL-1.2
  branch: main
repos:
  core-php:
    type: foundation
`
	_ = m.Write("/tmp/repos.yaml", yaml)

	reg, err := LoadRegistry(m, "/tmp/repos.yaml")
	require.NoError(t, err)

	repo, ok := reg.Get("core-php")
	require.True(t, ok)
	assert.Equal(t, "github-actions", repo.CI)
	assert.Equal(t, "main", repo.Branch)
	assert.Equal(t, "EUPL-1.2", reg.Defaults.License)
}

func TestLoadRegistry_Good_CustomRepoPath_Good(t *testing.T) {
	m := io.NewMockMedium()
	yaml := `
version: 1
org: host-uk
base_path: /tmp/repos
repos:
  special:
    type: module
    path: /opt/special-repo
`
	_ = m.Write("/tmp/repos.yaml", yaml)

	reg, err := LoadRegistry(m, "/tmp/repos.yaml")
	require.NoError(t, err)

	repo, ok := reg.Get("special")
	require.True(t, ok)
	assert.Equal(t, "/opt/special-repo", repo.Path)
}

func TestLoadRegistry_Good_CIOverride_Good(t *testing.T) {
	m := io.NewMockMedium()
	yaml := `
version: 1
org: test
base_path: /tmp
defaults:
  ci: default-ci
repos:
  a:
    type: module
  b:
    type: module
    ci: custom-ci
`
	_ = m.Write("/tmp/repos.yaml", yaml)

	reg, err := LoadRegistry(m, "/tmp/repos.yaml")
	require.NoError(t, err)

	a, _ := reg.Get("a")
	assert.Equal(t, "default-ci", a.CI)

	b, _ := reg.Get("b")
	assert.Equal(t, "custom-ci", b.CI)
}

func TestLoadRegistry_Bad_FileNotFound_Good(t *testing.T) {
	m := io.NewMockMedium()
	_, err := LoadRegistry(m, "/nonexistent/repos.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read")
}

func TestLoadRegistry_Bad_InvalidYAML_Good(t *testing.T) {
	m := io.NewMockMedium()
	_ = m.Write("/tmp/bad.yaml", "{{{{not yaml at all")

	_, err := LoadRegistry(m, "/tmp/bad.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}

// ── List / Get / ByType ────────────────────────────────────────────

func newTestRegistry(t *testing.T) *Registry {
	t.Helper()
	m := io.NewMockMedium()
	yaml := `
version: 1
org: host-uk
base_path: /tmp/repos
repos:
  core-php:
    type: foundation
    description: Foundation
  core-admin:
    type: module
    depends_on: [core-php]
    description: Admin
  core-tenant:
    type: module
    depends_on: [core-php]
    description: Tenancy
  core-bio:
    type: product
    depends_on: [core-php, core-tenant]
    description: Bio product
`
	_ = m.Write("/tmp/repos.yaml", yaml)
	reg, err := LoadRegistry(m, "/tmp/repos.yaml")
	require.NoError(t, err)
	return reg
}

func TestRegistry_List_Good(t *testing.T) {
	reg := newTestRegistry(t)
	repos := reg.List()
	assert.Len(t, repos, 4)
}

func TestRegistry_List_Good_SortedByName_Good(t *testing.T) {
	m := io.NewMockMedium()
	yaml := `
version: 1
org: host-uk
base_path: /tmp/repos
repos:
  zulu:
    type: module
  alpha:
    type: module
  mike:
    type: module
`
	_ = m.Write("/tmp/repos.yaml", yaml)

	reg, err := LoadRegistry(m, "/tmp/repos.yaml")
	require.NoError(t, err)

	repos := reg.List()
	require.Len(t, repos, 3)
	assert.Equal(t, "alpha", repos[0].Name)
	assert.Equal(t, "mike", repos[1].Name)
	assert.Equal(t, "zulu", repos[2].Name)
}

func TestRegistry_Get_Good(t *testing.T) {
	reg := newTestRegistry(t)
	repo, ok := reg.Get("core-php")
	assert.True(t, ok)
	assert.Equal(t, "core-php", repo.Name)
}

func TestRegistry_Get_Bad_NotFound_Good(t *testing.T) {
	reg := newTestRegistry(t)
	_, ok := reg.Get("nonexistent")
	assert.False(t, ok)
}

func TestRegistry_ByType_Good(t *testing.T) {
	reg := newTestRegistry(t)

	foundations := reg.ByType("foundation")
	assert.Len(t, foundations, 1)
	assert.Equal(t, "core-php", foundations[0].Name)

	modules := reg.ByType("module")
	assert.Len(t, modules, 2)

	products := reg.ByType("product")
	assert.Len(t, products, 1)
}

func TestRegistry_ByType_Good_SortedByName_Good(t *testing.T) {
	m := io.NewMockMedium()
	yaml := `
version: 1
org: host-uk
base_path: /tmp/repos
repos:
  zulu:
    type: module
  alpha:
    type: module
  mike:
    type: foundation
`
	_ = m.Write("/tmp/repos.yaml", yaml)

	reg, err := LoadRegistry(m, "/tmp/repos.yaml")
	require.NoError(t, err)

	modules := reg.ByType("module")
	require.Len(t, modules, 2)
	assert.Equal(t, "alpha", modules[0].Name)
	assert.Equal(t, "zulu", modules[1].Name)
}

func TestRegistry_ByType_Good_NoMatch_Good(t *testing.T) {
	reg := newTestRegistry(t)
	templates := reg.ByType("template")
	assert.Empty(t, templates)
}

func TestRegistry_Dependents_Good(t *testing.T) {
	reg := newTestRegistry(t)

	dependents := reg.Dependents("core-php")
	require.Len(t, dependents, 3)
	assert.Equal(t, "core-admin", dependents[0].Name)
	assert.Equal(t, "core-bio", dependents[1].Name)
	assert.Equal(t, "core-tenant", dependents[2].Name)
}

func TestRegistry_Impact_Good(t *testing.T) {
	reg := newTestRegistry(t)

	impacted, err := reg.Impact("core-php")
	require.NoError(t, err)
	require.Len(t, impacted, 3)
	assert.Equal(t, "core-admin", impacted[0].Name)
	assert.Equal(t, "core-bio", impacted[1].Name)
	assert.Equal(t, "core-tenant", impacted[2].Name)
}

func TestRegistry_Impact_Good_BreadthFirstOrder_Good(t *testing.T) {
	m := io.NewMockMedium()
	yaml := `
version: 1
org: core
base_path: /tmp/repos
repos:
  core:
    type: foundation
  alpha:
    type: module
    depends_on: [core]
  beta:
    type: module
    depends_on: [core]
  delta:
    type: module
    depends_on: [alpha]
  gamma:
    type: module
    depends_on: [beta]
`
	_ = m.Write("/tmp/repos.yaml", yaml)

	reg, err := LoadRegistry(m, "/tmp/repos.yaml")
	require.NoError(t, err)

	impacted, err := reg.Impact("core")
	require.NoError(t, err)
	require.Len(t, impacted, 4)
	assert.Equal(t, []string{"alpha", "beta", "delta", "gamma"}, []string{
		impacted[0].Name,
		impacted[1].Name,
		impacted[2].Name,
		impacted[3].Name,
	})
}

func TestRegistry_Impact_Bad_UnknownRepo_Good(t *testing.T) {
	reg := newTestRegistry(t)

	_, err := reg.Impact("missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown repo")
}

// ── TopologicalOrder ───────────────────────────────────────────────

func TestTopologicalOrder_Good(t *testing.T) {
	reg := newTestRegistry(t)
	order, err := TopologicalOrder(reg)
	require.NoError(t, err)
	assert.Len(t, order, 4)

	// core-php must come before everything that depends on it.
	phpIdx := -1
	for i, r := range order {
		if r.Name == "core-php" {
			phpIdx = i
			break
		}
	}
	require.GreaterOrEqual(t, phpIdx, 0, "core-php not found")

	for i, r := range order {
		for _, dep := range r.DependsOn {
			depIdx := -1
			for j, d := range order {
				if d.Name == dep {
					depIdx = j
					break
				}
			}
			assert.Less(t, depIdx, i, "%s should come before %s", dep, r.Name)
		}
	}
}

func TopologicalOrder(reg *Registry) ([]*Repo, error) {
	return reg.TopologicalOrder()
}

func TestTopologicalOrder_Bad_CircularDep_Good(t *testing.T) {
	m := io.NewMockMedium()
	yaml := `
version: 1
org: test
base_path: /tmp
repos:
  a:
    type: module
    depends_on: [b]
  b:
    type: module
    depends_on: [a]
`
	_ = m.Write("/tmp/repos.yaml", yaml)
	reg, err := LoadRegistry(m, "/tmp/repos.yaml")
	require.NoError(t, err)

	_, err = reg.TopologicalOrder()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestTopologicalOrder_Bad_UnknownDep_Good(t *testing.T) {
	m := io.NewMockMedium()
	yaml := `
version: 1
org: test
base_path: /tmp
repos:
  a:
    type: module
    depends_on: [nonexistent]
`
	_ = m.Write("/tmp/repos.yaml", yaml)
	reg, err := LoadRegistry(m, "/tmp/repos.yaml")
	require.NoError(t, err)

	_, err = reg.TopologicalOrder()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown repo")
}

func TestTopologicalOrder_Good_NoDeps_Good(t *testing.T) {
	m := io.NewMockMedium()
	yaml := `
version: 1
org: test
base_path: /tmp
repos:
  a:
    type: module
  b:
    type: module
`
	_ = m.Write("/tmp/repos.yaml", yaml)
	reg, err := LoadRegistry(m, "/tmp/repos.yaml")
	require.NoError(t, err)

	order, err := reg.TopologicalOrder()
	require.NoError(t, err)
	assert.Len(t, order, 2)
}

func TestTopologicalOrder_Good_NoDeps_Sorted_Good(t *testing.T) {
	m := io.NewMockMedium()
	yaml := `
version: 1
org: test
base_path: /tmp
repos:
  zulu:
    type: module
  alpha:
    type: module
`
	_ = m.Write("/tmp/repos.yaml", yaml)
	reg, err := LoadRegistry(m, "/tmp/repos.yaml")
	require.NoError(t, err)

	order, err := reg.TopologicalOrder()
	require.NoError(t, err)
	require.Len(t, order, 2)
	assert.Equal(t, "alpha", order[0].Name)
	assert.Equal(t, "zulu", order[1].Name)
}

// ── ScanDirectory ──────────────────────────────────────────────────

func TestScanDirectory_Good(t *testing.T) {
	m := io.NewMockMedium()

	// Create mock repos with .git dirs. MockMedium surfaces directories
	// via files beneath them, so seed a marker file in each .git dir.
	_ = m.EnsureDir("/workspace/repo-a/.git")
	_ = m.Write("/workspace/repo-a/.git/HEAD", "ref: refs/heads/main\n")
	_ = m.EnsureDir("/workspace/repo-b/.git")
	_ = m.Write("/workspace/repo-b/.git/HEAD", "ref: refs/heads/main\n")
	_ = m.EnsureDir("/workspace/not-a-repo") // No .git
	_ = m.Write("/workspace/not-a-repo/README.md", "hello")

	// Write a file (not a dir) at top level.
	_ = m.Write("/workspace/README.md", "hello")

	reg, err := ScanDirectory(m, "/workspace")
	require.NoError(t, err)

	assert.Len(t, reg.Repos, 2)

	a, ok := reg.Repos["repo-a"]
	assert.True(t, ok)
	assert.Equal(t, "/workspace/repo-a", a.Path)
	assert.Equal(t, "module", a.Type) // Default type.

	_, ok = reg.Repos["not-a-repo"]
	assert.False(t, ok)
}

func TestScanDirectory_Good_DetectsGitHubOrg_Good(t *testing.T) {
	m := io.NewMockMedium()

	_ = m.EnsureDir("/workspace/my-repo/.git")
	_ = m.Write("/workspace/my-repo/.git/config", `[core]
	repositoryformatversion = 0
[remote "origin"]
	url = git@github.com:host-uk/my-repo.git
	fetch = +refs/heads/*:refs/remotes/origin/*
`)

	reg, err := ScanDirectory(m, "/workspace")
	require.NoError(t, err)
	assert.Equal(t, "host-uk", reg.Org)
}

func TestScanDirectory_Good_DetectsHTTPSOrg_Good(t *testing.T) {
	m := io.NewMockMedium()

	_ = m.EnsureDir("/workspace/my-repo/.git")
	_ = m.Write("/workspace/my-repo/.git/config", `[remote "origin"]
	url = https://github.com/lethean-io/my-repo.git
`)

	reg, err := ScanDirectory(m, "/workspace")
	require.NoError(t, err)
	assert.Equal(t, "lethean-io", reg.Org)
}

func TestScanDirectory_Good_EmptyDir_Good(t *testing.T) {
	m := io.NewMockMedium()
	_ = m.EnsureDir("/empty")

	reg, err := ScanDirectory(m, "/empty")
	require.NoError(t, err)
	assert.Empty(t, reg.Repos)
	assert.Equal(t, "", reg.Org)
}

func TestScanDirectory_Bad_InvalidDir_Good(t *testing.T) {
	m := io.NewMockMedium()
	_, err := ScanDirectory(m, "/nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read directory")
}

// ── detectOrg ──────────────────────────────────────────────────────

func TestDetectOrg_Good_SSHRemote_Good(t *testing.T) {
	m := io.NewMockMedium()
	_ = m.Write("/repo/.git/config", `[remote "origin"]
	url = git@github.com:host-uk/core.git
`)
	assert.Equal(t, "host-uk", detectOrg(m, "/repo"))
}

func TestDetectOrg_Good_HTTPSRemote_Good(t *testing.T) {
	m := io.NewMockMedium()
	_ = m.Write("/repo/.git/config", `[remote "origin"]
	url = https://github.com/snider/project.git
`)
	assert.Equal(t, "snider", detectOrg(m, "/repo"))
}

func TestDetectOrg_Bad_NoConfig_Good(t *testing.T) {
	m := io.NewMockMedium()
	assert.Equal(t, "", detectOrg(m, "/nonexistent"))
}

func TestDetectOrg_Bad_NoRemote_Good(t *testing.T) {
	m := io.NewMockMedium()
	_ = m.Write("/repo/.git/config", `[core]
	repositoryformatversion = 0
`)
	assert.Equal(t, "", detectOrg(m, "/repo"))
}

func TestDetectOrg_Bad_NonGitHubRemote_Good(t *testing.T) {
	m := io.NewMockMedium()
	_ = m.Write("/repo/.git/config", `[remote "origin"]
	url = ssh://git@forge.lthn.ai:2223/core/go.git
`)
	assert.Equal(t, "core", detectOrg(m, "/repo"))
}

func TestDetectOrg_Good_SCPStyleRemote_Good(t *testing.T) {
	m := io.NewMockMedium()
	_ = m.Write("/repo/.git/config", `[remote "origin"]
	url = git@forge.lthn.ai:core/go.git
`)
	assert.Equal(t, "core", detectOrg(m, "/repo"))
}

// ── expandPath ─────────────────────────────────────────────────────

func TestExpandPath_Good_Tilde_Good(t *testing.T) {
	got := expandPath("~/Code/repos")
	assert.NotContains(t, got, "~")
	assert.Contains(t, got, "Code/repos")
}

func TestExpandPath_Good_NoTilde_Good(t *testing.T) {
	assert.Equal(t, "/absolute/path", expandPath("/absolute/path"))
	assert.Equal(t, "relative/path", expandPath("relative/path"))
}

// ── Repo.Exists / IsGitRepo ───────────────────────────────────────

func TestRepo_Exists_Good(t *testing.T) {
	m := io.NewMockMedium()
	reg := &Registry{
		medium:   m,
		BasePath: "/tmp/repos",
		Repos:    make(map[string]*Repo),
	}
	repo := &Repo{
		Name:     "core",
		Path:     "/tmp/repos/core",
		registry: reg,
	}

	assert.False(t, repo.Exists())

	_ = m.EnsureDir("/tmp/repos/core")
	assert.True(t, repo.Exists())
}

func TestRepo_IsGitRepo_Good(t *testing.T) {
	m := io.NewMockMedium()
	reg := &Registry{
		medium:   m,
		BasePath: "/tmp/repos",
		Repos:    make(map[string]*Repo),
	}
	repo := &Repo{
		Name:     "core",
		Path:     "/tmp/repos/core",
		registry: reg,
	}

	assert.False(t, repo.IsGitRepo())

	_ = m.EnsureDir("/tmp/repos/core/.git")
	assert.True(t, repo.IsGitRepo())
}

// ── getMedium fallback ─────────────────────────────────────────────

func TestGetMedium_Good_FallbackToLocal_Good(t *testing.T) {
	repo := &Repo{Name: "orphan", Path: "/tmp/orphan"}
	// No registry set — should fall back to io.Local.
	m := repo.getMedium()
	assert.Equal(t, io.Local, m)
}

func TestGetMedium_Good_NilMediumFallback_Good(t *testing.T) {
	reg := &Registry{} // medium is nil.
	repo := &Repo{Name: "test", registry: reg}
	m := repo.getMedium()
	assert.Equal(t, io.Local, m)
}

// ── Registry discovery ────────────────────────────────────────────

func TestFindRegistry_Good_CORE_REPOSEnv_Good(t *testing.T) {
	m := io.NewMockMedium()
	_ = m.Write("/custom/repos.yaml", `
version: 1
org: test
base_path: /tmp/repos
repos: {}
`)
	t.Setenv("CORE_REPOS", "/custom/repos.yaml")

	path, err := FindRegistry(m)
	require.NoError(t, err)
	assert.Equal(t, "/custom/repos.yaml", path)
}

func TestFindRegistry_Good_CORE_REPOSEnv_Tilde_Good(t *testing.T) {
	m := io.NewMockMedium()
	home := t.TempDir()
	t.Setenv("HOME", home)
	path := filepath.Join(home, "repos.yaml")
	_ = m.Write(path, `
version: 1
org: test
base_path: /tmp/repos
repos: {}
`)
	t.Setenv("CORE_REPOS", "~/repos.yaml")

	found, err := FindRegistry(m)
	require.NoError(t, err)
	assert.Equal(t, path, found)
}

func TestLoadRegistry_Good_RelativePathAndBranch_Good(t *testing.T) {
	m := io.NewMockMedium()
	yaml := `
version: 1
org: host-uk
base_path: /tmp/repos
defaults:
  branch: main
repos:
  go:
    type: module
    path: core/go
    branch: dev
`
	_ = m.Write("/tmp/repos.yaml", yaml)

	reg, err := LoadRegistry(m, "/tmp/repos.yaml")
	require.NoError(t, err)

	repo, ok := reg.Get("go")
	require.True(t, ok)
	assert.Equal(t, "go", repo.Name)
	assert.Equal(t, "/tmp/repos/core/go", repo.Path)
	assert.Equal(t, "dev", repo.Branch)
}

func TestLoadRegistry_Good_ListForm_Good(t *testing.T) {
	m := io.NewMockMedium()
	yaml := `
version: 1
org: host-uk
base_path: /tmp/repos
repos:
  - path: core/go
    remote: ssh://git@forge.lthn.ai:2223/core/go.git
    branch: dev
    type: module
    description: Go packages
`
	_ = m.Write("/tmp/repos.yaml", yaml)

	reg, err := LoadRegistry(m, "/tmp/repos.yaml")
	require.NoError(t, err)

	repo, ok := reg.Get("go")
	require.True(t, ok)
	assert.Equal(t, "go", repo.Name)
	assert.Equal(t, "/tmp/repos/core/go", repo.Path)
	assert.Equal(t, "ssh://git@forge.lthn.ai:2223/core/go.git", repo.Remote)
	assert.Equal(t, "dev", repo.Branch)
	assert.Equal(t, "Go packages", repo.Description)
}

func TestFindRegistries_Good_HomeCoreRepos_Good(t *testing.T) {
	oldWD, err := os.Getwd()
	require.NoError(t, err)
	cwd := t.TempDir()
	require.NoError(t, os.Chdir(cwd))
	t.Cleanup(func() {
		_ = os.Chdir(oldWD)
	})

	home := t.TempDir()
	t.Setenv("HOME", home)

	m := io.NewMockMedium()
	path := filepath.Join(home, ".core", "repos.yaml")
	_ = m.Write(path, `
version: 1
org: test
base_path: /tmp/repos
repos: {}
`)

	paths, err := FindRegistries(m)
	require.NoError(t, err)
	assert.Contains(t, paths, path)
}

func TestFindRegistries_Good_CORE_REPOSEnv_MultiplePaths_Good(t *testing.T) {
	oldWD, err := os.Getwd()
	require.NoError(t, err)
	cwd := t.TempDir()
	require.NoError(t, os.Chdir(cwd))
	t.Cleanup(func() {
		_ = os.Chdir(oldWD)
	})

	t.Setenv("HOME", t.TempDir())

	m := io.NewMockMedium()
	first := "/custom/one/repos.yaml"
	second := "/custom/two/repos.yaml"

	_ = m.Write(first, `
version: 1
org: one
base_path: /tmp/one
repos: {}
`)
	_ = m.Write(second, `
version: 1
org: two
base_path: /tmp/two
repos: {}
`)

	t.Setenv("CORE_REPOS", first+","+second)

	paths, err := FindRegistries(m)
	require.NoError(t, err)
	assert.Equal(t, []string{first, second}, paths[:2])
}

func TestLoadRegistries_Good_NoRegistriesReturnsEmpty_Good(t *testing.T) {
	oldWD, err := os.Getwd()
	require.NoError(t, err)
	cwd := t.TempDir()
	require.NoError(t, os.Chdir(cwd))
	t.Cleanup(func() {
		_ = os.Chdir(oldWD)
	})

	t.Setenv("HOME", t.TempDir())
	t.Setenv("CORE_REPOS", "")

	regs, err := LoadRegistries(io.NewMockMedium())
	require.NoError(t, err)
	assert.Empty(t, regs)
}

func TestMergeRegistries_Good_FirstOccurrenceWins_Good(t *testing.T) {
	m := io.NewMockMedium()
	require.NoError(t, m.Write("/tmp/one.yaml", `
version: 1
org: core
base_path: /tmp/one
repos:
  shared:
    type: module
  alpha:
    type: module
`))
	require.NoError(t, m.Write("/tmp/two.yaml", `
version: 1
org: other
base_path: /tmp/two
repos:
  shared:
    type: module
    path: elsewhere/shared
  beta:
    type: module
`))

	first, err := LoadRegistry(m, "/tmp/one.yaml")
	require.NoError(t, err)
	second, err := LoadRegistry(m, "/tmp/two.yaml")
	require.NoError(t, err)

	merged := MergeRegistries(first, second)
	repos := merged.List()
	require.Len(t, repos, 3)

	shared, ok := merged.Get("shared")
	require.True(t, ok)
	assert.Equal(t, "/tmp/one/shared", shared.Path)
	assert.Equal(t, "core", shared.Org)

	beta, ok := merged.Get("beta")
	require.True(t, ok)
	assert.Equal(t, "/tmp/two/beta", beta.Path)
	assert.Equal(t, "other", beta.Org)
}
