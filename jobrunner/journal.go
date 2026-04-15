// SPDX-License-Identifier: EUPL-1.2

package jobrunner

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Journal writes ActionResult entries to date-partitioned JSONL files.
type Journal struct {
	baseDir string
	mu      sync.Mutex
}

// JournalEntry is a single line in the JSONL audit log.
type JournalEntry struct {
	Timestamp string         `json:"ts"`
	Epic      int            `json:"epic"`
	Child     int            `json:"child"`
	PR        int            `json:"pr"`
	Repo      string         `json:"repo"`
	Action    string         `json:"action"`
	Signals   SignalSnapshot `json:"signals"`
	Result    ResultSnapshot `json:"result"`
	Cycle     int            `json:"cycle"`
}

// SignalSnapshot captures the structural state of a PR at the time of action.
type SignalSnapshot struct {
	PRState         string `json:"pr_state"`
	IsDraft         bool   `json:"is_draft"`
	CheckStatus     string `json:"check_status"`
	Mergeable       string `json:"mergeable"`
	ThreadsTotal    int    `json:"threads_total"`
	ThreadsResolved int    `json:"threads_resolved"`
}

// ResultSnapshot captures the outcome of an action.
type ResultSnapshot struct {
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
	DurationMs int64  `json:"duration_ms"`
}

// NewJournal creates a new Journal rooted at baseDir.
func NewJournal(baseDir string) (*Journal, error) {
	if baseDir == "" {
		return nil, errors.New("jobrunner.NewJournal: baseDir is required")
	}
	abs, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("jobrunner.NewJournal: resolve baseDir: %w", err)
	}
	return &Journal{baseDir: abs}, nil
}

// Append writes a journal entry for the given signal and result.
func (j *Journal) Append(signal *PipelineSignal, result *ActionResult) error {
	if j == nil {
		return errors.New("jobrunner.Journal.Append: journal is required")
	}
	if result == nil {
		return errors.New("jobrunner.Journal.Append: result is required")
	}

	ts := result.Timestamp
	if ts.IsZero() {
		ts = time.Now().UTC()
	} else {
		ts = ts.UTC()
	}

	entry := JournalEntry{
		Timestamp: ts.Format(time.RFC3339Nano),
		Epic:      result.EpicNumber,
		Child:     result.ChildNumber,
		PR:        result.PRNumber,
		Repo:      repoRef(signal),
		Action:    result.Action,
		Signals: SignalSnapshot{
			PRState:         signalValue(signal, func(s *PipelineSignal) string { return s.PRState }),
			IsDraft:         signalValue(signal, func(s *PipelineSignal) bool { return s.IsDraft }),
			CheckStatus:     signalValue(signal, func(s *PipelineSignal) string { return s.CheckStatus }),
			Mergeable:       signalValue(signal, func(s *PipelineSignal) string { return s.Mergeable }),
			ThreadsTotal:    signalValue(signal, func(s *PipelineSignal) int { return s.ThreadsTotal }),
			ThreadsResolved: signalValue(signal, func(s *PipelineSignal) int { return s.ThreadsResolved }),
		},
		Result: ResultSnapshot{
			Success:    result.Success,
			Error:      result.Error,
			DurationMs: result.Duration.Milliseconds(),
		},
		Cycle: result.Cycle,
	}

	payload, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("jobrunner.Journal.Append: marshal entry: %w", err)
	}

	filePath := filepath.Join(j.baseDir, ts.Format("2006"), ts.Format("01"), ts.Format("02")+".jsonl")

	j.mu.Lock()
	defer j.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return fmt.Errorf("jobrunner.Journal.Append: create directories: %w", err)
	}

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("jobrunner.Journal.Append: open journal: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(append(payload, '\n')); err != nil {
		return fmt.Errorf("jobrunner.Journal.Append: write journal: %w", err)
	}
	return nil
}

func repoRef(signal *PipelineSignal) string {
	if signal == nil {
		return ""
	}
	return signal.RepoFullName()
}

func signalValue[T any](signal *PipelineSignal, fn func(*PipelineSignal) T) T {
	var zero T
	if signal == nil {
		return zero
	}
	return fn(signal)
}
