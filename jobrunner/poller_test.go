// SPDX-License-Identifier: EUPL-1.2

package jobrunner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type stubSource struct {
	name      string
	signals   []*PipelineSignal
	reports   []*ActionResult
	reportErr error
}

func (s *stubSource) Name() string { return s.name }

func (s *stubSource) Poll(context.Context) ([]*PipelineSignal, error) {
	return s.signals, nil
}

func (s *stubSource) Report(_ context.Context, result *ActionResult) error {
	s.reports = append(s.reports, result)
	return s.reportErr
}

type stubHandler struct {
	name        string
	matchSignal *PipelineSignal
	result      *ActionResult
	executed    []*PipelineSignal
}

func (h *stubHandler) Name() string { return h.name }

func (h *stubHandler) Match(signal *PipelineSignal) bool {
	if h.matchSignal == nil || signal == nil {
		return false
	}
	return h.matchSignal.PRNumber == signal.PRNumber
}

func (h *stubHandler) Execute(_ context.Context, signal *PipelineSignal) (*ActionResult, error) {
	h.executed = append(h.executed, signal)
	return h.result, nil
}

func TestPollerRunOnceDispatchesAndReports(t *testing.T) {
	journal, err := NewJournal(t.TempDir())
	if err != nil {
		t.Fatalf("new journal: %v", err)
	}

	signal := &PipelineSignal{
		EpicNumber:      1,
		ChildNumber:     2,
		PRNumber:        3,
		RepoOwner:       "core",
		RepoName:        "go-scm",
		PRState:         "OPEN",
		CheckStatus:     "SUCCESS",
		Mergeable:       "MERGEABLE",
		ThreadsTotal:    1,
		ThreadsResolved: 1,
	}
	result := &ActionResult{
		Action:      "dispatch",
		RepoOwner:   "core",
		RepoName:    "go-scm",
		EpicNumber:  1,
		ChildNumber: 2,
		PRNumber:    3,
		Success:     true,
		Timestamp:   time.Date(2026, 4, 15, 9, 0, 0, 0, time.UTC),
	}

	source := &stubSource{name: "forgejo", signals: []*PipelineSignal{signal}}
	handler := &stubHandler{name: "dispatch", matchSignal: signal, result: result}
	poller := NewPoller(PollerConfig{
		Sources:      []JobSource{source},
		Handlers:     []JobHandler{handler},
		Journal:      journal,
		PollInterval: time.Second,
	})

	if err := poller.RunOnce(context.Background()); err != nil {
		t.Fatalf("run once: %v", err)
	}
	if poller.Cycle() != 1 {
		t.Fatalf("expected one completed cycle, got %d", poller.Cycle())
	}
	if len(handler.executed) != 1 {
		t.Fatalf("expected handler to execute once, got %d", len(handler.executed))
	}
	if len(source.reports) != 1 {
		t.Fatalf("expected source to report once, got %d", len(source.reports))
	}

	path := filepath.Join(journal.baseDir, "2026", "04", "15.jsonl")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected journal entry: %v", err)
	}
}

func TestPollerDryRunSkipsExecution(t *testing.T) {
	poller := NewPoller(PollerConfig{DryRun: true, PollInterval: time.Second})
	if !poller.DryRun() {
		t.Fatalf("expected dry run")
	}

	source := &stubSource{name: "forgejo", signals: []*PipelineSignal{{PRNumber: 1}}}
	handler := &stubHandler{name: "match", matchSignal: &PipelineSignal{PRNumber: 1}, result: &ActionResult{}}
	poller.AddSource(source)
	poller.AddHandler(handler)

	if err := poller.RunOnce(context.Background()); err != nil {
		t.Fatalf("run once: %v", err)
	}
	if len(handler.executed) != 0 {
		t.Fatalf("expected dry run to skip execution")
	}
	if len(source.reports) != 0 {
		t.Fatalf("expected dry run to skip reporting")
	}
}
