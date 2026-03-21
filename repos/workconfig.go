package repos

import (
	"path/filepath"
	"time"

	coreerr "dappco.re/go/core/log"
	"dappco.re/go/core/io"
	"gopkg.in/yaml.v3"
)

// WorkConfig holds sync policy for a workspace.
// Stored at .core/work.yaml and checked into git (shared across the team).
type WorkConfig struct {
	Version  int              `yaml:"version"`
	Sync     SyncConfig       `yaml:"sync"`
	Agents   AgentPolicy      `yaml:"agents"`
	Triggers []string         `yaml:"triggers,omitempty"`
}

// SyncConfig controls how and when repos are synced.
type SyncConfig struct {
	Interval     time.Duration `yaml:"interval"`
	AutoPull     bool          `yaml:"auto_pull"`
	AutoPush     bool          `yaml:"auto_push"`
	CloneMissing bool          `yaml:"clone_missing"`
}

// AgentPolicy controls multi-agent clash prevention.
type AgentPolicy struct {
	Heartbeat      time.Duration `yaml:"heartbeat"`
	StaleAfter     time.Duration `yaml:"stale_after"`
	WarnOnOverlap  bool          `yaml:"warn_on_overlap"`
}

// DefaultWorkConfig returns sensible defaults for workspace sync.
func DefaultWorkConfig() *WorkConfig {
	return &WorkConfig{
		Version: 1,
		Sync: SyncConfig{
			Interval:     5 * time.Minute,
			AutoPull:     true,
			AutoPush:     false,
			CloneMissing: true,
		},
		Agents: AgentPolicy{
			Heartbeat:     2 * time.Minute,
			StaleAfter:    10 * time.Minute,
			WarnOnOverlap: true,
		},
		Triggers: []string{"on_activate", "on_commit", "scheduled"},
	}
}

// LoadWorkConfig reads .core/work.yaml from the given workspace root directory.
// Returns defaults if the file does not exist.
func LoadWorkConfig(m io.Medium, root string) (*WorkConfig, error) {
	path := filepath.Join(root, ".core", "work.yaml")

	if !m.Exists(path) {
		return DefaultWorkConfig(), nil
	}

	content, err := m.Read(path)
	if err != nil {
		return nil, coreerr.E("repos.LoadWorkConfig", "failed to read work config", err)
	}

	wc := DefaultWorkConfig()
	if err := yaml.Unmarshal([]byte(content), wc); err != nil {
		return nil, coreerr.E("repos.LoadWorkConfig", "failed to parse work config", err)
	}

	return wc, nil
}

// SaveWorkConfig writes .core/work.yaml to the given workspace root directory.
func SaveWorkConfig(m io.Medium, root string, wc *WorkConfig) error {
	coreDir := filepath.Join(root, ".core")
	if err := m.EnsureDir(coreDir); err != nil {
		return coreerr.E("repos.SaveWorkConfig", "failed to create .core directory", err)
	}

	data, err := yaml.Marshal(wc)
	if err != nil {
		return coreerr.E("repos.SaveWorkConfig", "failed to marshal work config", err)
	}

	path := filepath.Join(coreDir, "work.yaml")
	if err := m.Write(path, string(data)); err != nil {
		return coreerr.E("repos.SaveWorkConfig", "failed to write work config", err)
	}

	return nil
}

// HasTrigger returns true if the given trigger name is in the triggers list.
func (wc *WorkConfig) HasTrigger(name string) bool {
	for _, t := range wc.Triggers {
		if t == name {
			return true
		}
	}
	return false
}
