// SPDX-License-Identifier: EUPL-1.2

package handlers

import (
	"context"
	`encoding/json`
	`fmt`
	`os/exec`
	`strings`
	"time"

	forgejo "codeberg.org/forgejo/go-sdk/forgejo"
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

func NewCompletionHandler(client *coreforge.Client) *CompletionHandler {
	return &CompletionHandler{forge: client}
}
func NewDismissReviewsHandler(f *coreforge.Client) *DismissReviewsHandler {
	return &DismissReviewsHandler{forge: f}
}
func NewDispatchHandler(client *coreforge.Client, forgeURL, token string, spinner *agentci.Spinner) *DispatchHandler {
	return &DispatchHandler{forge: client, forgeURL: forgeURL, token: token, spinner: spinner}
}
func NewEnableAutoMergeHandler(f *coreforge.Client) *EnableAutoMergeHandler {
	return &EnableAutoMergeHandler{forge: f}
}
func NewPublishDraftHandler(f *coreforge.Client) *PublishDraftHandler {
	return &PublishDraftHandler{forge: f}
}
func NewSendFixCommandHandler(f *coreforge.Client) *SendFixCommandHandler {
	return &SendFixCommandHandler{forge: f}
}
func NewTickParentHandler(f *coreforge.Client) *TickParentHandler {
	return &TickParentHandler{forge: f}
}

func (h *CompletionHandler) Name() string      { return "completion" }
func (h *DismissReviewsHandler) Name() string  { return "dismiss-reviews" }
func (h *DispatchHandler) Name() string        { return "dispatch" }
func (h *EnableAutoMergeHandler) Name() string { return "enable-auto-merge" }
func (h *PublishDraftHandler) Name() string    { return "publish-draft" }
func (h *SendFixCommandHandler) Name() string  { return "send-fix-command" }
func (h *TickParentHandler) Name() string      { return "tick-parent" }

func (h *CompletionHandler) Match(signal *jobrunner.PipelineSignal) bool {
	return signal != nil && signal.Type == "agent_completion"
}
func (h *DismissReviewsHandler) Match(signal *jobrunner.PipelineSignal) bool {
	return signal != nil && strings.EqualFold(signal.PRState, "OPEN") && signal.HasUnresolvedThreads()
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
	return signal != nil && strings.EqualFold(signal.PRState, "OPEN") && !signal.IsDraft && strings.EqualFold(signal.CheckStatus, "SUCCESS") && strings.EqualFold(signal.Mergeable, "MERGEABLE") && !signal.HasUnresolvedThreads()
}
func (h *PublishDraftHandler) Match(signal *jobrunner.PipelineSignal) bool {
	return signal != nil && strings.EqualFold(signal.PRState, "OPEN") && signal.IsDraft && strings.EqualFold(signal.CheckStatus, "SUCCESS")
}
func (h *SendFixCommandHandler) Match(signal *jobrunner.PipelineSignal) bool {
	return signal != nil && strings.EqualFold(signal.PRState, "OPEN") && (strings.EqualFold(signal.Mergeable, "CONFLICTING") || (signal.HasUnresolvedThreads() && !strings.EqualFold(signal.CheckStatus, "SUCCESS")))
}
func (h *TickParentHandler) Match(signal *jobrunner.PipelineSignal) bool {
	return signal != nil && strings.EqualFold(signal.PRState, "MERGED")
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

func (h *CompletionHandler) Execute(ctx context.Context, signal *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error)  /* v090-result-boundary */ {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return result(h.Name(), signal, false, err.Error()), err
		}
	}
	if h.forge == nil || signal == nil {
		err := fmt.Errorf("handlers.CompletionHandler.Execute: forge client and signal are required")
		return result(h.Name(), signal, false, err.Error()), err
	}
	body := completionComment(signal)
	if err := h.forge.CreateIssueComment(signal.RepoOwner, signal.RepoName, int64(signal.ChildNumber), body); err != nil {
		return result(h.Name(), signal, false, err.Error()), err
	}
	return result(h.Name(), signal, true, "completion noted"), nil
}
func (h *DismissReviewsHandler) Execute(ctx context.Context, signal *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error)  /* v090-result-boundary */ {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return result(h.Name(), signal, false, err.Error()), err
		}
	}
	if h.forge == nil || signal == nil {
		err := fmt.Errorf("handlers.DismissReviewsHandler.Execute: forge client and signal are required")
		return result(h.Name(), signal, false, err.Error()), err
	}
	reviews, err := h.forge.ListPRReviews(signal.RepoOwner, signal.RepoName, int64(signal.PRNumber))
	if err != nil {
		return result(h.Name(), signal, false, err.Error()), err
	}
	dismissed := 0
	for _, review := range reviews {
		if review == nil {
			continue
		}
		if !review.Stale || !strings.EqualFold(string(review.State), "REQUEST_CHANGES") {
			continue
		}
		if err := h.forge.DismissReview(signal.RepoOwner, signal.RepoName, int64(signal.PRNumber), review.ID, "stale request changes review"); err != nil {
			return result(h.Name(), signal, false, err.Error()), err
		}
		dismissed++
	}
	return result(h.Name(), signal, true, fmt.Sprintf("dismissed %d reviews", dismissed)), nil
}
func (h *DispatchHandler) Execute(ctx context.Context, signal *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error)  /* v090-result-boundary */ {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return result(h.Name(), signal, false, err.Error()), err
		}
	}
	if signal == nil {
		err := fmt.Errorf("handlers.DispatchHandler.Execute: signal is required")
		return result(h.Name(), nil, false, err.Error()), err
	}
	agentName, agent, ok := h.resolveAgent(signal.Assignee)
	if !ok {
		err := fmt.Errorf("handlers.DispatchHandler.Execute: unknown agent %q", signal.Assignee)
		return result(h.Name(), signal, false, err.Error()), err
	}
	if strings.TrimSpace(agent.Host) == "" {
		err := fmt.Errorf("handlers.DispatchHandler.Execute: agent %q has no host", agentName)
		return result(h.Name(), signal, false, err.Error()), err
	}
	if strings.TrimSpace(agent.QueueDir) == "" {
		err := fmt.Errorf("handlers.DispatchHandler.Execute: agent %q has no queue dir", agentName)
		return result(h.Name(), signal, false, err.Error()), err
	}
	payload, err := buildDispatchTicket(h.forgeURL, agentName, agent, signal)
	if err != nil {
		return result(h.Name(), signal, false, err.Error()), err
	}
	cmd, err := buildDispatchSSHCommand(ctx, agent, h.token, payload)
	if err != nil {
		return result(h.Name(), signal, false, err.Error()), err
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return result(h.Name(), signal, false, msg), fmt.Errorf("handlers.DispatchHandler.Execute: %s", msg)
	}
	return result(h.Name(), signal, true, fmt.Sprintf("dispatched to %s", agentName)), nil
}
func (h *EnableAutoMergeHandler) Execute(ctx context.Context, signal *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error)  /* v090-result-boundary */ {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return result(h.Name(), signal, false, err.Error()), err
		}
	}
	if h.forge == nil || signal == nil {
		err := fmt.Errorf("handlers.EnableAutoMergeHandler.Execute: forge client and signal are required")
		return result(h.Name(), signal, false, err.Error()), err
	}
	if err := h.forge.MergePullRequest(signal.RepoOwner, signal.RepoName, int64(signal.PRNumber), "squash"); err != nil {
		return result(h.Name(), signal, false, err.Error()), err
	}
	return result(h.Name(), signal, true, "pull request merged"), nil
}
func (h *PublishDraftHandler) Execute(ctx context.Context, signal *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error)  /* v090-result-boundary */ {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return result(h.Name(), signal, false, err.Error()), err
		}
	}
	if h.forge == nil || signal == nil {
		err := fmt.Errorf("handlers.PublishDraftHandler.Execute: forge client and signal are required")
		return result(h.Name(), signal, false, err.Error()), err
	}
	if err := h.forge.SetPRDraft(signal.RepoOwner, signal.RepoName, int64(signal.PRNumber), false); err != nil {
		return result(h.Name(), signal, false, err.Error()), err
	}
	return result(h.Name(), signal, true, "pull request published"), nil
}
func (h *SendFixCommandHandler) Execute(ctx context.Context, signal *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error)  /* v090-result-boundary */ {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return result(h.Name(), signal, false, err.Error()), err
		}
	}
	if h.forge == nil || signal == nil {
		err := fmt.Errorf("handlers.SendFixCommandHandler.Execute: forge client and signal are required")
		return result(h.Name(), signal, false, err.Error()), err
	}
	body := fixCommandComment(signal)
	if err := h.forge.CreateIssueComment(signal.RepoOwner, signal.RepoName, int64(signal.PRNumber), body); err != nil {
		return result(h.Name(), signal, false, err.Error()), err
	}
	return result(h.Name(), signal, true, "fix command posted"), nil
}
func (h *TickParentHandler) Execute(ctx context.Context, signal *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error)  /* v090-result-boundary */ {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return result(h.Name(), signal, false, err.Error()), err
		}
	}
	if h.forge == nil || signal == nil {
		err := fmt.Errorf("handlers.TickParentHandler.Execute: forge client and signal are required")
		return result(h.Name(), signal, false, err.Error()), err
	}
	epicBody, err := h.forge.GetIssueBody(signal.RepoOwner, signal.RepoName, int64(signal.EpicNumber))
	if err != nil {
		return result(h.Name(), signal, false, err.Error()), err
	}
	updatedBody, changed := tickCheckbox(epicBody, signal.ChildNumber)
	if !changed {
		err := fmt.Errorf("handlers.TickParentHandler.Execute: child %d not found in epic %d", signal.ChildNumber, signal.EpicNumber)
		return result(h.Name(), signal, false, err.Error()), err
	}
	body := updatedBody
	if _, err := h.forge.EditIssue(signal.RepoOwner, signal.RepoName, int64(signal.EpicNumber), forgejo.EditIssueOption{Body: &body}); err != nil {
		return result(h.Name(), signal, false, err.Error()), err
	}
	if err := h.forge.CloseIssue(signal.RepoOwner, signal.RepoName, int64(signal.ChildNumber)); err != nil {
		return result(h.Name(), signal, false, err.Error()), err
	}
	return result(h.Name(), signal, true, "parent epic updated"), nil
}

func (h *DispatchHandler) resolveAgent(name string) (string, agentci.AgentConfig, bool) {
	if h == nil || h.spinner == nil {
		return "", agentci.AgentConfig{}, false
	}
	if resolved, cfg, ok := h.spinner.FindByForgejoUser(name); ok {
		return resolved, cfg, true
	}
	if resolved, cfg, ok := h.spinner.FindByForgejoUser(strings.TrimSpace(name)); ok {
		return resolved, cfg, true
	}
	return "", agentci.AgentConfig{}, false
}

func buildDispatchTicket(forgeURL, agentName string, agent agentci.AgentConfig, signal *jobrunner.PipelineSignal) ([]byte, error)  /* v090-result-boundary */ {
	ticket := DispatchTicket{
		ID:           fmt.Sprintf("%s-%d-%d", signal.RepoName, signal.EpicNumber, signal.ChildNumber),
		RepoOwner:    signal.RepoOwner,
		RepoName:     signal.RepoName,
		IssueNumber:  signal.ChildNumber,
		IssueTitle:   signal.IssueTitle,
		IssueBody:    signal.IssueBody,
		TargetBranch: "dev",
		EpicNumber:   signal.EpicNumber,
		ForgeURL:     forgeURL,
		ForgeUser:    agent.ForgejoUser,
		Model:        agent.Model,
		Runner:       agent.Runner,
		VerifyModel:  agent.VerifyModel,
		DualRun:      agent.DualRun,
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
	}
	if ticket.ForgeUser == "" {
		ticket.ForgeUser = agentName
	}
	if ticket.TargetBranch == "" {
		ticket.TargetBranch = "dev"
	}
	return json.MarshalIndent(ticket, "", "  ")
}

func buildDispatchSSHCommand(ctx context.Context, agent agentci.AgentConfig, token string, payload []byte) (*exec.Cmd, error)  /* v090-result-boundary */ {
	queueDir, err := agentci.ValidateRemoteDir(agent.QueueDir)
	if err != nil {
		return nil, err
	}
	ticketName := fmt.Sprintf("%s-%d.json", sanitizeName(agent.ForgejoUser), time.Now().UTC().UnixNano())
	if agent.ForgejoUser == "" {
		ticketName = fmt.Sprintf("ticket-%d.json", time.Now().UTC().UnixNano())
	}
	ticketPath, err := agentci.JoinRemotePath(queueDir, ticketName)
	if err != nil {
		return nil, err
	}
	envPath, err := agentci.JoinRemotePath(queueDir, ".env")
	if err != nil {
		return nil, err
	}

	var shell strings.Builder
	shell.WriteString("mkdir -p ")
	shell.WriteString(agentci.EscapeShellArg(queueDir))
	shell.WriteString(" && cat > ")
	shell.WriteString(agentci.EscapeShellArg(ticketPath))
	shell.WriteString(" <<'EOF'\n")
	shell.Write(payload)
	shell.WriteString("\nEOF\n")
	shell.WriteString("cat > ")
	shell.WriteString(agentci.EscapeShellArg(envPath))
	shell.WriteString(" <<'EOF'\n")
	shell.WriteString("FORGE_TOKEN=")
	shell.WriteString(token)
	shell.WriteString("\nEOF\n")
	shell.WriteString("chmod 0600 ")
	shell.WriteString(agentci.EscapeShellArg(envPath))

	return agentci.SecureSSHCommandContext(ctx, agent.Host, shell.String()), nil
}

func completionComment(signal *jobrunner.PipelineSignal) string {
	return fmt.Sprintf("Agent completion recorded for PR #%d.\n\n%s", signal.PRNumber, signal.Message)
}

func fixCommandComment(signal *jobrunner.PipelineSignal) string {
	switch {
	case strings.EqualFold(signal.Mergeable, "CONFLICTING"):
		return "Please resolve the merge conflicts and push an updated branch."
	case signal.HasUnresolvedThreads():
		return "Please address the unresolved review threads and push an updated branch."
	default:
		return "Please make the requested fixes and push an updated branch."
	}
}

func tickCheckbox(body string, childNumber int) (string, bool) {
	lines := strings.Split(body, "\n")
	target := fmt.Sprintf("#%d", childNumber)
	changed := false
	for i, line := range lines {
		if changed {
			break
		}
		if !strings.Contains(line, target) {
			continue
		}
		if strings.Contains(line, "[ ]") {
			lines[i] = strings.Replace(line, "[ ]", "[x]", 1)
			changed = true
			break
		}
		if strings.Contains(line, "[X]") {
			changed = true
			break
		}
	}
	return strings.Join(lines, "\n"), changed
}

func sanitizeName(input string) string {
	if input == "" {
		return "ticket"
	}
	return strings.NewReplacer("/", "-", " ", "-", ":", "-", "@", "-").Replace(input)
}

// DispatchTicket is the JSON payload written to the agent's queue.
// The ForgeToken is transferred separately via a .env file with 0600 permissions.
type DispatchTicket struct {
	ID           string `json:"id"`
	RepoOwner    string `json:"repo_owner"`
	RepoName     string `json:"repo_name"`
	IssueNumber  int    `json:"issue_number"`
	IssueTitle   string `json:"issue_title"`
	IssueBody    string `json:"issue_body"`
	TargetBranch string `json:"target_branch"`
	EpicNumber   int    `json:"epic_number"`
	ForgeURL     string `json:"forge_url"`
	ForgeUser    string `json:"forgejo_user"`
	Model        string `json:"model,omitempty"`
	Runner       string `json:"runner,omitempty"`
	VerifyModel  string `json:"verify_model,omitempty"`
	DualRun      bool   `json:"dual_run"`
	CreatedAt    string `json:"created_at"`
}
