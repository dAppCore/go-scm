// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	json "dappco.re/go/core/scm/internal/ax/jsonx"
	"sync"
	"time"

	"dappco.re/go/core/io"
	core "dappco.re/go/core/log"
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
)

// State tracks collection progress for incremental runs.
// It persists entries to disk so that subsequent runs can resume
// where they left off.
type State struct {
	mu      sync.Mutex
	medium  io.Medium
	path    string
	entries map[string]*StateEntry
}

// StateEntry tracks state for one source.
type StateEntry struct {
	// Source identifies the collector.
	Source string `json:"source"`

	// LastRun is the timestamp of the last successful run.
	LastRun time.Time `json:"last_run"`

	// LastID is an opaque identifier for the last item processed.
	LastID string `json:"last_id,omitempty"`

	// Items is the total number of items collected so far.
	Items int `json:"items"`

	// Cursor is an opaque pagination cursor for resumption.
	Cursor string `json:"cursor,omitempty"`
}

// NewState creates a state tracker that persists to the given path
// using the provided storage medium.
// Usage: NewState(...)
func NewState(m io.Medium, path string) *State {
	if m == nil {
		m = io.Local
	}
	return &State{
		medium:  m,
		path:    path,
		entries: make(map[string]*StateEntry),
	}
}

// Load reads state from disk. If the file does not exist, the state
// is initialised as empty without error.
// Usage: Load(...)
func (s *State) Load() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.medium == nil {
		s.medium = io.Local
	}

	if !s.medium.IsFile(s.path) {
		if s.entries == nil {
			s.entries = make(map[string]*StateEntry)
		}
		return nil
	}

	data, err := s.medium.Read(s.path)
	if err != nil {
		return core.E("collect.State.Load", "failed to read state file", err)
	}

	var entries map[string]*StateEntry
	if err := json.Unmarshal([]byte(data), &entries); err != nil {
		return core.E("collect.State.Load", "failed to parse state file", err)
	}

	if entries == nil {
		entries = make(map[string]*StateEntry)
	}
	s.entries = entries
	return nil
}

// Save writes state to disk.
// Usage: Save(...)
func (s *State) Save() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.medium == nil {
		s.medium = io.Local
	}

	if s.entries == nil {
		s.entries = make(map[string]*StateEntry)
	}
	if err := s.medium.EnsureDir(filepath.Dir(s.path)); err != nil {
		return core.E("collect.State.Save", "failed to create state directory", err)
	}

	data, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return core.E("collect.State.Save", "failed to marshal state", err)
	}

	if err := s.medium.Write(s.path, string(data)); err != nil {
		return core.E("collect.State.Save", "failed to write state file", err)
	}

	return nil
}

// Get returns a copy of the state for a source. The second return value
// indicates whether the entry was found.
// Usage: Get(...)
func (s *State) Get(source string) (*StateEntry, bool) {
	if s == nil {
		return nil, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.entries == nil {
		return nil, false
	}
	entry, ok := s.entries[source]
	if !ok {
		return nil, false
	}
	// Return a copy to avoid callers mutating internal state.
	cp := *entry
	return &cp, true
}

// Set updates state for a source.
// Usage: Set(...)
func (s *State) Set(source string, entry *StateEntry) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.entries == nil {
		s.entries = make(map[string]*StateEntry)
	}
	if entry == nil {
		delete(s.entries, source)
		return
	}
	cp := *entry
	s.entries[source] = &cp
}
