// SPDX-License-Identifier: EUPL-1.2

package repos

import (
	"errors"
	"path/filepath"
	"strings"

	coreio "dappco.re/go/core/io"
	"dappco.re/go/scm/internal/ax/filepathx"
	"gopkg.in/yaml.v3"
)

type WikiConfig struct {
	Enabled bool   `yaml:"enabled"`
	Dir     string `yaml:"dir"`
	Remote  string `yaml:"remote"`
}

type KBSearch struct {
	QdrantHost string `yaml:"qdrant_host"`
	QdrantPort int    `yaml:"qdrant_port"`
	Collection string `yaml:"collection"`
	OllamaURL  string `yaml:"ollama_url"`
	EmbedModel string `yaml:"embed_model"`
	TopK       int    `yaml:"top_k"`
}

type KBConfig struct {
	Version int        `yaml:"version"`
	Wiki    WikiConfig `yaml:"wiki"`
	Search  KBSearch   `yaml:"search"`
}

func DefaultKBConfig() *KBConfig { return &KBConfig{Version: 1, Wiki: WikiConfig{Dir: ".core/wiki"}} }

func (kb *KBConfig) WikiRepoURL(repoName string) string {
	if kb == nil || kb.Wiki.Remote == "" || repoName == "" {
		return ""
	}
	return strings.TrimRight(kb.Wiki.Remote, "/") + "/" + repoName + ".wiki.git"
}

func (kb *KBConfig) WikiLocalPath(root, repoName string) string {
	dir := ".core/wiki"
	if kb != nil && kb.Wiki.Dir != "" {
		dir = kb.Wiki.Dir
	}
	return filepath.Join(root, dir, repoName)
}

func LoadKBConfig(m coreio.Medium, root string) (*KBConfig, error) {
	if m == nil {
		return nil, errors.New("repos.LoadKBConfig: medium is required")
	}
	raw, err := m.Read(filepathx.Join(root, ".core", "kb.yaml"))
	if err != nil {
		return DefaultKBConfig(), nil
	}
	var kb KBConfig
	if err := yaml.Unmarshal([]byte(raw), &kb); err != nil {
		return nil, err
	}
	if kb.Version == 0 {
		kb.Version = 1
	}
	return &kb, nil
}

func SaveKBConfig(m coreio.Medium, root string, kb *KBConfig) error {
	if m == nil {
		return errors.New("repos.SaveKBConfig: medium is required")
	}
	if kb == nil {
		kb = DefaultKBConfig()
	}
	raw, err := yaml.Marshal(kb)
	if err != nil {
		return err
	}
	return m.Write(filepathx.Join(root, ".core", "kb.yaml"), string(raw))
}
