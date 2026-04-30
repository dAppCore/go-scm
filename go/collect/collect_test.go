// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	"context"
	"errors"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"time"

	core "dappco.re/go"
	coreio "dappco.re/go/io"
)

const (
	sonarCollectTestStateJson = "state.json"
)

type ax7Collector struct {
	name   string
	result *Result
	err    error
}

func (c ax7Collector) Name() string { return c.name }

func (c ax7Collector) Collect(context.Context, *Config) (*Result, error) {
	return c.result, c.err
}

func ax7CollectConfig() *Config {
	cfg := NewConfigWithMedium(coreio.NewMockMedium(), "collect")
	cfg.Limiter.SetDelay("github", 0)
	cfg.Limiter.SetDelay("bitcointalk", 0)
	cfg.Limiter.SetDelay("coingecko", 0)
	cfg.Limiter.SetDelay("iacr", 0)
	cfg.Limiter.SetDelay("arxiv", 0)
	return cfg
}

func ax7FakeGH(t *core.T, body string) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gh")
	script := "#!/bin/sh\nprintf '%s\\n' '" + body + "'\n"
	core.RequireNoError(t, os.WriteFile(path, []byte(script), 0o700))
	t.Setenv("PATH", dir)
}

func TestCollect_NewConfig_Good(t *core.T) {
	cfg := NewConfig("out")
	core.AssertEqual(t, "out", cfg.OutputDir)
	core.AssertNotNil(t, cfg.Output)
	core.AssertNotNil(t, cfg.Dispatcher)
}

func TestCollect_NewConfig_Bad(t *core.T) {
	cfg := NewConfig("")
	core.AssertEqual(t, "collect", cfg.OutputDir)
	core.AssertNotNil(t, cfg.State)
}

func TestCollect_NewConfig_Ugly(t *core.T) {
	cfg := NewConfig("out/../out")
	core.AssertEqual(
		t, "out", cfg.OutputDir,
	)
}

func TestCollect_NewConfigWithMedium_Good(t *core.T) {
	medium := coreio.NewMockMedium()
	cfg := NewConfigWithMedium(medium, "out")
	core.AssertEqual(t, medium, cfg.Output)
	core.AssertEqual(t, "out", cfg.OutputDir)
}

func TestCollect_NewConfigWithMedium_Bad(t *core.T) {
	cfg := NewConfigWithMedium(nil, "")
	core.AssertNotNil(t, cfg.Output)
	core.AssertEqual(t, "collect", cfg.OutputDir)
}

func TestCollect_NewConfigWithMedium_Ugly(t *core.T) {
	cfg := NewConfigWithMedium(coreio.NewMockMedium(), "out/nested")
	core.AssertEqual(t, "out/nested", cfg.OutputDir)
	core.AssertNotNil(t, cfg.State)
}

func TestCollect_MergeResults_Good(t *core.T) {
	got := MergeResults("all", &Result{Items: 1, Files: []string{"a"}}, &Result{Errors: 2, Skipped: 3, Files: []string{"b"}})
	core.AssertEqual(t, "all", got.Source)
	core.AssertEqual(t, 1, got.Items)
	core.AssertEqual(t, 2, got.Errors)
	core.AssertEqual(t, 3, got.Skipped)
	core.AssertEqual(t, []string{"a", "b"}, got.Files)
}

func TestCollect_MergeResults_Bad(t *core.T) {
	got := MergeResults("all", nil)
	core.AssertEqual(t, "all", got.Source)
	core.AssertEqual(t, 0, got.Items)
}

func TestCollect_MergeResults_Ugly(t *core.T) {
	got := MergeResults("", &Result{Items: -1})
	core.AssertEqual(t, "", got.Source)
	core.AssertEqual(t, -1, got.Items)
}

func TestCollect_NewDispatcher_Good(t *core.T) {
	dispatcher := NewDispatcher()
	core.AssertNotNil(t, dispatcher)
	core.AssertNotNil(t, dispatcher.handlers)
}

func TestCollect_NewDispatcher_Bad(t *core.T) {
	dispatcher := NewDispatcher()
	core.AssertEmpty(
		t, dispatcher.handlers,
	)
}

func TestCollect_NewDispatcher_Ugly(t *core.T) {
	dispatcher := NewDispatcher()
	ax7RegisterStartHandler(dispatcher)
	core.AssertLen(t, dispatcher.handlers[EventStart], 1)
}

func ax7RegisterStartHandler(dispatcher *Dispatcher) {
	dispatcher.On(EventStart, func(Event) {
		// Empty handler verifies registration without side effects.
	})
}

func TestCollect_Dispatcher_On_Good(t *core.T) {
	dispatcher := NewDispatcher()
	dispatcher.On(EventStart, func(Event) {
		// Empty handler verifies registration without side effects.
	})
	core.AssertLen(t, dispatcher.handlers[EventStart], 1)
}

func TestCollect_Dispatcher_On_Bad(t *core.T) {
	dispatcher := NewDispatcher()
	dispatcher.On(EventStart, nil)
	core.AssertEmpty(t, dispatcher.handlers[EventStart])
}

func TestCollect_Dispatcher_On_Ugly(t *core.T) {
	var dispatcher *Dispatcher
	core.AssertNotPanics(t, func() {
		dispatcher.On(EventStart, func(Event) {
			// Empty handler verifies nil dispatcher handling.
		})
	})
}

func TestCollect_Dispatcher_Emit_Good(t *core.T) {
	dispatcher := NewDispatcher()
	var got Event
	dispatcher.On(EventStart, func(event Event) { got = event })
	dispatcher.Emit(Event{Type: EventStart, Source: "github", Message: "start"})
	core.AssertEqual(t, "github", got.Source)
	core.AssertFalse(t, got.Time.IsZero())
}

func TestCollect_Dispatcher_Emit_Bad(t *core.T) {
	dispatcher := NewDispatcher()
	dispatcher.Emit(Event{Type: "missing"})
	core.AssertEmpty(t, dispatcher.handlers["missing"])
}

func TestCollect_Dispatcher_Emit_Ugly(t *core.T) {
	var dispatcher *Dispatcher
	core.AssertNotPanics(t, func() {
		dispatcher.Emit(Event{Type: EventStart})
	})
}

func TestCollect_Dispatcher_EmitStart_Good(t *core.T) {
	dispatcher := NewDispatcher()
	var got Event
	dispatcher.On(EventStart, func(event Event) { got = event })
	dispatcher.EmitStart("github", "start")
	core.AssertEqual(t, EventStart, got.Type)
	core.AssertEqual(t, "start", got.Message)
}

func TestCollect_Dispatcher_EmitStart_Bad(t *core.T) {
	dispatcher := NewDispatcher()
	dispatcher.EmitStart("", "")
	core.AssertEmpty(t, dispatcher.handlers)
}

func TestCollect_Dispatcher_EmitStart_Ugly(t *core.T) {
	var dispatcher *Dispatcher
	core.AssertNotPanics(
		t, func() { dispatcher.EmitStart("x", "y") },
	)
}

func TestCollect_Dispatcher_EmitProgress_Good(t *core.T) {
	dispatcher := NewDispatcher()
	var got Event
	dispatcher.On(EventProgress, func(event Event) { got = event })
	dispatcher.EmitProgress("github", "progress", 1)
	core.AssertEqual(t, EventProgress, got.Type)
	core.AssertEqual(t, 1, got.Data)
}

func TestCollect_Dispatcher_EmitProgress_Bad(t *core.T) {
	dispatcher := NewDispatcher()
	dispatcher.EmitProgress("", "", nil)
	core.AssertEmpty(t, dispatcher.handlers)
}

func TestCollect_Dispatcher_EmitProgress_Ugly(t *core.T) {
	var dispatcher *Dispatcher
	core.AssertNotPanics(
		t, func() { dispatcher.EmitProgress("x", "y", nil) },
	)
}

func TestCollect_Dispatcher_EmitItem_Good(t *core.T) {
	dispatcher := NewDispatcher()
	var got Event
	dispatcher.On(EventItem, func(event Event) { got = event })
	dispatcher.EmitItem("github", "item", "data")
	core.AssertEqual(t, EventItem, got.Type)
	core.AssertEqual(t, "data", got.Data)
}

func TestCollect_Dispatcher_EmitItem_Bad(t *core.T) {
	dispatcher := NewDispatcher()
	dispatcher.EmitItem("", "", nil)
	core.AssertEmpty(t, dispatcher.handlers)
}

func TestCollect_Dispatcher_EmitItem_Ugly(t *core.T) {
	var dispatcher *Dispatcher
	core.AssertNotPanics(
		t, func() { dispatcher.EmitItem("x", "y", nil) },
	)
}

func TestCollect_Dispatcher_EmitError_Good(t *core.T) {
	dispatcher := NewDispatcher()
	var got Event
	dispatcher.On(EventError, func(event Event) { got = event })
	dispatcher.EmitError("github", "error", errors.New("boom"))
	core.AssertEqual(t, EventError, got.Type)
	core.AssertNotNil(t, got.Data)
}

func TestCollect_Dispatcher_EmitError_Bad(t *core.T) {
	dispatcher := NewDispatcher()
	dispatcher.EmitError("", "", nil)
	core.AssertEmpty(t, dispatcher.handlers)
}

func TestCollect_Dispatcher_EmitError_Ugly(t *core.T) {
	var dispatcher *Dispatcher
	core.AssertNotPanics(
		t, func() { dispatcher.EmitError("x", "y", nil) },
	)
}

func TestCollect_Dispatcher_EmitComplete_Good(t *core.T) {
	dispatcher := NewDispatcher()
	var got Event
	dispatcher.On(EventComplete, func(event Event) { got = event })
	dispatcher.EmitComplete("github", "done", &Result{Items: 1})
	core.AssertEqual(t, EventComplete, got.Type)
	core.AssertNotNil(t, got.Data)
}

func TestCollect_Dispatcher_EmitComplete_Bad(t *core.T) {
	dispatcher := NewDispatcher()
	dispatcher.EmitComplete("", "", nil)
	core.AssertEmpty(t, dispatcher.handlers)
}

func TestCollect_Dispatcher_EmitComplete_Ugly(t *core.T) {
	var dispatcher *Dispatcher
	core.AssertNotPanics(
		t, func() { dispatcher.EmitComplete("x", "y", nil) },
	)
}

func TestCollect_NewState_Good(t *core.T) {
	state := NewState(coreio.NewMockMedium(), sonarCollectTestStateJson)
	core.AssertEqual(t, sonarCollectTestStateJson, state.path)
	core.AssertNotNil(t, state.entries)
}

func TestCollect_NewState_Bad(t *core.T) {
	state := NewState(nil, "")
	core.AssertNotNil(t, state.medium)
	core.AssertEqual(t, "", state.path)
}

func TestCollect_NewState_Ugly(t *core.T) {
	state := NewState(coreio.NewMockMedium(), "dir/../state.json")
	core.AssertEqual(
		t, sonarCollectTestStateJson, state.path,
	)
}

func TestCollect_State_Get_Good(t *core.T) {
	state := NewState(coreio.NewMockMedium(), "")
	state.Set("github", &StateEntry{Items: 2})
	entry, ok := state.Get("github")
	core.AssertTrue(t, ok)
	core.AssertEqual(t, 2, entry.Items)
}

func TestCollect_State_Get_Bad(t *core.T) {
	state := NewState(coreio.NewMockMedium(), "")
	_, ok := state.Get("missing")
	core.AssertFalse(t, ok)
}

func TestCollect_State_Get_Ugly(t *core.T) {
	var state *State
	_, ok := state.Get("github")
	core.AssertFalse(t, ok)
}

func TestCollect_State_Set_Good(t *core.T) {
	state := NewState(coreio.NewMockMedium(), "")
	state.Set("github", &StateEntry{Items: 2})
	entry, ok := state.Get("github")
	core.AssertTrue(t, ok)
	core.AssertEqual(t, "github", entry.Source)
}

func TestCollect_State_Set_Bad(t *core.T) {
	state := NewState(coreio.NewMockMedium(), "")
	state.Set("github", nil)
	_, ok := state.Get("github")
	core.AssertFalse(t, ok)
}

func TestCollect_State_Set_Ugly(t *core.T) {
	var state *State
	core.AssertNotPanics(
		t, func() { state.Set("github", &StateEntry{}) },
	)
}

func TestCollect_State_Load_Good(t *core.T) {
	medium := coreio.NewMockMedium()
	core.RequireNoError(t, medium.Write(sonarCollectTestStateJson, `{"github":{"source":"github","items":2}}`))
	state := NewState(medium, sonarCollectTestStateJson)
	err := state.Load()
	core.AssertNoError(t, err)
	entry, ok := state.Get("github")
	core.AssertTrue(t, ok)
	core.AssertEqual(t, 2, entry.Items)
}

func TestCollect_State_Load_Bad(t *core.T) {
	var state *State
	err := state.Load()
	core.AssertError(t, err)
}

func TestCollect_State_Load_Ugly(t *core.T) {
	state := NewState(coreio.NewMockMedium(), "missing.json")
	err := state.Load()
	core.AssertNoError(t, err)
	core.AssertEmpty(t, state.entries)
}

func TestCollect_State_Save_Good(t *core.T) {
	medium := coreio.NewMockMedium()
	state := NewState(medium, sonarCollectTestStateJson)
	state.Set("github", &StateEntry{Items: 2})
	err := state.Save()
	core.AssertNoError(t, err)
	raw, readErr := medium.Read(sonarCollectTestStateJson)
	core.RequireNoError(t, readErr)
	core.AssertContains(t, raw, "github")
}

func TestCollect_State_Save_Bad(t *core.T) {
	var state *State
	err := state.Save()
	core.AssertError(t, err)
}

func TestCollect_State_Save_Ugly(t *core.T) {
	state := NewState(coreio.NewMockMedium(), "")
	state.Set("github", &StateEntry{Items: 2})
	err := state.Save()
	core.AssertNoError(t, err)
}

func TestCollect_GitHubCollector_Name_Good(t *core.T) {
	got := (&GitHubCollector{}).Name()
	core.AssertEqual(
		t, "github", got,
	)
}

func TestCollect_GitHubCollector_Name_Bad(t *core.T) {
	var collector *GitHubCollector
	got := collector.Name()
	core.AssertEqual(t, "github", got)
}

func TestCollect_GitHubCollector_Name_Ugly(t *core.T) {
	got := (&GitHubCollector{Org: "core"}).Name()
	core.AssertEqual(
		t, "github", got,
	)
}

func TestCollect_GitHubCollector_Collect_Good(t *core.T) {
	cfg := ax7CollectConfig()
	result, err := (&GitHubCollector{Org: "core", Repo: "go-scm"}).Collect(context.Background(), cfg)
	core.AssertNoError(t, err)
	core.AssertEqual(t, 1, result.Items)
	core.AssertContains(t, result.Files[0], "core/go-scm.md")
}

func TestCollect_GitHubCollector_Collect_Bad(t *core.T) {
	_, err := (&GitHubCollector{}).Collect(context.Background(), nil)
	core.AssertError(
		t, err,
	)
}

func TestCollect_GitHubCollector_Collect_Ugly(t *core.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := (&GitHubCollector{}).Collect(ctx, ax7CollectConfig())
	core.AssertErrorIs(t, err, context.Canceled)
}

func TestCollect_Excavator_Name_Good(t *core.T) {
	got := (&Excavator{}).Name()
	core.AssertEqual(
		t, "excavator", got,
	)
}

func TestCollect_Excavator_Name_Bad(t *core.T) {
	var excavator *Excavator
	got := excavator.Name()
	core.AssertEqual(t, "excavator", got)
}

func TestCollect_Excavator_Name_Ugly(t *core.T) {
	got := (&Excavator{ScanOnly: true}).Name()
	core.AssertEqual(
		t, "excavator", got,
	)
}

func TestCollect_Excavator_Run_Good(t *core.T) {
	cfg := ax7CollectConfig()
	excavator := &Excavator{Collectors: []Collector{ax7Collector{name: "mock", result: &Result{Items: 2}}}}
	result, err := excavator.Run(context.Background(), cfg)
	core.AssertNoError(t, err)
	core.AssertEqual(t, 2, result.Items)
}

func TestCollect_Excavator_Run_Bad(t *core.T) {
	_, err := (&Excavator{}).Run(context.Background(), nil)
	core.AssertError(
		t, err,
	)
}

func TestCollect_Excavator_Run_Ugly(t *core.T) {
	cfg := ax7CollectConfig()
	excavator := &Excavator{ScanOnly: true, Collectors: []Collector{ax7Collector{name: "mock", result: &Result{Items: 2}}}}
	result, err := excavator.Run(context.Background(), cfg)
	core.AssertNoError(t, err)
	core.AssertEqual(t, 0, result.Items)
}

func TestCollect_NewRateLimiter_Good(t *core.T) {
	limiter := NewRateLimiter()
	core.AssertEqual(
		t, 500*time.Millisecond, limiter.GetDelay("github"),
	)
}

func TestCollect_NewRateLimiter_Bad(t *core.T) {
	limiter := NewRateLimiter()
	limiter.delays = nil
	core.AssertEqual(t, 500*time.Millisecond, limiter.GetDelay("missing"))
}

func TestCollect_NewRateLimiter_Ugly(t *core.T) {
	limiter := NewRateLimiter()
	limiter.SetDelay("github", time.Nanosecond)
	core.AssertEqual(t, 500*time.Millisecond, defaultDelays["github"])
}

func TestCollect_RateLimiter_Wait_Good(t *core.T) {
	limiter := NewRateLimiter()
	limiter.SetDelay("unit", 0)
	err := limiter.Wait(context.Background(), "unit")
	core.AssertNoError(t, err)
}

func TestCollect_RateLimiter_Wait_Bad(t *core.T) {
	limiter := NewRateLimiter()
	limiter.SetDelay("unit", time.Hour)
	limiter.last["unit"] = time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := limiter.Wait(ctx, "unit")
	core.AssertErrorIs(t, err, context.Canceled)
}

func TestCollect_RateLimiter_Wait_Ugly(t *core.T) {
	var limiter *RateLimiter
	var ctx context.Context
	err := limiter.Wait(ctx, "unit")
	core.AssertNoError(t, err)
}

func TestCollect_RateLimiter_SetDelay_Good(t *core.T) {
	limiter := NewRateLimiter()
	limiter.SetDelay("unit", time.Second)
	core.AssertEqual(t, time.Second, limiter.GetDelay("unit"))
}

func TestCollect_RateLimiter_SetDelay_Bad(t *core.T) {
	var limiter *RateLimiter
	core.AssertNotPanics(
		t, func() { limiter.SetDelay("unit", time.Second) },
	)
}

func TestCollect_RateLimiter_SetDelay_Ugly(t *core.T) {
	limiter := &RateLimiter{}
	limiter.SetDelay("", -time.Second)
	core.AssertEqual(t, -time.Second, limiter.GetDelay(""))
}

func TestCollect_RateLimiter_GetDelay_Good(t *core.T) {
	limiter := NewRateLimiter()
	core.AssertEqual(
		t, 2*time.Second, limiter.GetDelay("bitcointalk"),
	)
}

func TestCollect_RateLimiter_GetDelay_Bad(t *core.T) {
	limiter := NewRateLimiter()
	core.AssertEqual(
		t, 500*time.Millisecond, limiter.GetDelay("missing"),
	)
}

func TestCollect_RateLimiter_GetDelay_Ugly(t *core.T) {
	var limiter *RateLimiter
	core.AssertEqual(
		t, 500*time.Millisecond, limiter.GetDelay("missing"),
	)
}

func TestCollect_RateLimiter_CheckGitHubRateLimit_Good(t *core.T) {
	ax7FakeGH(t, "10 100")
	limiter := NewRateLimiter()
	used, limit, err := limiter.CheckGitHubRateLimit()
	core.AssertNoError(t, err)
	core.AssertEqual(t, 10, used)
	core.AssertEqual(t, 100, limit)
}

func TestCollect_RateLimiter_CheckGitHubRateLimit_Bad(t *core.T) {
	t.Setenv("PATH", t.TempDir())
	_, _, err := NewRateLimiter().CheckGitHubRateLimit()
	core.AssertError(t, err)
}

func TestCollect_RateLimiter_CheckGitHubRateLimit_Ugly(t *core.T) {
	var limiter *RateLimiter
	used, limit, err := limiter.CheckGitHubRateLimit()
	core.AssertNoError(t, err)
	core.AssertEqual(t, 0, used)
	core.AssertEqual(t, 0, limit)
}

func TestCollect_RateLimiter_CheckGitHubRateLimitCtx_Good(t *core.T) {
	ax7FakeGH(t, "80 100")
	limiter := NewRateLimiter()
	used, limit, err := limiter.CheckGitHubRateLimitCtx(context.Background())
	core.AssertNoError(t, err)
	core.AssertEqual(t, 80, used)
	core.AssertEqual(t, 100, limit)
	core.AssertEqual(t, 5*time.Second, limiter.GetDelay("github"))
}

func TestCollect_RateLimiter_CheckGitHubRateLimitCtx_Bad(t *core.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err := NewRateLimiter().CheckGitHubRateLimitCtx(ctx)
	core.AssertError(t, err)
}

func TestCollect_RateLimiter_CheckGitHubRateLimitCtx_Ugly(t *core.T) {
	var limiter *RateLimiter
	var ctx context.Context
	used, limit, err := limiter.CheckGitHubRateLimitCtx(ctx)
	core.AssertNoError(t, err)
	core.AssertEqual(t, 0, used)
	core.AssertEqual(t, 0, limit)
}

func TestCollect_Processor_Name_Good(t *core.T) {
	got := (&Processor{}).Name()
	core.AssertEqual(
		t, "process", got,
	)
}

func TestCollect_Processor_Name_Bad(t *core.T) {
	var processor *Processor
	got := processor.Name()
	core.AssertEqual(t, "process", got)
}

func TestCollect_Processor_Name_Ugly(t *core.T) {
	got := (&Processor{Source: "raw"}).Name()
	core.AssertEqual(
		t, "process", got,
	)
}

func TestCollect_Processor_Process_Good(t *core.T) {
	cfg := ax7CollectConfig()
	core.RequireNoError(t, cfg.Output.Write("raw/page.html", "<h1>Hello</h1>"))
	result, err := (&Processor{Source: "raw"}).Process(context.Background(), cfg)
	core.AssertNoError(t, err)
	core.AssertEqual(t, 1, result.Items)
	core.AssertContains(t, result.Files[0], "process/page.md")
}

func TestCollect_Processor_Process_Bad(t *core.T) {
	_, err := (&Processor{}).Process(context.Background(), nil)
	core.AssertError(
		t, err,
	)
}

func TestCollect_Processor_Process_Ugly(t *core.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := (&Processor{Source: "raw"}).Process(ctx, ax7CollectConfig())
	core.AssertErrorIs(t, err, context.Canceled)
}

func TestCollect_HTMLToMarkdown_Good(t *core.T) {
	got, err := HTMLToMarkdown(`<h1>Title</h1><p><a href="https://example.test">Link</a></p>`)
	core.AssertNoError(t, err)
	core.AssertContains(t, got, "# Title")
	core.AssertContains(t, got, "[Link](https://example.test)")
}

func TestCollect_HTMLToMarkdown_Bad(t *core.T) {
	got, err := HTMLToMarkdown("")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "", got)
}

func TestCollect_HTMLToMarkdown_Ugly(t *core.T) {
	got, err := HTMLToMarkdown("<div>Unclosed")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "Unclosed", got)
}

func TestCollect_JSONToMarkdown_Good(t *core.T) {
	got, err := JSONToMarkdown(`{"agent":"codex"}`)
	core.AssertNoError(t, err)
	core.AssertContains(t, got, "```json")
	core.AssertContains(t, got, `"agent": "codex"`)
}

func TestCollect_JSONToMarkdown_Bad(t *core.T) {
	_, err := JSONToMarkdown(`{"agent":`)
	core.AssertError(t, err)
}

func TestCollect_JSONToMarkdown_Ugly(t *core.T) {
	got, err := JSONToMarkdown("{\"a\":1}\n{\"b\":2}")
	core.AssertNoError(t, err)
	core.AssertContains(t, got, `"a": 1`)
	core.AssertContains(t, got, `"b": 2`)
}

func TestCollect_SetHTTPClient_Good(t *core.T) {
	old := httpClient
	defer func() { httpClient = old }()
	client := &http.Client{Timeout: time.Second}
	SetHTTPClient(client)
	core.AssertEqual(t, client, httpClient)
}

func TestCollect_SetHTTPClient_Bad(t *core.T) {
	old := httpClient
	SetHTTPClient(nil)
	core.AssertEqual(t, old, httpClient)
}

func TestCollect_SetHTTPClient_Ugly(t *core.T) {
	old := httpClient
	defer func() { httpClient = old }()
	client := &http.Client{}
	SetHTTPClient(client)
	core.AssertEqual(t, client, httpClient)
}

func TestCollect_BitcoinTalkCollector_Name_Good(t *core.T) {
	got := (&BitcoinTalkCollector{}).Name()
	core.AssertEqual(
		t, "bitcointalk", got,
	)
}

func TestCollect_BitcoinTalkCollector_Name_Bad(t *core.T) {
	var collector *BitcoinTalkCollector
	got := collector.Name()
	core.AssertEqual(t, "bitcointalk", got)
}

func TestCollect_BitcoinTalkCollector_Name_Ugly(t *core.T) {
	got := (&BitcoinTalkCollector{TopicID: "1"}).Name()
	core.AssertEqual(
		t, "bitcointalk", got,
	)
}

func TestCollect_BitcoinTalkCollector_Collect_Good(t *core.T) {
	cfg := ax7CollectConfig()
	cfg.DryRun = true
	result, err := (&BitcoinTalkCollector{TopicID: "1"}).Collect(context.Background(), cfg)
	core.AssertNoError(t, err)
	core.AssertEqual(t, "bitcointalk", result.Source)
}

func TestCollect_BitcoinTalkCollector_Collect_Bad(t *core.T) {
	_, err := (&BitcoinTalkCollector{}).Collect(context.Background(), nil)
	core.AssertError(
		t, err,
	)
}

func TestCollect_BitcoinTalkCollector_Collect_Ugly(t *core.T) {
	result, err := (&BitcoinTalkCollector{}).Collect(context.Background(), ax7CollectConfig())
	core.AssertNoError(t, err)
	core.AssertEqual(t, 0, result.Items)
}

func TestCollect_BitcoinTalkCollectorWithFetcher_Name_Good(t *core.T) {
	got := (&BitcoinTalkCollectorWithFetcher{}).Name()
	core.AssertEqual(
		t, "bitcointalk", got,
	)
}

func TestCollect_BitcoinTalkCollectorWithFetcher_Name_Bad(t *core.T) {
	var collector *BitcoinTalkCollectorWithFetcher
	core.AssertPanics(
		t, func() { _ = collector.Name() },
	)
}

func TestCollect_BitcoinTalkCollectorWithFetcher_Name_Ugly(t *core.T) {
	got := (&BitcoinTalkCollectorWithFetcher{BitcoinTalkCollector: BitcoinTalkCollector{TopicID: "1"}}).Name()
	core.AssertEqual(
		t, "bitcointalk", got,
	)
}

func TestCollect_BitcoinTalkCollectorWithFetcher_Collect_Good(t *core.T) {
	cfg := ax7CollectConfig()
	collector := &BitcoinTalkCollectorWithFetcher{
		BitcoinTalkCollector: BitcoinTalkCollector{TopicID: "1", Pages: 1},
		Fetcher: func(context.Context, string) ([]btPost, error) {
			return []btPost{{Number: 1, Author: "satoshi", Content: "hello"}}, nil
		},
	}
	result, err := collector.Collect(context.Background(), cfg)
	core.AssertNoError(t, err)
	core.AssertEqual(t, 1, result.Items)
}

func TestCollect_BitcoinTalkCollectorWithFetcher_Collect_Bad(t *core.T) {
	_, err := (&BitcoinTalkCollectorWithFetcher{}).Collect(context.Background(), nil)
	core.AssertError(
		t, err,
	)
}

func TestCollect_BitcoinTalkCollectorWithFetcher_Collect_Ugly(t *core.T) {
	collector := &BitcoinTalkCollectorWithFetcher{
		BitcoinTalkCollector: BitcoinTalkCollector{TopicID: "1", Pages: 1},
		Fetcher: func(context.Context, string) ([]btPost, error) {
			return nil, errors.New("fetch failed")
		},
	}
	result, err := collector.Collect(context.Background(), ax7CollectConfig())
	core.AssertNoError(t, err)
	core.AssertEqual(t, 1, result.Errors)
}

func TestCollect_ParsePostsFromHTML_Good(t *core.T) {
	posts, err := ParsePostsFromHTML(`<div class="post"><span class="author">satoshi</span><span class="date">2009</span><p>Hello</p></div>`)
	core.AssertNoError(t, err)
	core.AssertLen(t, posts, 1)
	core.AssertEqual(t, "satoshi", posts[0].Author)
}

func TestCollect_ParsePostsFromHTML_Bad(t *core.T) {
	posts, err := ParsePostsFromHTML("")
	core.AssertNoError(t, err)
	core.AssertNil(t, posts)
}

func TestCollect_ParsePostsFromHTML_Ugly(t *core.T) {
	posts, err := ParsePostsFromHTML("<p>plain text</p>")
	core.AssertNoError(t, err)
	core.AssertLen(t, posts, 1)
	core.AssertContains(t, posts[0].Content, "plain text")
}

func TestCollect_FormatPostMarkdown_Good(t *core.T) {
	got := FormatPostMarkdown(1, "satoshi", "2009", "hello")
	core.AssertContains(t, got, "## Post 1")
	core.AssertContains(t, got, "satoshi")
}

func TestCollect_FormatPostMarkdown_Bad(t *core.T) {
	got := FormatPostMarkdown(0, "", "", "")
	core.AssertContains(
		t, got, "## Post 0",
	)
}

func TestCollect_FormatPostMarkdown_Ugly(t *core.T) {
	got := FormatPostMarkdown(-1, " author ", "", " content ")
	core.AssertContains(t, got, "## Post -1")
	core.AssertContains(t, got, "content")
}

func TestCollect_PapersCollector_Name_Good(t *core.T) {
	got := (&PapersCollector{}).Name()
	core.AssertEqual(
		t, "papers", got,
	)
}

func TestCollect_PapersCollector_Name_Bad(t *core.T) {
	var collector *PapersCollector
	got := collector.Name()
	core.AssertEqual(t, "papers", got)
}

func TestCollect_PapersCollector_Name_Ugly(t *core.T) {
	got := (&PapersCollector{Source: PaperSourceArXiv}).Name()
	core.AssertEqual(
		t, "papers", got,
	)
}

func TestCollect_PapersCollector_Collect_Good(t *core.T) {
	cfg := ax7CollectConfig()
	result, err := (&PapersCollector{Source: PaperSourceArXiv, Query: "zk"}).Collect(context.Background(), cfg)
	core.AssertNoError(t, err)
	core.AssertEqual(t, 1, result.Items)
}

func TestCollect_PapersCollector_Collect_Bad(t *core.T) {
	_, err := (&PapersCollector{}).Collect(context.Background(), nil)
	core.AssertError(
		t, err,
	)
}

func TestCollect_PapersCollector_Collect_Ugly(t *core.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := (&PapersCollector{}).Collect(ctx, ax7CollectConfig())
	core.AssertErrorIs(t, err, context.Canceled)
}

func TestCollect_FormatPaperMarkdown_Good(t *core.T) {
	got := FormatPaperMarkdown("Paper", []string{"Alice", "Bob"}, "2026", "https://example.test", "arxiv", "abstract")
	core.AssertContains(t, got, "# Paper")
	core.AssertContains(t, got, "Alice, Bob")
}

func TestCollect_FormatPaperMarkdown_Bad(t *core.T) {
	got := FormatPaperMarkdown("", nil, "", "", "", "")
	core.AssertContains(
		t, got, "# ",
	)
}

func TestCollect_FormatPaperMarkdown_Ugly(t *core.T) {
	got := FormatPaperMarkdown("  Paper  ", nil, "", "", "", "  abstract  ")
	core.AssertContains(t, got, "# Paper")
	core.AssertContains(t, got, "abstract")
}

func TestCollect_MarketCollector_Name_Good(t *core.T) {
	got := (&MarketCollector{}).Name()
	core.AssertEqual(
		t, "market", got,
	)
}

func TestCollect_MarketCollector_Name_Bad(t *core.T) {
	var collector *MarketCollector
	got := collector.Name()
	core.AssertEqual(t, "market", got)
}

func TestCollect_MarketCollector_Name_Ugly(t *core.T) {
	got := (&MarketCollector{CoinID: "bitcoin"}).Name()
	core.AssertEqual(
		t, "market", got,
	)
}

func TestCollect_MarketCollector_Collect_Good(t *core.T) {
	cfg := ax7CollectConfig()
	result, err := (&MarketCollector{CoinID: "bitcoin"}).Collect(context.Background(), cfg)
	core.AssertNoError(t, err)
	core.AssertEqual(t, 1, result.Items)
}

func TestCollect_MarketCollector_Collect_Bad(t *core.T) {
	_, err := (&MarketCollector{}).Collect(context.Background(), nil)
	core.AssertError(
		t, err,
	)
}

func TestCollect_MarketCollector_Collect_Ugly(t *core.T) {
	_, err := (&MarketCollector{Historical: true, FromDate: "bad-date"}).Collect(context.Background(), ax7CollectConfig())
	core.AssertError(
		t, err,
	)
}

func TestCollect_FormatMarketSummary_Good(t *core.T) {
	got := FormatMarketSummary(&coinData{Name: "Bitcoin", Symbol: "BTC", CurrentPrice: 1, MarketCap: 2, Volume: 3, Change24H: 4})
	core.AssertContains(t, got, "# Bitcoin (BTC)")
	core.AssertContains(t, got, "$1")
}

func TestCollect_FormatMarketSummary_Bad(t *core.T) {
	got := FormatMarketSummary(nil)
	core.AssertEqual(
		t, "", got,
	)
}

func TestCollect_FormatMarketSummary_Ugly(t *core.T) {
	got := FormatMarketSummary(&coinData{Name: "Bad", Symbol: "BAD", CurrentPrice: math.NaN()})
	core.AssertContains(
		t, got, "n/a",
	)
}
