// SPDX-License-Identifier: EUPL-1.2

package agentci

import (
	"context"
	"errors"
	"testing"

	"dappco.re/go/core/scm/jobrunner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpinner_Weave_Good_ExactMatch(t *testing.T) {
	spinner := NewSpinner(ClothoConfig{ValidationThreshold: 0.85}, nil)

	ok, err := spinner.Weave(context.Background(), []byte("alpha beta gamma"), []byte("alpha beta gamma"))
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestSpinner_Weave_Good_ThresholdMatch(t *testing.T) {
	spinner := NewSpinner(ClothoConfig{ValidationThreshold: 0.8}, nil)

	ok, err := spinner.Weave(
		context.Background(),
		[]byte("alpha beta gamma delta epsilon zeta"),
		[]byte("alpha beta gamma delta epsilon eta"),
	)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestSpinner_Weave_Bad_ThresholdMismatch(t *testing.T) {
	spinner := NewSpinner(ClothoConfig{ValidationThreshold: 0.9}, nil)

	ok, err := spinner.Weave(
		context.Background(),
		[]byte("alpha beta gamma delta epsilon zeta"),
		[]byte("alpha beta gamma delta epsilon eta"),
	)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestSpinner_Weave_Good_EmptyOutputs(t *testing.T) {
	spinner := NewSpinner(ClothoConfig{}, nil)

	ok, err := spinner.Weave(context.Background(), nil, nil)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestSpinner_Weave_Bad_ZeroValueConfigDefaults(t *testing.T) {
	spinner := NewSpinner(ClothoConfig{}, nil)

	ok, err := spinner.Weave(
		context.Background(),
		[]byte("alpha beta gamma delta epsilon zeta"),
		[]byte("alpha beta gamma delta epsilon eta"),
	)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestSpinner_Weave_Good_ExplicitZeroThresholdPreserved(t *testing.T) {
	spinner := NewSpinner(ClothoConfig{Strategy: "clotho-verified", ValidationThreshold: 0.0}, nil)

	ok, err := spinner.Weave(
		context.Background(),
		[]byte("alpha beta gamma delta epsilon zeta"),
		[]byte("alpha beta gamma delta epsilon eta"),
	)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestSpinner_Weave_Bad_ContextCancelled(t *testing.T) {
	spinner := NewSpinner(ClothoConfig{}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ok, err := spinner.Weave(ctx, []byte("alpha"), []byte("alpha"))
	assert.False(t, ok)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}

func TestSpinner_DeterminePlan_Good(t *testing.T) {
	spinner := NewSpinner(ClothoConfig{Strategy: "clotho-verified"}, map[string]AgentConfig{
		"charon": {DualRun: true},
	})

	ok := spinner.DeterminePlan(&jobrunner.PipelineSignal{RepoName: "docs"}, "charon")
	assert.Equal(t, ModeDual, ok)
}

func TestSpinner_DeterminePlan_Good_HighSecurityLevel_Good(t *testing.T) {
	spinner := NewSpinner(ClothoConfig{Strategy: "clotho-verified"}, map[string]AgentConfig{
		"charon": {SecurityLevel: "high"},
	})

	ok := spinner.DeterminePlan(&jobrunner.PipelineSignal{RepoName: "docs"}, "charon")
	assert.Equal(t, ModeDual, ok)
}

func TestSpinner_DeterminePlan_Good_CriticalRepoIsCaseInsensitive(t *testing.T) {
	spinner := NewSpinner(ClothoConfig{Strategy: "clotho-verified"}, map[string]AgentConfig{
		"charon": {},
	})

	ok := spinner.DeterminePlan(&jobrunner.PipelineSignal{RepoName: "Security-Toolkit"}, "charon")
	assert.Equal(t, ModeDual, ok)
}

func TestSpinner_DeterminePlan_Good_StrategyIsCaseInsensitive(t *testing.T) {
	spinner := NewSpinner(ClothoConfig{Strategy: "ClOtHo-VeRiFiEd"}, map[string]AgentConfig{
		"charon": {DualRun: true},
	})

	ok := spinner.DeterminePlan(&jobrunner.PipelineSignal{RepoName: "docs"}, "charon")
	assert.Equal(t, ModeDual, ok)
}

func TestSpinner_DeterminePlan_Good_ForgejoUserLookup(t *testing.T) {
	spinner := NewSpinner(ClothoConfig{Strategy: "clotho-verified"}, map[string]AgentConfig{
		"charon": {ForgejoUser: "Charon", DualRun: true},
	})

	ok := spinner.DeterminePlan(&jobrunner.PipelineSignal{RepoName: "docs"}, "charon")
	assert.Equal(t, ModeDual, ok)
}

func TestSpinner_GetVerifierModel_Good_ForgejoUserLookup(t *testing.T) {
	spinner := NewSpinner(ClothoConfig{}, map[string]AgentConfig{
		"charon": {ForgejoUser: "Charon", VerifyModel: "gpt-5.1"},
	})

	assert.Equal(t, "gpt-5.1", spinner.GetVerifierModel("Charon"))
}

func TestSpinner_FindByForgejoUser_Good_CaseInsensitive(t *testing.T) {
	spinner := NewSpinner(ClothoConfig{}, map[string]AgentConfig{
		"charon": {ForgejoUser: "Charon"},
	})

	name, agent, ok := spinner.FindByForgejoUser("charon")
	require.True(t, ok)
	assert.Equal(t, "charon", name)
	assert.Equal(t, "Charon", agent.ForgejoUser)
}

func TestSpinner_FindByForgejoUser_Good_CaseInsensitiveKey(t *testing.T) {
	spinner := NewSpinner(ClothoConfig{}, map[string]AgentConfig{
		"Charon": {ForgejoUser: "other"},
	})

	name, agent, ok := spinner.FindByForgejoUser("charon")
	require.True(t, ok)
	assert.Equal(t, "Charon", name)
	assert.Equal(t, "other", agent.ForgejoUser)
}

func TestSpinner_Good_NilReceiverDefaults(t *testing.T) {
	var spinner *Spinner

	assert.Equal(t, ModeStandard, spinner.DeterminePlan(&jobrunner.PipelineSignal{RepoName: "docs"}, "charon"))
	assert.Equal(t, "gemini-1.5-pro", spinner.GetVerifierModel("charon"))

	name, agent, ok := spinner.FindByForgejoUser("charon")
	assert.Empty(t, name)
	assert.Empty(t, agent)
	assert.False(t, ok)

	converged, err := spinner.Weave(context.Background(), []byte("alpha"), []byte("alpha"))
	require.NoError(t, err)
	assert.True(t, converged)
}
