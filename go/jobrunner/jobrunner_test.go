// SPDX-License-Identifier: EUPL-1.2

package jobrunner

import (
	"context"
	`encoding/json`
	`errors`
	`os`
	`path/filepath`
	"time"

	core "dappco.re/go"
)

const (
	sonarJobrunnerTestGoScm = "go-scm"
)

type ax7Source struct {
	name      string
	signals   []*PipelineSignal
	pollErr   error
	reportErr error
	polled    int
	reports   []*ActionResult
	onPoll    func()
}

func (s *ax7Source) Name() string {
	return s.name
}

func (s *ax7Source) Poll(context.Context) ([]*PipelineSignal, error) {
	s.polled++
	if s.onPoll != nil {
		s.onPoll()
	}
	return s.signals, s.pollErr
}

func (s *ax7Source) Report(_ context.Context, result *ActionResult) error {
	s.reports = append(s.reports, result)
	return s.reportErr
}

type ax7Handler struct {
	name     string
	matches  bool
	result   *ActionResult
	err      error
	executed []*PipelineSignal
}

func (h *ax7Handler) Name() string {
	return h.name
}

func (h *ax7Handler) Match(signal *PipelineSignal) bool {
	return signal != nil && h.matches
}

func (h *ax7Handler) Execute(_ context.Context, signal *PipelineSignal) (*ActionResult, error) {
	h.executed = append(h.executed, signal)
	return h.result, h.err
}

func TestJobrunner_ActionResult_MarshalJSON_Good(t *core.T) {
	result := ActionResult{Action: "dispatch", Success: true, Duration: 1500 * time.Millisecond}
	raw, err := result.MarshalJSON()
	core.AssertNoError(t, err)
	core.AssertContains(t, string(raw), `"duration_ms":1500`)
}

func TestJobrunner_ActionResult_MarshalJSON_Bad(t *core.T) {
	result := ActionResult{Action: "", Duration: -time.Millisecond}
	raw, err := result.MarshalJSON()
	core.AssertNoError(t, err)
	core.AssertContains(t, string(raw), `"duration_ms":-1`)
}

func TestJobrunner_ActionResult_MarshalJSON_Ugly(t *core.T) {
	result := ActionResult{Action: "zero"}
	raw, err := result.MarshalJSON()
	core.AssertNoError(t, err)
	core.AssertContains(t, string(raw), `"success":false`)
}

func TestJobrunner_ActionResult_UnmarshalJSON_Good(t *core.T) {
	var result ActionResult
	err := result.UnmarshalJSON([]byte(`{"action":"dispatch","duration_ms":1500,"success":true}`))
	core.AssertNoError(t, err)
	core.AssertEqual(t, 1500*time.Millisecond, result.Duration)
}

func TestJobrunner_ActionResult_UnmarshalJSON_Bad(t *core.T) {
	var result ActionResult
	err := result.UnmarshalJSON([]byte(`{"duration_ms":`))
	core.AssertError(t, err)
	core.AssertEqual(t, time.Duration(0), result.Duration)
}

func TestJobrunner_ActionResult_UnmarshalJSON_Ugly(t *core.T) {
	var result ActionResult
	err := result.UnmarshalJSON([]byte(`{}`))
	core.AssertNoError(t, err)
	core.AssertEqual(t, "", result.Action)
}

func TestJobrunner_PipelineSignal_HasUnresolvedThreads_Good(t *core.T) {
	signal := &PipelineSignal{ThreadsTotal: 3, ThreadsResolved: 2}
	got := signal.HasUnresolvedThreads()
	core.AssertTrue(t, got)
}

func TestJobrunner_PipelineSignal_HasUnresolvedThreads_Bad(t *core.T) {
	signal := &PipelineSignal{ThreadsTotal: 3, ThreadsResolved: 3}
	got := signal.HasUnresolvedThreads()
	core.AssertFalse(t, got)
}

func TestJobrunner_PipelineSignal_HasUnresolvedThreads_Ugly(t *core.T) {
	var signal *PipelineSignal
	got := signal.HasUnresolvedThreads()
	core.AssertFalse(t, got)
}

func TestJobrunner_PipelineSignal_RepoFullName_Good(t *core.T) {
	signal := &PipelineSignal{RepoOwner: "core", RepoName: sonarJobrunnerTestGoScm}
	got := signal.RepoFullName()
	core.AssertEqual(t, "core/go-scm", got)
}

func TestJobrunner_PipelineSignal_RepoFullName_Bad(t *core.T) {
	signal := &PipelineSignal{RepoOwner: "core"}
	got := signal.RepoFullName()
	core.AssertEqual(t, "core", got)
}

func TestJobrunner_PipelineSignal_RepoFullName_Ugly(t *core.T) {
	var signal *PipelineSignal
	got := signal.RepoFullName()
	core.AssertEqual(t, "", got)
}

func TestJobrunner_NewJournal_Good(t *core.T) {
	journal, err := NewJournal(t.TempDir())
	core.AssertNoError(t, err)
	core.AssertNotNil(t, journal)
}

func TestJobrunner_NewJournal_Bad(t *core.T) {
	journal, err := NewJournal("")
	core.AssertError(t, err)
	core.AssertNil(t, journal)
}

func TestJobrunner_NewJournal_Ugly(t *core.T) {
	t.Setenv("DIR_CWD", t.TempDir())
	journal, err := NewJournal("relative-journal")
	core.AssertNoError(t, err)
	core.AssertContains(t, journal.baseDir, "relative-journal")
}

func TestJobrunner_Journal_Append_Good(t *core.T) {
	journal, err := NewJournal(t.TempDir())
	core.RequireNoError(t, err)
	result := &ActionResult{Action: "dispatch", RepoOwner: "core", RepoName: sonarJobrunnerTestGoScm, Timestamp: time.Date(2026, 4, 15, 8, 0, 0, 0, time.UTC), Success: true}
	err = journal.Append(&PipelineSignal{RepoOwner: "core", RepoName: sonarJobrunnerTestGoScm}, result)
	core.AssertNoError(t, err)
	_, statErr := os.Stat(filepath.Join(journal.baseDir, "2026", "04", "15.jsonl"))
	core.AssertNoError(t, statErr)
}

func TestJobrunner_Journal_Append_Bad(t *core.T) {
	var journal *Journal
	err := journal.Append(&PipelineSignal{}, &ActionResult{Action: "dispatch"})
	core.AssertError(t, err)
	core.AssertContains(t, err.Error(), "journal is required")
}

func TestJobrunner_Journal_Append_Ugly(t *core.T) {
	journal, err := NewJournal(t.TempDir())
	core.RequireNoError(t, err)
	err = journal.Append(nil, nil)
	core.AssertError(t, err)
	core.AssertContains(t, err.Error(), "result is required")
}

func TestJobrunner_NewPoller_Good(t *core.T) {
	source := &ax7Source{name: "source"}
	poller := NewPoller(PollerConfig{Sources: []JobSource{source}, PollInterval: time.Second, DryRun: true})
	core.AssertTrue(t, poller.DryRun())
	core.AssertEqual(t, time.Second, poller.pollInterval())
}

func TestJobrunner_NewPoller_Bad(t *core.T) {
	poller := NewPoller(PollerConfig{PollInterval: 0})
	interval := poller.pollInterval()
	core.AssertEqual(t, time.Minute, interval)
}

func TestJobrunner_NewPoller_Ugly(t *core.T) {
	poller := NewPoller(PollerConfig{Sources: []JobSource{nil}, Handlers: []JobHandler{nil}})
	err := poller.RunOnce(context.Background())
	core.AssertNoError(t, err)
	core.AssertEqual(t, 1, poller.Cycle())
}

func TestJobrunner_Poller_AddHandler_Good(t *core.T) {
	poller := NewPoller(PollerConfig{})
	poller.AddHandler(&ax7Handler{name: "handler"})
	core.AssertLen(t, poller.handlers, 1)
}

func TestJobrunner_Poller_AddHandler_Bad(t *core.T) {
	poller := NewPoller(PollerConfig{})
	poller.AddHandler(nil)
	core.AssertEmpty(t, poller.handlers)
}

func TestJobrunner_Poller_AddHandler_Ugly(t *core.T) {
	var poller *Poller
	core.AssertNotPanics(t, func() { poller.AddHandler(&ax7Handler{name: "handler"}) })
	core.AssertNil(t, poller)
}

func TestJobrunner_Poller_AddSource_Good(t *core.T) {
	poller := NewPoller(PollerConfig{})
	poller.AddSource(&ax7Source{name: "source"})
	core.AssertLen(t, poller.sources, 1)
}

func TestJobrunner_Poller_AddSource_Bad(t *core.T) {
	poller := NewPoller(PollerConfig{})
	poller.AddSource(nil)
	core.AssertEmpty(t, poller.sources)
}

func TestJobrunner_Poller_AddSource_Ugly(t *core.T) {
	var poller *Poller
	core.AssertNotPanics(t, func() { poller.AddSource(&ax7Source{name: "source"}) })
	core.AssertNil(t, poller)
}

func TestJobrunner_Poller_Cycle_Good(t *core.T) {
	poller := NewPoller(PollerConfig{})
	err := poller.RunOnce(context.Background())
	core.AssertNoError(t, err)
	core.AssertEqual(t, 1, poller.Cycle())
}

func TestJobrunner_Poller_Cycle_Bad(t *core.T) {
	poller := NewPoller(PollerConfig{})
	got := poller.Cycle()
	core.AssertEqual(t, 0, got)
}

func TestJobrunner_Poller_Cycle_Ugly(t *core.T) {
	var poller *Poller
	got := poller.Cycle()
	core.AssertEqual(t, 0, got)
}

func TestJobrunner_Poller_DryRun_Good(t *core.T) {
	poller := NewPoller(PollerConfig{DryRun: true})
	got := poller.DryRun()
	core.AssertTrue(t, got)
}

func TestJobrunner_Poller_DryRun_Bad(t *core.T) {
	poller := NewPoller(PollerConfig{})
	got := poller.DryRun()
	core.AssertFalse(t, got)
}

func TestJobrunner_Poller_DryRun_Ugly(t *core.T) {
	var poller *Poller
	got := poller.DryRun()
	core.AssertFalse(t, got)
}

func TestJobrunner_Poller_SetDryRun_Good(t *core.T) {
	poller := NewPoller(PollerConfig{})
	poller.SetDryRun(true)
	core.AssertTrue(t, poller.DryRun())
}

func TestJobrunner_Poller_SetDryRun_Bad(t *core.T) {
	poller := NewPoller(PollerConfig{DryRun: true})
	poller.SetDryRun(false)
	core.AssertFalse(t, poller.DryRun())
}

func TestJobrunner_Poller_SetDryRun_Ugly(t *core.T) {
	var poller *Poller
	core.AssertNotPanics(t, func() { poller.SetDryRun(true) })
	core.AssertNil(t, poller)
}

func TestJobrunner_Poller_RunOnce_Good(t *core.T) {
	signal := &PipelineSignal{PRNumber: 7}
	result := &ActionResult{Action: "handled", Success: true, Timestamp: time.Now().UTC()}
	source := &ax7Source{name: "source", signals: []*PipelineSignal{signal}}
	handler := &ax7Handler{name: "handler", matches: true, result: result}
	poller := NewPoller(PollerConfig{Sources: []JobSource{source}, Handlers: []JobHandler{handler}})
	err := poller.RunOnce(context.Background())
	core.AssertNoError(t, err)
	core.AssertLen(t, handler.executed, 1)
}

func TestJobrunner_Poller_RunOnce_Bad(t *core.T) {
	sourceErr := errors.New("poll failed")
	source := &ax7Source{name: "source", pollErr: sourceErr}
	poller := NewPoller(PollerConfig{Sources: []JobSource{source}})
	err := poller.RunOnce(context.Background())
	core.AssertError(t, err)
	core.AssertContains(t, err.Error(), "poll source")
}

func TestJobrunner_Poller_RunOnce_Ugly(t *core.T) {
	var poller *Poller
	core.AssertPanics(t, func() { _ = poller.RunOnce(context.Background()) })
	core.AssertNil(t, poller)
}

func TestJobrunner_Poller_Run_Good(t *core.T) {
	ctx, cancel := context.WithCancel(context.Background())
	source := &ax7Source{name: "source", onPoll: cancel}
	poller := NewPoller(PollerConfig{Sources: []JobSource{source}, PollInterval: time.Hour})
	err := poller.Run(ctx)
	core.AssertError(t, err)
	core.AssertEqual(t, 1, poller.Cycle())
}

func TestJobrunner_Poller_Run_Bad(t *core.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	poller := NewPoller(PollerConfig{})
	err := poller.Run(ctx)
	core.AssertError(t, err)
	core.AssertEqual(t, 0, poller.Cycle())
}

func TestJobrunner_Poller_Run_Ugly(t *core.T) {
	var poller *Poller
	core.AssertPanics(t, func() { _ = poller.Run(context.Background()) })
	core.AssertNil(t, poller)
}

func TestJobrunner_ActionResult_JSONRoundTrip_Good(t *core.T) {
	original := ActionResult{Action: "dispatch", Duration: 2 * time.Second, Success: true}
	raw, err := original.MarshalJSON()
	core.RequireNoError(t, err)
	var decoded map[string]any
	core.RequireNoError(t, json.Unmarshal(raw, &decoded))
	core.AssertEqual(t, float64(2000), decoded["duration_ms"])
}
