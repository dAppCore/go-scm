// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	"time"

	"dappco.re/go/core/log"
)

// PRMeta holds structural signals from a pull request,
// used by the pipeline MetaReader for AI-driven workflows.
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

// Comment represents a comment with metadata.
type Comment struct {
	ID        int64
	Author    string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

const commentPageSize = 50

// GetPRMeta returns structural signals for a pull request.
// This is the Forgejo side of the dual MetaReader described in the pipeline design.
// Usage: GetPRMeta(...)
func (c *Client) GetPRMeta(owner, repo string, pr int64) (*PRMeta, error) {
	pull, _, err := c.api.GetPullRequest(owner, repo, pr)
	if err != nil {
		return nil, log.E("forge.GetPRMeta", "failed to get PR metadata", err)
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

	// Fetch comment count from the issue side (PRs are issues in Forgejo).
	// Paginate to get an accurate count.
	count := 0
	for _, err := range c.ListIssueCommentsIter(owner, repo, pr) {
		if err != nil {
			return nil, log.E("forge.GetPRMeta", "list issue comments", err)
		}
		count++
	}
	meta.CommentCount = count

	return meta, nil
}

// GetCommentBodies returns all comment bodies for a pull request.
// Usage: GetCommentBodies(...)
func (c *Client) GetCommentBodies(owner, repo string, pr int64) ([]Comment, error) {
	var comments []Comment
	for raw, err := range c.ListIssueCommentsIter(owner, repo, pr) {
		if err != nil {
			return nil, log.E("forge.GetCommentBodies", "failed to get PR comments", err)
		}

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

	return comments, nil
}

// GetIssueBody returns the body text of an issue.
// Usage: GetIssueBody(...)
func (c *Client) GetIssueBody(owner, repo string, issue int64) (string, error) {
	iss, _, err := c.api.GetIssue(owner, repo, issue)
	if err != nil {
		return "", log.E("forge.GetIssueBody", "failed to get issue body", err)
	}

	return iss.Body, nil
}
