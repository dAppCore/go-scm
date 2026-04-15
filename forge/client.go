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
	return &Client{url: url, token: token}, nil
}

func (c *Client) API() *forgejo.Client { if c == nil { return nil }; return c.api }
func (c *Client) URL() string          { if c == nil { return "" }; return c.url }
func (c *Client) Token() string        { if c == nil { return "" }; return c.token }

func (c *Client) GetCurrentUser() (*forgejo.User, error) { return nil, nil }
func (c *Client) CreatePullRequest(owner, repo string, opts forgejo.CreatePullRequestOption) (*forgejo.PullRequest, error) {
	return nil, nil
}
func (c *Client) ForkRepo(owner, repo string, org string) (*forgejo.Repository, error) { return nil, nil }
