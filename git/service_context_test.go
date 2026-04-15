// SPDX-License-Identifier: EUPL-1.2

package git

import (
	"context"
	"testing"

	"dappco.re/go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceHandleQuery_UsesCoreContext_Good(t *testing.T) {
	repo := createTaggedRepo(t, "service-query-repo",
		repoVersion{Version: "1.0.0"},
	)

	c := core.New()
	svc := &Service{ServiceRuntime: core.NewServiceRuntime(c, ServiceOptions{})}

	require.True(t, c.ServiceStartup(context.Background(), nil).OK)
	require.True(t, c.ServiceShutdown(context.Background()).OK)

	result := svc.handleQuery(c, QueryStatus{
		Paths: []string{repo},
		Names: map[string]string{repo: "service-query-repo"},
	})

	require.True(t, result.OK)

	statuses, ok := result.Value.([]RepoStatus)
	require.True(t, ok)
	require.Len(t, statuses, 1)
	require.Error(t, statuses[0].Error)
	assert.Contains(t, statuses[0].Error.Error(), "context canceled")
}

func TestServiceHandleAction_UsesCoreContext_Good(t *testing.T) {
	repo := createTaggedRepo(t, "service-action-repo",
		repoVersion{Version: "1.0.0"},
	)

	c := core.New()
	svc := &Service{ServiceRuntime: core.NewServiceRuntime(c, ServiceOptions{})}

	require.True(t, c.ServiceStartup(context.Background(), nil).OK)
	require.True(t, c.ServiceShutdown(context.Background()).OK)

	result := svc.handleAction(c, TaskCreateBranch{
		Path:   repo,
		Branch: "feature/context-aware",
	})

	require.False(t, result.OK)
	err, ok := result.Value.(error)
	require.True(t, ok)
	assert.Contains(t, err.Error(), "context canceled")
}
