package marketplace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	coreio "forge.lthn.ai/core/go-io"
	"forge.lthn.ai/core/go-scm/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeManifestYAML writes a .core/manifest.yaml for a module directory.
func writeManifestYAML(t *testing.T, dir, code, name, version string) {
	t.Helper()
	coreDir := filepath.Join(dir, ".core")
	require.NoError(t, os.MkdirAll(coreDir, 0755))
	yaml := "code: " + code + "\nname: " + name + "\nversion: " + version + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(coreDir, "manifest.yaml"), []byte(yaml), 0644))
}

// writeCoreJSON writes a core.json for a module directory.
func writeCoreJSON(t *testing.T, dir, code, name, version string) {
	t.Helper()
	cm := manifest.CompiledManifest{
		Manifest: manifest.Manifest{
			Code:    code,
			Name:    name,
			Version: version,
		},
		Commit: "abc123",
	}
	data, err := json.Marshal(cm)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "core.json"), data, 0644))
}

func TestBuildFromDirs_Good_ManifestYAML(t *testing.T) {
	root := t.TempDir()
	modDir := filepath.Join(root, "my-widget")
	require.NoError(t, os.MkdirAll(modDir, 0755))
	writeManifestYAML(t, modDir, "my-widget", "My Widget", "1.0.0")

	b := &Builder{BaseURL: "https://forge.lthn.ai", Org: "core"}
	idx, err := b.BuildFromDirs(root)
	require.NoError(t, err)

	require.Len(t, idx.Modules, 1)
	assert.Equal(t, "my-widget", idx.Modules[0].Code)
	assert.Equal(t, "My Widget", idx.Modules[0].Name)
	assert.Equal(t, "https://forge.lthn.ai/core/my-widget", idx.Modules[0].Repo)
	assert.Equal(t, IndexVersion, idx.Version)
}

func TestBuildFromDirs_Good_CoreJSON(t *testing.T) {
	root := t.TempDir()
	modDir := filepath.Join(root, "compiled-mod")
	require.NoError(t, os.MkdirAll(modDir, 0755))
	writeCoreJSON(t, modDir, "compiled-mod", "Compiled Module", "2.0.0")

	b := &Builder{}
	idx, err := b.BuildFromDirs(root)
	require.NoError(t, err)

	require.Len(t, idx.Modules, 1)
	assert.Equal(t, "compiled-mod", idx.Modules[0].Code)
	assert.Equal(t, "Compiled Module", idx.Modules[0].Name)
}

func TestBuildFromDirs_Good_PrefersCompiledOverSource(t *testing.T) {
	root := t.TempDir()
	modDir := filepath.Join(root, "dual-mod")
	require.NoError(t, os.MkdirAll(modDir, 0755))
	writeManifestYAML(t, modDir, "source-code", "Source Name", "1.0.0")
	writeCoreJSON(t, modDir, "compiled-code", "Compiled Name", "2.0.0")

	b := &Builder{}
	idx, err := b.BuildFromDirs(root)
	require.NoError(t, err)

	// core.json is preferred — its code should appear.
	require.Len(t, idx.Modules, 1)
	assert.Equal(t, "compiled-code", idx.Modules[0].Code)
}

func TestBuildFromDirs_Good_SkipsNoManifest(t *testing.T) {
	root := t.TempDir()
	// Directory with no manifest.
	require.NoError(t, os.MkdirAll(filepath.Join(root, "no-manifest"), 0755))
	// Directory with a manifest.
	modDir := filepath.Join(root, "has-manifest")
	require.NoError(t, os.MkdirAll(modDir, 0755))
	writeManifestYAML(t, modDir, "has-manifest", "Has Manifest", "0.1.0")

	b := &Builder{}
	idx, err := b.BuildFromDirs(root)
	require.NoError(t, err)
	assert.Len(t, idx.Modules, 1)
}

func TestBuildFromDirs_Good_Deduplicates(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	mod1 := filepath.Join(dir1, "shared")
	mod2 := filepath.Join(dir2, "shared")
	require.NoError(t, os.MkdirAll(mod1, 0755))
	require.NoError(t, os.MkdirAll(mod2, 0755))
	writeManifestYAML(t, mod1, "shared", "Shared V1", "1.0.0")
	writeManifestYAML(t, mod2, "shared", "Shared V2", "2.0.0")

	b := &Builder{}
	idx, err := b.BuildFromDirs(dir1, dir2)
	require.NoError(t, err)
	// First occurrence wins.
	assert.Len(t, idx.Modules, 1)
	assert.Equal(t, "shared", idx.Modules[0].Code)
}

func TestBuildFromDirs_Good_SortsByCode(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"charlie", "alpha", "bravo"} {
		d := filepath.Join(root, name)
		require.NoError(t, os.MkdirAll(d, 0755))
		writeManifestYAML(t, d, name, name, "1.0.0")
	}

	b := &Builder{}
	idx, err := b.BuildFromDirs(root)
	require.NoError(t, err)
	require.Len(t, idx.Modules, 3)
	assert.Equal(t, "alpha", idx.Modules[0].Code)
	assert.Equal(t, "bravo", idx.Modules[1].Code)
	assert.Equal(t, "charlie", idx.Modules[2].Code)
}

func TestBuildFromDirs_Good_EmptyDir(t *testing.T) {
	root := t.TempDir()
	b := &Builder{}
	idx, err := b.BuildFromDirs(root)
	require.NoError(t, err)
	assert.Empty(t, idx.Modules)
	assert.Equal(t, IndexVersion, idx.Version)
}

func TestBuildFromDirs_Good_NonexistentDir(t *testing.T) {
	b := &Builder{}
	idx, err := b.BuildFromDirs("/nonexistent/path")
	require.NoError(t, err)
	assert.Empty(t, idx.Modules)
}

func TestBuildFromDirs_Good_NoRepoURLWithoutConfig(t *testing.T) {
	root := t.TempDir()
	modDir := filepath.Join(root, "mod")
	require.NoError(t, os.MkdirAll(modDir, 0755))
	writeManifestYAML(t, modDir, "mod", "Module", "1.0.0")

	b := &Builder{} // No BaseURL or Org.
	idx, err := b.BuildFromDirs(root)
	require.NoError(t, err)
	assert.Empty(t, idx.Modules[0].Repo)
}

func TestBuildFromManifests_Good(t *testing.T) {
	manifests := []*manifest.Manifest{
		{Code: "bravo", Name: "Bravo"},
		{Code: "alpha", Name: "Alpha"},
	}
	idx := BuildFromManifests(manifests)
	require.Len(t, idx.Modules, 2)
	assert.Equal(t, "alpha", idx.Modules[0].Code)
	assert.Equal(t, "bravo", idx.Modules[1].Code)
	assert.Equal(t, IndexVersion, idx.Version)
}

func TestBuildFromManifests_Good_SkipsNil(t *testing.T) {
	manifests := []*manifest.Manifest{
		nil,
		{Code: "valid", Name: "Valid"},
		{Code: "", Name: "Empty Code"},
	}
	idx := BuildFromManifests(manifests)
	assert.Len(t, idx.Modules, 1)
	assert.Equal(t, "valid", idx.Modules[0].Code)
}

func TestBuildFromManifests_Good_Deduplicates(t *testing.T) {
	manifests := []*manifest.Manifest{
		{Code: "dup", Name: "First"},
		{Code: "dup", Name: "Second"},
	}
	idx := BuildFromManifests(manifests)
	assert.Len(t, idx.Modules, 1)
}

func TestWriteIndex_Good(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "marketplace", "index.json")

	idx := &Index{
		Version: 1,
		Modules: []Module{
			{Code: "test-mod", Name: "Test Module"},
		},
	}

	err := WriteIndex(coreio.Local, path, idx)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var parsed Index
	require.NoError(t, json.Unmarshal(data, &parsed))
	assert.Len(t, parsed.Modules, 1)
	assert.Equal(t, "test-mod", parsed.Modules[0].Code)
}

func TestWriteIndex_Good_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.json")

	root := t.TempDir()
	modDir := filepath.Join(root, "roundtrip")
	require.NoError(t, os.MkdirAll(modDir, 0755))
	writeManifestYAML(t, modDir, "roundtrip", "Roundtrip Module", "3.0.0")

	b := &Builder{BaseURL: "https://forge.lthn.ai", Org: "core"}
	idx, err := b.BuildFromDirs(root)
	require.NoError(t, err)

	require.NoError(t, WriteIndex(coreio.Local, path, idx))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	parsed, err := ParseIndex(data)
	require.NoError(t, err)

	require.Len(t, parsed.Modules, 1)
	assert.Equal(t, "roundtrip", parsed.Modules[0].Code)
	assert.Equal(t, "https://forge.lthn.ai/core/roundtrip", parsed.Modules[0].Repo)
}
