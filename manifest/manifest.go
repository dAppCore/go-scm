// SPDX-License-Identifier: EUPL-1.2

package manifest

import (
	"errors"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type Permissions struct {
	Read  []string `yaml:"read" json:"read"`
	Write []string `yaml:"write" json:"write"`
	Net   []string `yaml:"net" json:"net"`
	Run   []string `yaml:"run" json:"run"`
}

type ElementSpec struct {
	Tag    string `yaml:"tag" json:"tag"`
	Source string `yaml:"source" json:"source"`
}

type DaemonSpec struct {
	Binary  string   `yaml:"binary,omitempty" json:"binary,omitempty"`
	Args    []string `yaml:"args,omitempty" json:"args,omitempty"`
	Health  string   `yaml:"health,omitempty" json:"health,omitempty"`
	Default bool     `yaml:"default,omitempty" json:"default,omitempty"`
}

// BuildInfo captures metadata added when the manifest is compiled into core.json.
type BuildInfo struct {
	Targets   []string `yaml:"targets,omitempty" json:"targets,omitempty"`
	Checksums string   `yaml:"checksums,omitempty" json:"checksums,omitempty"`
	SHA256    string   `yaml:"sha256,omitempty" json:"sha256,omitempty"`
}

type Manifest struct {
	Code        string            `yaml:"code" json:"code"`
	Name        string            `yaml:"name" json:"name"`
	Description string            `yaml:"description,omitempty" json:"description,omitempty"`
	Version     string            `yaml:"version" json:"version"`
	Author      string            `yaml:"author,omitempty" json:"author,omitempty"`
	Licence     string            `yaml:"licence,omitempty" json:"licence,omitempty"`
	Sign        string            `yaml:"sign,omitempty" json:"sign,omitempty"`
	SignKey     string            `yaml:"sign_key,omitempty" json:"sign_key,omitempty"`
	Layout      string            `yaml:"layout,omitempty" json:"layout,omitempty"`
	Slots       map[string]string `yaml:"slots,omitempty" json:"slots,omitempty"`

	Namespace   string                `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	Port        int                   `yaml:"port,omitempty" json:"port,omitempty"`
	Binary      string                `yaml:"binary,omitempty" json:"binary,omitempty"`
	Args        []string              `yaml:"args,omitempty" json:"args,omitempty"`
	Element     *ElementSpec          `yaml:"element,omitempty" json:"element,omitempty"`
	Spec        string                `yaml:"spec,omitempty" json:"spec,omitempty"`
	Permissions Permissions           `yaml:"permissions,omitempty" json:"permissions,omitempty"`
	Modules     []string              `yaml:"modules,omitempty" json:"modules,omitempty"`
	Daemons     map[string]DaemonSpec `yaml:"daemons,omitempty" json:"daemons,omitempty"`
	Build       BuildInfo             `yaml:"build,omitempty" json:"build,omitempty"`
}

func Parse(data []byte) (*Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (m *Manifest) IsProvider() bool {
	if m == nil {
		return false
	}
	return strings.TrimSpace(m.Namespace) != "" && strings.TrimSpace(m.Binary) != ""
}

func (m *Manifest) SlotNames() []string {
	if m == nil || len(m.Slots) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(m.Slots))
	out := make([]string, 0, len(m.Slots))
	for _, v := range m.Slots {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

func (m *Manifest) DefaultDaemon() (string, DaemonSpec, bool) {
	if m == nil || len(m.Daemons) == 0 {
		return "", DaemonSpec{}, false
	}
	if len(m.Daemons) == 1 {
		for name, spec := range m.Daemons {
			return name, spec, true
		}
	}
	var (
		foundName string
		foundSpec DaemonSpec
		found     bool
	)
	for name, spec := range m.Daemons {
		if !spec.Default {
			continue
		}
		if found {
			return "", DaemonSpec{}, false
		}
		foundName, foundSpec, found = name, spec, true
	}
	if found {
		return foundName, foundSpec, true
	}
	return "", DaemonSpec{}, false
}

func validateManifest(m *Manifest) error {
	if m == nil {
		return errors.New("manifest is required")
	}
	if strings.TrimSpace(m.Code) == "" {
		return errors.New("manifest code is required")
	}
	if strings.TrimSpace(m.Name) == "" {
		return errors.New("manifest name is required")
	}
	if strings.TrimSpace(m.Version) == "" {
		return errors.New("manifest version is required")
	}
	return nil
}
