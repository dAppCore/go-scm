// SPDX-License-Identifier: EUPL-1.2

package repos

import (
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
	fmt "dappco.re/go/core/scm/internal/ax/fmtx"

	"dappco.re/go/core/io"
	coreerr "dappco.re/go/core/log"
	"gopkg.in/yaml.v3"
)

// KBConfig holds knowledge base configuration for a workspace.
// Stored at .core/kb.yaml and checked into git.
type KBConfig struct {
	Version int        `yaml:"version"`
	Wiki    WikiConfig `yaml:"wiki"`
	Search  KBSearch   `yaml:"search"`
}

// WikiConfig controls local wiki mirror behaviour.
type WikiConfig struct {
	// Enabled toggles wiki cloning on sync.
	Enabled bool `yaml:"enabled"`
	// Dir is the local directory for wiki clones, relative to .core/.
	Dir string `yaml:"dir"`
	// Remote is the SSH base URL for wiki repos (e.g. ssh://git@forge.lthn.ai:2223/core).
	// Repo wikis are at {Remote}/{name}.wiki.git
	Remote string `yaml:"remote"`
}

// KBSearch configures vector search against the OpenBrain Qdrant collection.
type KBSearch struct {
	// QdrantHost is the Qdrant server (gRPC).
	QdrantHost string `yaml:"qdrant_host"`
	// QdrantPort is the gRPC port.
	QdrantPort int `yaml:"qdrant_port"`
	// Collection is the Qdrant collection name.
	Collection string `yaml:"collection"`
	// OllamaURL is the Ollama API base URL for embedding queries.
	OllamaURL string `yaml:"ollama_url"`
	// EmbedModel is the Ollama model for embedding.
	EmbedModel string `yaml:"embed_model"`
	// TopK is the default number of results.
	TopK int `yaml:"top_k"`
}

// DefaultKBConfig returns sensible defaults for knowledge base config.
// Usage: DefaultKBConfig(...)
func DefaultKBConfig() *KBConfig {
	return &KBConfig{
		Version: 1,
		Wiki: WikiConfig{
			Enabled: true,
			Dir:     "kb/wiki",
			Remote:  "ssh://git@forge.lthn.ai:2223/core",
		},
		Search: KBSearch{
			QdrantHost: "qdrant.lthn.sh",
			QdrantPort: 6334,
			Collection: "openbrain",
			OllamaURL:  "https://ollama.lthn.sh",
			EmbedModel: "embeddinggemma",
			TopK:       5,
		},
	}
}

// LoadKBConfig reads .core/kb.yaml from the given workspace root directory.
// Returns defaults if the file does not exist.
// Usage: LoadKBConfig(...)
func LoadKBConfig(m io.Medium, root string) (*KBConfig, error) {
	path := filepath.Join(root, ".core", "kb.yaml")

	if !m.Exists(path) {
		return DefaultKBConfig(), nil
	}

	content, err := m.Read(path)
	if err != nil {
		return nil, coreerr.E("repos.LoadKBConfig", "failed to read kb config", err)
	}

	kb := DefaultKBConfig()
	if err := yaml.Unmarshal([]byte(content), kb); err != nil {
		return nil, coreerr.E("repos.LoadKBConfig", "failed to parse kb config", err)
	}

	return kb, nil
}

// SaveKBConfig writes .core/kb.yaml to the given workspace root directory.
// Usage: SaveKBConfig(...)
func SaveKBConfig(m io.Medium, root string, kb *KBConfig) error {
	coreDir := filepath.Join(root, ".core")
	if err := m.EnsureDir(coreDir); err != nil {
		return coreerr.E("repos.SaveKBConfig", "failed to create .core directory", err)
	}

	data, err := yaml.Marshal(kb)
	if err != nil {
		return coreerr.E("repos.SaveKBConfig", "failed to marshal kb config", err)
	}

	path := filepath.Join(coreDir, "kb.yaml")
	if err := m.Write(path, string(data)); err != nil {
		return coreerr.E("repos.SaveKBConfig", "failed to write kb config", err)
	}

	return nil
}

// WikiRepoURL returns the full clone URL for a repo's wiki.
// Usage: WikiRepoURL(...)
func (kb *KBConfig) WikiRepoURL(repoName string) string {
	return fmt.Sprintf("%s/%s.wiki.git", kb.Wiki.Remote, repoName)
}

// WikiLocalPath returns the local path for a repo's wiki clone.
// Usage: WikiLocalPath(...)
func (kb *KBConfig) WikiLocalPath(root, repoName string) string {
	return filepath.Join(root, ".core", kb.Wiki.Dir, repoName)
}
