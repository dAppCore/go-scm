// SPDX-License-Identifier: EUPL-1.2

package gitea

import (
	"errors"

	"code.gitea.io/sdk/gitea"
)

type Client struct {
	api   *gitea.Client
	url   string
	token string
}

func New(url, token string) (*Client, error) {
	if url == "" {
		return nil, errors.New("gitea.New: url is required")
	}
	api, err := gitea.NewClient(url, gitea.SetToken(token))
	if err != nil {
		return nil, err
	}
	return &Client{api: api, url: url, token: token}, nil
}

func (c *Client) API() *gitea.Client {
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

