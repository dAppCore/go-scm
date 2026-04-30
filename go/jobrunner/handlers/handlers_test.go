// SPDX-License-Identifier: EUPL-1.2

package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	core "dappco.re/go"
	"dappco.re/go/scm/agentci"
	coreforge "dappco.re/go/scm/forge"
	"dappco.re/go/scm/jobrunner"
)

const (
	sonarHandlersTestCodexBot        = "codex-bot"
	sonarHandlersTestDismissReviews  = "dismiss-reviews"
	sonarHandlersTestEnableAutoMerge = "enable-auto-merge"
	sonarHandlersTestHttpForgeTest   = "http://forge.test"
	sonarHandlersTestPublishDraft    = "publish-draft"
	sonarHandlersTestSendFixCommand  = "send-fix-command"
	sonarHandlersTestTickParent      = "tick-parent"
)

func testHandlersForgeClient(t *core.T) *coreforge.Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/version" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"version":"1.22.0"}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"blocked"}`))
	}))
	t.Cleanup(server.Close)
	client, err := coreforge.New(server.URL, "token")
	core.RequireNoError(t, err)
	return client
}

func testHandlersSignal() *jobrunner.PipelineSignal {
	return &jobrunner.PipelineSignal{
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
		Assignee:        sonarHandlersTestCodexBot,
		IssueTitle:      "Implement task",
		IssueBody:       "body",
	}
}

func testHandlersCanceled() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func testHandlersFakeSSH(t *core.T) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "ssh")
	core.RequireNoError(t, os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o700))
	t.Setenv("PATH", dir)
}

func testHandlersSpinner() *agentci.Spinner {
	return agentci.NewSpinner(agentci.ClothoConfig{}, map[string]agentci.AgentConfig{
		"codex": {ForgejoUser: sonarHandlersTestCodexBot, Host: "worker.example.test", QueueDir: "~/queue"},
	})
}

func TestHandlers_NewCompletionHandler_Good(t *core.T) {
	client := testHandlersForgeClient(t)
	handler := NewCompletionHandler(client)
	core.AssertEqual(t, client, handler.forge)
}

func TestHandlers_NewCompletionHandler_Bad(t *core.T) {
	handler := NewCompletionHandler(nil)
	core.AssertNil(
		t, handler.forge,
	)
}

func TestHandlers_NewCompletionHandler_Ugly(t *core.T) {
	handler := NewCompletionHandler(nil)
	got := handler.Name()
	core.AssertEqual(t, "completion", got)
}

func TestHandlers_NewDismissReviewsHandler_Good(t *core.T) {
	client := testHandlersForgeClient(t)
	handler := NewDismissReviewsHandler(client)
	core.AssertEqual(t, client, handler.forge)
}

func TestHandlers_NewDismissReviewsHandler_Bad(t *core.T) {
	handler := NewDismissReviewsHandler(nil)
	core.AssertNil(
		t, handler.forge,
	)
}

func TestHandlers_NewDismissReviewsHandler_Ugly(t *core.T) {
	handler := NewDismissReviewsHandler(nil)
	got := handler.Name()
	core.AssertEqual(t, sonarHandlersTestDismissReviews, got)
	core.AssertTrue(t, handler.Match(&jobrunner.PipelineSignal{PRState: "OPEN", ThreadsTotal: 1}))
}

func TestHandlers_NewDispatchHandler_Good(t *core.T) {
	spinner := testHandlersSpinner()
	handler := NewDispatchHandler(nil, sonarHandlersTestHttpForgeTest, "token", spinner)
	core.AssertEqual(t, sonarHandlersTestHttpForgeTest, handler.forgeURL)
	core.AssertEqual(t, spinner, handler.spinner)
}

func TestHandlers_NewDispatchHandler_Bad(t *core.T) {
	handler := NewDispatchHandler(nil, "", "", nil)
	core.AssertEqual(t, "", handler.forgeURL)
	core.AssertNil(t, handler.spinner)
}

func TestHandlers_NewDispatchHandler_Ugly(t *core.T) {
	handler := NewDispatchHandler(nil, sonarHandlersTestHttpForgeTest, "", testHandlersSpinner())
	got := handler.Name()
	core.AssertEqual(t, "dispatch", got)
}

func TestHandlers_NewEnableAutoMergeHandler_Good(t *core.T) {
	client := testHandlersForgeClient(t)
	handler := NewEnableAutoMergeHandler(client)
	core.AssertEqual(t, client, handler.forge)
}

func TestHandlers_NewEnableAutoMergeHandler_Bad(t *core.T) {
	handler := NewEnableAutoMergeHandler(nil)
	core.AssertNil(
		t, handler.forge,
	)
}

func TestHandlers_NewEnableAutoMergeHandler_Ugly(t *core.T) {
	handler := NewEnableAutoMergeHandler(nil)
	got := handler.Name()
	core.AssertEqual(t, sonarHandlersTestEnableAutoMerge, got)
	core.AssertTrue(t, handler.Match(&jobrunner.PipelineSignal{PRState: "OPEN", CheckStatus: "SUCCESS", Mergeable: "MERGEABLE"}))
}

func TestHandlers_NewPublishDraftHandler_Good(t *core.T) {
	client := testHandlersForgeClient(t)
	handler := NewPublishDraftHandler(client)
	core.AssertEqual(t, client, handler.forge)
}

func TestHandlers_NewPublishDraftHandler_Bad(t *core.T) {
	handler := NewPublishDraftHandler(nil)
	core.AssertNil(
		t, handler.forge,
	)
}

func TestHandlers_NewPublishDraftHandler_Ugly(t *core.T) {
	handler := NewPublishDraftHandler(nil)
	got := handler.Name()
	core.AssertEqual(t, sonarHandlersTestPublishDraft, got)
	core.AssertTrue(t, handler.Match(&jobrunner.PipelineSignal{PRState: "OPEN", IsDraft: true, CheckStatus: "SUCCESS"}))
}

func TestHandlers_NewSendFixCommandHandler_Good(t *core.T) {
	client := testHandlersForgeClient(t)
	handler := NewSendFixCommandHandler(client)
	core.AssertEqual(t, client, handler.forge)
}

func TestHandlers_NewSendFixCommandHandler_Bad(t *core.T) {
	handler := NewSendFixCommandHandler(nil)
	core.AssertNil(
		t, handler.forge,
	)
}

func TestHandlers_NewSendFixCommandHandler_Ugly(t *core.T) {
	handler := NewSendFixCommandHandler(nil)
	got := handler.Name()
	core.AssertEqual(t, sonarHandlersTestSendFixCommand, got)
	core.AssertTrue(t, handler.Match(&jobrunner.PipelineSignal{PRState: "OPEN", Mergeable: "CONFLICTING"}))
}

func TestHandlers_NewTickParentHandler_Good(t *core.T) {
	client := testHandlersForgeClient(t)
	handler := NewTickParentHandler(client)
	core.AssertEqual(t, client, handler.forge)
}

func TestHandlers_NewTickParentHandler_Bad(t *core.T) {
	handler := NewTickParentHandler(nil)
	core.AssertNil(
		t, handler.forge,
	)
}

func TestHandlers_NewTickParentHandler_Ugly(t *core.T) {
	handler := NewTickParentHandler(nil)
	got := handler.Name()
	core.AssertEqual(t, sonarHandlersTestTickParent, got)
	core.AssertTrue(t, handler.Match(&jobrunner.PipelineSignal{PRState: "MERGED"}))
}

func TestHandlers_CompletionHandler_Name_Good(t *core.T) {
	handler := NewCompletionHandler(nil)
	got := handler.Name()
	core.AssertEqual(t, "completion", got)
}

func TestHandlers_CompletionHandler_Name_Bad(t *core.T) {
	handler := &CompletionHandler{}
	got := handler.Name()
	core.AssertEqual(t, "completion", got)
}

func TestHandlers_CompletionHandler_Name_Ugly(t *core.T) {
	var handler *CompletionHandler
	got := handler.Name()
	core.AssertEqual(t, "completion", got)
}

func TestHandlers_DismissReviewsHandler_Name_Good(t *core.T) {
	handler := NewDismissReviewsHandler(nil)
	got := handler.Name()
	core.AssertEqual(t, sonarHandlersTestDismissReviews, got)
	core.AssertTrue(t, handler.Match(&jobrunner.PipelineSignal{PRState: "OPEN", ThreadsTotal: 2, ThreadsResolved: 1}))
}

func TestHandlers_DismissReviewsHandler_Name_Bad(t *core.T) {
	handler := &DismissReviewsHandler{}
	got := handler.Name()
	core.AssertEqual(t, sonarHandlersTestDismissReviews, got)
}

func TestHandlers_DismissReviewsHandler_Name_Ugly(t *core.T) {
	var handler *DismissReviewsHandler
	got := handler.Name()
	core.AssertEqual(t, sonarHandlersTestDismissReviews, got)
}

func TestHandlers_DispatchHandler_Name_Good(t *core.T) {
	handler := NewDispatchHandler(nil, "", "", nil)
	got := handler.Name()
	core.AssertEqual(t, "dispatch", got)
}

func TestHandlers_DispatchHandler_Name_Bad(t *core.T) {
	handler := &DispatchHandler{}
	got := handler.Name()
	core.AssertEqual(t, "dispatch", got)
}

func TestHandlers_DispatchHandler_Name_Ugly(t *core.T) {
	var handler *DispatchHandler
	got := handler.Name()
	core.AssertEqual(t, "dispatch", got)
}

func TestHandlers_EnableAutoMergeHandler_Name_Good(t *core.T) {
	handler := NewEnableAutoMergeHandler(nil)
	got := handler.Name()
	core.AssertEqual(t, sonarHandlersTestEnableAutoMerge, got)
	core.AssertTrue(t, handler.Match(&jobrunner.PipelineSignal{PRState: "OPEN", CheckStatus: "SUCCESS", Mergeable: "MERGEABLE"}))
}

func TestHandlers_EnableAutoMergeHandler_Name_Bad(t *core.T) {
	handler := &EnableAutoMergeHandler{}
	got := handler.Name()
	core.AssertEqual(t, sonarHandlersTestEnableAutoMerge, got)
}

func TestHandlers_EnableAutoMergeHandler_Name_Ugly(t *core.T) {
	var handler *EnableAutoMergeHandler
	got := handler.Name()
	core.AssertEqual(t, sonarHandlersTestEnableAutoMerge, got)
}

func TestHandlers_PublishDraftHandler_Name_Good(t *core.T) {
	handler := NewPublishDraftHandler(nil)
	got := handler.Name()
	core.AssertEqual(t, sonarHandlersTestPublishDraft, got)
	core.AssertTrue(t, handler.Match(&jobrunner.PipelineSignal{PRState: "OPEN", IsDraft: true, CheckStatus: "SUCCESS"}))
}

func TestHandlers_PublishDraftHandler_Name_Bad(t *core.T) {
	handler := &PublishDraftHandler{}
	got := handler.Name()
	core.AssertEqual(t, sonarHandlersTestPublishDraft, got)
}

func TestHandlers_PublishDraftHandler_Name_Ugly(t *core.T) {
	var handler *PublishDraftHandler
	got := handler.Name()
	core.AssertEqual(t, sonarHandlersTestPublishDraft, got)
}

func TestHandlers_SendFixCommandHandler_Name_Good(t *core.T) {
	handler := NewSendFixCommandHandler(nil)
	got := handler.Name()
	core.AssertEqual(t, sonarHandlersTestSendFixCommand, got)
	core.AssertTrue(t, handler.Match(&jobrunner.PipelineSignal{PRState: "OPEN", Mergeable: "CONFLICTING"}))
}

func TestHandlers_SendFixCommandHandler_Name_Bad(t *core.T) {
	handler := &SendFixCommandHandler{}
	got := handler.Name()
	core.AssertEqual(t, sonarHandlersTestSendFixCommand, got)
}

func TestHandlers_SendFixCommandHandler_Name_Ugly(t *core.T) {
	var handler *SendFixCommandHandler
	got := handler.Name()
	core.AssertEqual(t, sonarHandlersTestSendFixCommand, got)
}

func TestHandlers_TickParentHandler_Name_Good(t *core.T) {
	handler := NewTickParentHandler(nil)
	got := handler.Name()
	core.AssertEqual(t, sonarHandlersTestTickParent, got)
	core.AssertTrue(t, handler.Match(&jobrunner.PipelineSignal{PRState: "MERGED"}))
}

func TestHandlers_TickParentHandler_Name_Bad(t *core.T) {
	handler := &TickParentHandler{}
	got := handler.Name()
	core.AssertEqual(t, sonarHandlersTestTickParent, got)
}

func TestHandlers_TickParentHandler_Name_Ugly(t *core.T) {
	var handler *TickParentHandler
	got := handler.Name()
	core.AssertEqual(t, sonarHandlersTestTickParent, got)
}

func TestHandlers_CompletionHandler_Match_Good(t *core.T) {
	handler := NewCompletionHandler(nil)
	got := handler.Match(&jobrunner.PipelineSignal{Type: "agent_completion"})
	core.AssertTrue(t, got)
}

func TestHandlers_CompletionHandler_Match_Bad(t *core.T) {
	handler := NewCompletionHandler(nil)
	got := handler.Match(&jobrunner.PipelineSignal{Type: "other"})
	core.AssertFalse(t, got)
}

func TestHandlers_CompletionHandler_Match_Ugly(t *core.T) {
	handler := NewCompletionHandler(nil)
	got := handler.Match(nil)
	core.AssertFalse(t, got)
}

func TestHandlers_DismissReviewsHandler_Match_Good(t *core.T) {
	handler := NewDismissReviewsHandler(nil)
	got := handler.Match(&jobrunner.PipelineSignal{PRState: "open", ThreadsTotal: 2, ThreadsResolved: 1})
	core.AssertTrue(t, got)
}

func TestHandlers_DismissReviewsHandler_Match_Bad(t *core.T) {
	handler := NewDismissReviewsHandler(nil)
	got := handler.Match(&jobrunner.PipelineSignal{PRState: "closed", ThreadsTotal: 2, ThreadsResolved: 1})
	core.AssertFalse(t, got)
}

func TestHandlers_DismissReviewsHandler_Match_Ugly(t *core.T) {
	handler := NewDismissReviewsHandler(nil)
	got := handler.Match(nil)
	core.AssertFalse(t, got)
}

func TestHandlers_DispatchHandler_Match_Good(t *core.T) {
	handler := NewDispatchHandler(nil, "", "", nil)
	got := handler.Match(&jobrunner.PipelineSignal{NeedsCoding: true, Assignee: sonarHandlersTestCodexBot})
	core.AssertTrue(t, got)
}

func TestHandlers_DispatchHandler_Match_Bad(t *core.T) {
	handler := NewDispatchHandler(nil, "", "", nil)
	got := handler.Match(&jobrunner.PipelineSignal{NeedsCoding: false, Assignee: sonarHandlersTestCodexBot})
	core.AssertFalse(t, got)
}

func TestHandlers_DispatchHandler_Match_Ugly(t *core.T) {
	handler := NewDispatchHandler(nil, "", "", testHandlersSpinner())
	got := handler.Match(&jobrunner.PipelineSignal{NeedsCoding: true, Assignee: "missing"})
	core.AssertFalse(t, got)
}

func TestHandlers_EnableAutoMergeHandler_Match_Good(t *core.T) {
	handler := NewEnableAutoMergeHandler(nil)
	got := handler.Match(&jobrunner.PipelineSignal{PRState: "open", CheckStatus: "success", Mergeable: "mergeable"})
	core.AssertTrue(t, got)
}

func TestHandlers_EnableAutoMergeHandler_Match_Bad(t *core.T) {
	handler := NewEnableAutoMergeHandler(nil)
	got := handler.Match(&jobrunner.PipelineSignal{PRState: "open", IsDraft: true, CheckStatus: "success", Mergeable: "mergeable"})
	core.AssertFalse(t, got)
}

func TestHandlers_EnableAutoMergeHandler_Match_Ugly(t *core.T) {
	handler := NewEnableAutoMergeHandler(nil)
	got := handler.Match(nil)
	core.AssertFalse(t, got)
}

func TestHandlers_PublishDraftHandler_Match_Good(t *core.T) {
	handler := NewPublishDraftHandler(nil)
	got := handler.Match(&jobrunner.PipelineSignal{PRState: "open", IsDraft: true, CheckStatus: "success"})
	core.AssertTrue(t, got)
}

func TestHandlers_PublishDraftHandler_Match_Bad(t *core.T) {
	handler := NewPublishDraftHandler(nil)
	got := handler.Match(&jobrunner.PipelineSignal{PRState: "open", IsDraft: false, CheckStatus: "success"})
	core.AssertFalse(t, got)
}

func TestHandlers_PublishDraftHandler_Match_Ugly(t *core.T) {
	handler := NewPublishDraftHandler(nil)
	got := handler.Match(nil)
	core.AssertFalse(t, got)
}

func TestHandlers_SendFixCommandHandler_Match_Good(t *core.T) {
	handler := NewSendFixCommandHandler(nil)
	got := handler.Match(&jobrunner.PipelineSignal{PRState: "open", Mergeable: "conflicting"})
	core.AssertTrue(t, got)
}

func TestHandlers_SendFixCommandHandler_Match_Bad(t *core.T) {
	handler := NewSendFixCommandHandler(nil)
	got := handler.Match(&jobrunner.PipelineSignal{PRState: "open", Mergeable: "mergeable", CheckStatus: "success"})
	core.AssertFalse(t, got)
}

func TestHandlers_SendFixCommandHandler_Match_Ugly(t *core.T) {
	handler := NewSendFixCommandHandler(nil)
	got := handler.Match(nil)
	core.AssertFalse(t, got)
}

func TestHandlers_TickParentHandler_Match_Good(t *core.T) {
	handler := NewTickParentHandler(nil)
	got := handler.Match(&jobrunner.PipelineSignal{PRState: "merged"})
	core.AssertTrue(t, got)
}

func TestHandlers_TickParentHandler_Match_Bad(t *core.T) {
	handler := NewTickParentHandler(nil)
	got := handler.Match(&jobrunner.PipelineSignal{PRState: "open"})
	core.AssertFalse(t, got)
}

func TestHandlers_TickParentHandler_Match_Ugly(t *core.T) {
	handler := NewTickParentHandler(nil)
	got := handler.Match(nil)
	core.AssertFalse(t, got)
}

func TestHandlers_CompletionHandler_Execute_Good(t *core.T) {
	handler := NewCompletionHandler(testHandlersForgeClient(t))
	result, err := handler.Execute(context.Background(), testHandlersSignal())
	core.AssertError(t, err)
	core.AssertEqual(t, "completion", result.Action)
}

func TestHandlers_CompletionHandler_Execute_Bad(t *core.T) {
	handler := NewCompletionHandler(nil)
	result, err := handler.Execute(testHandlersCanceled(), testHandlersSignal())
	core.AssertError(t, err)
	core.AssertFalse(t, result.Success)
}

func TestHandlers_CompletionHandler_Execute_Ugly(t *core.T) {
	handler := NewCompletionHandler(nil)
	result, err := handler.Execute(context.Background(), nil)
	core.AssertError(t, err)
	core.AssertEqual(t, "completion", result.Action)
}

func TestHandlers_DismissReviewsHandler_Execute_Good(t *core.T) {
	handler := NewDismissReviewsHandler(testHandlersForgeClient(t))
	result, err := handler.Execute(context.Background(), testHandlersSignal())
	core.AssertError(t, err)
	core.AssertEqual(t, sonarHandlersTestDismissReviews, result.Action)
}

func TestHandlers_DismissReviewsHandler_Execute_Bad(t *core.T) {
	handler := NewDismissReviewsHandler(nil)
	result, err := handler.Execute(testHandlersCanceled(), testHandlersSignal())
	core.AssertError(t, err)
	core.AssertFalse(t, result.Success)
}

func TestHandlers_DismissReviewsHandler_Execute_Ugly(t *core.T) {
	handler := NewDismissReviewsHandler(nil)
	result, err := handler.Execute(context.Background(), nil)
	core.AssertError(t, err)
	core.AssertEqual(t, sonarHandlersTestDismissReviews, result.Action)
}

func TestHandlers_DispatchHandler_Execute_Good(t *core.T) {
	testHandlersFakeSSH(t)
	handler := NewDispatchHandler(nil, sonarHandlersTestHttpForgeTest, "token", testHandlersSpinner())
	result, err := handler.Execute(context.Background(), testHandlersSignal())
	core.AssertNoError(t, err)
	core.AssertTrue(t, result.Success)
}

func TestHandlers_DispatchHandler_Execute_Bad(t *core.T) {
	handler := NewDispatchHandler(nil, "", "", nil)
	result, err := handler.Execute(context.Background(), testHandlersSignal())
	core.AssertError(t, err)
	core.AssertFalse(t, result.Success)
}

func TestHandlers_DispatchHandler_Execute_Ugly(t *core.T) {
	handler := NewDispatchHandler(nil, "", "", nil)
	result, err := handler.Execute(context.Background(), nil)
	core.AssertError(t, err)
	core.AssertEqual(t, "dispatch", result.Action)
}

func TestHandlers_EnableAutoMergeHandler_Execute_Good(t *core.T) {
	handler := NewEnableAutoMergeHandler(testHandlersForgeClient(t))
	result, err := handler.Execute(context.Background(), testHandlersSignal())
	core.AssertNoError(t, err)
	core.AssertEqual(t, sonarHandlersTestEnableAutoMerge, result.Action)
}

func TestHandlers_EnableAutoMergeHandler_Execute_Bad(t *core.T) {
	handler := NewEnableAutoMergeHandler(nil)
	result, err := handler.Execute(testHandlersCanceled(), testHandlersSignal())
	core.AssertError(t, err)
	core.AssertFalse(t, result.Success)
}

func TestHandlers_EnableAutoMergeHandler_Execute_Ugly(t *core.T) {
	handler := NewEnableAutoMergeHandler(nil)
	result, err := handler.Execute(context.Background(), nil)
	core.AssertError(t, err)
	core.AssertEqual(t, sonarHandlersTestEnableAutoMerge, result.Action)
}

func TestHandlers_PublishDraftHandler_Execute_Good(t *core.T) {
	handler := NewPublishDraftHandler(testHandlersForgeClient(t))
	signal := testHandlersSignal()
	signal.IsDraft = true
	result, err := handler.Execute(context.Background(), signal)
	core.AssertError(t, err)
	core.AssertEqual(t, sonarHandlersTestPublishDraft, result.Action)
}

func TestHandlers_PublishDraftHandler_Execute_Bad(t *core.T) {
	handler := NewPublishDraftHandler(nil)
	result, err := handler.Execute(testHandlersCanceled(), testHandlersSignal())
	core.AssertError(t, err)
	core.AssertFalse(t, result.Success)
}

func TestHandlers_PublishDraftHandler_Execute_Ugly(t *core.T) {
	handler := NewPublishDraftHandler(nil)
	result, err := handler.Execute(context.Background(), nil)
	core.AssertError(t, err)
	core.AssertEqual(t, sonarHandlersTestPublishDraft, result.Action)
}

func TestHandlers_SendFixCommandHandler_Execute_Good(t *core.T) {
	handler := NewSendFixCommandHandler(testHandlersForgeClient(t))
	signal := testHandlersSignal()
	signal.Mergeable = "CONFLICTING"
	result, err := handler.Execute(context.Background(), signal)
	core.AssertError(t, err)
	core.AssertEqual(t, sonarHandlersTestSendFixCommand, result.Action)
}

func TestHandlers_SendFixCommandHandler_Execute_Bad(t *core.T) {
	handler := NewSendFixCommandHandler(nil)
	result, err := handler.Execute(testHandlersCanceled(), testHandlersSignal())
	core.AssertError(t, err)
	core.AssertFalse(t, result.Success)
}

func TestHandlers_SendFixCommandHandler_Execute_Ugly(t *core.T) {
	handler := NewSendFixCommandHandler(nil)
	result, err := handler.Execute(context.Background(), nil)
	core.AssertError(t, err)
	core.AssertEqual(t, sonarHandlersTestSendFixCommand, result.Action)
}

func TestHandlers_TickParentHandler_Execute_Good(t *core.T) {
	handler := NewTickParentHandler(testHandlersForgeClient(t))
	result, err := handler.Execute(context.Background(), testHandlersSignal())
	core.AssertError(t, err)
	core.AssertEqual(t, sonarHandlersTestTickParent, result.Action)
}

func TestHandlers_TickParentHandler_Execute_Bad(t *core.T) {
	handler := NewTickParentHandler(nil)
	result, err := handler.Execute(testHandlersCanceled(), testHandlersSignal())
	core.AssertError(t, err)
	core.AssertFalse(t, result.Success)
}

func TestHandlers_TickParentHandler_Execute_Ugly(t *core.T) {
	handler := NewTickParentHandler(nil)
	result, err := handler.Execute(context.Background(), nil)
	core.AssertError(t, err)
	core.AssertEqual(t, sonarHandlersTestTickParent, result.Action)
}
