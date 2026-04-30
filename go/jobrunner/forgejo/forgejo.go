// SPDX-License-Identifier: EUPL-1.2

package forgejo

import (
	"context"
	`fmt`
	"regexp"
	"strconv"
	`strings`
	"time"

	forgejo "codeberg.org/forgejo/go-sdk/forgejo"
	coreforge "dappco.re/go/scm/forge"
	"dappco.re/go/scm/jobrunner"
)

type Config struct {
	Repos []string
}

type ForgejoSource struct {
	repos []string
	forge *coreforge.Client
}

func New(cfg Config, client *coreforge.Client) *ForgejoSource {
	return &ForgejoSource{repos: append([]string(nil), cfg.Repos...), forge: client}
}

func (s *ForgejoSource) Name() string { return "forgejo" }

func (s *ForgejoSource) Poll(ctx context.Context) ([]*jobrunner.PipelineSignal, error)  /* v090-result-boundary */ {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	if s == nil || s.forge == nil {
		return nil, nil
	}

	var signals []*jobrunner.PipelineSignal
	for _, repoRef := range s.repos {
		repoSignals, err := s.pollRepo(ctx, repoRef)
		if err != nil {
			return nil, err
		}
		signals = append(signals, repoSignals...)
	}
	return signals, nil
}

func (s *ForgejoSource) pollRepo(ctx context.Context, repoRef string) ([]*jobrunner.PipelineSignal, error)  /* v090-result-boundary */ {
	owner, repo, err := splitRepoRef(repoRef)
	if err != nil {
		return nil, err
	}
	issues, err := s.forge.ListIssues(owner, repo, coreforge.ListIssuesOpts{State: "open", Limit: 100})
	if err != nil {
		return nil, err
	}
	var signals []*jobrunner.PipelineSignal
	for _, epic := range issues {
		signals = append(signals, s.signalsForEpic(ctx, owner, repo, epic)...)
	}
	return signals, nil
}

func (s *ForgejoSource) signalsForEpic(ctx context.Context, owner, repo string, epic *forgejo.Issue) []*jobrunner.PipelineSignal {
	if epic == nil || strings.TrimSpace(epic.Body) == "" {
		return nil
	}
	childNumbers := parseChildIssueNumbers(epic.Body)
	signals := make([]*jobrunner.PipelineSignal, 0, len(childNumbers))
	for _, childNumber := range childNumbers {
		childSignal, err := s.signalForChild(ctx, owner, repo, epic.Index, childNumber)
		if err != nil {
			continue
		}
		signals = append(signals, childSignal)
	}
	return signals
}

func (s *ForgejoSource) Report(ctx context.Context, result *jobrunner.ActionResult) error  /* v090-result-boundary */ {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	if s == nil || s.forge == nil || result == nil {
		return nil
	}
	body := fmt.Sprintf(
		"Action `%s` finished.\n\n- Success: `%t`\n- Error: `%s`\n- Cycle: `%d`\n- Duration: `%s`",
		result.Action,
		result.Success,
		result.Error,
		result.Cycle,
		result.Duration,
	)
	return s.forge.CreateIssueComment(result.RepoOwner, result.RepoName, int64(result.EpicNumber), body)
}

func (s *ForgejoSource) signalForChild(ctx context.Context, owner, repo string, epicNumber, childNumber int64) (*jobrunner.PipelineSignal, error)  /* v090-result-boundary */ {
	pr, err := s.forge.GetPullRequest(owner, repo, childNumber)
	if err != nil {
		return nil, err
	}
	childIssue, err := s.forge.GetIssue(owner, repo, childNumber)
	if err != nil {
		return nil, err
	}
	prMeta, err := s.forge.GetPRMeta(owner, repo, childNumber)
	if err != nil {
		return nil, err
	}

	reviews, _ := s.forge.ListPRReviews(owner, repo, childNumber)
	requestChanges := 0
	lastReviewAt := time.Time{}
	for _, review := range reviews {
		if review == nil {
			continue
		}
		if strings.EqualFold(string(review.State), "REQUEST_CHANGES") {
			requestChanges++
		}
		if review.Submitted.After(lastReviewAt) {
			lastReviewAt = review.Submitted
		}
	}

	signal := &jobrunner.PipelineSignal{
		EpicNumber:      int(epicNumber),
		ChildNumber:     int(childNumber),
		PRNumber:        int(childNumber),
		RepoOwner:       owner,
		RepoName:        repo,
		PRState:         pullState(pr),
		IsDraft:         false,
		Mergeable:       mergeableState(pr),
		CheckStatus:     s.combinedStatusState(owner, repo, pullCommitSHA(pr)),
		ThreadsTotal:    len(reviews),
		ThreadsResolved: len(reviews) - requestChanges,
		LastCommitSHA:   pullCommitSHA(pr),
		LastCommitAt:    pullCommitTime(pr),
		LastReviewAt:    lastReviewAt,
		NeedsCoding:     false,
		Assignee:        issueAssignee(childIssue),
		IssueTitle:      childIssue.Title,
		IssueBody:       childIssue.Body,
		Type:            "forgejo_epic_child",
	}
	if prMeta != nil && prMeta.CommentCount > signal.ThreadsTotal {
		signal.ThreadsTotal = prMeta.CommentCount
		if signal.ThreadsResolved > signal.ThreadsTotal {
			signal.ThreadsResolved = signal.ThreadsTotal
		}
	}
	if signal.ThreadsResolved < 0 {
		signal.ThreadsResolved = 0
	}
	return signal, nil
}

func splitRepoRef(ref string) (owner, repo string, err error)  /* v090-result-boundary */ {
	parts := strings.Split(ref, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("jobrunner.forgejo: invalid repo reference %q", ref)
	}
	owner = strings.TrimSpace(parts[0])
	repo = strings.TrimSpace(parts[1])
	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("jobrunner.forgejo: invalid repo reference %q", ref)
	}
	return owner, repo, nil
}

var childCheckboxRE = regexp.MustCompile(`(?mi)^\s*[-*]\s*\[\s*\]\s*(?:#|issue\s*)?(\d+)\b`)

func parseChildIssueNumbers(body string) []int64 {
	if strings.TrimSpace(body) == "" {
		return nil
	}
	matches := childCheckboxRE.FindAllStringSubmatch(body, -1)
	if len(matches) == 0 {
		return nil
	}
	seen := map[int64]struct{}{}
	out := make([]int64, 0, len(matches))
	for _, match := range matches {
		if len(match) != 2 {
			continue
		}
		num, err := strconv.ParseInt(match[1], 10, 64)
		if err != nil {
			continue
		}
		if _, ok := seen[num]; ok {
			continue
		}
		seen[num] = struct{}{}
		out = append(out, num)
	}
	return out
}

func pullState(pr *forgejo.PullRequest) string {
	if pr == nil {
		return "UNKNOWN"
	}
	if pr.HasMerged {
		return "MERGED"
	}
	switch strings.ToLower(string(pr.State)) {
	case "open":
		return "OPEN"
	case "closed":
		return "CLOSED"
	default:
		return strings.ToUpper(string(pr.State))
	}
}

func mergeableState(pr *forgejo.PullRequest) string {
	if pr == nil {
		return "UNKNOWN"
	}
	if pr.HasMerged || pr.Mergeable {
		return "MERGEABLE"
	}
	if pr.Head == nil || pr.Base == nil {
		return "UNKNOWN"
	}
	return "CONFLICTING"
}

func (s *ForgejoSource) combinedStatusState(owner, repo, ref string) string {
	if strings.TrimSpace(ref) == "" {
		return "UNKNOWN"
	}
	status, err := s.forge.GetCombinedStatus(owner, repo, ref)
	if err != nil || status == nil {
		return "UNKNOWN"
	}
	switch strings.ToLower(string(status.State)) {
	case "success":
		return "SUCCESS"
	case "failure":
		return "FAILURE"
	case "pending":
		return "PENDING"
	default:
		return "UNKNOWN"
	}
}

func issueAssignee(issue *forgejo.Issue) string {
	if issue == nil || len(issue.Assignees) == 0 || issue.Assignees[0] == nil {
		return ""
	}
	return issue.Assignees[0].UserName
}

func pullCommitSHA(pr *forgejo.PullRequest) string {
	if pr == nil || pr.Head == nil {
		return ""
	}
	return pr.Head.Sha
}

func pullCommitTime(pr *forgejo.PullRequest) time.Time {
	if pr == nil {
		return time.Time{}
	}
	if pr.Updated != nil {
		return *pr.Updated
	}
	if pr.Created != nil {
		return *pr.Created
	}
	return time.Time{}
}
