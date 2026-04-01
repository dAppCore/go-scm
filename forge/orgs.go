// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	"iter"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"

	"dappco.re/go/core/log"
)

// ListMyOrgs returns all organisations for the authenticated user.
// Usage: ListMyOrgs(...)
func (c *Client) ListMyOrgs() ([]*forgejo.Organization, error) {
	var all []*forgejo.Organization
	page := 1

	for {
		orgs, resp, err := c.api.ListMyOrgs(forgejo.ListOrgsOptions{
			ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
		})
		if err != nil {
			return nil, log.E("forge.ListMyOrgs", "failed to list orgs", err)
		}

		all = append(all, orgs...)

		if resp == nil || page >= resp.LastPage {
			break
		}
		page++
	}

	return all, nil
}

// ListMyOrgsIter returns an iterator over organisations for the authenticated user.
// Usage: ListMyOrgsIter(...)
func (c *Client) ListMyOrgsIter() iter.Seq2[*forgejo.Organization, error] {
	return func(yield func(*forgejo.Organization, error) bool) {
		page := 1

		for {
			orgs, resp, err := c.api.ListMyOrgs(forgejo.ListOrgsOptions{
				ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
			})
			if err != nil {
				yield(nil, log.E("forge.ListMyOrgs", "failed to list orgs", err))
				return
			}

			for _, org := range orgs {
				if !yield(org, nil) {
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

// GetOrg returns a single organisation by name.
// Usage: GetOrg(...)
func (c *Client) GetOrg(name string) (*forgejo.Organization, error) {
	org, _, err := c.api.GetOrg(name)
	if err != nil {
		return nil, log.E("forge.GetOrg", "failed to get org", err)
	}

	return org, nil
}

// CreateOrg creates a new organisation.
// Usage: CreateOrg(...)
func (c *Client) CreateOrg(opts forgejo.CreateOrgOption) (*forgejo.Organization, error) {
	org, _, err := c.api.CreateOrg(opts)
	if err != nil {
		return nil, log.E("forge.CreateOrg", "failed to create org", err)
	}

	return org, nil
}
