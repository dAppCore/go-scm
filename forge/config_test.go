// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// isolateConfigEnv sets up a clean environment for config resolution tests.
// Clears FORGE_* env vars and points HOME to a temp dir so no config file is loaded.
func isolateConfigEnv(t *testing.T) {
	t.Helper()
	t.Setenv("FORGE_URL", "")
	t.Setenv("FORGE_TOKEN", "")
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
	t.Setenv("FORGE_URL", "https://env-url.example.com")
	t.Setenv("FORGE_TOKEN", "env-token-abc")

	url, token, err := ResolveConfig("https://flag-url.example.com", "flag-token-xyz")
	require.NoError(t, err)
	assert.Equal(t, "https://flag-url.example.com", url, "flag URL should override env")
	assert.Equal(t, "flag-token-xyz", token, "flag token should override env")
}

func TestResolveConfig_Good_EnvVarsOverrideConfig(t *testing.T) {
	isolateConfigEnv(t)
	t.Setenv("FORGE_URL", "https://env-url.example.com")
	t.Setenv("FORGE_TOKEN", "env-token-123")

	url, token, err := ResolveConfig("", "")
	require.NoError(t, err)
	assert.Equal(t, "https://env-url.example.com", url)
	assert.Equal(t, "env-token-123", token)
}

func TestResolveConfig_Good_PartialOverrides(t *testing.T) {
	isolateConfigEnv(t)
	// Set only env URL, flag token.
	t.Setenv("FORGE_URL", "https://env-only.example.com")

	url, token, err := ResolveConfig("", "flag-only-token")
	require.NoError(t, err)
	assert.Equal(t, "https://env-only.example.com", url, "env URL should be used")
	assert.Equal(t, "flag-only-token", token, "flag token should be used")
}

func TestResolveConfig_Good_URLDefaultsWhenEmpty(t *testing.T) {
	isolateConfigEnv(t)
	t.Setenv("FORGE_TOKEN", "some-token")

	url, token, err := ResolveConfig("", "")
	require.NoError(t, err)
	assert.Equal(t, DefaultURL, url, "URL should fall back to default")
	assert.Equal(t, "some-token", token)
}

func TestConstants_Good(t *testing.T) {
	assert.Equal(t, "forge.url", ConfigKeyURL)
	assert.Equal(t, "forge.token", ConfigKeyToken)
	assert.Equal(t, "http://localhost:4000", DefaultURL)
}

func TestNewFromConfig_Bad_NoToken(t *testing.T) {
	isolateConfigEnv(t)

	_, err := NewFromConfig("", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no API token configured")
}

func TestNewFromConfig_Good_WithFlagToken(t *testing.T) {
	isolateConfigEnv(t)

	// The Forgejo SDK NewClient validates the token by calling /api/v1/version,
	// so we need a mock HTTP server.
	srv := newMockForgejoServer(t)
	defer srv.Close()

	client, err := NewFromConfig(srv.URL, "test-token")
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, srv.URL, client.URL())
	assert.Equal(t, "test-token", client.Token())
}

func TestNewFromConfig_Good_EnvToken(t *testing.T) {
	isolateConfigEnv(t)

	srv := newMockForgejoServer(t)
	defer srv.Close()

	t.Setenv("FORGE_URL", srv.URL)
	t.Setenv("FORGE_TOKEN", "env-test-token")

	client, err := NewFromConfig("", "")
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, srv.URL, client.URL())
}
