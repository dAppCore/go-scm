// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	"iter"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"

	"dappco.re/go/core/log"
)

// ListIssuesOpts configures issue listing.
type ListIssuesOpts struct {
	State  string   // "open", "closed", "all"
	Labels []string // filter by label names
	Page   int
	Limit  int
}

// ListIssues returns issues for the given repository.
// Usage: ListIssues(...)
func (c *Client) ListIssues(owner, repo string, opts ListIssuesOpts) ([]*forgejo.Issue, error) {
	state := forgejo.StateOpen
	switch opts.State {
	case "closed":
		state = forgejo.StateClosed
	case "all":
		state = forgejo.StateAll
	}

	limit := opts.Limit
	if limit == 0 {
		limit = 50
	}

	page := opts.Page
	if page == 0 {
		page = 1
	}

	var all []*forgejo.Issue

	for {
		listOpt := forgejo.ListIssueOption{
			ListOptions: forgejo.ListOptions{Page: page, PageSize: limit},
			State:       state,
			Type:        forgejo.IssueTypeIssue,
			Labels:      opts.Labels,
		}

		issues, resp, err := c.api.ListRepoIssues(owner, repo, listOpt)
		if err != nil {
			return nil, log.E("forge.ListIssues", "failed to list issues", err)
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
func (c *Client) ListIssuesIter(owner, repo string, opts ListIssuesOpts) iter.Seq2[*forgejo.Issue, error] {
	state := forgejo.StateOpen
	switch opts.State {
	case "closed":
		state = forgejo.StateClosed
	case "all":
		state = forgejo.StateAll
	}

	limit := opts.Limit
	if limit == 0 {
		limit = 50
	}

	page := opts.Page
	if page == 0 {
		page = 1
	}

	return func(yield func(*forgejo.Issue, error) bool) {
		for {
			issues, resp, err := c.api.ListRepoIssues(owner, repo, forgejo.ListIssueOption{
				ListOptions: forgejo.ListOptions{Page: page, PageSize: limit},
				State:       state,
				Type:        forgejo.IssueTypeIssue,
				Labels:      opts.Labels,
			})
			if err != nil {
				yield(nil, log.E("forge.ListIssues", "failed to list issues", err))
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
func (c *Client) GetIssue(owner, repo string, number int64) (*forgejo.Issue, error) {
	issue, _, err := c.api.GetIssue(owner, repo, number)
	if err != nil {
		return nil, log.E("forge.GetIssue", "failed to get issue", err)
	}

	return issue, nil
}

// CreateIssue creates a new issue in the given repository.
// Usage: CreateIssue(...)
func (c *Client) CreateIssue(owner, repo string, opts forgejo.CreateIssueOption) (*forgejo.Issue, error) {
	issue, _, err := c.api.CreateIssue(owner, repo, opts)
	if err != nil {
		return nil, log.E("forge.CreateIssue", "failed to create issue", err)
	}

	return issue, nil
}

// EditIssue edits an existing issue.
// Usage: EditIssue(...)
func (c *Client) EditIssue(owner, repo string, number int64, opts forgejo.EditIssueOption) (*forgejo.Issue, error) {
	issue, _, err := c.api.EditIssue(owner, repo, number, opts)
	if err != nil {
		return nil, log.E("forge.EditIssue", "failed to edit issue", err)
	}

	return issue, nil
}

// AssignIssue assigns an issue to the specified users.
// Usage: AssignIssue(...)
func (c *Client) AssignIssue(owner, repo string, number int64, assignees []string) error {
	_, _, err := c.api.EditIssue(owner, repo, number, forgejo.EditIssueOption{
		Assignees: assignees,
	})
	if err != nil {
		return log.E("forge.AssignIssue", "failed to assign issue", err)
	}
	return nil
}

// ListPullRequests returns pull requests for the given repository.
// Usage: ListPullRequests(...)
func (c *Client) ListPullRequests(owner, repo string, state string) ([]*forgejo.PullRequest, error) {
	st := forgejo.StateOpen
	switch state {
	case "closed":
		st = forgejo.StateClosed
	case "all":
		st = forgejo.StateAll
	}

	var all []*forgejo.PullRequest
	page := 1

	for {
		prs, resp, err := c.api.ListRepoPullRequests(owner, repo, forgejo.ListPullRequestsOptions{
			ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
			State:       st,
		})
		if err != nil {
			return nil, log.E("forge.ListPullRequests", "failed to list pull requests", err)
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
func (c *Client) ListPullRequestsIter(owner, repo string, state string) iter.Seq2[*forgejo.PullRequest, error] {
	st := forgejo.StateOpen
	switch state {
	case "closed":
		st = forgejo.StateClosed
	case "all":
		st = forgejo.StateAll
	}

	return func(yield func(*forgejo.PullRequest, error) bool) {
		page := 1
		for {
			prs, resp, err := c.api.ListRepoPullRequests(owner, repo, forgejo.ListPullRequestsOptions{
				ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
				State:       st,
			})
			if err != nil {
				yield(nil, log.E("forge.ListPullRequests", "failed to list pull requests", err))
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

// GetPullRequest returns a single pull request by number.
// Usage: GetPullRequest(...)
func (c *Client) GetPullRequest(owner, repo string, number int64) (*forgejo.PullRequest, error) {
	pr, _, err := c.api.GetPullRequest(owner, repo, number)
	if err != nil {
		return nil, log.E("forge.GetPullRequest", "failed to get pull request", err)
	}

	return pr, nil
}

// CreateIssueComment posts a comment on an issue or pull request.
// Usage: CreateIssueComment(...)
func (c *Client) CreateIssueComment(owner, repo string, issue int64, body string) error {
	_, _, err := c.api.CreateIssueComment(owner, repo, issue, forgejo.CreateIssueCommentOption{
		Body: body,
	})
	if err != nil {
		return log.E("forge.CreateIssueComment", "failed to create comment", err)
	}
	return nil
}

// GetIssueLabels returns the labels currently attached to an issue.
// Usage: GetIssueLabels(...)
func (c *Client) GetIssueLabels(owner, repo string, number int64) ([]*forgejo.Label, error) {
	labels, _, err := c.api.GetIssueLabels(owner, repo, number, forgejo.ListLabelsOptions{})
	if err != nil {
		return nil, log.E("forge.GetIssueLabels", "failed to get issue labels", err)
	}

	return labels, nil
}

// ListIssueComments returns comments for an issue.
// Usage: ListIssueComments(...)
func (c *Client) ListIssueComments(owner, repo string, number int64) ([]*forgejo.Comment, error) {
	var all []*forgejo.Comment
	page := 1

	for {
		comments, resp, err := c.api.ListIssueComments(owner, repo, number, forgejo.ListIssueCommentOptions{
			ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
		})
		if err != nil {
			return nil, log.E("forge.ListIssueComments", "failed to list comments", err)
		}

		all = append(all, comments...)

		if resp == nil || page >= resp.LastPage {
			break
		}
		page++
	}

	return all, nil
}

// ListIssueCommentsIter returns an iterator over comments for an issue.
// Usage: ListIssueCommentsIter(...)
func (c *Client) ListIssueCommentsIter(owner, repo string, number int64) iter.Seq2[*forgejo.Comment, error] {
	return func(yield func(*forgejo.Comment, error) bool) {
		page := 1
		for {
			comments, resp, err := c.api.ListIssueComments(owner, repo, number, forgejo.ListIssueCommentOptions{
				ListOptions: forgejo.ListOptions{Page: page, PageSize: commentPageSize},
			})
			if err != nil {
				yield(nil, log.E("forge.ListIssueComments", "failed to list comments", err))
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

// CloseIssue closes an issue by setting its state to closed.
// Usage: CloseIssue(...)
func (c *Client) CloseIssue(owner, repo string, number int64) error {
	closed := forgejo.StateClosed
	_, _, err := c.api.EditIssue(owner, repo, number, forgejo.EditIssueOption{
		State: &closed,
	})
	if err != nil {
		return log.E("forge.CloseIssue", "failed to close issue", err)
	}
	return nil
}
