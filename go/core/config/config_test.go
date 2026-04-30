// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	sonarConfigTestAgentName  = "agent.name"
	sonarConfigTestConfigYaml = "config.yaml"
	sonarConfigTestNewConfigV = "new config: %v"
)

func TestConfig_SetGetCommit(t *testing.T) {
	path := filepath.Join(t.TempDir(), sonarConfigTestConfigYaml)
	cfg := NewWithPath(path)

	if err := cfg.Set("agents.alpha.active", true); err != nil {
		t.Fatalf("set active: %v", err)
	}
	if err := cfg.Set("clotho.strategy", "clotho-verified"); err != nil {
		t.Fatalf("set clotho: %v", err)
	}

	var active bool
	if err := cfg.Get("agents.alpha.active", &active); err != nil {
		t.Fatalf("get active: %v", err)
	}
	if !active {
		t.Fatalf("expected active true")
	}

	var clotho struct {
		Strategy string `yaml:"strategy"`
	}
	if err := cfg.Get("clotho", &clotho); err != nil {
		t.Fatalf("get clotho: %v", err)
	}
	if clotho.Strategy != "clotho-verified" {
		t.Fatalf("expected strategy clotho-verified, got %q", clotho.Strategy)
	}

	if err := cfg.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
}

func TestConfig_WithPath_Good(t *testing.T) {
	cfg, err := New(WithPath(sonarConfigTestConfigYaml))
	if err != nil {
		t.Fatalf(sonarConfigTestNewConfigV, err)
	}
	if cfg.Path != sonarConfigTestConfigYaml {
		t.Fatalf("expected configured path, got %q", cfg.Path)
	}
}

func TestConfig_WithPath_Bad(t *testing.T) {
	cfg, err := New(WithPath(""))
	if err != nil {
		t.Fatalf(sonarConfigTestNewConfigV, err)
	}
	if cfg.Path != "" {
		t.Fatalf("expected empty path, got %q", cfg.Path)
	}
}

func TestConfig_WithPath_Ugly(t *testing.T) {
	cfg, err := New(WithPath(filepath.Join("..", sonarConfigTestConfigYaml)))
	if err != nil {
		t.Fatalf(sonarConfigTestNewConfigV, err)
	}
	if cfg.Path != filepath.Join("..", sonarConfigTestConfigYaml) {
		t.Fatalf("expected relative path preserved, got %q", cfg.Path)
	}
}

func TestConfig_New_Good(t *testing.T) {
	cfg, err := New(WithPath(sonarConfigTestConfigYaml))
	if err != nil {
		t.Fatalf(sonarConfigTestNewConfigV, err)
	}
	if cfg == nil || cfg.data == nil {
		t.Fatalf("expected initialized config")
	}
}

func TestConfig_New_Bad(t *testing.T) {
	panicked := false
	func() {
		defer func() {
			panicked = recover() != nil
		}()
		_, _ = New(nil)
	}()
	if !panicked {
		t.Fatalf("expected nil option to panic")
	}
}

func TestConfig_New_Ugly(t *testing.T) {
	cfg, err := New(WithPath("first.yaml"), WithPath("second.yaml"))
	if err != nil {
		t.Fatalf(sonarConfigTestNewConfigV, err)
	}
	if cfg.Path != "second.yaml" {
		t.Fatalf("expected later option to win, got %q", cfg.Path)
	}
}

func TestConfig_NewWithPath_Good(t *testing.T) {
	cfg := NewWithPath(sonarConfigTestConfigYaml)
	if cfg.Path != sonarConfigTestConfigYaml {
		t.Fatalf("expected configured path, got %q", cfg.Path)
	}
	if cfg.data == nil {
		t.Fatalf("expected initialized data")
	}
}

func TestConfig_NewWithPath_Bad(t *testing.T) {
	cfg := NewWithPath("")
	if cfg.Path != "" {
		t.Fatalf("expected empty path, got %q", cfg.Path)
	}
	if cfg.data == nil {
		t.Fatalf("expected initialized data")
	}
}

func TestConfig_NewWithPath_Ugly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", sonarConfigTestConfigYaml)
	cfg := NewWithPath(path)
	if cfg.Path != path {
		t.Fatalf("expected temp path, got %q", cfg.Path)
	}
}

func TestConfig_Config_Set_Good(t *testing.T) {
	cfg := NewWithPath("")
	if err := cfg.Set(sonarConfigTestAgentName, "codex"); err != nil {
		t.Fatalf("set dotted key: %v", err)
	}
	var got string
	if err := cfg.Get(sonarConfigTestAgentName, &got); err != nil {
		t.Fatalf("get dotted key: %v", err)
	}
	if got != "codex" {
		t.Fatalf("expected codex, got %q", got)
	}
}

func TestConfig_Config_Set_Bad(t *testing.T) {
	cfg := NewWithPath("")
	err := cfg.Set("agent..name", "codex")
	if err == nil {
		t.Fatalf("expected empty key segment error")
	}
}

func TestConfig_Config_Set_Ugly(t *testing.T) {
	cfg := NewWithPath("")
	err := cfg.Set("", map[string]any{"agent": map[string]any{"name": "codex"}})
	if err != nil {
		t.Fatalf("set root map: %v", err)
	}
	var got string
	if err := cfg.Get(sonarConfigTestAgentName, &got); err != nil {
		t.Fatalf("get cloned root value: %v", err)
	}
	if got != "codex" {
		t.Fatalf("expected codex, got %q", got)
	}
}

func TestConfig_Config_Get_Good(t *testing.T) {
	cfg := NewWithPath("")
	if err := cfg.Set("agent.active", true); err != nil {
		t.Fatalf("set active: %v", err)
	}
	var got bool
	if err := cfg.Get("agent.active", &got); err != nil {
		t.Fatalf("get active: %v", err)
	}
	if !got {
		t.Fatalf("expected active")
	}
}

func TestConfig_Config_Get_Bad(t *testing.T) {
	cfg := NewWithPath("")
	var got string
	err := cfg.Get("missing", &got)
	if err == nil {
		t.Fatalf("expected missing key error")
	}
}

func TestConfig_Config_Get_Ugly(t *testing.T) {
	cfg := NewWithPath("")
	err := cfg.Get("", nil)
	if err == nil {
		t.Fatalf("expected nil output error")
	}
}

func TestConfig_Config_Commit_Good(t *testing.T) {
	path := filepath.Join(t.TempDir(), sonarConfigTestConfigYaml)
	cfg := NewWithPath(path)
	if err := cfg.Set(sonarConfigTestAgentName, "codex"); err != nil {
		t.Fatalf("set agent: %v", err)
	}
	if err := cfg.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected committed file: %v", err)
	}
}

func TestConfig_Config_Commit_Bad(t *testing.T) {
	cfg := NewWithPath("")
	err := cfg.Commit()
	if err == nil {
		t.Fatalf("expected missing path error")
	}
}

func TestConfig_Config_Commit_Ugly(t *testing.T) {
	cfg := NewWithPath(filepath.Join(t.TempDir(), "nested", sonarConfigTestConfigYaml))
	if err := cfg.Commit(); err != nil {
		t.Fatalf("commit empty config: %v", err)
	}
}
