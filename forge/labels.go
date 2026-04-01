// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	"iter"

	strings "dappco.re/go/core/scm/internal/ax/stringsx"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"

	"dappco.re/go/core/log"
)

// ListOrgLabels returns all unique labels across repos in the given organisation.
// Note: The Forgejo SDK does not have a dedicated org-level labels endpoint.
// We aggregate labels from each repo and deduplicate them by name, preserving
// the first seen label metadata.
// Usage: ListOrgLabels(...)
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
		for _, label := range labels {
			key := strings.ToLower(label.Name)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			all = append(all, label)
		}
	}

	return all, nil
}

// ListRepoLabels returns all labels for a repository.
// Usage: ListRepoLabels(...)
func (c *Client) ListRepoLabels(owner, repo string) ([]*forgejo.Label, error) {
	var all []*forgejo.Label
	page := 1

	for {
		labels, resp, err := c.api.ListRepoLabels(owner, repo, forgejo.ListLabelsOptions{
			ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
		})
		if err != nil {
			return nil, log.E("forge.ListRepoLabels", "failed to list repo labels", err)
		}

		all = append(all, labels...)

		if resp == nil || page >= resp.LastPage {
			break
		}
		page++
	}

	return all, nil
}

// ListRepoLabelsIter returns an iterator over labels for a repository.
// Usage: ListRepoLabelsIter(...)
func (c *Client) ListRepoLabelsIter(owner, repo string) iter.Seq2[*forgejo.Label, error] {
	return func(yield func(*forgejo.Label, error) bool) {
		page := 1

		for {
			labels, resp, err := c.api.ListRepoLabels(owner, repo, forgejo.ListLabelsOptions{
				ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
			})
			if err != nil {
				yield(nil, log.E("forge.ListRepoLabels", "failed to list repo labels", err))
				return
			}

			for _, label := range labels {
				if !yield(label, nil) {
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

// CreateRepoLabel creates a label on a repository.
// Usage: CreateRepoLabel(...)
func (c *Client) CreateRepoLabel(owner, repo string, opts forgejo.CreateLabelOption) (*forgejo.Label, error) {
	label, _, err := c.api.CreateLabel(owner, repo, opts)
	if err != nil {
		return nil, log.E("forge.CreateRepoLabel", "failed to create repo label", err)
	}

	return label, nil
}

// GetLabelByName retrieves a specific label by name from a repository.
// Usage: GetLabelByName(...)
func (c *Client) GetLabelByName(owner, repo, name string) (*forgejo.Label, error) {
	labels, err := c.ListRepoLabels(owner, repo)
	if err != nil {
		return nil, err
	}

	for _, l := range labels {
		if strings.EqualFold(l.Name, name) {
			return l, nil
		}
	}

	return nil, log.E("forge.GetLabelByName", "label "+name+" not found in "+owner+"/"+repo, nil)
}

// EnsureLabel checks if a label exists, and creates it if it doesn't.
// Usage: EnsureLabel(...)
func (c *Client) EnsureLabel(owner, repo, name, color string) (*forgejo.Label, error) {
	label, err := c.GetLabelByName(owner, repo, name)
	if err == nil {
		return label, nil
	}

	return c.CreateRepoLabel(owner, repo, forgejo.CreateLabelOption{
		Name:  name,
		Color: color,
	})
}

// AddIssueLabels adds labels to an issue.
// Usage: AddIssueLabels(...)
func (c *Client) AddIssueLabels(owner, repo string, number int64, labelIDs []int64) error {
	_, _, err := c.api.AddIssueLabels(owner, repo, number, forgejo.IssueLabelsOption{
		Labels: labelIDs,
	})
	if err != nil {
		return log.E("forge.AddIssueLabels", "failed to add labels to issue", err)
	}
	return nil
}

// RemoveIssueLabel removes a label from an issue.
// Usage: RemoveIssueLabel(...)
func (c *Client) RemoveIssueLabel(owner, repo string, number int64, labelID int64) error {
	_, err := c.api.DeleteIssueLabel(owner, repo, number, labelID)
	if err != nil {
		return log.E("forge.RemoveIssueLabel", "failed to remove label from issue", err)
	}
	return nil
}
