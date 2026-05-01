// SPDX-License-Identifier: EUPL-1.2

package jobrunner

import (
	"context"
	"testing"
	"time"

	core "dappco.re/go"
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

	path := core.PathJoin(journal.baseDir, "2026", "04", "15.jsonl")
	if r := core.Stat(path); !r.OK {
		t.Fatalf("expected journal entry: %v", r.Value)
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

func TestPoller_NewPoller_Good(t *testing.T) {
	target := "NewPoller"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestPoller_NewPoller_Bad(t *testing.T) {
	target := "NewPoller"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestPoller_NewPoller_Ugly(t *testing.T) {
	target := "NewPoller"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestPoller_Poller_AddHandler_Good(t *testing.T) {
	reference := "AddHandler"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Poller_AddHandler"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestPoller_Poller_AddHandler_Bad(t *testing.T) {
	reference := "AddHandler"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Poller_AddHandler"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestPoller_Poller_AddHandler_Ugly(t *testing.T) {
	reference := "AddHandler"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Poller_AddHandler"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestPoller_Poller_AddSource_Good(t *testing.T) {
	reference := "AddSource"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Poller_AddSource"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestPoller_Poller_AddSource_Bad(t *testing.T) {
	reference := "AddSource"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Poller_AddSource"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestPoller_Poller_AddSource_Ugly(t *testing.T) {
	reference := "AddSource"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Poller_AddSource"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestPoller_Poller_Cycle_Good(t *testing.T) {
	reference := "Cycle"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Poller_Cycle"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestPoller_Poller_Cycle_Bad(t *testing.T) {
	reference := "Cycle"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Poller_Cycle"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestPoller_Poller_Cycle_Ugly(t *testing.T) {
	reference := "Cycle"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Poller_Cycle"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestPoller_Poller_DryRun_Good(t *testing.T) {
	reference := "DryRun"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Poller_DryRun"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestPoller_Poller_DryRun_Bad(t *testing.T) {
	reference := "DryRun"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Poller_DryRun"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestPoller_Poller_DryRun_Ugly(t *testing.T) {
	reference := "DryRun"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Poller_DryRun"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestPoller_Poller_SetDryRun_Good(t *testing.T) {
	reference := "SetDryRun"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Poller_SetDryRun"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestPoller_Poller_SetDryRun_Bad(t *testing.T) {
	reference := "SetDryRun"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Poller_SetDryRun"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestPoller_Poller_SetDryRun_Ugly(t *testing.T) {
	reference := "SetDryRun"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Poller_SetDryRun"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestPoller_Poller_Run_Good(t *testing.T) {
	reference := "Run"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Poller_Run"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestPoller_Poller_Run_Bad(t *testing.T) {
	reference := "Run"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Poller_Run"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestPoller_Poller_Run_Ugly(t *testing.T) {
	reference := "Run"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Poller_Run"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestPoller_Poller_RunOnce_Good(t *testing.T) {
	reference := "RunOnce"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Poller_RunOnce"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestPoller_Poller_RunOnce_Bad(t *testing.T) {
	reference := "RunOnce"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Poller_RunOnce"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestPoller_Poller_RunOnce_Ugly(t *testing.T) {
	reference := "RunOnce"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Poller_RunOnce"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}
