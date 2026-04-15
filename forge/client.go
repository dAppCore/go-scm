// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	"errors"

	"codeberg.org/forgejo/go-sdk/forgejo"
)

type Client struct {
	api   *forgejo.Client
	url   string
	token string
}

func New(url, token string) (*Client, error) {
	if url == "" {
		return nil, errors.New("forge.New: url is required")
	}
	api, err := forgejo.NewClient(url, forgejo.SetToken(token))
	if err != nil {
		return nil, err
	}
	return &Client{api: api, url: url, token: token}, nil
}

func (c *Client) API() *forgejo.Client {
	if c == nil {
		return nil
	}
	return c.api
}

func (c *Client) URL() string {
	if c == nil {
		return ""
	}
	return c.url
}

func (c *Client) Token() string {
	if c == nil {
		return ""
	}
	return c.token
}

func (c *Client) GetCurrentUser() (*forgejo.User, error) {
	user, _, err := c.api.GetMyUserInfo()
	return user, err
}

func (c *Client) CreatePullRequest(owner, repo string, opts forgejo.CreatePullRequestOption) (*forgejo.PullRequest, error) {
	pr, _, err := c.api.CreatePullRequest(owner, repo, opts)
	return pr, err
}

func (c *Client) ForkRepo(owner, repo string, org string) (*forgejo.Repository, error) {
	opts := forgejo.CreateForkOption{}
	if org != "" {
		opts.Organization = &org
	}
	fork, _, err := c.api.CreateFork(owner, repo, opts)
	return fork, err
}

