package gitea

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// isolateConfigEnv sets up a clean environment for config resolution tests.
func isolateConfigEnv(t *testing.T) {
	t.Helper()
	t.Setenv("GITEA_URL", "")
	t.Setenv("GITEA_TOKEN", "")
	t.Setenv("HOME", t.TempDir())
}

func TestResolveConfig_Good_Defaults(t *testing.T) {
	isolateConfigEnv(t)

	url, token, err := ResolveConfig("", "")
	require.NoError(t, err)
	assert.Equal(t, DefaultURL, url, "URL should default to DefaultURL")
	assert.Empty(t, token, "token should be empty when nothing configured")
}

func TestResolveConfig_Good_FlagsOverrideAll(t *testing.T) {
	isolateConfigEnv(t)
	t.Setenv("GITEA_URL", "https://env-url.example.com")
	t.Setenv("GITEA_TOKEN", "env-token-abc")

	url, token, err := ResolveConfig("https://flag-url.example.com", "flag-token-xyz")
	require.NoError(t, err)
	assert.Equal(t, "https://flag-url.example.com", url, "flag URL should override env")
	assert.Equal(t, "flag-token-xyz", token, "flag token should override env")
}

func TestResolveConfig_Good_EnvVarsOverrideConfig(t *testing.T) {
	isolateConfigEnv(t)
	t.Setenv("GITEA_URL", "https://env-url.example.com")
	t.Setenv("GITEA_TOKEN", "env-token-123")

	url, token, err := ResolveConfig("", "")
	require.NoError(t, err)
	assert.Equal(t, "https://env-url.example.com", url)
	assert.Equal(t, "env-token-123", token)
}

func TestResolveConfig_Good_PartialOverrides(t *testing.T) {
	isolateConfigEnv(t)
	t.Setenv("GITEA_URL", "https://env-only.example.com")

	url, token, err := ResolveConfig("", "flag-only-token")
	require.NoError(t, err)
	assert.Equal(t, "https://env-only.example.com", url, "env URL should be used")
	assert.Equal(t, "flag-only-token", token, "flag token should be used")
}

func TestResolveConfig_Good_URLDefaultsWhenEmpty(t *testing.T) {
	isolateConfigEnv(t)
	t.Setenv("GITEA_TOKEN", "some-token")

	url, token, err := ResolveConfig("", "")
	require.NoError(t, err)
	assert.Equal(t, DefaultURL, url, "URL should fall back to default")
	assert.Equal(t, "some-token", token)
}

func TestConstants(t *testing.T) {
	assert.Equal(t, "gitea.url", ConfigKeyURL)
	assert.Equal(t, "gitea.token", ConfigKeyToken)
	assert.Equal(t, "https://gitea.snider.dev", DefaultURL)
}

func TestNewFromConfig_Bad_NoToken(t *testing.T) {
	isolateConfigEnv(t)

	_, err := NewFromConfig("", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no API token configured")
}

func TestNewFromConfig_Good_WithFlagToken(t *testing.T) {
	isolateConfigEnv(t)

	srv := newMockGiteaServer(t)
	defer srv.Close()

	client, err := NewFromConfig(srv.URL, "test-token")
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, srv.URL, client.URL())
}

func TestNewFromConfig_Good_EnvToken(t *testing.T) {
	isolateConfigEnv(t)

	srv := newMockGiteaServer(t)
	defer srv.Close()

	t.Setenv("GITEA_URL", srv.URL)
	t.Setenv("GITEA_TOKEN", "env-test-token")

	client, err := NewFromConfig("", "")
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, srv.URL, client.URL())
}
