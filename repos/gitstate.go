package repos

import (
	"path/filepath"
	"time"

	coreerr "forge.lthn.ai/core/go-log"
	"forge.lthn.ai/core/go-io"
	"gopkg.in/yaml.v3"
)

// GitState holds per-machine git sync state for a workspace.
// Stored at .core/git.yaml and .gitignored (not shared across machines).
type GitState struct {
	Version int                        `yaml:"version"`
	Repos   map[string]*RepoGitState   `yaml:"repos,omitempty"`
	Agents  map[string]*AgentState     `yaml:"agents,omitempty"`
}

// RepoGitState tracks the last known git state for a single repo.
type RepoGitState struct {
	LastPull time.Time `yaml:"last_pull,omitempty"`
	LastPush time.Time `yaml:"last_push,omitempty"`
	Branch   string    `yaml:"branch,omitempty"`
	Remote   string    `yaml:"remote,omitempty"`
	Ahead    int       `yaml:"ahead,omitempty"`
	Behind   int       `yaml:"behind,omitempty"`
}

// AgentState tracks which agent last touched which repos.
type AgentState struct {
	LastSeen time.Time `yaml:"last_seen"`
	Active   []string  `yaml:"active,omitempty"`
}

// LoadGitState reads .core/git.yaml from the given workspace root directory.
// Returns a new empty GitState if the file does not exist.
func LoadGitState(m io.Medium, root string) (*GitState, error) {
	path := filepath.Join(root, ".core", "git.yaml")

	if !m.Exists(path) {
		return NewGitState(), nil
	}

	content, err := m.Read(path)
	if err != nil {
		return nil, coreerr.E("repos.LoadGitState", "failed to read git state", err)
	}

	var gs GitState
	if err := yaml.Unmarshal([]byte(content), &gs); err != nil {
		return nil, coreerr.E("repos.LoadGitState", "failed to parse git state", err)
	}

	if gs.Repos == nil {
		gs.Repos = make(map[string]*RepoGitState)
	}
	if gs.Agents == nil {
		gs.Agents = make(map[string]*AgentState)
	}

	return &gs, nil
}

// SaveGitState writes .core/git.yaml to the given workspace root directory.
func SaveGitState(m io.Medium, root string, gs *GitState) error {
	coreDir := filepath.Join(root, ".core")
	if err := m.EnsureDir(coreDir); err != nil {
		return coreerr.E("repos.SaveGitState", "failed to create .core directory", err)
	}

	data, err := yaml.Marshal(gs)
	if err != nil {
		return coreerr.E("repos.SaveGitState", "failed to marshal git state", err)
	}

	path := filepath.Join(coreDir, "git.yaml")
	if err := m.Write(path, string(data)); err != nil {
		return coreerr.E("repos.SaveGitState", "failed to write git state", err)
	}

	return nil
}

// NewGitState returns a new empty GitState with version 1.
func NewGitState() *GitState {
	return &GitState{
		Version: 1,
		Repos:   make(map[string]*RepoGitState),
		Agents:  make(map[string]*AgentState),
	}
}

// Touch records a pull timestamp for the named repo.
func (gs *GitState) TouchPull(name string) {
	gs.ensureRepo(name).LastPull = time.Now()
}

// TouchPush records a push timestamp for the named repo.
func (gs *GitState) TouchPush(name string) {
	gs.ensureRepo(name).LastPush = time.Now()
}

// UpdateRepo records the current git status for a repo.
func (gs *GitState) UpdateRepo(name, branch, remote string, ahead, behind int) {
	r := gs.ensureRepo(name)
	r.Branch = branch
	r.Remote = remote
	r.Ahead = ahead
	r.Behind = behind
}

// Heartbeat records an agent's presence and active packages.
func (gs *GitState) Heartbeat(agentName string, active []string) {
	if gs.Agents == nil {
		gs.Agents = make(map[string]*AgentState)
	}
	gs.Agents[agentName] = &AgentState{
		LastSeen: time.Now(),
		Active:   active,
	}
}

// StaleAgents returns agent names whose last heartbeat is older than the given duration.
func (gs *GitState) StaleAgents(staleAfter time.Duration) []string {
	cutoff := time.Now().Add(-staleAfter)
	var stale []string
	for name, agent := range gs.Agents {
		if agent.LastSeen.Before(cutoff) {
			stale = append(stale, name)
		}
	}
	return stale
}

// ActiveAgentsFor returns agent names that have the given repo in their active list
// and are not stale.
func (gs *GitState) ActiveAgentsFor(repoName string, staleAfter time.Duration) []string {
	cutoff := time.Now().Add(-staleAfter)
	var agents []string
	for name, agent := range gs.Agents {
		if agent.LastSeen.Before(cutoff) {
			continue
		}
		for _, r := range agent.Active {
			if r == repoName {
				agents = append(agents, name)
				break
			}
		}
	}
	return agents
}

// NeedsPull returns true if the repo has never been pulled or was pulled before the given duration.
func (gs *GitState) NeedsPull(name string, maxAge time.Duration) bool {
	r, ok := gs.Repos[name]
	if !ok {
		return true
	}
	if r.LastPull.IsZero() {
		return true
	}
	return time.Since(r.LastPull) > maxAge
}

func (gs *GitState) ensureRepo(name string) *RepoGitState {
	if gs.Repos == nil {
		gs.Repos = make(map[string]*RepoGitState)
	}
	r, ok := gs.Repos[name]
	if !ok {
		r = &RepoGitState{}
		gs.Repos[name] = r
	}
	return r
}
