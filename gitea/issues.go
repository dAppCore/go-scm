// SPDX-License-Identifier: EUPL-1.2

package gitea

import "code.gitea.io/sdk/gitea"

type ListIssuesOpts struct {
	State string
	Page  int
	Limit int
}

func (c *Client) CreateIssue(owner, repo string, opts gitea.CreateIssueOption) (*gitea.Issue, error) { return nil, nil }
func (c *Client) GetIssue(owner, repo string, number int64) (*gitea.Issue, error)                   { return nil, nil }
func (c *Client) GetPullRequest(owner, repo string, number int64) (*gitea.PullRequest, error)       { return nil, nil }
func (c *Client) ListIssues(owner, repo string, opts ListIssuesOpts) ([]*gitea.Issue, error)        { return nil, nil }
func (c *Client) ListPullRequests(owner, repo string, state string) ([]*gitea.PullRequest, error)    { return nil, nil }
