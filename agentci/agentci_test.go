// SPDX-License-Identifier: EUPL-1.2

package agentci

import (
	"context"
	"path/filepath"

	core "dappco.re/go"
	"dappco.re/go/config"
	"dappco.re/go/scm/jobrunner"
	"gopkg.in/yaml.v3"
)

func ax7AgentConfig(t *core.T) *config.Config {
	r := config.New(config.WithPath(filepath.Join(t.TempDir(), "config.yaml")))
	core.RequireNoError(t, configResultError(r))
	return core.MustCast[*config.Config](r)
}

func TestAgentci_LoadAgents_Good(t *core.T) {
	cfg := ax7AgentConfig(t)
	core.RequireNoError(t, configResultError(cfg.Set("agents", map[string]AgentConfig{"codex": {Active: true, Roles: []string{"coder"}}})))
	agents, err := LoadAgents(cfg)
	core.AssertNoError(t, err)
	core.AssertTrue(t, agents["codex"].Active)
}

func TestAgentci_LoadAgents_Bad(t *core.T) {
	agents, err := LoadAgents(nil)
	core.AssertNoError(t, err)
	core.AssertEmpty(t, agents)
}

func TestAgentci_LoadAgents_Ugly(t *core.T) {
	cfg := ax7AgentConfig(t)
	core.RequireNoError(t, configResultError(cfg.Set("agents", map[string]AgentConfig{"codex": {Roles: []string{"coder"}}})))
	agents, err := LoadAgents(cfg)
	core.AssertNoError(t, err)
	agents["codex"] = AgentConfig{Roles: []string{"mutated"}}
	again, err := LoadAgents(cfg)
	core.AssertNoError(t, err)
	core.AssertEqual(t, []string{"coder"}, again["codex"].Roles)
}

func TestAgentci_ListAgents_Good(t *core.T) {
	cfg := ax7AgentConfig(t)
	core.RequireNoError(t, configResultError(cfg.Set("agents", map[string]AgentConfig{"codex": {Active: true}})))
	agents, err := ListAgents(cfg)
	core.AssertNoError(t, err)
	core.AssertLen(t, agents, 1)
}

func TestAgentci_ListAgents_Bad(t *core.T) {
	agents, err := ListAgents(nil)
	core.AssertNoError(t, err)
	core.AssertEmpty(t, agents)
}

func TestAgentci_ListAgents_Ugly(t *core.T) {
	cfg := ax7AgentConfig(t)
	agents, err := ListAgents(cfg)
	core.AssertNoError(t, err)
	core.AssertEmpty(t, agents)
}

func TestAgentci_LoadActiveAgents_Good(t *core.T) {
	cfg := ax7AgentConfig(t)
	core.RequireNoError(t, configResultError(cfg.Set("agents", map[string]AgentConfig{"codex": {Active: true}, "idle": {Active: false}})))
	agents, err := LoadActiveAgents(cfg)
	core.AssertNoError(t, err)
	core.AssertLen(t, agents, 1)
	core.AssertTrue(t, agents["codex"].Active)
}

func TestAgentci_LoadActiveAgents_Bad(t *core.T) {
	agents, err := LoadActiveAgents(nil)
	core.AssertNoError(t, err)
	core.AssertEmpty(t, agents)
}

func TestAgentci_LoadActiveAgents_Ugly(t *core.T) {
	cfg := ax7AgentConfig(t)
	core.RequireNoError(t, configResultError(cfg.Set("agents", map[string]AgentConfig{"idle": {Active: false}})))
	agents, err := LoadActiveAgents(cfg)
	core.AssertNoError(t, err)
	core.AssertEmpty(t, agents)
}

func TestAgentci_LoadClothoConfig_Good(t *core.T) {
	cfg := ax7AgentConfig(t)
	core.RequireNoError(t, configResultError(cfg.Set("clotho", map[string]any{"strategy": "clotho-verified", "validation_threshold": 0.75})))
	got, err := LoadClothoConfig(cfg)
	core.AssertNoError(t, err)
	core.AssertEqual(t, "clotho-verified", got.Strategy)
	core.AssertEqual(t, 0.75, got.ValidationThreshold)
}

func TestAgentci_LoadClothoConfig_Bad(t *core.T) {
	cfg := ax7AgentConfig(t)
	core.RequireNoError(t, configResultError(cfg.Set("clotho", map[string]any{"strategy": "unknown"})))
	_, err := LoadClothoConfig(cfg)
	core.AssertError(t, err)
}

func TestAgentci_LoadClothoConfig_Ugly(t *core.T) {
	got, err := LoadClothoConfig(nil)
	core.AssertNoError(t, err)
	core.AssertEqual(t, "direct", got.Strategy)
	core.AssertEqual(t, 0.5, got.ValidationThreshold)
}

func TestAgentci_SaveAgent_Good(t *core.T) {
	cfg := ax7AgentConfig(t)
	err := SaveAgent(cfg, "codex", AgentConfig{Host: "agent.local", Active: true})
	core.AssertNoError(t, err)
	agents, loadErr := LoadAgents(cfg)
	core.RequireNoError(t, loadErr)
	core.AssertEqual(t, "agent.local", agents["codex"].Host)
}

func TestAgentci_SaveAgent_Bad(t *core.T) {
	err := SaveAgent(nil, "codex", AgentConfig{})
	core.AssertError(
		t, err,
	)
}

func TestAgentci_SaveAgent_Ugly(t *core.T) {
	err := SaveAgent(ax7AgentConfig(t), "", AgentConfig{})
	core.AssertError(
		t, err,
	)
}

func TestAgentci_RemoveAgent_Good(t *core.T) {
	cfg := ax7AgentConfig(t)
	core.RequireNoError(t, SaveAgent(cfg, "codex", AgentConfig{Active: true}))
	err := RemoveAgent(cfg, "codex")
	core.AssertNoError(t, err)
	agents, loadErr := LoadAgents(cfg)
	core.RequireNoError(t, loadErr)
	_, ok := agents["codex"]
	core.AssertFalse(t, ok)
}

func TestAgentci_RemoveAgent_Bad(t *core.T) {
	err := RemoveAgent(nil, "codex")
	core.AssertError(
		t, err,
	)
}

func TestAgentci_RemoveAgent_Ugly(t *core.T) {
	err := RemoveAgent(ax7AgentConfig(t), "")
	core.AssertError(
		t, err,
	)
}

func TestAgentci_AgentConfig_MarshalYAML_Good(t *core.T) {
	raw, err := yaml.Marshal(AgentConfig{Host: "agent.local", Roles: []string{"coder"}})
	core.AssertNoError(t, err)
	core.AssertContains(t, string(raw), "agent.local")
}

func TestAgentci_AgentConfig_MarshalYAML_Bad(t *core.T) {
	raw, err := yaml.Marshal(AgentConfig{})
	core.AssertNoError(t, err)
	core.AssertContains(t, string(raw), "host")
}

func TestAgentci_AgentConfig_MarshalYAML_Ugly(t *core.T) {
	value, err := (AgentConfig{Active: true}).MarshalYAML()
	core.AssertNoError(t, err)
	core.AssertNotNil(t, value)
}

func TestAgentci_AgentConfig_UnmarshalYAML_Good(t *core.T) {
	var agent AgentConfig
	err := yaml.Unmarshal([]byte("host: agent.local\nactive: true\n"), &agent)
	core.AssertNoError(t, err)
	core.AssertEqual(t, "agent.local", agent.Host)
	core.AssertTrue(t, agent.Active)
}

func TestAgentci_AgentConfig_UnmarshalYAML_Bad(t *core.T) {
	var agent AgentConfig
	err := yaml.Unmarshal([]byte("active: not-bool\n"), &agent)
	core.AssertError(t, err)
}

func TestAgentci_AgentConfig_UnmarshalYAML_Ugly(t *core.T) {
	var agent AgentConfig
	err := yaml.Unmarshal([]byte("{}"), &agent)
	core.AssertNoError(t, err)
	core.AssertEqual(t, "", agent.Host)
}

func TestAgentci_NewSpinner_Good(t *core.T) {
	spinner := NewSpinner(ClothoConfig{Strategy: "direct"}, map[string]AgentConfig{"codex": {Roles: []string{"coder"}}})
	core.AssertEqual(t, "direct", spinner.Config.Strategy)
	core.AssertEqual(t, []string{"coder"}, spinner.Agents["codex"].Roles)
}

func TestAgentci_NewSpinner_Bad(t *core.T) {
	spinner := NewSpinner(ClothoConfig{}, nil)
	core.AssertNotNil(t, spinner)
	core.AssertEmpty(t, spinner.Agents)
}

func TestAgentci_NewSpinner_Ugly(t *core.T) {
	agents := map[string]AgentConfig{"codex": {Roles: []string{"coder"}}}
	spinner := NewSpinner(ClothoConfig{}, agents)
	agents["codex"] = AgentConfig{Roles: []string{"mutated"}}
	core.AssertEqual(t, []string{"coder"}, spinner.Agents["codex"].Roles)
}

func TestAgentci_Spinner_DeterminePlan_Good(t *core.T) {
	spinner := NewSpinner(ClothoConfig{Strategy: "clotho-verified"}, map[string]AgentConfig{})
	got := spinner.DeterminePlan(&jobrunner.PipelineSignal{NeedsCoding: true}, "codex")
	core.AssertEqual(t, RunModeClothoVerified, got)
}

func TestAgentci_Spinner_DeterminePlan_Bad(t *core.T) {
	var spinner *Spinner
	got := spinner.DeterminePlan(&jobrunner.PipelineSignal{NeedsCoding: true}, "codex")
	core.AssertEqual(t, RunModeDirect, got)
}

func TestAgentci_Spinner_DeterminePlan_Ugly(t *core.T) {
	spinner := NewSpinner(ClothoConfig{Strategy: "direct"}, map[string]AgentConfig{"codex": {DualRun: true}})
	got := spinner.DeterminePlan(nil, "codex")
	core.AssertEqual(t, RunModeClothoVerified, got)
}

func TestAgentci_Spinner_FindByForgejoUser_Good(t *core.T) {
	spinner := NewSpinner(ClothoConfig{}, map[string]AgentConfig{"codex": {ForgejoUser: "codex-bot"}})
	name, agent, ok := spinner.FindByForgejoUser("codex-bot")
	core.AssertTrue(t, ok)
	core.AssertEqual(t, "codex", name)
	core.AssertEqual(t, "codex-bot", agent.ForgejoUser)
}

func TestAgentci_Spinner_FindByForgejoUser_Bad(t *core.T) {
	spinner := NewSpinner(ClothoConfig{}, map[string]AgentConfig{})
	_, _, ok := spinner.FindByForgejoUser("missing")
	core.AssertFalse(t, ok)
}

func TestAgentci_Spinner_FindByForgejoUser_Ugly(t *core.T) {
	var spinner *Spinner
	_, _, ok := spinner.FindByForgejoUser("codex")
	core.AssertFalse(t, ok)
}

func TestAgentci_Spinner_GetVerifierModel_Good(t *core.T) {
	spinner := NewSpinner(ClothoConfig{}, map[string]AgentConfig{"codex": {VerifyModel: "gpt-5.3"}})
	got := spinner.GetVerifierModel("codex")
	core.AssertEqual(t, "gpt-5.3", got)
}

func TestAgentci_Spinner_GetVerifierModel_Bad(t *core.T) {
	spinner := NewSpinner(ClothoConfig{}, map[string]AgentConfig{})
	got := spinner.GetVerifierModel("missing")
	core.AssertEqual(t, "", got)
}

func TestAgentci_Spinner_GetVerifierModel_Ugly(t *core.T) {
	var spinner *Spinner
	got := spinner.GetVerifierModel("codex")
	core.AssertEqual(t, "", got)
}

func TestAgentci_Spinner_Weave_Good(t *core.T) {
	ok, err := NewSpinner(ClothoConfig{}, nil).Weave(context.Background(), []byte("same\n"), []byte(" same "))
	core.AssertNoError(t, err)
	core.AssertTrue(t, ok)
}

func TestAgentci_Spinner_Weave_Bad(t *core.T) {
	ok, err := NewSpinner(ClothoConfig{}, nil).Weave(context.Background(), []byte("primary"), []byte("signed"))
	core.AssertNoError(t, err)
	core.AssertFalse(t, ok)
}

func TestAgentci_Spinner_Weave_Ugly(t *core.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ok, err := NewSpinner(ClothoConfig{}, nil).Weave(ctx, []byte("same"), []byte("same"))
	core.AssertErrorIs(t, err, context.Canceled)
	core.AssertFalse(t, ok)
}

func TestAgentci_SanitizePath_Good(t *core.T) {
	got, err := SanitizePath("agent-01.yaml")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "agent-01.yaml", got)
}

func TestAgentci_SanitizePath_Bad(t *core.T) {
	_, err := SanitizePath("../agent")
	core.AssertError(
		t, err,
	)
}

func TestAgentci_SanitizePath_Ugly(t *core.T) {
	_, err := SanitizePath("")
	core.AssertError(
		t, err,
	)
}

func TestAgentci_ValidatePathElement_Good(t *core.T) {
	got, err := ValidatePathElement("agent_01")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "agent_01", got)
}

func TestAgentci_ValidatePathElement_Bad(t *core.T) {
	_, err := ValidatePathElement("agent/01")
	core.AssertError(
		t, err,
	)
}

func TestAgentci_ValidatePathElement_Ugly(t *core.T) {
	_, err := ValidatePathElement(".")
	core.AssertError(
		t, err,
	)
}

func TestAgentci_ResolvePathWithinRoot_Good(t *core.T) {
	root := t.TempDir()
	name, path, err := ResolvePathWithinRoot(root, "agent.yaml")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "agent.yaml", name)
	core.AssertEqual(t, filepath.Join(root, "agent.yaml"), path)
}

func TestAgentci_ResolvePathWithinRoot_Bad(t *core.T) {
	_, _, err := ResolvePathWithinRoot("", "agent.yaml")
	core.AssertError(
		t, err,
	)
}

func TestAgentci_ResolvePathWithinRoot_Ugly(t *core.T) {
	_, _, err := ResolvePathWithinRoot(t.TempDir(), "..")
	core.AssertError(
		t, err,
	)
}

func TestAgentci_ValidateRemoteDir_Good(t *core.T) {
	got, err := ValidateRemoteDir("~/queue/tasks")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "~/queue/tasks", got)
}

func TestAgentci_ValidateRemoteDir_Bad(t *core.T) {
	_, err := ValidateRemoteDir("../queue")
	core.AssertError(
		t, err,
	)
}

func TestAgentci_ValidateRemoteDir_Ugly(t *core.T) {
	got, err := ValidateRemoteDir("/")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "/", got)
}

func TestAgentci_JoinRemotePath_Good(t *core.T) {
	got, err := JoinRemotePath("~/queue", "tasks", "job.json")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "~/queue/tasks/job.json", got)
}

func TestAgentci_JoinRemotePath_Bad(t *core.T) {
	_, err := JoinRemotePath("~/queue", "..")
	core.AssertError(
		t, err,
	)
}

func TestAgentci_JoinRemotePath_Ugly(t *core.T) {
	got, err := JoinRemotePath("~")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "~", got)
}

func TestAgentci_EscapeShellArg_Good(t *core.T) {
	got := EscapeShellArg("agent ready")
	core.AssertEqual(
		t, "'agent ready'", got,
	)
}

func TestAgentci_EscapeShellArg_Bad(t *core.T) {
	got := EscapeShellArg("agent'ready")
	core.AssertEqual(
		t, "'agent'\\''ready'", got,
	)
}

func TestAgentci_EscapeShellArg_Ugly(t *core.T) {
	got := EscapeShellArg("")
	core.AssertEqual(
		t, "''", got,
	)
}

func TestAgentci_SecureSSHCommand_Good(t *core.T) {
	cmd := SecureSSHCommand("agent.local", "echo ready")
	core.AssertEqual(t, "ssh", filepath.Base(cmd.Path))
	core.AssertContains(t, cmd.Args, "agent.local")
}

func TestAgentci_SecureSSHCommand_Bad(t *core.T) {
	cmd := SecureSSHCommand("", "")
	core.AssertEqual(t, "ssh", filepath.Base(cmd.Path))
	core.AssertContains(t, cmd.Args, "")
}

func TestAgentci_SecureSSHCommand_Ugly(t *core.T) {
	cmd := SecureSSHCommand("agent.local", "printf 'x'")
	core.AssertContains(
		t, cmd.Args, "StrictHostKeyChecking=yes",
	)
}

func TestAgentci_SecureSSHCommandContext_Good(t *core.T) {
	cmd := SecureSSHCommandContext(context.Background(), "agent.local", "echo ready")
	core.AssertEqual(t, "ssh", filepath.Base(cmd.Path))
	core.AssertContains(t, cmd.Args, "echo ready")
}

func TestAgentci_SecureSSHCommandContext_Bad(t *core.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd := SecureSSHCommandContext(ctx, "agent.local", "echo ready")
	core.AssertNotNil(t, cmd)
	core.AssertNotNil(t, cmd.Cancel)
}

func TestAgentci_SecureSSHCommandContext_Ugly(t *core.T) {
	cmd := SecureSSHCommandContext(nil, "agent.local", "echo ready")
	core.AssertNotNil(t, cmd)
	core.AssertContains(t, cmd.Args, "BatchMode=yes")
}

func TestAgentci_MaskToken_Good(t *core.T) {
	got := MaskToken("abcd1234wxyz")
	core.AssertEqual(
		t, "abcd****wxyz", got,
	)
}

func TestAgentci_MaskToken_Bad(t *core.T) {
	got := MaskToken("short")
	core.AssertEqual(
		t, "*****", got,
	)
}

func TestAgentci_MaskToken_Ugly(t *core.T) {
	got := MaskToken("")
	core.AssertEqual(
		t, "*****", got,
	)
}
