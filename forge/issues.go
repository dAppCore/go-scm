// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	"iter"

	"codeberg.org/forgejo/go-sdk/forgejo"
)

type ListIssuesOpts struct {
	State  string
	Labels []string
	Page   int
	Limit  int
}

func (c *Client) GetIssue(owner, repo string, number int64) (*forgejo.Issue, error) {
	issue, _, err := c.api.GetIssue(owner, repo, number)
	return issue, err
}

func (c *Client) EditIssue(owner, repo string, number int64, opts forgejo.EditIssueOption) (*forgejo.Issue, error) {
	issue, _, err := c.api.EditIssue(owner, repo, number, opts)
	return issue, err
}

func (c *Client) CloseIssue(owner, repo string, number int64) error {
	closed := forgejo.StateClosed
	_, _, err := c.api.EditIssue(owner, repo, number, forgejo.EditIssueOption{State: &closed})
	return err
}

func (c *Client) CreateIssue(owner, repo string, opts forgejo.CreateIssueOption) (*forgejo.Issue, error) {
	issue, _, err := c.api.CreateIssue(owner, repo, opts)
	return issue, err
}

func (c *Client) CreateIssueComment(owner, repo string, issue int64, body string) error {
	_, _, err := c.api.CreateIssueComment(owner, repo, issue, forgejo.CreateIssueCommentOption{Body: body})
	return err
}

func (c *Client) AssignIssue(owner, repo string, number int64, assignees []string) error {
	_, _, err := c.api.EditIssue(owner, repo, number, forgejo.EditIssueOption{Assignees: assignees})
	return err
}

func (c *Client) ListIssueComments(owner, repo string, number int64) ([]*forgejo.Comment, error) {
	var all []*forgejo.Comment
	page := 1
	for {
		comments, resp, err := c.api.ListIssueComments(owner, repo, number, forgejo.ListIssueCommentOptions{
			ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
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

func (c *Client) ListIssueCommentsIter(owner, repo string, number int64) iter.Seq2[*forgejo.Comment, error] {
	return func(yield func(*forgejo.Comment, error) bool) {
		page := 1
		for {
			comments, resp, err := c.api.ListIssueComments(owner, repo, number, forgejo.ListIssueCommentOptions{
				ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
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

func (c *Client) GetIssueLabels(owner, repo string, number int64) ([]*forgejo.Label, error) {
	labels, _, err := c.api.GetIssueLabels(owner, repo, number, forgejo.ListLabelsOptions{})
	return labels, err
}

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
		issues, resp, err := c.api.ListRepoIssues(owner, repo, forgejo.ListIssueOption{
			ListOptions: forgejo.ListOptions{Page: page, PageSize: limit},
			State:       state,
			Type:        forgejo.IssueTypeIssue,
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

func (c *Client) GetPullRequest(owner, repo string, number int64) (*forgejo.PullRequest, error) {
	pr, _, err := c.api.GetPullRequest(owner, repo, number)
	return pr, err
}
