// SPDX-License-Identifier: EUPL-1.2

package jobrunner

import (
	// Note: AX-6 — Journal appends are serialized by a process-local mutex.
	"sync"
	// Note: AX-6 — Journal entries use UTC timestamps and date partitions.
	"time"

	// Note: AX-6 — Core provides JSON, path, filesystem, and structured error primitives.
	core "dappco.re/go"
)

const (
	sonarJournalJobrunnerJournalAppend = "jobrunner.Journal.Append"
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
		return nil, core.E("jobrunner.NewJournal", "baseDir is required", nil)
	}
	return &Journal{baseDir: absoluteJournalPath(baseDir)}, nil
}

// Append writes a journal entry for the given signal and result.
func (j *Journal) Append(signal *PipelineSignal, result *ActionResult) error {
	if j == nil {
		return core.E(sonarJournalJobrunnerJournalAppend, "journal is required", nil)
	}
	if result == nil {
		return core.E(sonarJournalJobrunnerJournalAppend, "result is required", nil)
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

	marshalResult := core.JSONMarshal(entry)
	if !marshalResult.OK {
		return core.E(sonarJournalJobrunnerJournalAppend, "marshal entry", resultCause(marshalResult))
	}
	payload, ok := marshalResult.Value.([]byte)
	if !ok {
		return core.E(sonarJournalJobrunnerJournalAppend, "marshal entry returned invalid payload", nil)
	}

	filePath := core.Path(j.baseDir, ts.Format("2006"), ts.Format("01"), ts.Format("02")+".jsonl")

	j.mu.Lock()
	defer j.mu.Unlock()

	fs := (&core.Fs{}).NewUnrestricted()
	if r := fs.EnsureDir(core.PathDir(filePath)); !r.OK {
		return core.E(sonarJournalJobrunnerJournalAppend, "create directories", resultCause(r))
	}
	if !fs.Exists(filePath) {
		if r := fs.WriteMode(filePath, "", 0o600); !r.OK {
			return core.E(sonarJournalJobrunnerJournalAppend, "create journal", resultCause(r))
		}
	}

	openResult := fs.Append(filePath)
	if !openResult.OK {
		return core.E(sonarJournalJobrunnerJournalAppend, "open journal", resultCause(openResult))
	}
	f, ok := openResult.Value.(journalWriteCloser)
	if !ok {
		return core.E(sonarJournalJobrunnerJournalAppend, "open journal returned invalid writer", nil)
	}
	defer func() {
		_ = f.Close()
	}()

	if _, err := f.Write(append(payload, '\n')); err != nil {
		return core.E(sonarJournalJobrunnerJournalAppend, "write journal", err)
	}
	return nil
}

func absoluteJournalPath(path string) string {
	if core.PathIsAbs(path) {
		return core.Path(path)
	}
	return core.Path(core.Env("DIR_CWD"), path)
}

func resultCause(r core.Result) error {
	if err, ok := r.Value.(error); ok {
		return err
	}
	return nil
}

type journalWriteCloser interface {
	Write([]byte) (int, error)
	Close() error
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
