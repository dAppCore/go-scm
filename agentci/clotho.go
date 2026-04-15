// SPDX-License-Identifier: EUPL-1.2

package agentci

import (
	"context"
	strings "dappco.re/go/core/scm/internal/ax/stringsx"
	"math"

	"dappco.re/go/core/scm/jobrunner"
)

// RunMode determines the execution strategy for a dispatched task.
type RunMode string

const (
	//
	ModeStandard RunMode = "standard"
	//
	ModeDual RunMode = "dual" // The Clotho Protocol — dual-run verification
)

// Spinner is the Clotho orchestrator that determines the fate of each task.
type Spinner struct {
	Config ClothoConfig
	Agents map[string]AgentConfig
}

// NewSpinner creates a new Clotho orchestrator.
// Usage: NewSpinner(...)
func NewSpinner(cfg ClothoConfig, agents map[string]AgentConfig) *Spinner {
	cfg = normalizeClothoConfig(cfg)
	return &Spinner{
		Config: cfg,
		Agents: agents,
	}
}

// DeterminePlan decides if a signal requires dual-run verification based on
// the global strategy, agent configuration, and repository criticality.
// Usage: DeterminePlan(...)
func (s *Spinner) DeterminePlan(signal *jobrunner.PipelineSignal, agentName string) RunMode {
	if s == nil || signal == nil {
		return ModeStandard
	}
	if !strings.EqualFold(strings.TrimSpace(s.Config.Strategy), "clotho-verified") {
		return ModeStandard
	}

	agent, ok := s.Agents[agentName]
	if !ok {
		_, agent, ok = s.FindByForgejoUser(agentName)
	}
	if !ok {
		return ModeStandard
	}
	if agent.DualRun {
		return ModeDual
	}
	if strings.EqualFold(agent.SecurityLevel, "high") {
		return ModeDual
	}

	// Protect critical repos with dual-run (Axiom 1).
	repoID := strings.ToLower(strings.TrimSpace(signal.RepoFullName()))
	if strings.EqualFold(signal.RepoName, "core") || strings.Contains(repoID, "security") {
		return ModeDual
	}

	return ModeStandard
}

// GetVerifierModel returns the model for the secondary "signed" verification run.
// Usage: GetVerifierModel(...)
func (s *Spinner) GetVerifierModel(agentName string) string {
	if s == nil {
		return "gemini-1.5-pro"
	}
	agent, ok := s.Agents[agentName]
	if !ok {
		_, agent, ok = s.FindByForgejoUser(agentName)
	}
	if !ok || agent.VerifyModel == "" {
		return "gemini-1.5-pro"
	}
	return agent.VerifyModel
}

// FindByForgejoUser resolves a Forgejo username to the agent config key and config.
// This decouples agent naming (mythological roles) from Forgejo identity.
// Usage: FindByForgejoUser(...)
func (s *Spinner) FindByForgejoUser(forgejoUser string) (string, AgentConfig, bool) {
	if s == nil {
		return "", AgentConfig{}, false
	}
	if forgejoUser == "" {
		return "", AgentConfig{}, false
	}
	// Direct match on config key first.
	if agent, ok := s.Agents[forgejoUser]; ok {
		return forgejoUser, agent, true
	}
	for name, agent := range s.Agents {
		if strings.EqualFold(name, forgejoUser) {
			return name, agent, true
		}
	}
	// Search by ForgejoUser field.
	for name, agent := range s.Agents {
		if agent.ForgejoUser != "" && strings.EqualFold(agent.ForgejoUser, forgejoUser) {
			return name, agent, true
		}
	}
	return "", AgentConfig{}, false
}

// Weave compares primary and verifier outputs. Returns true if they converge.
// The comparison is a coarse token-overlap check controlled by the configured
// validation threshold. It is intentionally deterministic and fast; richer
// semantic diffing can replace it later without changing the signature.
// Usage: Weave(...)
func (s *Spinner) Weave(ctx context.Context, primaryOutput, signedOutput []byte) (bool, error) {
	if ctx != nil {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}
	}

	primary := tokenizeWeaveOutput(primaryOutput)
	signed := tokenizeWeaveOutput(signedOutput)

	if len(primary) == 0 && len(signed) == 0 {
		return true, nil
	}

	threshold := 0.85
	if s != nil {
		threshold = s.Config.ValidationThreshold
	}
	if threshold < 0 || threshold > 1 {
		threshold = 0.85
	}

	similarity := weaveDiceSimilarity(primary, signed)
	return similarity >= threshold, nil
}

func normalizeClothoConfig(cfg ClothoConfig) ClothoConfig {
	if cfg == (ClothoConfig{}) {
		cfg.Strategy = "direct"
		cfg.ValidationThreshold = 0.85
		return cfg
	}
	cfg.Strategy = strings.TrimSpace(cfg.Strategy)
	if cfg.Strategy == "" {
		cfg.Strategy = "direct"
	}
	if cfg.ValidationThreshold < 0 || cfg.ValidationThreshold > 1 {
		cfg.ValidationThreshold = 0.85
	}
	return cfg
}

func tokenizeWeaveOutput(output []byte) []string {
	fields := strings.Fields(strings.ReplaceAll(string(output), "\r\n", "\n"))
	if len(fields) == 0 {
		return nil
	}
	return fields
}

func weaveDiceSimilarity(primary, signed []string) float64 {
	if len(primary) == 0 || len(signed) == 0 {
		return 0
	}

	counts := make(map[string]int, len(primary))
	for _, token := range primary {
		counts[token]++
	}

	common := 0
	for _, token := range signed {
		if counts[token] == 0 {
			continue
		}
		counts[token]--
		common++
	}

	return math.Min(1, (2*float64(common))/float64(len(primary)+len(signed)))
}
