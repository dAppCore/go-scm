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

func giteaStateFromString(state string) gitea.StateType {
	switch state {
	case "closed":
		return gitea.StateClosed
	case "all":
		return gitea.StateAll
	default:
		return gitea.StateOpen
	}
}

func normalizeGiteaListIssuesOpts(opts ListIssuesOpts) (gitea.StateType, int, int) {
	limit := opts.Limit
	if limit == 0 {
		limit = 50
	}
	page := opts.Page
	if page == 0 {
		page = 1
	}
	return giteaStateFromString(opts.State), page, limit
}

func (c *Client) ListIssues(owner, repo string, opts ListIssuesOpts) ([]*gitea.Issue, error)  /* v090-result-boundary */ {
	state, page, limit := normalizeGiteaListIssuesOpts(opts)
	return collectGiteaLimitedPages(page, limit, func(page int) ([]*gitea.Issue, *gitea.Response, error) {
		return c.api.ListRepoIssues(owner, repo, gitea.ListIssueOption{
			ListOptions: gitea.ListOptions{Page: page, PageSize: limit},
			State:       state,
			Type:        gitea.IssueTypeIssue,
			Labels:      opts.Labels,
		})
	})
}

func (c *Client) ListIssuesIter(owner, repo string, opts ListIssuesOpts) iter.Seq2[*gitea.Issue, error] {
	state, page, limit := normalizeGiteaListIssuesOpts(opts)
	return func(yield func(*gitea.Issue, error) bool) {
		yieldGiteaLimitedPages(yield, page, limit, func(page int) ([]*gitea.Issue, *gitea.Response, error) {
			return c.api.ListRepoIssues(owner, repo, gitea.ListIssueOption{
				ListOptions: gitea.ListOptions{Page: page, PageSize: limit},
				State:       state,
				Type:        gitea.IssueTypeIssue,
				Labels:      opts.Labels,
			})
		})
	}
}

func (c *Client) GetIssue(owner, repo string, number int64) (*gitea.Issue, error)  /* v090-result-boundary */ {
	issue, _, err := c.api.GetIssue(owner, repo, number)
	return issue, err
}

func (c *Client) CreateIssue(owner, repo string, opts gitea.CreateIssueOption) (*gitea.Issue, error)  /* v090-result-boundary */ {
	issue, _, err := c.api.CreateIssue(owner, repo, opts)
	return issue, err
}

func (c *Client) EditIssue(owner, repo string, number int64, opts gitea.EditIssueOption) (*gitea.Issue, error)  /* v090-result-boundary */ {
	issue, _, err := c.api.EditIssue(owner, repo, number, opts)
	return issue, err
}

func (c *Client) AssignIssue(owner, repo string, number int64, assignees []string) error  /* v090-result-boundary */ {
	_, _, err := c.api.EditIssue(owner, repo, number, gitea.EditIssueOption{Assignees: assignees})
	return err
}

func (c *Client) CreateIssueComment(owner, repo string, issue int64, body string) error  /* v090-result-boundary */ {
	_, _, err := c.api.CreateIssueComment(owner, repo, issue, gitea.CreateIssueCommentOption{Body: body})
	return err
}

func (c *Client) ListIssueComments(owner, repo string, number int64) ([]*gitea.Comment, error)  /* v090-result-boundary */ {
	return collectGiteaPages(func(page int) ([]*gitea.Comment, *gitea.Response, error) {
		return c.api.ListIssueComments(owner, repo, number, gitea.ListIssueCommentOptions{
			ListOptions: gitea.ListOptions{Page: page, PageSize: commentPageSize},
		})
	})
}

func (c *Client) ListIssueCommentsIter(owner, repo string, number int64) iter.Seq2[*gitea.Comment, error] {
	return func(yield func(*gitea.Comment, error) bool) {
		yieldGiteaPages(yield, func(page int) ([]*gitea.Comment, *gitea.Response, error) {
			return c.api.ListIssueComments(owner, repo, number, gitea.ListIssueCommentOptions{
				ListOptions: gitea.ListOptions{Page: page, PageSize: commentPageSize},
			})
		})
	}
}

func (c *Client) GetIssueLabels(owner, repo string, number int64) ([]*gitea.Label, error)  /* v090-result-boundary */ {
	labels, _, err := c.api.GetIssueLabels(owner, repo, number, gitea.ListLabelsOptions{})
	return labels, err
}

func (c *Client) AddIssueLabels(owner, repo string, number int64, labelIDs []int64) error  /* v090-result-boundary */ {
	_, _, err := c.api.AddIssueLabels(owner, repo, number, gitea.IssueLabelsOption{Labels: labelIDs})
	return err
}

func (c *Client) RemoveIssueLabel(owner, repo string, number, labelID int64) error  /* v090-result-boundary */ {
	_, err := c.api.DeleteIssueLabel(owner, repo, number, labelID)
	return err
}

func (c *Client) CloseIssue(owner, repo string, number int64) error  /* v090-result-boundary */ {
	closed := gitea.StateClosed
	_, _, err := c.api.EditIssue(owner, repo, number, gitea.EditIssueOption{State: &closed})
	return err
}

func (c *Client) GetPullRequest(owner, repo string, number int64) (*gitea.PullRequest, error)  /* v090-result-boundary */ {
	pr, _, err := c.api.GetPullRequest(owner, repo, number)
	return pr, err
}

func (c *Client) ListPullRequests(owner, repo string, state string) ([]*gitea.PullRequest, error)  /* v090-result-boundary */ {
	st := giteaStateFromString(state)
	return collectGiteaPages(func(page int) ([]*gitea.PullRequest, *gitea.Response, error) {
		return c.api.ListRepoPullRequests(owner, repo, gitea.ListPullRequestsOptions{
			ListOptions: gitea.ListOptions{Page: page, PageSize: 50},
			State:       st,
		})
	})
}

func (c *Client) ListPullRequestsIter(owner, repo string, state string) iter.Seq2[*gitea.PullRequest, error] {
	st := giteaStateFromString(state)
	return func(yield func(*gitea.PullRequest, error) bool) {
		yieldGiteaPages(yield, func(page int) ([]*gitea.PullRequest, *gitea.Response, error) {
			return c.api.ListRepoPullRequests(owner, repo, gitea.ListPullRequestsOptions{
				ListOptions: gitea.ListOptions{Page: page, PageSize: 50},
				State:       st,
			})
		})
	}
}
