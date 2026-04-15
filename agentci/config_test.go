// SPDX-License-Identifier: EUPL-1.2

package agentci

import (
	"path/filepath"
	"testing"

	"dappco.re/go/core/config"
)

func TestLoadAgentsAndRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	cfg, err := config.New(config.WithPath(path))
	if err != nil {
		t.Fatalf("new config: %v", err)
	}

	agent := AgentConfig{
		Host:          "forgejo",
		QueueDir:      "queue",
		ForgejoUser:   "cladius",
		Model:         "gpt-5.4",
		Runner:        "codex",
		VerifyModel:   "gpt-5.3-codex-spark",
		SecurityLevel: "high",
		Roles:         []string{"dispatch", "review"},
		DualRun:       true,
		Active:        true,
	}

	if err := SaveAgent(cfg, "cladius", agent); err != nil {
		t.Fatalf("save agent: %v", err)
	}

	agents, err := LoadAgents(cfg)
	if err != nil {
		t.Fatalf("load agents: %v", err)
	}
	if got := agents["cladius"]; got.Host != agent.Host || got.VerifyModel != agent.VerifyModel {
		t.Fatalf("unexpected agent round-trip: %#v", got)
	}

	active, err := LoadActiveAgents(cfg)
	if err != nil {
		t.Fatalf("load active agents: %v", err)
	}
	if _, ok := active["cladius"]; !ok {
		t.Fatalf("expected active agent")
	}

	if err := RemoveAgent(cfg, "cladius"); err != nil {
		t.Fatalf("remove agent: %v", err)
	}
	agents, err = LoadAgents(cfg)
	if err != nil {
		t.Fatalf("reload agents: %v", err)
	}
	if _, ok := agents["cladius"]; ok {
		t.Fatalf("expected agent removal")
	}
}

func TestLoadClothoConfigDefaults(t *testing.T) {
	cfg, err := config.New()
	if err != nil {
		t.Fatalf("new config: %v", err)
	}

	clotho, err := LoadClothoConfig(cfg)
	if err != nil {
		t.Fatalf("load clotho: %v", err)
	}
	if clotho.Strategy != "direct" {
		t.Fatalf("expected direct strategy, got %q", clotho.Strategy)
	}
}

func TestLoadAgentsReturnsErrorForInvalidData(t *testing.T) {
	cfg, err := config.New()
	if err != nil {
		t.Fatalf("new config: %v", err)
	}

	if err := cfg.Set("agents", "not-a-map"); err != nil {
		t.Fatalf("set agents: %v", err)
	}

	if _, err := LoadAgents(cfg); err == nil {
		t.Fatalf("expected load agents error")
	}
}

func TestLoadClothoConfigReturnsErrorForInvalidData(t *testing.T) {
	cfg, err := config.New()
	if err != nil {
		t.Fatalf("new config: %v", err)
	}

	if err := cfg.Set("clotho", "not-a-map"); err != nil {
		t.Fatalf("set clotho: %v", err)
	}

	if _, err := LoadClothoConfig(cfg); err == nil {
		t.Fatalf("expected load clotho error")
	}
}
