// SPDX-License-Identifier: EUPL-1.2

package config

import (
	// Note: errors.New is retained because core/config is the local replacement for dappco.re/go/config and cannot depend on downstream core helpers.
	`errors`
	// Note: os filesystem calls are retained because core/config persists its own backing file before higher-level core filesystem APIs are available.
	`os`
	// Note: filepath is retained for OS-specific config path assembly in this standalone config module.
	`path/filepath`
	// Note: strings.Split is retained for dotted config-key traversal without adding a downstream core dependency.
	`strings`
	// Note: sync protects the in-memory config store and has no core equivalent in this low-level module.
	"sync"

	"gopkg.in/yaml.v3"
)

// Config is a small YAML-backed configuration store.
//
// It intentionally mirrors the subset of the Core config API needed by go-scm.
type Config struct {
	mu   sync.RWMutex
	Path string
	data map[string]any
}

// Option configures a Config instance.
type Option func(*Config)

// WithPath sets the backing file path.
func WithPath(path string) Option {
	return func(c *Config) {
		c.Path = path
	}
}

// New creates an empty config store.
func New(opts ...Option) (*Config, error)  /* v090-result-boundary */ {
	cfg := &Config{data: make(map[string]any)}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg, nil
}

// NewWithPath creates an empty config store with a backing file path.
func NewWithPath(path string) *Config {
	return &Config{Path: path, data: make(map[string]any)}
}

func (c *Config) ensure() {
	if c.data == nil {
		c.data = make(map[string]any)
	}
}

// Set stores a dotted key path into the config map.
func (c *Config) Set(key string, v any) error  /* v090-result-boundary */ {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ensure()
	if key == "" {
		if m, ok := v.(map[string]any); ok {
			c.data = cloneMap(m)
			return nil
		}
		return errors.New("config.Set: empty key requires map value")
	}

	parts := strings.Split(key, ".")
	cur := c.data
	for i, part := range parts {
		if part == "" {
			return errors.New("config.Set: empty key segment")
		}
		if i == len(parts)-1 {
			cur[part] = v
			return nil
		}
		next, ok := cur[part]
		if !ok {
			child := make(map[string]any)
			cur[part] = child
			cur = child
			continue
		}
		child, ok := next.(map[string]any)
		if !ok {
			child = make(map[string]any)
			cur[part] = child
		}
		cur = child
	}
	return nil
}

// Get unmarshals a config value into out.
func (c *Config) Get(key string, out any) error  /* v090-result-boundary */ {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if out == nil {
		return errors.New("config.Get: output is required")
	}
	if c.data == nil {
		return errors.New("config.Get: key not found")
	}

	var target any
	if key == "" {
		target = c.data
	} else {
		var ok bool
		target, ok = lookup(c.data, strings.Split(key, "."))
		if !ok {
			return errors.New("config.Get: key not found")
		}
	}

	raw, err := yaml.Marshal(target)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(raw, out)
}

// Commit writes the config to disk as YAML.
func (c *Config) Commit() error  /* v090-result-boundary */ {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.Path == "" {
		return errors.New("config.Commit: path is required")
	}
	if err := os.MkdirAll(filepath.Dir(c.Path), 0o755); err != nil {
		return err
	}
	raw, err := yaml.Marshal(c.data)
	if err != nil {
		return err
	}
	return os.WriteFile(c.Path, raw, 0o600)
}

func lookup(m map[string]any, parts []string) (any, bool) {
	cur := any(m)
	for _, part := range parts {
		if part == "" {
			return nil, false
		}
		obj, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		next, ok := obj[part]
		if !ok {
			return nil, false
		}
		cur = next
	}
	return cur, true
}

func cloneMap(src map[string]any) map[string]any {
	dst := make(map[string]any, len(src))
	for k, v := range src {
		if child, ok := v.(map[string]any); ok {
			dst[k] = cloneMap(child)
			continue
		}
		dst[k] = v
	}
	return dst
}
