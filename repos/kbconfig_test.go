// SPDX-License-Identifier: EUPL-1.2

package repos

import (
	"testing"

	"dappco.re/go/core/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── DefaultKBConfig ────────────────────────────────────────────────

func TestDefaultKBConfig_Good(t *testing.T) {
	kb := DefaultKBConfig()
	assert.Equal(t, 1, kb.Version)
	assert.True(t, kb.Wiki.Enabled)
	assert.Equal(t, "kb/wiki", kb.Wiki.Dir)
	assert.Contains(t, kb.Wiki.Remote, "forge.lthn.ai")
	assert.Equal(t, "qdrant.lthn.sh", kb.Search.QdrantHost)
	assert.Equal(t, 6334, kb.Search.QdrantPort)
	assert.Equal(t, "openbrain", kb.Search.Collection)
	assert.Equal(t, "embeddinggemma", kb.Search.EmbedModel)
	assert.Equal(t, 5, kb.Search.TopK)
}

// ── Load / Save round-trip ─────────────────────────────────────────

func TestKBConfig_LoadSave_Good(t *testing.T) {
	m := io.NewMockMedium()
	_ = m.EnsureDir("/workspace/.core")

	kb := DefaultKBConfig()
	kb.Search.TopK = 10
	kb.Search.Collection = "custom_brain"

	err := SaveKBConfig(m, "/workspace", kb)
	require.NoError(t, err)

	loaded, err := LoadKBConfig(m, "/workspace")
	require.NoError(t, err)

	assert.Equal(t, 1, loaded.Version)
	assert.Equal(t, 10, loaded.Search.TopK)
	assert.Equal(t, "custom_brain", loaded.Search.Collection)
	assert.True(t, loaded.Wiki.Enabled)
}

func TestKBConfig_Load_Good_NoFile(t *testing.T) {
	m := io.NewMockMedium()
	_ = m.EnsureDir("/workspace/.core")

	kb, err := LoadKBConfig(m, "/workspace")
	require.NoError(t, err)
	assert.Equal(t, DefaultKBConfig().Search.Collection, kb.Search.Collection)
}

func TestKBConfig_Load_Good_PartialOverride(t *testing.T) {
	m := io.NewMockMedium()
	_ = m.Write("/workspace/.core/kb.yaml", `
version: 1
search:
  top_k: 20
  collection: my_brain
`)

	kb, err := LoadKBConfig(m, "/workspace")
	require.NoError(t, err)

	assert.Equal(t, 20, kb.Search.TopK)
	assert.Equal(t, "my_brain", kb.Search.Collection)
	// Defaults preserved
	assert.True(t, kb.Wiki.Enabled)
	assert.Equal(t, "embeddinggemma", kb.Search.EmbedModel)
}

func TestKBConfig_Load_Bad_InvalidYAML(t *testing.T) {
	m := io.NewMockMedium()
	_ = m.Write("/workspace/.core/kb.yaml", "{{{{broken")

	_, err := LoadKBConfig(m, "/workspace")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}

// ── WikiRepoURL ────────────────────────────────────────────────────

func TestKBConfig_WikiRepoURL_Good(t *testing.T) {
	kb := DefaultKBConfig()
	url := kb.WikiRepoURL("go-scm")
	assert.Equal(t, "ssh://git@forge.lthn.ai:2223/core/go-scm.wiki.git", url)
}

func TestKBConfig_WikiRepoURL_Good_CustomRemote(t *testing.T) {
	kb := &KBConfig{
		Wiki: WikiConfig{Remote: "ssh://git@git.example.com/org"},
	}
	url := kb.WikiRepoURL("my-repo")
	assert.Equal(t, "ssh://git@git.example.com/org/my-repo.wiki.git", url)
}

// ── WikiLocalPath ──────────────────────────────────────────────────

func TestKBConfig_WikiLocalPath_Good(t *testing.T) {
	kb := DefaultKBConfig()
	path := kb.WikiLocalPath("/workspace", "go-scm")
	assert.Equal(t, "/workspace/.core/kb/wiki/go-scm", path)
}
