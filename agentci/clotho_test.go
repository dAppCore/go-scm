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
