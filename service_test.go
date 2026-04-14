// SPDX-License-Identifier: EUPL-1.2

package scm

import (
	"context"
	"testing"

	"dappco.re/go/core"
	"dappco.re/go/core/io"
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
