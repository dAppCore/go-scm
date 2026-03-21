package repos

import (
	"testing"
	"time"

	"dappco.re/go/core/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── DefaultWorkConfig ──────────────────────────────────────────────

func TestDefaultWorkConfig_Good(t *testing.T) {
	wc := DefaultWorkConfig()
	assert.Equal(t, 1, wc.Version)
	assert.Equal(t, 5*time.Minute, wc.Sync.Interval)
	assert.True(t, wc.Sync.AutoPull)
	assert.False(t, wc.Sync.AutoPush)
	assert.True(t, wc.Sync.CloneMissing)
	assert.Equal(t, 2*time.Minute, wc.Agents.Heartbeat)
	assert.Equal(t, 10*time.Minute, wc.Agents.StaleAfter)
	assert.True(t, wc.Agents.WarnOnOverlap)
	assert.Contains(t, wc.Triggers, "on_activate")
	assert.Contains(t, wc.Triggers, "on_commit")
	assert.Contains(t, wc.Triggers, "scheduled")
}

// ── Load / Save round-trip ─────────────────────────────────────────

func TestWorkConfig_LoadSave_Good(t *testing.T) {
	m := io.NewMockMedium()
	_ = m.EnsureDir("/workspace/.core")

	wc := DefaultWorkConfig()
	wc.Sync.Interval = 10 * time.Minute
	wc.Sync.AutoPush = true

	err := SaveWorkConfig(m, "/workspace", wc)
	require.NoError(t, err)

	loaded, err := LoadWorkConfig(m, "/workspace")
	require.NoError(t, err)

	assert.Equal(t, 1, loaded.Version)
	assert.Equal(t, 10*time.Minute, loaded.Sync.Interval)
	assert.True(t, loaded.Sync.AutoPush)
	assert.True(t, loaded.Sync.AutoPull)
}

func TestWorkConfig_Load_Good_NoFile(t *testing.T) {
	m := io.NewMockMedium()
	_ = m.EnsureDir("/workspace/.core")

	wc, err := LoadWorkConfig(m, "/workspace")
	require.NoError(t, err)
	assert.Equal(t, DefaultWorkConfig().Sync.Interval, wc.Sync.Interval)
}

func TestWorkConfig_Load_Good_PartialOverride(t *testing.T) {
	m := io.NewMockMedium()
	_ = m.Write("/workspace/.core/work.yaml", `
version: 1
sync:
  interval: 30s
  auto_push: true
`)

	wc, err := LoadWorkConfig(m, "/workspace")
	require.NoError(t, err)

	assert.Equal(t, 30*time.Second, wc.Sync.Interval)
	assert.True(t, wc.Sync.AutoPush)
	// Defaults preserved for unset fields
	assert.True(t, wc.Sync.AutoPull)
	assert.True(t, wc.Sync.CloneMissing)
}

func TestWorkConfig_Load_Bad_InvalidYAML(t *testing.T) {
	m := io.NewMockMedium()
	_ = m.Write("/workspace/.core/work.yaml", "{{{{broken")

	_, err := LoadWorkConfig(m, "/workspace")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}

// ── HasTrigger ─────────────────────────────────────────────────────

func TestWorkConfig_HasTrigger_Good(t *testing.T) {
	wc := DefaultWorkConfig()
	assert.True(t, wc.HasTrigger("on_activate"))
	assert.True(t, wc.HasTrigger("on_commit"))
	assert.True(t, wc.HasTrigger("scheduled"))
}

func TestWorkConfig_HasTrigger_Bad_NotFound(t *testing.T) {
	wc := DefaultWorkConfig()
	assert.False(t, wc.HasTrigger("on_deploy"))
}

func TestWorkConfig_HasTrigger_Good_CustomTriggers(t *testing.T) {
	wc := &WorkConfig{
		Version:  1,
		Triggers: []string{"on_pr", "manual"},
	}
	assert.True(t, wc.HasTrigger("on_pr"))
	assert.True(t, wc.HasTrigger("manual"))
	assert.False(t, wc.HasTrigger("on_commit"))
}
