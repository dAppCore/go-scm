// SPDX-License-Identifier: EUPL-1.2

package gitea

import (
	"errors"

	"code.gitea.io/sdk/gitea"
)

type Client struct {
	api *gitea.Client
	url string
}

func New(url string) (*Client, error) {
	if url == "" {
		return nil, errors.New("gitea.New: url is required")
	}
	return &Client{url: url}, nil
}

func (c *Client) API() *gitea.Client { if c == nil { return nil }; return c.api }
func (c *Client) URL() string        { if c == nil { return "" }; return c.url }
