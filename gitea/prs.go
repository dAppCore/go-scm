// SPDX-License-Identifier: EUPL-1.2

package gitea

import (
	"bytes"
	"iter"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"dappco.re/go/core/log"
	"dappco.re/go/core/scm/agentci"
	"dappco.re/go/core/scm/internal/ax/jsonx"

	"code.gitea.io/sdk/gitea"
)

// httpClient is a package-level client with a timeout to avoid hanging indefinitely.
var httpClient = &http.Client{Timeout: 30 * time.Second}

// MergePullRequest merges a pull request with the given method ("squash", "rebase", "merge").
// Usage: MergePullRequest(...)
func (c *Client) MergePullRequest(owner, repo string, index int64, method string) error {
	style := gitea.MergeStyleMerge
	switch method {
	case "squash":
		style = gitea.MergeStyleSquash
	case "rebase":
		style = gitea.MergeStyleRebase
	case "rebase-merge":
		style = gitea.MergeStyleRebaseMerge
	}

	merged, _, err := c.api.MergePullRequest(owner, repo, index, gitea.MergePullRequestOption{
		Style:                  style,
		DeleteBranchAfterMerge: true,
	})
	if err != nil {
		return log.E("gitea.MergePullRequest", "failed to merge pull request", err)
	}
	if !merged {
		return log.E("gitea.MergePullRequest", "failed to merge pull request", nil)
	}
	return nil
}

// SetPRDraft sets or clears the draft status on a pull request.
// The Gitea SDK exposes draft state on the model, but the edit option does not
// currently include a draft field, so we use a raw PATCH request.
// Usage: SetPRDraft(...)
func (c *Client) SetPRDraft(owner, repo string, index int64, draft bool) error {
	safeOwner, err := agentci.ValidatePathElement(owner)
	if err != nil {
		return log.E("gitea.SetPRDraft", "invalid owner", err)
	}
	safeRepo, err := agentci.ValidatePathElement(repo)
	if err != nil {
		return log.E("gitea.SetPRDraft", "invalid repo", err)
	}

	payload := map[string]bool{"draft": draft}
	body, err := jsonx.Marshal(payload)
	if err != nil {
		return log.E("gitea.SetPRDraft", "marshal payload", err)
	}

	path, err := url.JoinPath(c.url, "api", "v1", "repos", safeOwner, safeRepo, "pulls", strconv.FormatInt(index, 10))
	if err != nil {
		return log.E("gitea.SetPRDraft", "failed to build request path", err)
	}

	req, err := http.NewRequest(http.MethodPatch, path, bytes.NewReader(body))
	if err != nil {
		return log.E("gitea.SetPRDraft", "create request", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "token "+c.Token())

	resp, err := httpClient.Do(req)
	if err != nil {
		return log.E("gitea.SetPRDraft", "failed to update draft status", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return log.E("gitea.SetPRDraft", "unexpected status "+strconv.Itoa(resp.StatusCode), nil)
	}
	return nil
}

// ListPRReviews returns all reviews for a pull request.
// Usage: ListPRReviews(...)
func (c *Client) ListPRReviews(owner, repo string, index int64) ([]*gitea.PullReview, error) {
	var all []*gitea.PullReview
	page := 1

	for {
		reviews, resp, err := c.api.ListPullReviews(owner, repo, index, gitea.ListPullReviewsOptions{
			ListOptions: gitea.ListOptions{Page: page, PageSize: 50},
		})
		if err != nil {
			return nil, log.E("gitea.ListPRReviews", "failed to list reviews", err)
		}

		all = append(all, reviews...)

		if resp == nil || page >= resp.LastPage {
			break
		}
		page++
	}

	return all, nil
}

// ListPRReviewsIter returns an iterator over reviews for a pull request.
// Usage: ListPRReviewsIter(...)
func (c *Client) ListPRReviewsIter(owner, repo string, index int64) iter.Seq2[*gitea.PullReview, error] {
	return func(yield func(*gitea.PullReview, error) bool) {
		page := 1
		for {
			reviews, resp, err := c.api.ListPullReviews(owner, repo, index, gitea.ListPullReviewsOptions{
				ListOptions: gitea.ListOptions{Page: page, PageSize: 50},
			})
			if err != nil {
				yield(nil, log.E("gitea.ListPRReviews", "failed to list reviews", err))
				return
			}
			for _, review := range reviews {
				if !yield(review, nil) {
					return
				}
			}
			if resp == nil || page >= resp.LastPage {
				break
			}
			page++
		}
	}
}

// GetCombinedStatus returns the combined commit status for a ref (SHA or branch).
// Usage: GetCombinedStatus(...)
func (c *Client) GetCombinedStatus(owner, repo string, ref string) (*gitea.CombinedStatus, error) {
	status, _, err := c.api.GetCombinedStatus(owner, repo, ref)
	if err != nil {
		return nil, log.E("gitea.GetCombinedStatus", "failed to get combined status", err)
	}
	return status, nil
}

// DismissReview dismisses a pull request review by ID.
// Usage: DismissReview(...)
func (c *Client) DismissReview(owner, repo string, index, reviewID int64, message string) error {
	_, err := c.api.DismissPullReview(owner, repo, index, reviewID, gitea.DismissPullReviewOptions{
		Message: message,
	})
	if err != nil {
		return log.E("gitea.DismissReview", "failed to dismiss review", err)
	}
	return nil
}

// UndismissReview removes a dismissal from a pull request review.
// Usage: UndismissReview(...)
func (c *Client) UndismissReview(owner, repo string, index, reviewID int64) error {
	_, err := c.api.UnDismissPullReview(owner, repo, index, reviewID)
	if err != nil {
		return log.E("gitea.UndismissReview", "failed to undismiss review", err)
	}
	return nil
}
