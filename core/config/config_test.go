// SPDX-License-Identifier: EUPL-1.2

package config

import (
	// Note: filepath.Join is retained in tests to build temporary config paths without changing production dependencies.
	"path/filepath"
	// Note: testing is the standard Go test harness.
	"testing"
)

func TestConfig_SetGetCommit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
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
