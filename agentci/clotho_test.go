// SPDX-License-Identifier: EUPL-1.2

package agentci

import (
	"context"
	"testing"

	"dappco.re/go/scm/jobrunner"
)

func TestSpinnerDeterministicBehaviour(t *testing.T) {
	s := NewSpinner(ClothoConfig{Strategy: "clotho-verified"}, map[string]AgentConfig{
		"charon": {
			ForgejoUser:   "forge",
			Model:         "gpt-5.4",
			VerifyModel:   "gpt-5.3-codex-spark",
			SecurityLevel: "high",
		},
	})

	signal := &jobrunner.PipelineSignal{
		RepoOwner:       "core",
		RepoName:        "go-scm",
		PRState:         "OPEN",
		ThreadsTotal:    2,
		ThreadsResolved: 1,
		CheckStatus:     "PENDING",
		Mergeable:       "UNKNOWN",
		IsDraft:         true,
	}

	if got := s.DeterminePlan(signal, "charon"); got != RunModeClothoVerified {
		t.Fatalf("expected verified mode, got %q", got)
	}
	if name, _, ok := s.FindByForgejoUser("forge"); !ok || name != "charon" {
		t.Fatalf("expected forgejo lookup to resolve charon")
	}
	if got := s.GetVerifierModel("charon"); got != "gpt-5.3-codex-spark" {
		t.Fatalf("unexpected verifier model: %q", got)
	}
	ok, err := s.Weave(context.Background(), []byte("same"), []byte("same\n"))
	if err != nil || !ok {
		t.Fatalf("expected weave convergence: ok=%v err=%v", ok, err)
	}
}

func TestSpinnerResolveByForgejoUser(t *testing.T) {
	s := NewSpinner(ClothoConfig{}, map[string]AgentConfig{
		"charon": {
			ForgejoUser: "forge",
			Model:       "gpt-5.4",
			VerifyModel: "gpt-5.3-codex-spark",
		},
	})

	if got := s.GetVerifierModel("forge"); got != "gpt-5.3-codex-spark" {
		t.Fatalf("expected verifier model by forgejo user, got %q", got)
	}
}

func TestSpinnerDeterminePlanHonorsAgentOverridesUnderDirectStrategy(t *testing.T) {
	s := NewSpinner(ClothoConfig{Strategy: "direct"}, map[string]AgentConfig{
		"charon": {
			ForgejoUser:   "forge",
			Model:         "gpt-5.4",
			VerifyModel:   "gpt-5.3-codex-spark",
			SecurityLevel: "high",
		},
	})

	if got := s.DeterminePlan(&jobrunner.PipelineSignal{}, "charon"); got != RunModeClothoVerified {
		t.Fatalf("expected agent override to force verified mode, got %q", got)
	}
}

func TestSpinnerDeterminePlanIgnoresResolvedThreadsWhenOtherwiseClean(t *testing.T) {
	s := NewSpinner(ClothoConfig{Strategy: "clotho-verified"}, nil)

	signal := &jobrunner.PipelineSignal{
		PRState:        "OPEN",
		ThreadsTotal:    1,
		ThreadsResolved: 1,
		CheckStatus:     "SUCCESS",
		Mergeable:       "MERGEABLE",
	}

	if got := s.DeterminePlan(signal, "charon"); got != RunModeDirect {
		t.Fatalf("expected resolved threads to stay in direct mode, got %q", got)
	}
}
