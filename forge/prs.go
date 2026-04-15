// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	"iter"

	"codeberg.org/forgejo/go-sdk/forgejo"
)

func (c *Client) DismissReview(owner, repo string, index, reviewID int64, message string) error { return nil }
func (c *Client) GetCombinedStatus(owner, repo string, ref string) (*forgejo.CombinedStatus, error) {
	return nil, nil
}
func (c *Client) ListPRReviews(owner, repo string, index int64) ([]*forgejo.PullReview, error) { return nil, nil }
func (c *Client) MergePullRequest(owner, repo string, index int64, method string) error        { return nil }
func (c *Client) SetPRDraft(owner, repo string, index int64, draft bool) error                 { return nil }
func (c *Client) ListPullRequestsIter(owner, repo string, state string) iter.Seq2[*forgejo.PullRequest, error] {
	return func(yield func(*forgejo.PullRequest, error) bool) {}
}
