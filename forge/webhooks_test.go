// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	"net/http"
	"net/http/httptest"
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

func TestClient_CreateRepoWebhook_Bad_ServerError_Good(t *testing.T) {
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

func TestClient_ListRepoWebhooks_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.ListRepoWebhooks("test-org", "org-repo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list repo webhooks")
}

func TestClient_ListRepoWebhooksIter_Good_Paginates_Good(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]string{"version": "1.21.0"})
	})
	mux.HandleFunc("/api/v1/repos/test-org/org-repo/hooks", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("page") {
		case "2":
			jsonResponse(w, []map[string]any{
				{"id": 2, "type": "forgejo", "active": true, "config": map[string]any{"url": "https://example.com/second"}},
			})
		case "3":
			jsonResponse(w, []map[string]any{})
		default:
			w.Header().Set("Link", "<http://"+r.Host+"/api/v1/repos/test-org/org-repo/hooks?page=2>; rel=\"next\", <http://"+r.Host+"/api/v1/repos/test-org/org-repo/hooks?page=2>; rel=\"last\"")
			jsonResponse(w, []map[string]any{
				{"id": 1, "type": "forgejo", "active": true, "config": map[string]any{"url": "https://example.com/hook"}},
			})
		}
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := New(srv.URL, "test-token")
	require.NoError(t, err)

	var urls []string
	for hook, err := range client.ListRepoWebhooksIter("test-org", "org-repo") {
		require.NoError(t, err)
		urls = append(urls, hook.Config["url"])
	}

	require.Len(t, urls, 2)
	assert.Equal(t, []string{"https://example.com/hook", "https://example.com/second"}, urls)
}
