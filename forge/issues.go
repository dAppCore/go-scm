// SPDX-License-Identifier: EUPL-1.2

package forge

import "codeberg.org/forgejo/go-sdk/forgejo"

type ListIssuesOpts struct {
	State  string
	Labels []string
	Page   int
	Limit  int
}

func (c *Client) GetIssue(owner, repo string, number int64) (*forgejo.Issue, error) { return nil, nil }
func (c *Client) EditIssue(owner, repo string, number int64, opts forgejo.EditIssueOption) (*forgejo.Issue, error) {
	return nil, nil
}
func (c *Client) CloseIssue(owner, repo string, number int64) error { return nil }
func (c *Client) CreateIssue(owner, repo string, opts forgejo.CreateIssueOption) (*forgejo.Issue, error) {
	return nil, nil
}
func (c *Client) CreateIssueComment(owner, repo string, issue int64, body string) error { return nil }
func (c *Client) AssignIssue(owner, repo string, number int64, assignees []string) error { return nil }
func (c *Client) ListIssueComments(owner, repo string, number int64) ([]*forgejo.Comment, error) {
	return nil, nil
}
func (c *Client) ListIssues(owner, repo string, opts ListIssuesOpts) ([]*forgejo.Issue, error) { return nil, nil }
func (c *Client) GetPullRequest(owner, repo string, number int64) (*forgejo.PullRequest, error) { return nil, nil }
func (c *Client) ListPullRequests(owner, repo string, state string) ([]*forgejo.PullRequest, error) { return nil, nil }
