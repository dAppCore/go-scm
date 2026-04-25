// SPDX-License-Identifier: EUPL-1.2

package repos

import (
	// Note: AX-6 — Config APIs return standard errors for nil storage media.
	"errors"
	// Note: AX-6 — Trigger names are matched case-insensitively.
	"strings"
	// Note: AX-6 — Work config exposes duration fields and defaults.
	"time"

	coreio "dappco.re/go/io"
	"dappco.re/go/scm/internal/ax/filepathx"
	"gopkg.in/yaml.v3"
)

type AgentPolicy struct {
	Heartbeat     time.Duration `yaml:"heartbeat"`
	StaleAfter    time.Duration `yaml:"stale_after"`
	WarnOnOverlap bool          `yaml:"warn_on_overlap"`
}

type SyncConfig struct {
	Interval     time.Duration `yaml:"interval"`
	AutoPull     bool          `yaml:"auto_pull"`
	AutoPush     bool          `yaml:"auto_push"`
	CloneMissing bool          `yaml:"clone_missing"`
}

type WorkConfig struct {
	Version  int         `yaml:"version"`
	Sync     SyncConfig  `yaml:"sync"`
	Agents   AgentPolicy `yaml:"agents"`
	Triggers []string    `yaml:"triggers,omitempty"`
}

func DefaultWorkConfig() *WorkConfig {
	return &WorkConfig{
		Version: 1,
		Sync:    SyncConfig{Interval: time.Minute, AutoPull: true, AutoPush: true, CloneMissing: true},
		Agents:  AgentPolicy{Heartbeat: time.Minute, StaleAfter: 5 * time.Minute, WarnOnOverlap: true},
	}
}

func (wc *WorkConfig) HasTrigger(name string) bool {
	if wc == nil {
		return false
	}
	for _, trigger := range wc.Triggers {
		if strings.EqualFold(trigger, name) {
			return true
		}
	}
	return false
}

func LoadWorkConfig(m coreio.Medium, root string) (*WorkConfig, error) {
	if m == nil {
		return nil, errors.New("repos.LoadWorkConfig: medium is required")
	}
	raw, err := m.Read(filepathx.Join(root, ".core", "work.yaml"))
	if err != nil {
		return DefaultWorkConfig(), nil
	}
	var wc WorkConfig
	if err := yaml.Unmarshal([]byte(raw), &wc); err != nil {
		return nil, err
	}
	if wc.Version == 0 {
		wc.Version = 1
	}
	return &wc, nil
}

func SaveWorkConfig(m coreio.Medium, root string, wc *WorkConfig) error {
	if m == nil {
		return errors.New("repos.SaveWorkConfig: medium is required")
	}
	if wc == nil {
		wc = DefaultWorkConfig()
	}
	raw, err := yaml.Marshal(wc)
	if err != nil {
		return err
	}
	return m.Write(filepathx.Join(root, ".core", "work.yaml"), string(raw))
}
