// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	// Note: iter.Seq2 is retained because the forge client exposes lazy paginated iterators directly.
	"iter"
	// Note: strings helpers are retained for label de-duplication and case-insensitive matching.
	"strings"

	"codeberg.org/forgejo/go-sdk/forgejo"
)

func (c *Client) CreateRepoLabel(owner, repo string, opts forgejo.CreateLabelOption) (*forgejo.Label, error) {
	label, _, err := c.api.CreateLabel(owner, repo, opts)
	return label, err
}

func (c *Client) ListRepoLabels(owner, repo string) ([]*forgejo.Label, error) {
	return collectForgePages(func(page int) ([]*forgejo.Label, *forgeResponse, error) {
		return c.api.ListRepoLabels(owner, repo, forgejo.ListLabelsOptions{
			ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
		})
	})
}

func (c *Client) ListRepoLabelsIter(owner, repo string) iter.Seq2[*forgejo.Label, error] {
	return func(yield func(*forgejo.Label, error) bool) {
		yieldForgePages(yield, func(page int) ([]*forgejo.Label, *forgeResponse, error) {
			return c.api.ListRepoLabels(owner, repo, forgejo.ListLabelsOptions{
				ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
			})
		})
	}
}

func (c *Client) ListOrgLabels(org string) ([]*forgejo.Label, error) {
	repos, err := c.ListOrgRepos(org)
	if err != nil {
		return nil, err
	}
	if len(repos) == 0 {
		return nil, nil
	}

	seen := make(map[string]struct{}, len(repos))
	var all []*forgejo.Label
	for _, repo := range repos {
		labels, err := c.ListRepoLabels(repo.Owner.UserName, repo.Name)
		if err != nil {
			return nil, err
		}
		all = appendUniqueLabels(all, seen, labels)
	}
	return all, nil
}

func appendUniqueLabels(all []*forgejo.Label, seen map[string]struct{}, labels []*forgejo.Label) []*forgejo.Label {
	for _, label := range labels {
		key := strings.ToLower(label.Name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		all = append(all, label)
	}
	return all
}

func (c *Client) ListOrgLabelsIter(org string) iter.Seq2[*forgejo.Label, error] {
	return func(yield func(*forgejo.Label, error) bool) {
		seen := make(map[string]struct{})
		yieldForgePages(func(repo *forgejo.Repository, err error) bool {
			if err != nil {
				return yield(nil, err)
			}
			return c.yieldRepoLabels(yield, seen, repo)
		}, func(page int) ([]*forgejo.Repository, *forgeResponse, error) {
			return c.api.ListOrgRepos(org, forgejo.ListOrgReposOptions{
				ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
			})
		})
	}
}

func (c *Client) yieldRepoLabels(yield func(*forgejo.Label, error) bool, seen map[string]struct{}, repo *forgejo.Repository) bool {
	labels, err := c.ListRepoLabels(repo.Owner.UserName, repo.Name)
	if err != nil {
		return yield(nil, err)
	}
	for _, label := range labels {
		key := strings.ToLower(label.Name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		if !yield(label, nil) {
			return false
		}
	}
	return true
}

func (c *Client) GetLabelByName(owner, repo, name string) (*forgejo.Label, error) {
	labels, err := c.ListRepoLabels(owner, repo)
	if err != nil {
		return nil, err
	}
	for _, label := range labels {
		if strings.EqualFold(label.Name, name) {
			return label, nil
		}
	}
	return nil, nil
}

func (c *Client) EnsureLabel(owner, repo, name, color string) (*forgejo.Label, error) {
	label, err := c.GetLabelByName(owner, repo, name)
	if err == nil && label != nil {
		return label, nil
	}
	return c.CreateRepoLabel(owner, repo, forgejo.CreateLabelOption{Name: name, Color: color})
}

func (c *Client) AddIssueLabels(owner, repo string, number int64, labelIDs []int64) error {
	_, _, err := c.api.AddIssueLabels(owner, repo, number, forgejo.IssueLabelsOption{Labels: labelIDs})
	return err
}

func (c *Client) RemoveIssueLabel(owner, repo string, number, labelID int64) error {
	_, err := c.api.DeleteIssueLabel(owner, repo, number, labelID)
	return err
}
