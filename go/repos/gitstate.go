// SPDX-License-Identifier: EUPL-1.2

package repos

import (
	"time"

	core "dappco.re/go"
	coreio "dappco.re/go/io"
	"gopkg.in/yaml.v3"
)

type RepoGitState struct {
	LastPull time.Time `yaml:"last_pull,omitempty"`
	LastPush time.Time `yaml:"last_push,omitempty"`
	Branch   string    `yaml:"branch,omitempty"`
	Remote   string    `yaml:"remote,omitempty"`
	Ahead    int       `yaml:"ahead,omitempty"`
	Behind   int       `yaml:"behind,omitempty"`
}

type AgentState struct {
	LastSeen time.Time `yaml:"last_seen"`
	Active   []string  `yaml:"active,omitempty"`
}

type GitState struct {
	Version int                      `yaml:"version"`
	Repos   map[string]*RepoGitState `yaml:"repos,omitempty"`
	Agents  map[string]*AgentState   `yaml:"agents,omitempty"`
}

func NewGitState() *GitState {
	return &GitState{Version: 1, Repos: map[string]*RepoGitState{}, Agents: map[string]*AgentState{}}
}

func (gs *GitState) ensure() {
	if gs.Repos == nil {
		gs.Repos = map[string]*RepoGitState{}
	}
	if gs.Agents == nil {
		gs.Agents = map[string]*AgentState{}
	}
	if gs.Version == 0 {
		gs.Version = 1
	}
}

func (gs *GitState) TouchPull(name string) {
	gs.ensure()
	st := gs.repo(name)
	st.LastPull = time.Now().UTC()
}
func (gs *GitState) TouchPush(name string) {
	gs.ensure()
	st := gs.repo(name)
	st.LastPush = time.Now().UTC()
}

func (gs *GitState) UpdateRepo(name, branch, remote string, ahead, behind int) {
	gs.ensure()
	st := gs.repo(name)
	st.Branch, st.Remote, st.Ahead, st.Behind = branch, remote, ahead, behind
}

func (gs *GitState) repo(name string) *RepoGitState {
	if st, ok := gs.Repos[name]; ok && st != nil {
		return st
	}
	st := &RepoGitState{}
	gs.Repos[name] = st
	return st
}

func (gs *GitState) Heartbeat(agentName string, active []string) {
	gs.ensure()
	gs.Agents[agentName] = &AgentState{LastSeen: time.Now().UTC(), Active: append([]string(nil), active...)}
}

func (gs *GitState) StaleAgents(staleAfter time.Duration) []string {
	if gs == nil {
		return nil
	}
	cutoff := time.Now().UTC().Add(-staleAfter)
	var out []string
	for name, agent := range gs.Agents {
		if agent == nil || agent.LastSeen.Before(cutoff) {
			out = append(out, name)
		}
	}
	return out
}

func (gs *GitState) ActiveAgentsFor(repoName string, staleAfter time.Duration) []string {
	if gs == nil {
		return nil
	}
	cutoff := time.Now().UTC().Add(-staleAfter)
	var out []string
	for name, agent := range gs.Agents {
		if agent == nil || agent.LastSeen.Before(cutoff) {
			continue
		}
		for _, active := range agent.Active {
			if active == repoName {
				out = append(out, name)
				break
			}
		}
	}
	return out
}

func (gs *GitState) NeedsPull(name string, maxAge time.Duration) bool {
	if gs == nil {
		return true
	}
	st, ok := gs.Repos[name]
	if !ok || st == nil || st.LastPull.IsZero() {
		return true
	}
	return time.Since(st.LastPull) > maxAge
}

func LoadGitState(m coreio.Medium, root string) (*GitState, error)  /* v090-result-boundary */ {
	if m == nil {
		return nil, core.E("repos.LoadGitState", "medium is required", nil)
	}
	raw, err := m.Read(core.PathJoin(root, ".core", "git.yaml"))
	if err != nil {
		return NewGitState(), nil
	}
	var gs GitState
	if err := yaml.Unmarshal([]byte(raw), &gs); err != nil {
		return nil, err
	}
	gs.ensure()
	return &gs, nil
}

func SaveGitState(m coreio.Medium, root string, gs *GitState) error  /* v090-result-boundary */ {
	if m == nil {
		return core.E("repos.SaveGitState", "medium is required", nil)
	}
	if gs == nil {
		gs = NewGitState()
	}
	raw, err := yaml.Marshal(gs)
	if err != nil {
		return err
	}
	return m.Write(core.PathJoin(root, ".core", "git.yaml"), string(raw))
}
