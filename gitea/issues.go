// SPDX-License-Identifier: EUPL-1.2

package gitea

import (
	"iter"

	"code.gitea.io/sdk/gitea"

	"dappco.re/go/core/log"
)

// ListIssuesOpts configures issue listing.
type ListIssuesOpts struct {
	State string // "open", "closed", "all"
	Page  int
	Limit int
}

// ListIssues returns issues for the given repository.
// Usage: ListIssues(...)
func (c *Client) ListIssues(owner, repo string, opts ListIssuesOpts) ([]*gitea.Issue, error) {
	state := gitea.StateOpen
	switch opts.State {
	case "closed":
		state = gitea.StateClosed
	case "all":
		state = gitea.StateAll
	}

	limit := opts.Limit
	if limit == 0 {
		limit = 50
	}

	page := opts.Page
	if page == 0 {
		page = 1
	}

	var all []*gitea.Issue

	for {
		issues, resp, err := c.api.ListRepoIssues(owner, repo, gitea.ListIssueOption{
			ListOptions: gitea.ListOptions{Page: page, PageSize: limit},
			State:       state,
			Type:        gitea.IssueTypeIssue,
		})
		if err != nil {
			return nil, log.E("gitea.ListIssues", "failed to list issues", err)
		}

		all = append(all, issues...)
		if len(issues) < limit || len(issues) == 0 {
			break
		}
		if resp != nil && resp.LastPage > 0 && page >= resp.LastPage {
			break
		}
		page++
	}

	return all, nil
}

// ListIssuesIter returns an iterator over issues for the given repository.
// Usage: ListIssuesIter(...)
func (c *Client) ListIssuesIter(owner, repo string, opts ListIssuesOpts) iter.Seq2[*gitea.Issue, error] {
	state := gitea.StateOpen
	switch opts.State {
	case "closed":
		state = gitea.StateClosed
	case "all":
		state = gitea.StateAll
	}

	limit := opts.Limit
	if limit == 0 {
		limit = 50
	}

	page := opts.Page
	if page == 0 {
		page = 1
	}

	return func(yield func(*gitea.Issue, error) bool) {
		for {
			issues, resp, err := c.api.ListRepoIssues(owner, repo, gitea.ListIssueOption{
				ListOptions: gitea.ListOptions{Page: page, PageSize: limit},
				State:       state,
				Type:        gitea.IssueTypeIssue,
			})
			if err != nil {
				yield(nil, log.E("gitea.ListIssues", "failed to list issues", err))
				return
			}
			for _, issue := range issues {
				if !yield(issue, nil) {
					return
				}
			}
			if len(issues) < limit || len(issues) == 0 {
				break
			}
			if resp != nil && resp.LastPage > 0 && page >= resp.LastPage {
				break
			}
			page++
		}
	}
}

// GetIssue returns a single issue by number.
// Usage: GetIssue(...)
func (c *Client) GetIssue(owner, repo string, number int64) (*gitea.Issue, error) {
	issue, _, err := c.api.GetIssue(owner, repo, number)
	if err != nil {
		return nil, log.E("gitea.GetIssue", "failed to get issue", err)
	}

	return issue, nil
}

// CreateIssue creates a new issue in the given repository.
// Usage: CreateIssue(...)
func (c *Client) CreateIssue(owner, repo string, opts gitea.CreateIssueOption) (*gitea.Issue, error) {
	issue, _, err := c.api.CreateIssue(owner, repo, opts)
	if err != nil {
		return nil, log.E("gitea.CreateIssue", "failed to create issue", err)
	}

	return issue, nil
}

// EditIssue edits an existing issue.
// Usage: EditIssue(...)
func (c *Client) EditIssue(owner, repo string, number int64, opts gitea.EditIssueOption) (*gitea.Issue, error) {
	issue, _, err := c.api.EditIssue(owner, repo, number, opts)
	if err != nil {
		return nil, log.E("gitea.EditIssue", "failed to edit issue", err)
	}

	return issue, nil
}

// AssignIssue assigns an issue to the specified users.
// Usage: AssignIssue(...)
func (c *Client) AssignIssue(owner, repo string, number int64, assignees []string) error {
	_, _, err := c.api.EditIssue(owner, repo, number, gitea.EditIssueOption{
		Assignees: assignees,
	})
	if err != nil {
		return log.E("gitea.AssignIssue", "failed to assign issue", err)
	}
	return nil
}

// CreateIssueComment posts a comment on an issue or pull request.
// Usage: CreateIssueComment(...)
func (c *Client) CreateIssueComment(owner, repo string, issue int64, body string) error {
	_, _, err := c.api.CreateIssueComment(owner, repo, issue, gitea.CreateIssueCommentOption{
		Body: body,
	})
	if err != nil {
		return log.E("gitea.CreateIssueComment", "failed to create comment", err)
	}
	return nil
}

// ListIssueComments returns all comments for an issue.
// Usage: ListIssueComments(...)
func (c *Client) ListIssueComments(owner, repo string, number int64) ([]*gitea.Comment, error) {
	var all []*gitea.Comment
	page := 1

	for {
		comments, resp, err := c.api.ListIssueComments(owner, repo, number, gitea.ListIssueCommentOptions{
			ListOptions: gitea.ListOptions{Page: page, PageSize: commentPageSize},
		})
		if err != nil {
			return nil, log.E("gitea.ListIssueComments", "failed to list comments", err)
		}

		all = append(all, comments...)

		if resp == nil || page >= resp.LastPage {
			break
		}
		page++
	}

	return all, nil
}

// ListPullRequests returns pull requests for the given repository.
// Usage: ListPullRequests(...)
func (c *Client) ListPullRequests(owner, repo string, state string) ([]*gitea.PullRequest, error) {
	st := gitea.StateOpen
	switch state {
	case "closed":
		st = gitea.StateClosed
	case "all":
		st = gitea.StateAll
	}

	var all []*gitea.PullRequest
	page := 1

	for {
		prs, resp, err := c.api.ListRepoPullRequests(owner, repo, gitea.ListPullRequestsOptions{
			ListOptions: gitea.ListOptions{Page: page, PageSize: 50},
			State:       st,
		})
		if err != nil {
			return nil, log.E("gitea.ListPullRequests", "failed to list pull requests", err)
		}

		all = append(all, prs...)

		if resp == nil || page >= resp.LastPage {
			break
		}
		page++
	}

	return all, nil
}

// ListPullRequestsIter returns an iterator over pull requests for the given repository.
// Usage: ListPullRequestsIter(...)
func (c *Client) ListPullRequestsIter(owner, repo string, state string) iter.Seq2[*gitea.PullRequest, error] {
	st := gitea.StateOpen
	switch state {
	case "closed":
		st = gitea.StateClosed
	case "all":
		st = gitea.StateAll
	}

	return func(yield func(*gitea.PullRequest, error) bool) {
		page := 1
		for {
			prs, resp, err := c.api.ListRepoPullRequests(owner, repo, gitea.ListPullRequestsOptions{
				ListOptions: gitea.ListOptions{Page: page, PageSize: 50},
				State:       st,
			})
			if err != nil {
				yield(nil, log.E("gitea.ListPullRequests", "failed to list pull requests", err))
				return
			}
			for _, pr := range prs {
				if !yield(pr, nil) {
					return
				}
			}
			if resp == nil || page >= resp.LastPage {
				break
			}
			page++
		}
	}
}

// ListIssueCommentsIter returns an iterator over comments for an issue.
// Usage: ListIssueCommentsIter(...)
func (c *Client) ListIssueCommentsIter(owner, repo string, number int64) iter.Seq2[*gitea.Comment, error] {
	return func(yield func(*gitea.Comment, error) bool) {
		page := 1
		for {
			comments, resp, err := c.api.ListIssueComments(owner, repo, number, gitea.ListIssueCommentOptions{
				ListOptions: gitea.ListOptions{Page: page, PageSize: commentPageSize},
			})
			if err != nil {
				yield(nil, log.E("gitea.ListIssueComments", "failed to list comments", err))
				return
			}
			for _, comment := range comments {
				if !yield(comment, nil) {
					return
				}
			}
			if resp == nil || page >= resp.LastPage {
				break
			}
			page++
		}
	}
}

// GetIssueLabels returns the labels currently attached to an issue.
// Usage: GetIssueLabels(...)
func (c *Client) GetIssueLabels(owner, repo string, number int64) ([]*gitea.Label, error) {
	labels, _, err := c.api.GetIssueLabels(owner, repo, number, gitea.ListLabelsOptions{})
	if err != nil {
		return nil, log.E("gitea.GetIssueLabels", "failed to get issue labels", err)
	}

	return labels, nil
}

// AddIssueLabels adds labels to an issue.
// Usage: AddIssueLabels(...)
func (c *Client) AddIssueLabels(owner, repo string, number int64, labelIDs []int64) error {
	_, _, err := c.api.AddIssueLabels(owner, repo, number, gitea.IssueLabelsOption{
		Labels: labelIDs,
	})
	if err != nil {
		return log.E("gitea.AddIssueLabels", "failed to add labels to issue", err)
	}
	return nil
}

// RemoveIssueLabel removes a label from an issue.
// Usage: RemoveIssueLabel(...)
func (c *Client) RemoveIssueLabel(owner, repo string, number, labelID int64) error {
	_, err := c.api.DeleteIssueLabel(owner, repo, number, labelID)
	if err != nil {
		return log.E("gitea.RemoveIssueLabel", "failed to remove label from issue", err)
	}
	return nil
}

// CloseIssue closes an issue by setting its state to closed.
// Usage: CloseIssue(...)
func (c *Client) CloseIssue(owner, repo string, number int64) error {
	closed := gitea.StateClosed
	_, _, err := c.api.EditIssue(owner, repo, number, gitea.EditIssueOption{
		State: &closed,
	})
	if err != nil {
		return log.E("gitea.CloseIssue", "failed to close issue", err)
	}
	return nil
}

// GetPullRequest returns a single pull request by number.
// Usage: GetPullRequest(...)
func (c *Client) GetPullRequest(owner, repo string, number int64) (*gitea.PullRequest, error) {
	pr, _, err := c.api.GetPullRequest(owner, repo, number)
	if err != nil {
		return nil, log.E("gitea.GetPullRequest", "failed to get pull request", err)
	}

	return pr, nil
}
