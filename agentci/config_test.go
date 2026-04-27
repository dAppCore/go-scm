// SPDX-License-Identifier: EUPL-1.2

package agentci

import (
	// Note: filepath.Join is retained in tests to build temporary config paths without touching production path helpers.
	"path/filepath"
	// Note: testing is the standard Go test harness.
	"testing"

	"dappco.re/go/config"
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

	agents["cladius"] = AgentConfig{Host: "mutated"}
	agents["new"] = AgentConfig{Host: "extra"}
	agentsAgain, err := LoadAgents(cfg)
	if err != nil {
		t.Fatalf("reload agents after mutation: %v", err)
	}
	if got := agentsAgain["cladius"]; got.Host != agent.Host {
		t.Fatalf("expected load result to be detached from caller mutations, got %#v", got)
	}
	if _, ok := agentsAgain["new"]; ok {
		t.Fatalf("expected load result not to include caller-added entries")
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
	if clotho.ValidationThreshold != 0.5 {
		t.Fatalf("expected default validation threshold, got %v", clotho.ValidationThreshold)
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

func TestLoadClothoConfigHandlesNullConfig(t *testing.T) {
	cfg, err := config.New()
	if err != nil {
		t.Fatalf("new config: %v", err)
	}

	if err := cfg.Set("clotho", nil); err != nil {
		t.Fatalf("set clotho: %v", err)
	}

	clotho, err := LoadClothoConfig(cfg)
	if err != nil {
		t.Fatalf("load clotho: %v", err)
	}
	if clotho.Strategy != "direct" {
		t.Fatalf("expected default strategy, got %q", clotho.Strategy)
	}
	if clotho.ValidationThreshold != 0.5 {
		t.Fatalf("expected default validation threshold, got %v", clotho.ValidationThreshold)
	}
}

func TestLoadClothoConfigRejectsOutOfRangeThreshold(t *testing.T) {
	cfg, err := config.New()
	if err != nil {
		t.Fatalf("new config: %v", err)
	}

	if err := cfg.Set("clotho", map[string]any{
		"validation_threshold": 1.5,
	}); err != nil {
		t.Fatalf("set clotho: %v", err)
	}

	if _, err := LoadClothoConfig(cfg); err == nil {
		t.Fatalf("expected load clotho error for invalid threshold")
	}
}

func TestLoadClothoConfigRejectsUnknownStrategy(t *testing.T) {
	cfg, err := config.New()
	if err != nil {
		t.Fatalf("new config: %v", err)
	}

	if err := cfg.Set("clotho", map[string]any{
		"strategy": "experimental",
	}); err != nil {
		t.Fatalf("set clotho: %v", err)
	}

	if _, err := LoadClothoConfig(cfg); err == nil {
		t.Fatalf("expected load clotho error for invalid strategy")
	}
}

func TestSaveAndRemoveAgentPropagateLoadErrors(t *testing.T) {
	cfg, err := config.New()
	if err != nil {
		t.Fatalf("new config: %v", err)
	}

	if err := cfg.Set("agents", "not-a-map"); err != nil {
		t.Fatalf("set agents: %v", err)
	}

	if err := SaveAgent(cfg, "charon", AgentConfig{}); err == nil {
		t.Fatalf("expected save agent error")
	}
	if err := RemoveAgent(cfg, "charon"); err == nil {
		t.Fatalf("expected remove agent error")
	}
}
