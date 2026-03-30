// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	"context"
	"testing"
	"time"

	"dappco.re/go/core/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExcavator_Run_Good_ResumeSkipsCompleted(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	// Pre-populate state so source-a looks completed.
	cfg.State.Set("source-a", &StateEntry{
		Source:  "source-a",
		LastRun: time.Now().Add(-1 * time.Hour),
		Items:   10,
	})

	c1 := &mockCollector{name: "source-a", items: 10}
	c2 := &mockCollector{name: "source-b", items: 5}

	e := &Excavator{
		Collectors: []Collector{c1, c2},
		Resume:     true,
	}

	result, err := e.Run(context.Background(), cfg)

	require.NoError(t, err)
	assert.False(t, c1.called, "source-a should be skipped (already completed)")
	assert.True(t, c2.called, "source-b should run")
	assert.Equal(t, 5, result.Items)
	assert.Equal(t, 1, result.Skipped)
}

func TestExcavator_Run_Good_ResumeRunsIncomplete(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	// Pre-populate state with 0 items (incomplete).
	cfg.State.Set("source-a", &StateEntry{
		Source:  "source-a",
		LastRun: time.Now(),
		Items:   0,
	})

	c1 := &mockCollector{name: "source-a", items: 5}

	e := &Excavator{
		Collectors: []Collector{c1},
		Resume:     true,
	}

	result, err := e.Run(context.Background(), cfg)

	require.NoError(t, err)
	assert.True(t, c1.called, "source-a should run (0 items in previous run)")
	assert.Equal(t, 5, result.Items)
}

func TestExcavator_Run_Good_NilState(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.State = nil
	cfg.Limiter = nil

	c1 := &mockCollector{name: "source-a", items: 3}

	e := &Excavator{
		Collectors: []Collector{c1},
	}

	result, err := e.Run(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 3, result.Items)
}

func TestExcavator_Run_Good_NilDispatcher(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.Dispatcher = nil
	cfg.Limiter = nil

	c1 := &mockCollector{name: "source-a", items: 2}

	e := &Excavator{
		Collectors: []Collector{c1},
	}

	result, err := e.Run(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 2, result.Items)
}

func TestExcavator_Run_Good_ProgressEvents(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	var progressMsgs []string
	cfg.Dispatcher.On(EventProgress, func(e Event) {
		progressMsgs = append(progressMsgs, e.Message)
	})

	c1 := &mockCollector{name: "source-a", items: 1}
	c2 := &mockCollector{name: "source-b", items: 1}

	e := &Excavator{
		Collectors: []Collector{c1, c2},
	}

	_, err := e.Run(context.Background(), cfg)
	require.NoError(t, err)

	assert.Len(t, progressMsgs, 2)
	assert.Contains(t, progressMsgs[0], "1/2")
	assert.Contains(t, progressMsgs[1], "2/2")
}
