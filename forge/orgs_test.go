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

func TestClient_ListMyOrgs_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	orgs, err := client.ListMyOrgs()
	require.NoError(t, err)
	require.Len(t, orgs, 1)
	assert.Equal(t, "test-org", orgs[0].UserName)
}

func TestClient_ListMyOrgsIter_Good_Paginates_Good(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]string{"version": "1.21.0"})
	})
	mux.HandleFunc("/api/v1/user/orgs", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("page") {
		case "2":
			jsonResponse(w, []map[string]any{
				{"id": 101, "login": "second-org", "username": "second-org", "full_name": "Second Organisation"},
			})
		default:
			w.Header().Set("Link", "<http://"+r.Host+"/api/v1/user/orgs?page=2>; rel=\"next\", <http://"+r.Host+"/api/v1/user/orgs?page=2>; rel=\"last\"")
			jsonResponse(w, []map[string]any{
				{"id": 100, "login": "test-org", "username": "test-org", "full_name": "Test Organisation"},
			})
		}
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := New(srv.URL, "test-token")
	require.NoError(t, err)

	var names []string
	for org, err := range client.ListMyOrgsIter() {
		require.NoError(t, err)
		names = append(names, org.UserName)
	}

	require.Len(t, names, 2)
	assert.Equal(t, []string{"test-org", "second-org"}, names)
}

func TestClient_ListMyOrgs_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.ListMyOrgs()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list orgs")
}

func TestClient_ListMyOrgsIter_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	for _, err := range client.ListMyOrgsIter() {
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list orgs")
		break
	}
}

func TestClient_GetOrg_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	org, err := client.GetOrg("test-org")
	require.NoError(t, err)
	assert.Equal(t, "test-org", org.UserName)
}

func TestClient_GetOrg_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.GetOrg("test-org")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get org")
}

func TestClient_CreateOrg_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	org, err := client.CreateOrg(forgejo.CreateOrgOption{
		Name:       "new-org",
		FullName:   "New Organisation",
		Visibility: "private",
	})
	require.NoError(t, err)
	assert.NotNil(t, org)
}

func TestClient_CreateOrg_Bad_ServerError_Good(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.CreateOrg(forgejo.CreateOrgOption{
		Name: "new-org",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create org")
}
