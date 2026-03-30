// SPDX-License-Identifier: EUPL-1.2

package jobrunner

import (
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
	json "dappco.re/go/core/scm/internal/ax/jsonx"
	os "dappco.re/go/core/scm/internal/ax/osx"
	strings "dappco.re/go/core/scm/internal/ax/stringsx"
	"regexp"
	"sync"

	coreio "dappco.re/go/core/io"
	coreerr "dappco.re/go/core/log"
)

// validPathComponent matches safe repo owner/name characters (alphanumeric, hyphen, underscore, dot).
var validPathComponent = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)

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

// Journal writes ActionResult entries to date-partitioned JSONL files.
type Journal struct {
	baseDir string
	mu      sync.Mutex
}

// NewJournal creates a new Journal rooted at baseDir.
// Usage: NewJournal(...)
func NewJournal(baseDir string) (*Journal, error) {
	if baseDir == "" {
		return nil, coreerr.E("jobrunner.NewJournal", "base directory is required", nil)
	}
	return &Journal{baseDir: baseDir}, nil
}

// sanitizePathComponent validates a single path component (owner or repo name)
// to prevent path traversal attacks. It rejects "..", empty strings, paths
// containing separators, and any value outside the safe character set.
func sanitizePathComponent(name string) (string, error) {
	// Reject empty or whitespace-only values.
	if name == "" || strings.TrimSpace(name) == "" {
		return "", coreerr.E("jobrunner.sanitizePathComponent", "invalid path component: "+name, nil)
	}

	// Reject inputs containing path separators (directory traversal attempt).
	if strings.ContainsAny(name, `/\`) {
		return "", coreerr.E("jobrunner.sanitizePathComponent", "path component contains directory separator: "+name, nil)
	}

	// Use filepath.Clean to normalize (e.g., collapse redundant dots).
	clean := filepath.Clean(name)

	// Reject traversal components.
	if clean == "." || clean == ".." {
		return "", coreerr.E("jobrunner.sanitizePathComponent", "invalid path component: "+name, nil)
	}

	// Validate against the safe character set.
	if !validPathComponent.MatchString(clean) {
		return "", coreerr.E("jobrunner.sanitizePathComponent", "path component contains invalid characters: "+name, nil)
	}

	return clean, nil
}

// Append writes a journal entry for the given signal and result.
// Usage: Append(...)
func (j *Journal) Append(signal *PipelineSignal, result *ActionResult) error {
	if signal == nil {
		return coreerr.E("jobrunner.Journal.Append", "signal is required", nil)
	}
	if result == nil {
		return coreerr.E("jobrunner.Journal.Append", "result is required", nil)
	}

	entry := JournalEntry{
		Timestamp: result.Timestamp.UTC().Format("2006-01-02T15:04:05Z"),
		Epic:      signal.EpicNumber,
		Child:     signal.ChildNumber,
		PR:        signal.PRNumber,
		Repo:      signal.RepoFullName(),
		Action:    result.Action,
		Signals: SignalSnapshot{
			PRState:         signal.PRState,
			IsDraft:         signal.IsDraft,
			CheckStatus:     signal.CheckStatus,
			Mergeable:       signal.Mergeable,
			ThreadsTotal:    signal.ThreadsTotal,
			ThreadsResolved: signal.ThreadsResolved,
		},
		Result: ResultSnapshot{
			Success:    result.Success,
			Error:      result.Error,
			DurationMs: result.Duration.Milliseconds(),
		},
		Cycle: result.Cycle,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return coreerr.E("jobrunner.Journal.Append", "marshal journal entry", err)
	}
	data = append(data, '\n')

	// Sanitize path components to prevent path traversal (CVE: issue #46).
	owner, err := sanitizePathComponent(signal.RepoOwner)
	if err != nil {
		return coreerr.E("jobrunner.Journal.Append", "invalid repo owner", err)
	}
	repo, err := sanitizePathComponent(signal.RepoName)
	if err != nil {
		return coreerr.E("jobrunner.Journal.Append", "invalid repo name", err)
	}

	date := result.Timestamp.UTC().Format("2006-01-02")
	dir := filepath.Join(j.baseDir, owner, repo)

	// Resolve to absolute path and verify it stays within baseDir.
	absBase, err := filepath.Abs(j.baseDir)
	if err != nil {
		return coreerr.E("jobrunner.Journal.Append", "resolve base directory", err)
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return coreerr.E("jobrunner.Journal.Append", "resolve journal directory", err)
	}
	if !strings.HasPrefix(absDir, absBase+string(filepath.Separator)) {
		return coreerr.E("jobrunner.Journal.Append", "journal path escapes base directory", nil)
	}

	j.mu.Lock()
	defer j.mu.Unlock()

	if err := coreio.Local.EnsureDir(dir); err != nil {
		return coreerr.E("jobrunner.Journal.Append", "create journal directory", err)
	}

	path := filepath.Join(dir, date+".jsonl")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return coreerr.E("jobrunner.Journal.Append", "open journal file", err)
	}
	defer func() { _ = f.Close() }()

	_, err = f.Write(data)
	return err
}
