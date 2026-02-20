package forge

import (
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

func TestClient_ListMyOrgs_Bad_ServerError(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.ListMyOrgs()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list orgs")
}

func TestClient_GetOrg_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	org, err := client.GetOrg("test-org")
	require.NoError(t, err)
	assert.Equal(t, "test-org", org.UserName)
}

func TestClient_GetOrg_Bad_ServerError(t *testing.T) {
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

func TestClient_CreateOrg_Bad_ServerError(t *testing.T) {
	client, srv := newErrorServer(t)
	defer srv.Close()

	_, err := client.CreateOrg(forgejo.CreateOrgOption{
		Name: "new-org",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create org")
}
