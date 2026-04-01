// SPDX-License-Identifier: EUPL-1.2

package marketplace

import (
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
	json "dappco.re/go/core/scm/internal/ax/jsonx"
	os "dappco.re/go/core/scm/internal/ax/osx"
	"testing"

	"dappco.re/go/core/io"
	"dappco.re/go/core/scm/manifest"
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

// writeManifestYAMLWithSign writes a .core/manifest.yaml with a signing key.
func writeManifestYAMLWithSign(t *testing.T, dir, code, name, version, sign string) {
	t.Helper()
	coreDir := filepath.Join(dir, ".core")
	require.NoError(t, os.MkdirAll(coreDir, 0755))
	yaml := "code: " + code + "\nname: " + name + "\nversion: " + version + "\nsign: " + sign + "\n"
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

func TestBuildFromDirs_Good_ManifestYAML_Good(t *testing.T) {
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
	assert.Equal(t, "https://forge.lthn.ai/core/my-widget.git", idx.Modules[0].Repo)
	assert.Equal(t, IndexVersion, idx.Version)
}

func TestBuildFromDirs_Good_IndexesRootDirectory_Good(t *testing.T) {
	root := t.TempDir()
	writeManifestYAML(t, root, "root-mod", "Root Module", "1.0.0")

	b := &Builder{BaseURL: "https://forge.lthn.ai", Org: "core"}
	idx, err := b.BuildFromDirs(root)
	require.NoError(t, err)

	require.Len(t, idx.Modules, 1)
	assert.Equal(t, "root-mod", idx.Modules[0].Code)
	assert.Equal(t, "Root Module", idx.Modules[0].Name)
	assert.Equal(t, "https://forge.lthn.ai/core/root-mod.git", idx.Modules[0].Repo)
}

func TestBuildFromDirs_Good_CarriesSignKey_Good(t *testing.T) {
	root := t.TempDir()
	modDir := filepath.Join(root, "signed-mod")
	require.NoError(t, os.MkdirAll(modDir, 0755))
	writeManifestYAMLWithSign(t, modDir, "signed-mod", "Signed Module", "1.0.0", "abc123")

	b := &Builder{}
	idx, err := b.BuildFromDirs(root)
	require.NoError(t, err)

	require.Len(t, idx.Modules, 1)
	assert.Equal(t, "abc123", idx.Modules[0].SignKey)
}

func TestBuildFromDirs_Good_CoreJSON_Good(t *testing.T) {
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

func TestBuildFromDirs_Good_PrefersCompiledOverSource_Good(t *testing.T) {
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

func TestBuildFromDirs_Good_SkipsNoManifest_Good(t *testing.T) {
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

func TestBuildFromDirs_Good_Deduplicates_Good(t *testing.T) {
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

func TestBuildFromDirs_Good_SortsByCode_Good(t *testing.T) {
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

func TestBuildFromDirs_Good_EmptyDir_Good(t *testing.T) {
	root := t.TempDir()
	b := &Builder{}
	idx, err := b.BuildFromDirs(root)
	require.NoError(t, err)
	assert.Empty(t, idx.Modules)
	assert.Equal(t, IndexVersion, idx.Version)
}

func TestBuildFromDirs_Good_NonexistentDir_Good(t *testing.T) {
	b := &Builder{}
	idx, err := b.BuildFromDirs("/nonexistent/path")
	require.NoError(t, err)
	assert.Empty(t, idx.Modules)
}

func TestBuildFromDirs_Good_NoRepoURLWithoutConfig_Good(t *testing.T) {
	root := t.TempDir()
	modDir := filepath.Join(root, "mod")
	require.NoError(t, os.MkdirAll(modDir, 0755))
	writeManifestYAML(t, modDir, "mod", "Module", "1.0.0")

	b := &Builder{} // No BaseURL or Org.
	idx, err := b.BuildFromDirs(root)
	require.NoError(t, err)
	assert.Empty(t, idx.Modules[0].Repo)
}

func TestBuildFromDirs_Good_DefaultForgeURL_Good(t *testing.T) {
	root := t.TempDir()
	modDir := filepath.Join(root, "mod")
	require.NoError(t, os.MkdirAll(modDir, 0755))
	writeManifestYAML(t, modDir, "mod", "Module", "1.0.0")

	b := &Builder{Org: "core"}
	idx, err := b.BuildFromDirs(root)
	require.NoError(t, err)

	require.Len(t, idx.Modules, 1)
	assert.Equal(t, "https://forge.lthn.ai/core/mod.git", idx.Modules[0].Repo)
}

func TestBuildFromManifests_Good(t *testing.T) {
	manifests := []*manifest.Manifest{
		{Code: "bravo", Name: "Bravo", Sign: "key-bravo"},
		{Code: "alpha", Name: "Alpha", Sign: "key-alpha"},
	}
	idx := BuildFromManifests(manifests)
	require.Len(t, idx.Modules, 2)
	assert.Equal(t, "alpha", idx.Modules[0].Code)
	assert.Equal(t, "bravo", idx.Modules[1].Code)
	assert.Equal(t, IndexVersion, idx.Version)
	assert.Equal(t, "key-alpha", idx.Modules[0].SignKey)
	assert.Equal(t, "key-bravo", idx.Modules[1].SignKey)
}

func TestBuildFromManifests_Good_SkipsNil_Good(t *testing.T) {
	manifests := []*manifest.Manifest{
		nil,
		{Code: "valid", Name: "Valid"},
		{Code: "", Name: "Empty Code"},
	}
	idx := BuildFromManifests(manifests)
	assert.Len(t, idx.Modules, 1)
	assert.Equal(t, "valid", idx.Modules[0].Code)
}

func TestBuildFromManifests_Good_Deduplicates_Good(t *testing.T) {
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

	err := WriteIndex(path, idx)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var parsed Index
	require.NoError(t, json.Unmarshal(data, &parsed))
	assert.Len(t, parsed.Modules, 1)
	assert.Equal(t, "test-mod", parsed.Modules[0].Code)
}

func TestWriteIndex_Good_RoundTrip_Good(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.json")

	root := t.TempDir()
	modDir := filepath.Join(root, "roundtrip")
	require.NoError(t, os.MkdirAll(modDir, 0755))
	writeManifestYAML(t, modDir, "roundtrip", "Roundtrip Module", "3.0.0")

	b := &Builder{BaseURL: "https://forge.lthn.ai", Org: "core"}
	idx, err := b.BuildFromDirs(root)
	require.NoError(t, err)

	require.NoError(t, WriteIndex(path, idx))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	parsed, err := ParseIndex(data)
	require.NoError(t, err)

	require.Len(t, parsed.Modules, 1)
	assert.Equal(t, "roundtrip", parsed.Modules[0].Code)
	assert.Equal(t, "https://forge.lthn.ai/core/roundtrip.git", parsed.Modules[0].Repo)
}

func TestLoadIndex_Good(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.json")

	idx := &Index{
		Version: 1,
		Modules: []Module{
			{Code: "refresh", Name: "Refresh Module"},
		},
	}
	require.NoError(t, WriteIndex(path, idx))

	loaded, err := LoadIndex(io.Local, path)
	require.NoError(t, err)
	require.Len(t, loaded.Modules, 1)
	assert.Equal(t, "refresh", loaded.Modules[0].Code)
}

func TestLoadIndex_Bad_NotFound_Good(t *testing.T) {
	_, err := LoadIndex(io.Local, "/missing/index.json")
	require.Error(t, err)
}
