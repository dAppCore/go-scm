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
