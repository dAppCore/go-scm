// SPDX-License-Identifier: EUPL-1.2

package gitea

import (
	// Note: time.Time mirrors Gitea metadata timestamps in public structs.
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

func collectGiteaPages[T any](fetch func(page int) ([]T, *gitea.Response, error)) ([]T, error) {
	var all []T
	for page := 1; ; page++ {
		items, resp, err := fetch(page)
		if err != nil {
			return nil, err
		}
		all = append(all, items...)
		if !hasNextGiteaPage(resp, page) {
			return all, nil
		}
	}
}

func collectGiteaLimitedPages[T any](page, limit int, fetch func(page int) ([]T, *gitea.Response, error)) ([]T, error) {
	var all []T
	for {
		items, resp, err := fetch(page)
		if err != nil {
			return nil, err
		}
		all = append(all, items...)
		if !hasMoreGiteaItems(items, resp, page, limit) {
			return all, nil
		}
		page++
	}
}

func yieldGiteaPages[T any](yield func(T, error) bool, fetch func(page int) ([]T, *gitea.Response, error)) {
	for page := 1; ; page++ {
		items, resp, err := fetch(page)
		if err != nil {
			var zero T
			yield(zero, err)
			return
		}
		for _, item := range items {
			if !yield(item, nil) {
				return
			}
		}
		if !hasNextGiteaPage(resp, page) {
			return
		}
	}
}

func yieldGiteaLimitedPages[T any](yield func(T, error) bool, page, limit int, fetch func(page int) ([]T, *gitea.Response, error)) {
	for {
		items, resp, err := fetch(page)
		if err != nil {
			var zero T
			yield(zero, err)
			return
		}
		for _, item := range items {
			if !yield(item, nil) {
				return
			}
		}
		if !hasMoreGiteaItems(items, resp, page, limit) {
			return
		}
		page++
	}
}

func hasNextGiteaPage(resp *gitea.Response, page int) bool {
	return resp != nil && page < resp.LastPage
}

func hasMoreGiteaItems[T any](items []T, resp *gitea.Response, page, limit int) bool {
	if len(items) == 0 || len(items) < limit {
		return false
	}
	return resp == nil || resp.LastPage <= 0 || page < resp.LastPage
}

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
