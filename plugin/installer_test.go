// SPDX-License-Identifier: EUPL-1.2

package plugin

import (
	"context"
	"testing"

	"dappco.re/go/core/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── NewInstaller ───────────────────────────────────────────────────

func TestNewInstaller_Good(t *testing.T) {
	m := io.NewMockMedium()
	reg := NewRegistry(m, "/plugins")
	inst := NewInstaller(m, reg)

	assert.NotNil(t, inst)
	assert.Equal(t, m, inst.medium)
	assert.Equal(t, reg, inst.registry)
}

// ── Install error paths ────────────────────────────────────────────

func TestInstall_InvalidSource_Bad(t *testing.T) {
	m := io.NewMockMedium()
	reg := NewRegistry(m, "/plugins")
	inst := NewInstaller(m, reg)

	err := inst.Install(context.Background(), "bad-source")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid source")
}

func TestInstall_AlreadyInstalled_Bad(t *testing.T) {
	m := io.NewMockMedium()
	reg := NewRegistry(m, "/plugins")
	_ = reg.Add(&PluginConfig{Name: "my-plugin", Version: "1.0.0"})

	inst := NewInstaller(m, reg)
	err := inst.Install(context.Background(), "org/my-plugin")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already installed")
}

func TestInstall_Bad_PathTraversalSource(t *testing.T) {
	m := io.NewMockMedium()
	reg := NewRegistry(m, "/plugins")
	inst := NewInstaller(m, reg)

	err := inst.Install(context.Background(), "../repo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid source")
	assert.False(t, m.Exists("/repo"))
}

// ── Remove ─────────────────────────────────────────────────────────

func TestRemove_Good(t *testing.T) {
	m := io.NewMockMedium()
	reg := NewRegistry(m, "/plugins")
	_ = reg.Add(&PluginConfig{Name: "removable", Version: "1.0.0"})

	// Create plugin directory.
	_ = m.EnsureDir("/plugins/removable")
	_ = m.Write("/plugins/removable/plugin.json", `{"name":"removable"}`)

	inst := NewInstaller(m, reg)
	err := inst.Remove("removable")
	require.NoError(t, err)

	// Plugin removed from registry.
	_, ok := reg.Get("removable")
	assert.False(t, ok)

	// Directory cleaned up.
	assert.False(t, m.Exists("/plugins/removable"))
}

func TestRemove_DirAlreadyGone_Good(t *testing.T) {
	m := io.NewMockMedium()
	reg := NewRegistry(m, "/plugins")
	_ = reg.Add(&PluginConfig{Name: "ghost", Version: "1.0.0"})
	// No directory exists — should still succeed.

	inst := NewInstaller(m, reg)
	err := inst.Remove("ghost")
	require.NoError(t, err)

	_, ok := reg.Get("ghost")
	assert.False(t, ok)
}

func TestRemove_NotFound_Bad(t *testing.T) {
	m := io.NewMockMedium()
	reg := NewRegistry(m, "/plugins")
	inst := NewInstaller(m, reg)

	err := inst.Remove("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin not found")
}

func TestRemove_Bad_PathTraversalName(t *testing.T) {
	m := io.NewMockMedium()
	reg := NewRegistry(m, "/plugins")
	_ = reg.Add(&PluginConfig{Name: "safe", Version: "1.0.0"})
	_ = m.EnsureDir("/escape")

	inst := NewInstaller(m, reg)
	err := inst.Remove("../escape")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid plugin name")
	assert.True(t, m.Exists("/escape"))
}

// ── Update error paths ─────────────────────────────────────────────

func TestUpdate_NotFound_Bad(t *testing.T) {
	m := io.NewMockMedium()
	reg := NewRegistry(m, "/plugins")
	inst := NewInstaller(m, reg)

	err := inst.Update(context.Background(), "missing")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin not found")
}

// ── ParseSource ────────────────────────────────────────────────────

func TestParseSource_OrgRepo_Good(t *testing.T) {
	org, repo, version, err := ParseSource("host-uk/core-plugin")
	assert.NoError(t, err)
	assert.Equal(t, "host-uk", org)
	assert.Equal(t, "core-plugin", repo)
	assert.Equal(t, "", version)
}

func TestParseSource_OrgRepoVersion_Good(t *testing.T) {
	org, repo, version, err := ParseSource("host-uk/core-plugin@v1.0.0")
	assert.NoError(t, err)
	assert.Equal(t, "host-uk", org)
	assert.Equal(t, "core-plugin", repo)
	assert.Equal(t, "v1.0.0", version)
}

func TestParseSource_VersionWithoutPrefix_Good(t *testing.T) {
	org, repo, version, err := ParseSource("org/repo@1.2.3")
	assert.NoError(t, err)
	assert.Equal(t, "org", org)
	assert.Equal(t, "repo", repo)
	assert.Equal(t, "1.2.3", version)
}

func TestParseSource_Empty_Bad(t *testing.T) {
	_, _, _, err := ParseSource("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source is empty")
}

func TestParseSource_NoSlash_Bad(t *testing.T) {
	_, _, _, err := ParseSource("just-a-name")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "org/repo")
}

func TestParseSource_TooManySlashes_Bad(t *testing.T) {
	_, _, _, err := ParseSource("a/b/c")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "org/repo")
}

func TestParseSource_EmptyOrg_Bad(t *testing.T) {
	_, _, _, err := ParseSource("/repo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "org/repo")
}

func TestParseSource_EmptyRepo_Bad(t *testing.T) {
	_, _, _, err := ParseSource("org/")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "org/repo")
}

func TestParseSource_EmptyVersion_Bad(t *testing.T) {
	_, _, _, err := ParseSource("org/repo@")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "version is empty")
}

func TestParseSource_Bad_PathTraversal(t *testing.T) {
	_, _, _, err := ParseSource("org/../repo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "org/repo")
}

func TestParseSource_Bad_PathTraversalEncoded(t *testing.T) {
	_, _, _, err := ParseSource("org%2f..%2frepo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "org/repo")
}

func TestInstall_Bad_EncodedPathTraversal(t *testing.T) {
	m := io.NewMockMedium()
	reg := NewRegistry(m, "/plugins")
	inst := NewInstaller(m, reg)

	err := inst.Install(context.Background(), "org%2f..%2frepo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid source")
}
