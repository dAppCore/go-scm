// SPDX-License-Identifier: EUPL-1.2

package agentci

import (
	"bytes"
	"context"
	"strings"

	"dappco.re/go/scm/jobrunner"
)

// RunMode determines the execution strategy for a dispatched task.
type RunMode string

const (
	RunModeDirect         RunMode = "direct"
	RunModeClothoVerified RunMode = "clotho-verified"
)

// Spinner is the Clotho orchestrator that determines the fate of each task.
type Spinner struct {
	Config ClothoConfig
	Agents map[string]AgentConfig
}

func (s *Spinner) resolveAgent(agentName string) (string, AgentConfig, bool) {
	if s == nil {
		return "", AgentConfig{}, false
	}

	if agent, ok := s.Agents[agentName]; ok {
		return agentName, agent, true
	}

	for name, cfg := range s.Agents {
		if strings.EqualFold(name, agentName) || strings.EqualFold(cfg.ForgejoUser, agentName) {
			return name, cfg, true
		}
	}

	return "", AgentConfig{}, false
}

// NewSpinner creates a new Clotho orchestrator.
func NewSpinner(cfg ClothoConfig, agents map[string]AgentConfig) *Spinner {
	cp := make(map[string]AgentConfig, len(agents))
	for k, v := range agents {
		cp[k] = v
	}
	return &Spinner{Config: cfg, Agents: cp}
}

// DeterminePlan decides if a signal requires dual-run verification based on
// the global strategy, agent configuration, and repository criticality.
func (s *Spinner) DeterminePlan(signal *jobrunner.PipelineSignal, agentName string) RunMode {
	if s == nil {
		return RunModeDirect
	}

	_, agent, ok := s.resolveAgent(agentName)

	critical := false
	if signal != nil {
		critical = signal.IsDraft ||
			signal.HasUnresolvedThreads() ||
			(signal.CheckStatus != "" && !strings.EqualFold(signal.CheckStatus, "SUCCESS")) ||
			(signal.Mergeable != "" && !strings.EqualFold(signal.Mergeable, "MERGEABLE")) ||
			signal.NeedsCoding
	}

	if ok {
		if agent.DualRun || strings.EqualFold(agent.SecurityLevel, "high") {
			return RunModeClothoVerified
		}
	}

	if strings.EqualFold(s.Config.Strategy, string(RunModeClothoVerified)) && critical {
		return RunModeClothoVerified
	}
	return RunModeDirect
}

// FindByForgejoUser resolves a Forgejo username to the agent config key and config.
func (s *Spinner) FindByForgejoUser(forgejoUser string) (string, AgentConfig, bool) {
	if s == nil {
		return "", AgentConfig{}, false
	}
	for name, agent := range s.Agents {
		if strings.EqualFold(name, forgejoUser) || strings.EqualFold(agent.ForgejoUser, forgejoUser) {
			return name, agent, true
		}
	}
	return "", AgentConfig{}, false
}

// GetVerifierModel returns the model for the secondary "signed" verification run.
func (s *Spinner) GetVerifierModel(agentName string) string {
	_, agent, ok := s.resolveAgent(agentName)
	if !ok {
		return ""
	}
	if agent.VerifyModel != "" {
		return agent.VerifyModel
	}
	if agent.Model != "" {
		return agent.Model
	}
	return ""
}

// Weave compares primary and verifier outputs. Returns true if they converge.
func (s *Spinner) Weave(ctx context.Context, primaryOutput, signedOutput []byte) (bool, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return false, err
		}
	}
	return bytes.Equal(bytes.TrimSpace(primaryOutput), bytes.TrimSpace(signedOutput)), nil
}
