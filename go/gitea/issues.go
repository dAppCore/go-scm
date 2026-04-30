// SPDX-License-Identifier: EUPL-1.2

package gitea

import (
	// Note: iter.Seq2 is retained because the Gitea client exposes lazy paginated iterators directly.
	"iter"

	"code.gitea.io/sdk/gitea"
)

type ListIssuesOpts struct {
	State  string
	Labels []string
	Page   int
	Limit  int
}

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
			Labels:      opts.Labels,
		})
		if err != nil {
			return nil, err
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
				Labels:      opts.Labels,
			})
			if err != nil {
				yield(nil, err)
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

func (c *Client) GetIssue(owner, repo string, number int64) (*gitea.Issue, error) {
	issue, _, err := c.api.GetIssue(owner, repo, number)
	return issue, err
}

func (c *Client) CreateIssue(owner, repo string, opts gitea.CreateIssueOption) (*gitea.Issue, error) {
	issue, _, err := c.api.CreateIssue(owner, repo, opts)
	return issue, err
}

func (c *Client) EditIssue(owner, repo string, number int64, opts gitea.EditIssueOption) (*gitea.Issue, error) {
	issue, _, err := c.api.EditIssue(owner, repo, number, opts)
	return issue, err
}

func (c *Client) AssignIssue(owner, repo string, number int64, assignees []string) error {
	_, _, err := c.api.EditIssue(owner, repo, number, gitea.EditIssueOption{Assignees: assignees})
	return err
}

func (c *Client) CreateIssueComment(owner, repo string, issue int64, body string) error {
	_, _, err := c.api.CreateIssueComment(owner, repo, issue, gitea.CreateIssueCommentOption{Body: body})
	return err
}

func (c *Client) ListIssueComments(owner, repo string, number int64) ([]*gitea.Comment, error) {
	var all []*gitea.Comment
	page := 1
	for {
		comments, resp, err := c.api.ListIssueComments(owner, repo, number, gitea.ListIssueCommentOptions{
			ListOptions: gitea.ListOptions{Page: page, PageSize: commentPageSize},
		})
		if err != nil {
			return nil, err
		}
		all = append(all, comments...)
		if resp == nil || page >= resp.LastPage {
			break
		}
		page++
	}
	return all, nil
}

func (c *Client) ListIssueCommentsIter(owner, repo string, number int64) iter.Seq2[*gitea.Comment, error] {
	return func(yield func(*gitea.Comment, error) bool) {
		page := 1
		for {
			comments, resp, err := c.api.ListIssueComments(owner, repo, number, gitea.ListIssueCommentOptions{
				ListOptions: gitea.ListOptions{Page: page, PageSize: commentPageSize},
			})
			if err != nil {
				yield(nil, err)
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

func (c *Client) GetIssueLabels(owner, repo string, number int64) ([]*gitea.Label, error) {
	labels, _, err := c.api.GetIssueLabels(owner, repo, number, gitea.ListLabelsOptions{})
	return labels, err
}

func (c *Client) AddIssueLabels(owner, repo string, number int64, labelIDs []int64) error {
	_, _, err := c.api.AddIssueLabels(owner, repo, number, gitea.IssueLabelsOption{Labels: labelIDs})
	return err
}

func (c *Client) RemoveIssueLabel(owner, repo string, number, labelID int64) error {
	_, err := c.api.DeleteIssueLabel(owner, repo, number, labelID)
	return err
}

func (c *Client) CloseIssue(owner, repo string, number int64) error {
	closed := gitea.StateClosed
	_, _, err := c.api.EditIssue(owner, repo, number, gitea.EditIssueOption{State: &closed})
	return err
}

func (c *Client) GetPullRequest(owner, repo string, number int64) (*gitea.PullRequest, error) {
	pr, _, err := c.api.GetPullRequest(owner, repo, number)
	return pr, err
}

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
			return nil, err
		}
		all = append(all, prs...)
		if resp == nil || page >= resp.LastPage {
			break
		}
		page++
	}
	return all, nil
}

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
				yield(nil, err)
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
