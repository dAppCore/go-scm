// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	// Note: bytes.NewReader is retained for constructing the small PATCH payload body.
	"bytes"
	// Note: iter.Seq2 is retained because the forge client exposes lazy paginated iterators directly.
	"iter"
	// Note: net/http is retained for the raw PATCH endpoint not covered by the Forgejo SDK.
	"net/http"
	// Note: net/url.JoinPath is retained for safe API endpoint assembly.
	"net/url"
	// Note: strconv is retained for bool/int path and JSON literal formatting in the raw PATCH call.
	"strconv"

	"codeberg.org/forgejo/go-sdk/forgejo"
)

func (c *Client) MergePullRequest(owner, repo string, index int64, method string) error {
	style := forgejo.MergeStyleMerge
	switch method {
	case "squash":
		style = forgejo.MergeStyleSquash
	case "rebase":
		style = forgejo.MergeStyleRebase
	}
	deleteAfter := true
	_, _, err := c.api.MergePullRequest(owner, repo, index, forgejo.MergePullRequestOption{
		Style:                  style,
		DeleteBranchAfterMerge: &deleteAfter,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) SetPRDraft(owner, repo string, index int64, draft bool) error {
	payload := []byte(`{"draft":` + strconv.FormatBool(draft) + `}`)
	path, err := url.JoinPath(c.url, "api", "v1", "repos", owner, repo, "pulls", strconv.FormatInt(index, 10))
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPatch, path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "token "+c.token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &httpError{status: resp.StatusCode}
	}
	return nil
}

type httpError struct {
	status int
}

func (e *httpError) Error() string {
	return "unexpected HTTP status"
}

func (c *Client) ListPRReviews(owner, repo string, index int64) ([]*forgejo.PullReview, error) {
	return collectForgePages(func(page int) ([]*forgejo.PullReview, *forgeResponse, error) {
		return c.api.ListPullReviews(owner, repo, index, forgejo.ListPullReviewsOptions{
			ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
		})
	})
}

func (c *Client) ListPRReviewsIter(owner, repo string, index int64) iter.Seq2[*forgejo.PullReview, error] {
	return func(yield func(*forgejo.PullReview, error) bool) {
		yieldForgePages(yield, func(page int) ([]*forgejo.PullReview, *forgeResponse, error) {
			return c.api.ListPullReviews(owner, repo, index, forgejo.ListPullReviewsOptions{
				ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
			})
		})
	}
}

func (c *Client) GetCombinedStatus(owner, repo string, ref string) (*forgejo.CombinedStatus, error) {
	status, _, err := c.api.GetCombinedStatus(owner, repo, ref)
	return status, err
}

func (c *Client) DismissReview(owner, repo string, index, reviewID int64, message string) error {
	_, err := c.api.DismissPullReview(owner, repo, index, reviewID, forgejo.DismissPullReviewOptions{Message: message})
	return err
}

func (c *Client) UndismissReview(owner, repo string, index, reviewID int64) error {
	_, err := c.api.UnDismissPullReview(owner, repo, index, reviewID)
	return err
}
