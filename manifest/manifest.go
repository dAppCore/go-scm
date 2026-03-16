package manifest

import (
	coreerr "forge.lthn.ai/core/go-log"
	"gopkg.in/yaml.v3"
)

// Manifest represents a .core/manifest.yaml application manifest.
type Manifest struct {
	Code        string            `yaml:"code" json:"code"`
	Name        string            `yaml:"name" json:"name"`
	Description string            `yaml:"description,omitempty" json:"description,omitempty"`
	Version     string            `yaml:"version" json:"version"`
	Author      string            `yaml:"author,omitempty" json:"author,omitempty"`
	Licence     string            `yaml:"licence,omitempty" json:"licence,omitempty"`
	Sign        string            `yaml:"sign,omitempty" json:"sign,omitempty"`
	Layout      string            `yaml:"layout,omitempty" json:"layout,omitempty"`
	Slots       map[string]string `yaml:"slots,omitempty" json:"slots,omitempty"`

	// Provider fields — used by runtime provider loading.
	Namespace string       `yaml:"namespace,omitempty" json:"namespace,omitempty"` // API route prefix, e.g. /api/v1/cool-widget
	Port      int          `yaml:"port,omitempty" json:"port,omitempty"`           // Listen port (0 = auto-assign)
	Binary    string       `yaml:"binary,omitempty" json:"binary,omitempty"`       // Path to provider binary (relative to provider dir)
	Args      []string     `yaml:"args,omitempty" json:"args,omitempty"`           // Additional CLI args for the binary
	Element   *ElementSpec `yaml:"element,omitempty" json:"element,omitempty"`     // Custom element for GUI rendering
	Spec      string       `yaml:"spec,omitempty" json:"spec,omitempty"`           // Path to OpenAPI spec file

	Permissions Permissions            `yaml:"permissions,omitempty" json:"permissions,omitempty"`
	Modules     []string               `yaml:"modules,omitempty" json:"modules,omitempty"`
	Daemons     map[string]DaemonSpec  `yaml:"daemons,omitempty" json:"daemons,omitempty"`
}

// ElementSpec describes a web component for GUI rendering.
type ElementSpec struct {
	// Tag is the custom element tag name, e.g. "core-cool-widget".
	Tag string `yaml:"tag" json:"tag"`

	// Source is the path to the JS bundle (relative to provider dir).
	Source string `yaml:"source" json:"source"`
}

// IsProvider returns true if this manifest declares provider fields
// (namespace and binary), indicating it is a runtime provider.
func (m *Manifest) IsProvider() bool {
	return m.Namespace != "" && m.Binary != ""
}

// Permissions declares the I/O capabilities a module requires.
type Permissions struct {
	Read  []string `yaml:"read" json:"read"`
	Write []string `yaml:"write" json:"write"`
	Net   []string `yaml:"net" json:"net"`
	Run   []string `yaml:"run" json:"run"`
}

// DaemonSpec describes a long-running process managed by the runtime.
type DaemonSpec struct {
	Binary  string   `yaml:"binary,omitempty" json:"binary,omitempty"`
	Args    []string `yaml:"args,omitempty" json:"args,omitempty"`
	Health  string   `yaml:"health,omitempty" json:"health,omitempty"`
	Default bool     `yaml:"default,omitempty" json:"default,omitempty"`
}

// Parse decodes YAML bytes into a Manifest.
func Parse(data []byte) (*Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, coreerr.E("manifest.Parse", "unmarshal failed", err)
	}
	return &m, nil
}

// SlotNames returns a deduplicated list of component names from slots.
func (m *Manifest) SlotNames() []string {
	seen := make(map[string]bool)
	var names []string
	for _, name := range m.Slots {
		if !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	}
	return names
}

// DefaultDaemon returns the name, spec, and true for the default daemon.
// A daemon is the default if it has Default:true, or if it is the only daemon
// in the map. If multiple daemons have Default:true, returns false (ambiguous).
// Returns empty values and false if no default can be determined.
func (m *Manifest) DefaultDaemon() (string, DaemonSpec, bool) {
	if len(m.Daemons) == 0 {
		return "", DaemonSpec{}, false
	}

	// Look for an explicit default; reject ambiguous multiple defaults.
	var defaultName string
	var defaultSpec DaemonSpec
	for name, spec := range m.Daemons {
		if spec.Default {
			if defaultName != "" {
				// Multiple defaults — ambiguous.
				return "", DaemonSpec{}, false
			}
			defaultName = name
			defaultSpec = spec
		}
	}
	if defaultName != "" {
		return defaultName, defaultSpec, true
	}

	// If exactly one daemon exists, treat it as the implicit default.
	if len(m.Daemons) == 1 {
		for name, spec := range m.Daemons {
			return name, spec, true
		}
	}

	return "", DaemonSpec{}, false
}
