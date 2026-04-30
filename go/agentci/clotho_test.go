// SPDX-License-Identifier: EUPL-1.2

package agentci

import (
	// Note: context.Context is retained in tests to exercise Spinner.Weave through its public API.
	"context"
	// Note: testing is the standard Go test harness.
	"testing"

	"dappco.re/go/scm/jobrunner"
)

const (
	sonarClothoTestGpt53CodexSpark = "gpt-5.3-codex-spark"
	sonarClothoTestGpt54           = "gpt-5.4"
)

func TestSpinnerDeterministicBehaviour(t *testing.T) {
	s := NewSpinner(ClothoConfig{Strategy: "clotho-verified"}, map[string]AgentConfig{
		"charon": {
			ForgejoUser:   "forge",
			Model:         sonarClothoTestGpt54,
			VerifyModel:   sonarClothoTestGpt53CodexSpark,
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
	if got := s.GetVerifierModel("charon"); got != sonarClothoTestGpt53CodexSpark {
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
			Model:       sonarClothoTestGpt54,
			VerifyModel: sonarClothoTestGpt53CodexSpark,
		},
	})

	if got := s.GetVerifierModel("forge"); got != sonarClothoTestGpt53CodexSpark {
		t.Fatalf("expected verifier model by forgejo user, got %q", got)
	}
}

func TestSpinnerGetVerifierModelReturnsOnlySecondaryModel(t *testing.T) {
	s := NewSpinner(ClothoConfig{}, map[string]AgentConfig{
		"charon": {
			ForgejoUser: "forge",
			Model:       sonarClothoTestGpt54,
		},
	})

	if got := s.GetVerifierModel("charon"); got != "" {
		t.Fatalf("expected empty verifier model when no secondary model is configured, got %q", got)
	}
}

func TestNewSpinnerCopiesAgentSlices(t *testing.T) {
	agents := map[string]AgentConfig{
		"charon": {
			ForgejoUser: "forge",
			Roles:       []string{"dispatch", "review"},
		},
	}

	s := NewSpinner(ClothoConfig{}, agents)
	agents["charon"] = AgentConfig{ForgejoUser: "mutated", Roles: []string{"other"}}
	agents["new"] = AgentConfig{ForgejoUser: "extra"}

	got, ok := s.Agents["charon"]
	if !ok {
		t.Fatal("expected spinner to keep original agent")
	}
	if got.ForgejoUser != "forge" {
		t.Fatalf("expected spinner to retain original agent data, got %#v", got)
	}

	mutated := agents["charon"]
	mutated.Roles[0] = "mutated"
	agents["charon"] = mutated
	if got.Roles[0] != "dispatch" || got.Roles[1] != "review" {
		t.Fatalf("expected spinner roles to be detached from caller mutations, got %#v", got.Roles)
	}
}

func TestSpinnerDeterminePlanHonorsAgentOverridesUnderDirectStrategy(t *testing.T) {
	s := NewSpinner(ClothoConfig{Strategy: "direct"}, map[string]AgentConfig{
		"charon": {
			ForgejoUser:   "forge",
			Model:         sonarClothoTestGpt54,
			VerifyModel:   sonarClothoTestGpt53CodexSpark,
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
		PRState:         "OPEN",
		ThreadsTotal:    1,
		ThreadsResolved: 1,
		CheckStatus:     "SUCCESS",
		Mergeable:       "MERGEABLE",
	}

	if got := s.DeterminePlan(signal, "charon"); got != RunModeDirect {
		t.Fatalf("expected resolved threads to stay in direct mode, got %q", got)
	}
}
