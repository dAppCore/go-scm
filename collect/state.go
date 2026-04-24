// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	// Note: encoding/json is retained for persisted state compatibility; core.JSON helpers do not expose MarshalIndent or streaming behavior.
	"encoding/json"
	// Note: errors.Is/New are retained for fs.ErrNotExist handling and stable state validation errors.
	"errors"
	// Note: fmt.Errorf is retained for wrapped state persistence errors.
	"fmt"
	// Note: io/fs is retained for fs.ErrNotExist from the configured coreio medium.
	"io/fs"
	// Note: filepath is retained for OS-specific state file normalization.
	"path/filepath"
	// Note: strings.TrimSpace is retained for state path validation without refactoring persistence setup.
	"strings"
	// Note: sync.Mutex protects the persisted state map and has no core equivalent.
	"sync"
	// Note: time.Time is retained for state timestamps serialized to disk.
	"time"

	coreio "dappco.re/go/core/io"
)

// State tracks collection progress for incremental runs.
type State struct {
	mu      sync.Mutex
	medium  coreio.Medium
	path    string
	entries map[string]*StateEntry
}

// StateEntry tracks state for one source.
type StateEntry struct {
	Source  string    `json:"source"`
	LastRun time.Time `json:"last_run"`
	LastID  string    `json:"last_id,omitempty"`
	Items   int       `json:"items"`
	Cursor  string    `json:"cursor,omitempty"`
}

// NewState creates a state tracker that persists to the given path.
func NewState(m coreio.Medium, path string) *State {
	if m == nil {
		m = coreio.NewMemoryMedium()
	}
	statePath := strings.TrimSpace(path)
	if statePath != "" {
		statePath = filepath.Clean(statePath)
	}
	return &State{medium: m, path: statePath, entries: make(map[string]*StateEntry)}
}

// Get returns a copy of the state for a source.
func (s *State) Get(source string) (*StateEntry, bool) {
	if s == nil {
		return nil, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[source]
	if !ok || entry == nil {
		return nil, false
	}
	cp := *entry
	return &cp, true
}

// Set updates state for a source.
func (s *State) Set(source string, entry *StateEntry) {
	if s == nil || entry == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *entry
	cp.Source = source
	if s.entries == nil {
		s.entries = make(map[string]*StateEntry)
	}
	s.entries[source] = &cp
}

// Load reads state from disk.
func (s *State) Load() error {
	if s == nil {
		return errors.New("collect.State.Load: state is required")
	}
	if s.medium == nil || s.path == "" {
		return nil
	}
	raw, err := s.medium.Read(s.path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			s.mu.Lock()
			s.entries = make(map[string]*StateEntry)
			s.mu.Unlock()
			return nil
		}
		return fmt.Errorf("collect.State.Load: read: %w", err)
	}
	var data map[string]*StateEntry
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return fmt.Errorf("collect.State.Load: unmarshal: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = data
	if s.entries == nil {
		s.entries = make(map[string]*StateEntry)
	}
	return nil
}

// Save writes state to disk.
func (s *State) Save() error {
	if s == nil {
		return errors.New("collect.State.Save: state is required")
	}
	if s.medium == nil || s.path == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	raw, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return fmt.Errorf("collect.State.Save: marshal: %w", err)
	}
	dir := filepath.Dir(s.path)
	if dir != "." {
		if err := s.medium.EnsureDir(dir); err != nil {
			return fmt.Errorf("collect.State.Save: ensure dir: %w", err)
		}
	}
	if err := s.medium.Write(s.path, string(raw)); err != nil {
		return fmt.Errorf("collect.State.Save: write: %w", err)
	}
	return nil
}
