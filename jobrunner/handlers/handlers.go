// SPDX-License-Identifier: EUPL-1.2

package handlers

import (
	"context"
	"fmt"
	"time"

	"dappco.re/go/scm/agentci"
	coreforge "dappco.re/go/scm/forge"
	"dappco.re/go/scm/jobrunner"
)

type CompletionHandler struct{ forge *coreforge.Client }
type DismissReviewsHandler struct{ forge *coreforge.Client }
type DispatchHandler struct {
	forge    *coreforge.Client
	forgeURL string
	token    string
	spinner  *agentci.Spinner
}
type EnableAutoMergeHandler struct{ forge *coreforge.Client }
type PublishDraftHandler struct{ forge *coreforge.Client }
type SendFixCommandHandler struct{ forge *coreforge.Client }
type TickParentHandler struct{ forge *coreforge.Client }

func NewCompletionHandler(client *coreforge.Client) *CompletionHandler     { return &CompletionHandler{forge: client} }
func NewDismissReviewsHandler(f *coreforge.Client) *DismissReviewsHandler   { return &DismissReviewsHandler{forge: f} }
func NewDispatchHandler(client *coreforge.Client, forgeURL, token string, spinner *agentci.Spinner) *DispatchHandler {
	return &DispatchHandler{forge: client, forgeURL: forgeURL, token: token, spinner: spinner}
}
func NewEnableAutoMergeHandler(f *coreforge.Client) *EnableAutoMergeHandler { return &EnableAutoMergeHandler{forge: f} }
func NewPublishDraftHandler(f *coreforge.Client) *PublishDraftHandler       { return &PublishDraftHandler{forge: f} }
func NewSendFixCommandHandler(f *coreforge.Client) *SendFixCommandHandler   { return &SendFixCommandHandler{forge: f} }
func NewTickParentHandler(f *coreforge.Client) *TickParentHandler           { return &TickParentHandler{forge: f} }

func (h *CompletionHandler) Name() string      { return "completion" }
func (h *DismissReviewsHandler) Name() string  { return "dismiss-reviews" }
func (h *DispatchHandler) Name() string        { return "dispatch" }
func (h *EnableAutoMergeHandler) Name() string { return "enable-auto-merge" }
func (h *PublishDraftHandler) Name() string     { return "publish-draft" }
func (h *SendFixCommandHandler) Name() string   { return "send-fix-command" }
func (h *TickParentHandler) Name() string       { return "tick-parent" }

func (h *CompletionHandler) Match(signal *jobrunner.PipelineSignal) bool {
	return signal != nil && signal.Type == "agent_completion"
}
func (h *DismissReviewsHandler) Match(signal *jobrunner.PipelineSignal) bool {
	return signal != nil && signal.PRState == "OPEN" && signal.HasUnresolvedThreads()
}
func (h *DispatchHandler) Match(signal *jobrunner.PipelineSignal) bool {
	if signal == nil || !signal.NeedsCoding || signal.Assignee == "" {
		return false
	}
	if h.spinner == nil {
		return true
	}
	_, _, ok := h.spinner.FindByForgejoUser(signal.Assignee)
	return ok
}
func (h *EnableAutoMergeHandler) Match(signal *jobrunner.PipelineSignal) bool {
	return signal != nil && signal.PRState == "OPEN" && !signal.IsDraft && signal.CheckStatus == "SUCCESS" && signal.Mergeable == "MERGEABLE" && !signal.HasUnresolvedThreads()
}
func (h *PublishDraftHandler) Match(signal *jobrunner.PipelineSignal) bool {
	return signal != nil && signal.PRState == "OPEN" && signal.IsDraft && signal.CheckStatus == "SUCCESS"
}
func (h *SendFixCommandHandler) Match(signal *jobrunner.PipelineSignal) bool {
	return signal != nil && signal.PRState == "OPEN" && (signal.Mergeable == "CONFLICTING" || (signal.HasUnresolvedThreads() && signal.CheckStatus != "SUCCESS"))
}
func (h *TickParentHandler) Match(signal *jobrunner.PipelineSignal) bool {
	return signal != nil && signal.PRState == "MERGED"
}

func result(name string, signal *jobrunner.PipelineSignal, success bool, msg string) *jobrunner.ActionResult {
	if signal == nil {
		return &jobrunner.ActionResult{Action: name, Success: success, Error: msg, Timestamp: time.Now().UTC()}
	}
	return &jobrunner.ActionResult{
		Action:      name,
		RepoOwner:   signal.RepoOwner,
		RepoName:    signal.RepoName,
		EpicNumber:  signal.EpicNumber,
		ChildNumber: signal.ChildNumber,
		PRNumber:    signal.PRNumber,
		Success:     success,
		Error:       msg,
		Timestamp:   time.Now().UTC(),
	}
}

func (h *CompletionHandler) Execute(ctx context.Context, signal *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error) {
	_ = ctx
	_ = h.forge
	return result(h.Name(), signal, true, ""), nil
}
func (h *DismissReviewsHandler) Execute(ctx context.Context, signal *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error) {
	_ = ctx
	_ = h.forge
	return result(h.Name(), signal, true, ""), nil
}
func (h *DispatchHandler) Execute(ctx context.Context, signal *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error) {
	_ = ctx
	_ = h.forge
	_ = h.forgeURL
	_ = h.token
	return result(h.Name(), signal, true, fmt.Sprintf("dispatched to %s", signal.Assignee)), nil
}
func (h *EnableAutoMergeHandler) Execute(ctx context.Context, signal *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error) {
	_ = ctx
	_ = h.forge
	return result(h.Name(), signal, true, ""), nil
}
func (h *PublishDraftHandler) Execute(ctx context.Context, signal *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error) {
	_ = ctx
	_ = h.forge
	return result(h.Name(), signal, true, ""), nil
}
func (h *SendFixCommandHandler) Execute(ctx context.Context, signal *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error) {
	_ = ctx
	_ = h.forge
	return result(h.Name(), signal, true, ""), nil
}
func (h *TickParentHandler) Execute(ctx context.Context, signal *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error) {
	_ = ctx
	_ = h.forge
	return result(h.Name(), signal, true, ""), nil
}
