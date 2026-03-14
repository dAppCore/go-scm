package marketplace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createProviderDir creates a provider directory with a .core/manifest.yaml.
func createProviderDir(t *testing.T, baseDir, code string, manifestYAML string) string {
	t.Helper()
	provDir := filepath.Join(baseDir, code)
	coreDir := filepath.Join(provDir, ".core")
	require.NoError(t, os.MkdirAll(coreDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(coreDir, "manifest.yaml"),
		[]byte(manifestYAML), 0644,
	))
	return provDir
}

func TestDiscoverProviders_Good(t *testing.T) {
	dir := t.TempDir()

	createProviderDir(t, dir, "cool-widget", `
code: cool-widget
name: Cool Widget
version: 1.0.0
namespace: /api/v1/cool-widget
binary: ./cool-widget
element:
  tag: core-cool-widget
  source: ./assets/core-cool-widget.js
`)

	createProviderDir(t, dir, "data-viz", `
code: data-viz
name: Data Visualiser
version: 0.2.0
namespace: /api/v1/data-viz
binary: ./data-viz
`)

	providers, err := DiscoverProviders(dir)
	require.NoError(t, err)
	assert.Len(t, providers, 2)

	codes := map[string]bool{}
	for _, p := range providers {
		codes[p.Manifest.Code] = true
	}
	assert.True(t, codes["cool-widget"])
	assert.True(t, codes["data-viz"])
}

func TestDiscoverProviders_Good_SkipNonProvider(t *testing.T) {
	dir := t.TempDir()

	// This has a valid manifest but no namespace/binary — not a provider.
	createProviderDir(t, dir, "plain-module", `
code: plain-module
name: Plain Module
version: 1.0.0
`)

	// This IS a provider.
	createProviderDir(t, dir, "real-provider", `
code: real-provider
name: Real Provider
version: 1.0.0
namespace: /api/v1/real
binary: ./real-provider
`)

	providers, err := DiscoverProviders(dir)
	require.NoError(t, err)
	assert.Len(t, providers, 1)
	assert.Equal(t, "real-provider", providers[0].Manifest.Code)
}

func TestDiscoverProviders_Good_SkipNoManifest(t *testing.T) {
	dir := t.TempDir()

	// Directory with no manifest.
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "no-manifest"), 0755))

	// Directory with a valid provider manifest.
	createProviderDir(t, dir, "good-provider", `
code: good-provider
name: Good Provider
version: 1.0.0
namespace: /api/v1/good
binary: ./good-provider
`)

	providers, err := DiscoverProviders(dir)
	require.NoError(t, err)
	assert.Len(t, providers, 1)
	assert.Equal(t, "good-provider", providers[0].Manifest.Code)
}

func TestDiscoverProviders_Good_SkipInvalidManifest(t *testing.T) {
	dir := t.TempDir()

	// Directory with invalid YAML.
	provDir := filepath.Join(dir, "bad-yaml")
	coreDir := filepath.Join(provDir, ".core")
	require.NoError(t, os.MkdirAll(coreDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(coreDir, "manifest.yaml"),
		[]byte("not: valid: yaml: ["), 0644,
	))

	providers, err := DiscoverProviders(dir)
	require.NoError(t, err)
	assert.Empty(t, providers)
}

func TestDiscoverProviders_Good_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	providers, err := DiscoverProviders(dir)
	require.NoError(t, err)
	assert.Empty(t, providers)
}

func TestDiscoverProviders_Good_NonexistentDir(t *testing.T) {
	providers, err := DiscoverProviders("/tmp/nonexistent-discovery-test-dir")
	require.NoError(t, err)
	assert.Nil(t, providers)
}

func TestDiscoverProviders_Good_SkipFiles(t *testing.T) {
	dir := t.TempDir()

	// Create a regular file (not a directory).
	require.NoError(t, os.WriteFile(filepath.Join(dir, "readme.md"), []byte("# readme"), 0644))

	providers, err := DiscoverProviders(dir)
	require.NoError(t, err)
	assert.Empty(t, providers)
}

func TestDiscoverProviders_Good_ProviderDir(t *testing.T) {
	dir := t.TempDir()

	createProviderDir(t, dir, "test-prov", `
code: test-prov
name: Test Provider
version: 1.0.0
namespace: /api/v1/test-prov
binary: ./test-prov
`)

	providers, err := DiscoverProviders(dir)
	require.NoError(t, err)
	require.Len(t, providers, 1)
	assert.Equal(t, filepath.Join(dir, "test-prov"), providers[0].Dir)
}

// -- ProviderRegistryFile tests -----------------------------------------------

func TestProviderRegistry_LoadSave_Good(t *testing.T) {
	path := filepath.Join(t.TempDir(), "registry.yaml")

	reg := &ProviderRegistryFile{
		Version:   1,
		Providers: map[string]ProviderRegistryEntry{},
	}
	reg.Add("cool-widget", ProviderRegistryEntry{
		Installed: "2026-03-14T12:00:00Z",
		Version:   "1.0.0",
		Source:    "forge.lthn.ai/someone/cool-widget",
		AutoStart: true,
	})

	err := SaveProviderRegistry(path, reg)
	require.NoError(t, err)

	loaded, err := LoadProviderRegistry(path)
	require.NoError(t, err)
	assert.Equal(t, 1, loaded.Version)
	assert.Len(t, loaded.Providers, 1)

	entry, ok := loaded.Get("cool-widget")
	require.True(t, ok)
	assert.Equal(t, "1.0.0", entry.Version)
	assert.Equal(t, "forge.lthn.ai/someone/cool-widget", entry.Source)
	assert.True(t, entry.AutoStart)
}

func TestProviderRegistry_Load_Good_NonexistentFile(t *testing.T) {
	reg, err := LoadProviderRegistry("/tmp/nonexistent-registry-test.yaml")
	require.NoError(t, err)
	assert.Equal(t, 1, reg.Version)
	assert.Empty(t, reg.Providers)
}

func TestProviderRegistry_Add_Good(t *testing.T) {
	reg := &ProviderRegistryFile{
		Version:   1,
		Providers: map[string]ProviderRegistryEntry{},
	}

	reg.Add("widget-a", ProviderRegistryEntry{Version: "1.0.0", AutoStart: true})
	reg.Add("widget-b", ProviderRegistryEntry{Version: "2.0.0", AutoStart: false})

	assert.Len(t, reg.Providers, 2)

	a, ok := reg.Get("widget-a")
	require.True(t, ok)
	assert.Equal(t, "1.0.0", a.Version)
}

func TestProviderRegistry_Remove_Good(t *testing.T) {
	reg := &ProviderRegistryFile{
		Version: 1,
		Providers: map[string]ProviderRegistryEntry{
			"widget-a": {Version: "1.0.0"},
			"widget-b": {Version: "2.0.0"},
		},
	}

	reg.Remove("widget-a")
	assert.Len(t, reg.Providers, 1)

	_, ok := reg.Get("widget-a")
	assert.False(t, ok)
}

func TestProviderRegistry_Get_Bad_NotFound(t *testing.T) {
	reg := &ProviderRegistryFile{
		Version:   1,
		Providers: map[string]ProviderRegistryEntry{},
	}

	_, ok := reg.Get("nonexistent")
	assert.False(t, ok)
}

func TestProviderRegistry_List_Good(t *testing.T) {
	reg := &ProviderRegistryFile{
		Version: 1,
		Providers: map[string]ProviderRegistryEntry{
			"a": {Version: "1.0"},
			"b": {Version: "2.0"},
		},
	}

	codes := reg.List()
	assert.Len(t, codes, 2)
	assert.Contains(t, codes, "a")
	assert.Contains(t, codes, "b")
}

func TestProviderRegistry_AutoStartProviders_Good(t *testing.T) {
	reg := &ProviderRegistryFile{
		Version: 1,
		Providers: map[string]ProviderRegistryEntry{
			"auto-a":   {Version: "1.0", AutoStart: true},
			"manual-b": {Version: "2.0", AutoStart: false},
			"auto-c":   {Version: "3.0", AutoStart: true},
		},
	}

	auto := reg.AutoStartProviders()
	assert.Len(t, auto, 2)
	assert.Contains(t, auto, "auto-a")
	assert.Contains(t, auto, "auto-c")
}
