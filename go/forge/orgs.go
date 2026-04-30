// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	// Note: iter.Seq2 is retained because the forge client exposes lazy paginated iterators directly.
	"iter"

	"codeberg.org/forgejo/go-sdk/forgejo"
)

func (c *Client) CreateOrg(opts forgejo.CreateOrgOption) (*forgejo.Organization, error) {
	org, _, err := c.api.CreateOrg(opts)
	return org, err
}

func (c *Client) GetOrg(name string) (*forgejo.Organization, error) {
	org, _, err := c.api.GetOrg(name)
	return org, err
}

func (c *Client) ListMyOrgs() ([]*forgejo.Organization, error) {
	return collectForgePages(func(page int) ([]*forgejo.Organization, *forgeResponse, error) {
		return c.api.ListMyOrgs(forgejo.ListOrgsOptions{
			ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
		})
	})
}

func (c *Client) ListMyOrgsIter() iter.Seq2[*forgejo.Organization, error] {
	return func(yield func(*forgejo.Organization, error) bool) {
		yieldForgePages(yield, func(page int) ([]*forgejo.Organization, *forgeResponse, error) {
			return c.api.ListMyOrgs(forgejo.ListOrgsOptions{
				ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
			})
		})
	}
}
