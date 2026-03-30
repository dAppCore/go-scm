// SPDX-Licence-Identifier: EUPL-1.2

package gitea

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_Good(t *testing.T) {
	srv := newMockGiteaServer(t)
	defer srv.Close()

	client, err := New(srv.URL, "test-token-123")
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.API())
	assert.Equal(t, srv.URL, client.URL())
}

func TestNew_Bad_InvalidURL(t *testing.T) {
	_, err := New("://invalid-url", "token")
	assert.Error(t, err)
}

func TestClient_API_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	assert.NotNil(t, client.API(), "API() should return the underlying SDK client")
}

func TestClient_URL_Good(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	assert.Equal(t, srv.URL, client.URL())
}
