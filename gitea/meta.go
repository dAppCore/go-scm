// SPDX-License-Identifier: EUPL-1.2

package gitea

import (
	"time"
)

type Comment struct {
	ID        int64
	Author    string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PRMeta struct {
	Number       int64
	Title        string
	State        string
	Author       string
	Branch       string
	BaseBranch   string
	Labels       []string
	Assignees    []string
	IsMerged     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CommentCount int
}

func (c *Client) GetIssueBody(owner, repo string, issue int64) (string, error) { return "", nil }
func (c *Client) GetCommentBodies(owner, repo string, pr int64) ([]Comment, error) {
	return nil, nil
}
func (c *Client) GetPRMeta(owner, repo string, pr int64) (*PRMeta, error) { return &PRMeta{}, nil }
