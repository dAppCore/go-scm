// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	"testing"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_CreateRepoWebhook_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	hook, err := client.CreateRepoWebhook("test-org", "org-repo", forgejo.CreateHookOption{
		Type: "forgejo",
		Config: map[string]string{
			"url": "https://example.com/hook",
		},
	})
	require.NoError(t, err)
	assert.NotNil(t, hook)
}

func TestClient_CreateRepoWebhook_Bad_ServerError(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.CreateRepoWebhook("test-org", "org-repo", forgejo.CreateHookOption{
		Type: "forgejo",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create repo webhook")
}

func TestClient_ListRepoWebhooks_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	hooks, err := client.ListRepoWebhooks("test-org", "org-repo")
	require.NoError(t, err)
	require.Len(t, hooks, 1)
}

func TestClient_ListRepoWebhooks_Bad_ServerError(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.ListRepoWebhooks("test-org", "org-repo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list repo webhooks")
}
