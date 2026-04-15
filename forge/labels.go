// SPDX-License-Identifier: EUPL-1.2

package forge

import "codeberg.org/forgejo/go-sdk/forgejo"

func (c *Client) CreateRepoLabel(owner, repo string, opts forgejo.CreateLabelOption) (*forgejo.Label, error) {
	return nil, nil
}
func (c *Client) EnsureLabel(owner, repo, name, color string) (*forgejo.Label, error) { return nil, nil }
func (c *Client) GetLabelByName(owner, repo, name string) (*forgejo.Label, error)     { return nil, nil }
func (c *Client) ListRepoLabels(owner, repo string) ([]*forgejo.Label, error)         { return nil, nil }
func (c *Client) ListOrgLabels(org string) ([]*forgejo.Label, error)                  { return nil, nil }
func (c *Client) AddIssueLabels(owner, repo string, number int64, labelIDs []int64) error { return nil }
func (c *Client) RemoveIssueLabel(owner, repo string, number int64, labelID int64) error { return nil }
