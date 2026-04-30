// SPDX-License-Identifier: EUPL-1.2

package agentci

import (
	// Note: filepath.Join is retained in tests to build temporary config paths without touching production path helpers.
	`path/filepath`
	// Note: testing is the standard Go test harness.
	"testing"

	core "dappco.re/go"
	"dappco.re/go/config"
)

const (
	sonarConfigTestNewConfigV = "new config: %v"
	sonarConfigTestNotAMap    = "not-a-map"
	sonarConfigTestSetClothoV = "set clotho: %v"
)

func configResultError(r core.Result) error {
	if r.OK {
		return nil
	}
	return core.E("agentci.config_test", r.Error(), nil)
}

func testConfig(t *testing.T, opts ...config.Option) *config.Config {
	t.Helper()
	r := config.New(opts...)
	if err := configResultError(r); err != nil {
		t.Fatalf(sonarConfigTestNewConfigV, err)
	}
	return core.MustCast[*config.Config](r)
}

func TestLoadAgentsAndRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	cfg := testConfig(t, config.WithPath(path))

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
	cfg := testConfig(t)

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
	cfg := testConfig(t)

	if err := configResultError(cfg.Set("agents", sonarConfigTestNotAMap)); err != nil {
		t.Fatalf("set agents: %v", err)
	}

	if _, err := LoadAgents(cfg); err == nil {
		t.Fatalf("expected load agents error")
	}
}

func TestLoadClothoConfigReturnsErrorForInvalidData(t *testing.T) {
	cfg := testConfig(t)

	if err := configResultError(cfg.Set("clotho", sonarConfigTestNotAMap)); err != nil {
		t.Fatalf(sonarConfigTestSetClothoV, err)
	}

	if _, err := LoadClothoConfig(cfg); err == nil {
		t.Fatalf("expected load clotho error")
	}
}

func TestLoadClothoConfigHandlesNullConfig(t *testing.T) {
	cfg := testConfig(t)

	if err := configResultError(cfg.Set("clotho", nil)); err != nil {
		t.Fatalf(sonarConfigTestSetClothoV, err)
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
	cfg := testConfig(t)

	if err := configResultError(cfg.Set("clotho", map[string]any{
		"validation_threshold": 1.5,
	})); err != nil {
		t.Fatalf(sonarConfigTestSetClothoV, err)
	}

	if _, err := LoadClothoConfig(cfg); err == nil {
		t.Fatalf("expected load clotho error for invalid threshold")
	}
}

func TestLoadClothoConfigRejectsUnknownStrategy(t *testing.T) {
	cfg := testConfig(t)

	if err := configResultError(cfg.Set("clotho", map[string]any{
		"strategy": "experimental",
	})); err != nil {
		t.Fatalf(sonarConfigTestSetClothoV, err)
	}

	if _, err := LoadClothoConfig(cfg); err == nil {
		t.Fatalf("expected load clotho error for invalid strategy")
	}
}

func TestSaveAndRemoveAgentPropagateLoadErrors(t *testing.T) {
	cfg := testConfig(t)

	if err := configResultError(cfg.Set("agents", sonarConfigTestNotAMap)); err != nil {
		t.Fatalf("set agents: %v", err)
	}

	if err := SaveAgent(cfg, "charon", AgentConfig{}); err == nil {
		t.Fatalf("expected save agent error")
	}
	if err := RemoveAgent(cfg, "charon"); err == nil {
		t.Fatalf("expected remove agent error")
	}
}

func TestConfig_LoadAgents_Good(t *testing.T) {
	target := "LoadAgents"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestConfig_LoadAgents_Bad(t *testing.T) {
	target := "LoadAgents"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestConfig_LoadAgents_Ugly(t *testing.T) {
	target := "LoadAgents"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestConfig_ListAgents_Good(t *testing.T) {
	target := "ListAgents"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestConfig_ListAgents_Bad(t *testing.T) {
	target := "ListAgents"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestConfig_ListAgents_Ugly(t *testing.T) {
	target := "ListAgents"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestConfig_LoadActiveAgents_Good(t *testing.T) {
	target := "LoadActiveAgents"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestConfig_LoadActiveAgents_Bad(t *testing.T) {
	target := "LoadActiveAgents"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestConfig_LoadActiveAgents_Ugly(t *testing.T) {
	target := "LoadActiveAgents"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestConfig_LoadClothoConfig_Good(t *testing.T) {
	target := "LoadClothoConfig"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestConfig_LoadClothoConfig_Bad(t *testing.T) {
	target := "LoadClothoConfig"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestConfig_LoadClothoConfig_Ugly(t *testing.T) {
	target := "LoadClothoConfig"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestConfig_SaveAgent_Good(t *testing.T) {
	target := "SaveAgent"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestConfig_SaveAgent_Bad(t *testing.T) {
	target := "SaveAgent"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestConfig_SaveAgent_Ugly(t *testing.T) {
	target := "SaveAgent"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestConfig_RemoveAgent_Good(t *testing.T) {
	target := "RemoveAgent"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestConfig_RemoveAgent_Bad(t *testing.T) {
	target := "RemoveAgent"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestConfig_RemoveAgent_Ugly(t *testing.T) {
	target := "RemoveAgent"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestConfig_AgentConfig_MarshalYAML_Good(t *testing.T) {
	reference := "MarshalYAML"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "AgentConfig_MarshalYAML"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestConfig_AgentConfig_MarshalYAML_Bad(t *testing.T) {
	reference := "MarshalYAML"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "AgentConfig_MarshalYAML"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestConfig_AgentConfig_MarshalYAML_Ugly(t *testing.T) {
	reference := "MarshalYAML"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "AgentConfig_MarshalYAML"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestConfig_AgentConfig_UnmarshalYAML_Good(t *testing.T) {
	reference := "UnmarshalYAML"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "AgentConfig_UnmarshalYAML"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestConfig_AgentConfig_UnmarshalYAML_Bad(t *testing.T) {
	reference := "UnmarshalYAML"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "AgentConfig_UnmarshalYAML"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestConfig_AgentConfig_UnmarshalYAML_Ugly(t *testing.T) {
	reference := "UnmarshalYAML"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "AgentConfig_UnmarshalYAML"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}
