// SPDX-License-Identifier: EUPL-1.2

package scm

import (
	"context"
	"testing"

	"dappco.re/go/core"
	"dappco.re/go/core/io"
	"dappco.re/go/core/scm/repos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
