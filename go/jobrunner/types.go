// SPDX-License-Identifier: EUPL-1.2

package jobrunner

import (
	"context"
	"time"

	core "dappco.re/go"
)

// ActionResult carries the outcome of a handler execution.
type ActionResult struct {
	Action      string        `json:"action"`
	RepoOwner   string        `json:"repo_owner"`
	RepoName    string        `json:"repo_name"`
	EpicNumber  int           `json:"epic"`
	ChildNumber int           `json:"child"`
	PRNumber    int           `json:"pr"`
	Success     bool          `json:"success"`
	Error       string        `json:"error,omitempty"`
	Timestamp   time.Time     `json:"ts"`
	Duration    time.Duration `json:"duration_ms"`
	Cycle       int           `json:"cycle"`
}

// MarshalJSON encodes Duration in milliseconds so the JSON form matches the tag
// and the journal snapshots.
func (a ActionResult) MarshalJSON() ([]byte, error)  /* v090-result-boundary */ {
	type alias struct {
		Action      string    `json:"action"`
		RepoOwner   string    `json:"repo_owner"`
		RepoName    string    `json:"repo_name"`
		EpicNumber  int       `json:"epic"`
		ChildNumber int       `json:"child"`
		PRNumber    int       `json:"pr"`
		Success     bool      `json:"success"`
		Error       string    `json:"error,omitempty"`
		Timestamp   time.Time `json:"ts"`
		DurationMs  int64     `json:"duration_ms"`
		Cycle       int       `json:"cycle"`
	}

	r := core.JSONMarshal(alias{
		Action:      a.Action,
		RepoOwner:   a.RepoOwner,
		RepoName:    a.RepoName,
		EpicNumber:  a.EpicNumber,
		ChildNumber: a.ChildNumber,
		PRNumber:    a.PRNumber,
		Success:     a.Success,
		Error:       a.Error,
		Timestamp:   a.Timestamp,
		DurationMs:  a.Duration.Milliseconds(),
		Cycle:       a.Cycle,
	})
	if !r.OK {
		err, _ := r.Value.(error)
		return nil, err
	}
	data, _ := r.Value.([]byte)
	return data, nil
}

// UnmarshalJSON decodes Duration from milliseconds in the JSON form.
func (a *ActionResult) UnmarshalJSON(data []byte) error  /* v090-result-boundary */ {
	type alias struct {
		Action      string    `json:"action"`
		RepoOwner   string    `json:"repo_owner"`
		RepoName    string    `json:"repo_name"`
		EpicNumber  int       `json:"epic"`
		ChildNumber int       `json:"child"`
		PRNumber    int       `json:"pr"`
		Success     bool      `json:"success"`
		Error       string    `json:"error,omitempty"`
		Timestamp   time.Time `json:"ts"`
		DurationMs  int64     `json:"duration_ms"`
		Cycle       int       `json:"cycle"`
	}

	var out alias
	if r := core.JSONUnmarshal(data, &out); !r.OK {
		err, _ := r.Value.(error)
		return err
	}

	*a = ActionResult{
		Action:      out.Action,
		RepoOwner:   out.RepoOwner,
		RepoName:    out.RepoName,
		EpicNumber:  out.EpicNumber,
		ChildNumber: out.ChildNumber,
		PRNumber:    out.PRNumber,
		Success:     out.Success,
		Error:       out.Error,
		Timestamp:   out.Timestamp,
		Duration:    time.Duration(out.DurationMs) * time.Millisecond,
		Cycle:       out.Cycle,
	}
	return nil
}

// JobHandler processes a single pipeline signal.
type JobHandler interface {
	Name() string
	Match(signal *PipelineSignal) bool
	Execute(ctx context.Context, signal *PipelineSignal) (*ActionResult, error)
}

// JobSource produces pipeline signals.
type JobSource interface {
	Name() string
	Poll(ctx context.Context) ([]*PipelineSignal, error)
	Report(ctx context.Context, result *ActionResult) error
}

// PipelineSignal is the structural snapshot of a child issue/PR.
type PipelineSignal struct {
	EpicNumber      int
	ChildNumber     int
	PRNumber        int
	RepoOwner       string
	RepoName        string
	PRState         string
	IsDraft         bool
	Mergeable       string
	CheckStatus     string
	ThreadsTotal    int
	ThreadsResolved int
	LastCommitSHA   string
	LastCommitAt    time.Time
	LastReviewAt    time.Time
	NeedsCoding     bool
	Assignee        string
	IssueTitle      string
	IssueBody       string
	Type            string
	Success         bool
	Error           string
	Message         string
}

// HasUnresolvedThreads returns true if there are unresolved review threads.
func (s *PipelineSignal) HasUnresolvedThreads() bool {
	if s == nil {
		return false
	}
	return s.ThreadsTotal > s.ThreadsResolved
}

// RepoFullName returns "owner/repo".
func (s *PipelineSignal) RepoFullName() string {
	if s == nil {
		return ""
	}
	switch {
	case s.RepoOwner != "" && s.RepoName != "":
		return s.RepoOwner + "/" + s.RepoName
	case s.RepoOwner != "":
		return s.RepoOwner
	default:
		return s.RepoName
	}
}
