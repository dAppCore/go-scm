// SPDX-License-Identifier: EUPL-1.2

package gitea

import (
	"time"

	"code.gitea.io/sdk/gitea"
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

const commentPageSize = 50

func (c *Client) GetIssueBody(owner, repo string, issue int64) (string, error) {
	iss, _, err := c.api.GetIssue(owner, repo, issue)
	if err != nil {
		return "", err
	}
	return iss.Body, nil
}

func (c *Client) GetCommentBodies(owner, repo string, pr int64) ([]Comment, error) {
	var comments []Comment
	page := 1
	for {
		rawComments, resp, err := c.api.ListIssueComments(owner, repo, pr, gitea.ListIssueCommentOptions{
			ListOptions: gitea.ListOptions{Page: page, PageSize: commentPageSize},
		})
		if err != nil {
			return nil, err
		}
		for _, raw := range rawComments {
			comment := Comment{
				ID:        raw.ID,
				Body:      raw.Body,
				CreatedAt: raw.Created,
				UpdatedAt: raw.Updated,
			}
			if raw.Poster != nil {
				comment.Author = raw.Poster.UserName
			}
			comments = append(comments, comment)
		}
		if resp == nil || page >= resp.LastPage {
			break
		}
		page++
	}
	return comments, nil
}

func (c *Client) GetPRMeta(owner, repo string, pr int64) (*PRMeta, error) {
	pull, _, err := c.api.GetPullRequest(owner, repo, pr)
	if err != nil {
		return nil, err
	}

	meta := &PRMeta{
		Number:     pull.Index,
		Title:      pull.Title,
		State:      string(pull.State),
		Branch:     pull.Head.Ref,
		BaseBranch: pull.Base.Ref,
		IsMerged:   pull.HasMerged,
	}
	if pull.Created != nil {
		meta.CreatedAt = *pull.Created
	}
	if pull.Updated != nil {
		meta.UpdatedAt = *pull.Updated
	}
	if pull.Poster != nil {
		meta.Author = pull.Poster.UserName
	}
	for _, label := range pull.Labels {
		meta.Labels = append(meta.Labels, label.Name)
	}
	for _, assignee := range pull.Assignees {
		meta.Assignees = append(meta.Assignees, assignee.UserName)
	}
	count := 0
	page := 1
	for {
		rawComments, resp, err := c.api.ListIssueComments(owner, repo, pr, gitea.ListIssueCommentOptions{
			ListOptions: gitea.ListOptions{Page: page, PageSize: commentPageSize},
		})
		if err != nil {
			return nil, err
		}
		count += len(rawComments)
		if resp == nil || page >= resp.LastPage {
			break
		}
		page++
	}
	meta.CommentCount = count
	return meta, nil
}
