package repos

import (
	"testing"
	"time"

	"dappco.re/go/core/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── NewGitState ────────────────────────────────────────────────────

func TestNewGitState_Good(t *testing.T) {
	gs := NewGitState()
	assert.Equal(t, 1, gs.Version)
	assert.NotNil(t, gs.Repos)
	assert.NotNil(t, gs.Agents)
	assert.Empty(t, gs.Repos)
	assert.Empty(t, gs.Agents)
}

// ── Load / Save round-trip ─────────────────────────────────────────

func TestGitState_LoadSave_Good(t *testing.T) {
	m := io.NewMockMedium()
	_ = m.EnsureDir("/workspace/.core")

	gs := NewGitState()
	gs.UpdateRepo("core-php", "main", "origin", 2, 0)
	gs.Heartbeat("cladius", []string{"core-php", "core-tenant"})

	err := SaveGitState(m, "/workspace", gs)
	require.NoError(t, err)

	loaded, err := LoadGitState(m, "/workspace")
	require.NoError(t, err)

	assert.Equal(t, 1, loaded.Version)
	assert.Contains(t, loaded.Repos, "core-php")
	assert.Equal(t, "main", loaded.Repos["core-php"].Branch)
	assert.Equal(t, "origin", loaded.Repos["core-php"].Remote)
	assert.Equal(t, 2, loaded.Repos["core-php"].Ahead)
	assert.Equal(t, 0, loaded.Repos["core-php"].Behind)

	assert.Contains(t, loaded.Agents, "cladius")
	assert.Equal(t, []string{"core-php", "core-tenant"}, loaded.Agents["cladius"].Active)
}

func TestGitState_Load_Good_NoFile(t *testing.T) {
	m := io.NewMockMedium()
	_ = m.EnsureDir("/workspace/.core")

	gs, err := LoadGitState(m, "/workspace")
	require.NoError(t, err)
	assert.Equal(t, 1, gs.Version)
	assert.Empty(t, gs.Repos)
	assert.Empty(t, gs.Agents)
}

func TestGitState_Load_Bad_InvalidYAML(t *testing.T) {
	m := io.NewMockMedium()
	_ = m.Write("/workspace/.core/git.yaml", "{{{{not yaml")

	_, err := LoadGitState(m, "/workspace")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}

// ── TouchPull / TouchPush ──────────────────────────────────────────

func TestGitState_TouchPull_Good(t *testing.T) {
	gs := NewGitState()
	before := time.Now()

	gs.TouchPull("core-php")

	r := gs.Repos["core-php"]
	require.NotNil(t, r)
	assert.False(t, r.LastPull.IsZero())
	assert.True(t, r.LastPull.After(before) || r.LastPull.Equal(before))
}

func TestGitState_TouchPush_Good(t *testing.T) {
	gs := NewGitState()
	before := time.Now()

	gs.TouchPush("core-php")

	r := gs.Repos["core-php"]
	require.NotNil(t, r)
	assert.False(t, r.LastPush.IsZero())
	assert.True(t, r.LastPush.After(before) || r.LastPush.Equal(before))
}

// ── UpdateRepo ─────────────────────────────────────────────────────

func TestGitState_UpdateRepo_Good(t *testing.T) {
	gs := NewGitState()
	gs.UpdateRepo("core-admin", "develop", "upstream", 3, 1)

	r := gs.Repos["core-admin"]
	require.NotNil(t, r)
	assert.Equal(t, "develop", r.Branch)
	assert.Equal(t, "upstream", r.Remote)
	assert.Equal(t, 3, r.Ahead)
	assert.Equal(t, 1, r.Behind)
}

func TestGitState_UpdateRepo_Good_Overwrite(t *testing.T) {
	gs := NewGitState()
	gs.UpdateRepo("core-php", "main", "origin", 1, 0)
	gs.UpdateRepo("core-php", "main", "origin", 0, 0)

	assert.Equal(t, 0, gs.Repos["core-php"].Ahead)
}

// ── Heartbeat ──────────────────────────────────────────────────────

func TestGitState_Heartbeat_Good(t *testing.T) {
	gs := NewGitState()
	before := time.Now()

	gs.Heartbeat("athena", []string{"core-bio"})

	agent := gs.Agents["athena"]
	require.NotNil(t, agent)
	assert.Equal(t, []string{"core-bio"}, agent.Active)
	assert.True(t, agent.LastSeen.After(before) || agent.LastSeen.Equal(before))
}

func TestGitState_Heartbeat_Good_Updates(t *testing.T) {
	gs := NewGitState()
	gs.Heartbeat("cladius", []string{"core-php"})
	gs.Heartbeat("cladius", []string{"core-php", "core-tenant"})

	assert.Equal(t, []string{"core-php", "core-tenant"}, gs.Agents["cladius"].Active)
}

// ── StaleAgents ────────────────────────────────────────────────────

func TestGitState_StaleAgents_Good(t *testing.T) {
	gs := NewGitState()
	gs.Agents["fresh"] = &AgentState{
		LastSeen: time.Now(),
		Active:   []string{"core-php"},
	}
	gs.Agents["stale"] = &AgentState{
		LastSeen: time.Now().Add(-20 * time.Minute),
		Active:   []string{"core-bio"},
	}

	stale := gs.StaleAgents(10 * time.Minute)
	assert.Contains(t, stale, "stale")
	assert.NotContains(t, stale, "fresh")
}

func TestGitState_StaleAgents_Good_NoneStale(t *testing.T) {
	gs := NewGitState()
	gs.Heartbeat("cladius", []string{"core-php"})

	stale := gs.StaleAgents(10 * time.Minute)
	assert.Empty(t, stale)
}

// ── ActiveAgentsFor ────────────────────────────────────────────────

func TestGitState_ActiveAgentsFor_Good(t *testing.T) {
	gs := NewGitState()
	gs.Heartbeat("cladius", []string{"core-php", "core-tenant"})
	gs.Heartbeat("athena", []string{"core-bio"})

	agents := gs.ActiveAgentsFor("core-php", 10*time.Minute)
	assert.Contains(t, agents, "cladius")
	assert.NotContains(t, agents, "athena")
}

func TestGitState_ActiveAgentsFor_Good_IgnoresStale(t *testing.T) {
	gs := NewGitState()
	gs.Agents["gone"] = &AgentState{
		LastSeen: time.Now().Add(-20 * time.Minute),
		Active:   []string{"core-php"},
	}

	agents := gs.ActiveAgentsFor("core-php", 10*time.Minute)
	assert.Empty(t, agents)
}

func TestGitState_ActiveAgentsFor_Good_NoMatch(t *testing.T) {
	gs := NewGitState()
	gs.Heartbeat("cladius", []string{"core-php"})

	agents := gs.ActiveAgentsFor("core-bio", 10*time.Minute)
	assert.Empty(t, agents)
}

// ── NeedsPull ──────────────────────────────────────────────────────

func TestGitState_NeedsPull_Good_NeverPulled(t *testing.T) {
	gs := NewGitState()
	assert.True(t, gs.NeedsPull("core-php", 5*time.Minute))
}

func TestGitState_NeedsPull_Good_RecentPull(t *testing.T) {
	gs := NewGitState()
	gs.TouchPull("core-php")
	assert.False(t, gs.NeedsPull("core-php", 5*time.Minute))
}

func TestGitState_NeedsPull_Good_StalePull(t *testing.T) {
	gs := NewGitState()
	gs.Repos["core-php"] = &RepoGitState{
		LastPull: time.Now().Add(-10 * time.Minute),
	}
	assert.True(t, gs.NeedsPull("core-php", 5*time.Minute))
}
