// SPDX-License-Identifier: EUPL-1.2

package jobrunner

import (
	"bufio"
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
	json "dappco.re/go/core/scm/internal/ax/jsonx"
	os "dappco.re/go/core/scm/internal/ax/osx"
	strings "dappco.re/go/core/scm/internal/ax/stringsx"
	"regexp"
	"sort"
	"sync"
	"time"

	coreio "dappco.re/go/core/io"
	coreerr "dappco.re/go/core/log"
)

const (
	journalDateLayout      = "2006-01-02"
	journalTimestampLayout = "2006-01-02T15:04:05Z"
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

// JournalQueryOptions filters replay results from the journal.
type JournalQueryOptions struct {
	RepoOwner    string
	RepoName     string
	RepoFullName string
	Action       string
	Since        time.Time
	Until        time.Time
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
		Timestamp: result.Timestamp.UTC().Format(journalTimestampLayout),
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

	date := result.Timestamp.UTC().Format(journalDateLayout)
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

type journalQueryHit struct {
	entry     JournalEntry
	timestamp time.Time
	repo      string
}

// Query replays journal entries from the archive and applies the requested filters.
// Usage: Query(...)
func (j *Journal) Query(opts JournalQueryOptions) ([]JournalEntry, error) {
	if j == nil {
		return nil, coreerr.E("jobrunner.Journal.Query", "journal is required", nil)
	}

	j.mu.Lock()
	defer j.mu.Unlock()

	ownerFilter, repoFilter, err := normaliseJournalQueryRepo(opts)
	if err != nil {
		return nil, coreerr.E("jobrunner.Journal.Query", "normalise repo filter", err)
	}

	hits, err := j.collectQueryHits(opts, ownerFilter, repoFilter)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(hits, func(i, k int) bool {
		if hits[i].timestamp.Equal(hits[k].timestamp) {
			if hits[i].repo == hits[k].repo {
				if hits[i].entry.Action == hits[k].entry.Action {
					if hits[i].entry.Epic == hits[k].entry.Epic {
						if hits[i].entry.Child == hits[k].entry.Child {
							if hits[i].entry.PR == hits[k].entry.PR {
								return hits[i].entry.Cycle < hits[k].entry.Cycle
							}
							return hits[i].entry.PR < hits[k].entry.PR
						}
						return hits[i].entry.Child < hits[k].entry.Child
					}
					return hits[i].entry.Epic < hits[k].entry.Epic
				}
				return hits[i].entry.Action < hits[k].entry.Action
			}
			return hits[i].repo < hits[k].repo
		}
		return hits[i].timestamp.Before(hits[k].timestamp)
	})

	entries := make([]JournalEntry, len(hits))
	for i, hit := range hits {
		entries[i] = hit.entry
	}
	return entries, nil
}

func (j *Journal) collectQueryHits(opts JournalQueryOptions, ownerFilter, repoFilter string) ([]journalQueryHit, error) {
	entries, err := os.ReadDir(j.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, coreerr.E("jobrunner.Journal.Query", "read journal base directory", err)
	}
	sort.Slice(entries, func(i, k int) bool { return entries[i].Name() < entries[k].Name() })

	var hits []journalQueryHit
	for _, ownerEntry := range entries {
		if !ownerEntry.IsDir() {
			continue
		}
		owner := ownerEntry.Name()
		if ownerFilter != "" && owner != ownerFilter {
			continue
		}

		repoPath := filepath.Join(j.baseDir, owner)
		repos, err := os.ReadDir(repoPath)
		if err != nil {
			return nil, coreerr.E("jobrunner.Journal.Query", "read owner directory", err)
		}
		sort.Slice(repos, func(i, k int) bool { return repos[i].Name() < repos[k].Name() })

		for _, repoEntry := range repos {
			if !repoEntry.IsDir() {
				continue
			}
			repo := repoEntry.Name()
			if repoFilter != "" && repo != repoFilter {
				continue
			}

			repoDir := filepath.Join(repoPath, repo)
			files, err := os.ReadDir(repoDir)
			if err != nil {
				return nil, coreerr.E("jobrunner.Journal.Query", "read repo directory", err)
			}
			sort.Slice(files, func(i, k int) bool { return files[i].Name() < files[k].Name() })

			for _, fileEntry := range files {
				if fileEntry.IsDir() || !strings.HasSuffix(fileEntry.Name(), ".jsonl") {
					continue
				}

				path := filepath.Join(repoDir, fileEntry.Name())
				fileHits, err := j.readQueryFile(path, opts, owner+"/"+repo)
				if err != nil {
					return nil, err
				}
				hits = append(hits, fileHits...)
			}
		}
	}

	return hits, nil
}

func (j *Journal) readQueryFile(path string, opts JournalQueryOptions, repo string) ([]journalQueryHit, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, coreerr.E("jobrunner.Journal.Query", "read journal file", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	var hits []journalQueryHit
	for scanner.Scan() {
		var entry JournalEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			return nil, coreerr.E("jobrunner.Journal.Query", "decode journal entry", err)
		}

		ts, err := time.Parse(journalTimestampLayout, entry.Timestamp)
		if err != nil {
			return nil, coreerr.E("jobrunner.Journal.Query", "parse journal timestamp", err)
		}
		if !journalEntryMatches(opts, entry, ts) {
			continue
		}

		hits = append(hits, journalQueryHit{
			entry:     entry,
			timestamp: ts,
			repo:      repo,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, coreerr.E("jobrunner.Journal.Query", "scan journal file", err)
	}

	return hits, nil
}

func normaliseJournalQueryRepo(opts JournalQueryOptions) (string, string, error) {
	owner := opts.RepoOwner
	repo := opts.RepoName

	if opts.RepoFullName != "" {
		parts := strings.SplitN(opts.RepoFullName, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return "", "", coreerr.E("jobrunner.normaliseJournalQueryRepo", "repo full name must be owner/repo", nil)
		}
		if owner != "" && owner != parts[0] {
			return "", "", coreerr.E("jobrunner.normaliseJournalQueryRepo", "repo owner does not match repo full name", nil)
		}
		if repo != "" && repo != parts[1] {
			return "", "", coreerr.E("jobrunner.normaliseJournalQueryRepo", "repo name does not match repo full name", nil)
		}
		owner = parts[0]
		repo = parts[1]
	}

	if owner != "" {
		clean, err := sanitizePathComponent(owner)
		if err != nil {
			return "", "", err
		}
		owner = clean
	}
	if repo != "" {
		clean, err := sanitizePathComponent(repo)
		if err != nil {
			return "", "", err
		}
		repo = clean
	}

	return owner, repo, nil
}

func journalEntryMatches(opts JournalQueryOptions, entry JournalEntry, ts time.Time) bool {
	if opts.Action != "" && entry.Action != opts.Action {
		return false
	}
	if !opts.Since.IsZero() && ts.Before(opts.Since) {
		return false
	}
	if !opts.Until.IsZero() && ts.After(opts.Until) {
		return false
	}
	return true
}
