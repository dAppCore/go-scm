// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	// Note: iter.Seq2 is retained because the forge client exposes lazy paginated iterators directly.
	"iter"

	"codeberg.org/forgejo/go-sdk/forgejo"
)

type ListIssuesOpts struct {
	State  string
	Labels []string
	Page   int
	Limit  int
}

func forgeStateFromString(state string) forgejo.StateType {
	switch state {
	case "closed":
		return forgejo.StateClosed
	case "all":
		return forgejo.StateAll
	default:
		return forgejo.StateOpen
	}
}

func normalizeForgeListIssuesOpts(opts ListIssuesOpts) (forgejo.StateType, int, int) {
	limit := opts.Limit
	if limit == 0 {
		limit = 50
	}
	page := opts.Page
	if page == 0 {
		page = 1
	}
	return forgeStateFromString(opts.State), page, limit
}

func (c *Client) GetIssue(owner, repo string, number int64) (*forgejo.Issue, error)  /* v090-result-boundary */ {
	issue, _, err := c.api.GetIssue(owner, repo, number)
	return issue, err
}

func (c *Client) EditIssue(owner, repo string, number int64, opts forgejo.EditIssueOption) (*forgejo.Issue, error)  /* v090-result-boundary */ {
	issue, _, err := c.api.EditIssue(owner, repo, number, opts)
	return issue, err
}

func (c *Client) CloseIssue(owner, repo string, number int64) error  /* v090-result-boundary */ {
	closed := forgejo.StateClosed
	_, _, err := c.api.EditIssue(owner, repo, number, forgejo.EditIssueOption{State: &closed})
	return err
}

func (c *Client) CreateIssue(owner, repo string, opts forgejo.CreateIssueOption) (*forgejo.Issue, error)  /* v090-result-boundary */ {
	issue, _, err := c.api.CreateIssue(owner, repo, opts)
	return issue, err
}

func (c *Client) CreateIssueComment(owner, repo string, issue int64, body string) error  /* v090-result-boundary */ {
	_, _, err := c.api.CreateIssueComment(owner, repo, issue, forgejo.CreateIssueCommentOption{Body: body})
	return err
}

func (c *Client) AssignIssue(owner, repo string, number int64, assignees []string) error  /* v090-result-boundary */ {
	_, _, err := c.api.EditIssue(owner, repo, number, forgejo.EditIssueOption{Assignees: assignees})
	return err
}

func (c *Client) ListIssueComments(owner, repo string, number int64) ([]*forgejo.Comment, error)  /* v090-result-boundary */ {
	return collectForgePages(func(page int) ([]*forgejo.Comment, *forgeResponse, error) {
		return c.api.ListIssueComments(owner, repo, number, forgejo.ListIssueCommentOptions{
			ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
		})
	})
}

func (c *Client) ListIssueCommentsIter(owner, repo string, number int64) iter.Seq2[*forgejo.Comment, error] {
	return func(yield func(*forgejo.Comment, error) bool) {
		yieldForgePages(yield, func(page int) ([]*forgejo.Comment, *forgeResponse, error) {
			return c.api.ListIssueComments(owner, repo, number, forgejo.ListIssueCommentOptions{
				ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
			})
		})
	}
}

func (c *Client) GetIssueLabels(owner, repo string, number int64) ([]*forgejo.Label, error)  /* v090-result-boundary */ {
	labels, _, err := c.api.GetIssueLabels(owner, repo, number, forgejo.ListLabelsOptions{})
	return labels, err
}

func (c *Client) ListIssues(owner, repo string, opts ListIssuesOpts) ([]*forgejo.Issue, error)  /* v090-result-boundary */ {
	state, page, limit := normalizeForgeListIssuesOpts(opts)
	return collectForgeLimitedPages(page, limit, func(page int) ([]*forgejo.Issue, *forgeResponse, error) {
		return c.api.ListRepoIssues(owner, repo, forgejo.ListIssueOption{
			ListOptions: forgejo.ListOptions{Page: page, PageSize: limit},
			State:       state,
			Type:        forgejo.IssueTypeIssue,
			Labels:      opts.Labels,
		})
	})
}

func (c *Client) ListPullRequests(owner, repo string, state string) ([]*forgejo.PullRequest, error)  /* v090-result-boundary */ {
	st := forgeStateFromString(state)
	return collectForgePages(func(page int) ([]*forgejo.PullRequest, *forgeResponse, error) {
		return c.api.ListRepoPullRequests(owner, repo, forgejo.ListPullRequestsOptions{
			ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
			State:       st,
		})
	})
}

func (c *Client) ListPullRequestsIter(owner, repo string, state string) iter.Seq2[*forgejo.PullRequest, error] {
	st := forgeStateFromString(state)
	return func(yield func(*forgejo.PullRequest, error) bool) {
		yieldForgePages(yield, func(page int) ([]*forgejo.PullRequest, *forgeResponse, error) {
			return c.api.ListRepoPullRequests(owner, repo, forgejo.ListPullRequestsOptions{
				ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
				State:       st,
			})
		})
	}
}

func (c *Client) GetPullRequest(owner, repo string, number int64) (*forgejo.PullRequest, error)  /* v090-result-boundary */ {
	pr, _, err := c.api.GetPullRequest(owner, repo, number)
	return pr, err
}
