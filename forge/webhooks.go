// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	"iter"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"

	"dappco.re/go/core/log"
)

// CreateRepoWebhook creates a webhook on a repository.
// Usage: CreateRepoWebhook(...)
func (c *Client) CreateRepoWebhook(owner, repo string, opts forgejo.CreateHookOption) (*forgejo.Hook, error) {
	hook, _, err := c.api.CreateRepoHook(owner, repo, opts)
	if err != nil {
		return nil, log.E("forge.CreateRepoWebhook", "failed to create repo webhook", err)
	}

	return hook, nil
}

// ListRepoWebhooks returns all webhooks for a repository.
// Usage: ListRepoWebhooks(...)
func (c *Client) ListRepoWebhooks(owner, repo string) ([]*forgejo.Hook, error) {
	var all []*forgejo.Hook
	page := 1

	for {
		hooks, resp, err := c.api.ListRepoHooks(owner, repo, forgejo.ListHooksOptions{
			ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
		})
		if err != nil {
			return nil, log.E("forge.ListRepoWebhooks", "failed to list repo webhooks", err)
		}

		all = append(all, hooks...)

		if resp == nil || page >= resp.LastPage {
			break
		}
		page++
	}

	return all, nil
}

// ListRepoWebhooksIter returns an iterator over webhooks for a repository.
// Usage: ListRepoWebhooksIter(...)
func (c *Client) ListRepoWebhooksIter(owner, repo string) iter.Seq2[*forgejo.Hook, error] {
	return func(yield func(*forgejo.Hook, error) bool) {
		page := 1
		for {
			hooks, resp, err := c.api.ListRepoHooks(owner, repo, forgejo.ListHooksOptions{
				ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
			})
			if err != nil {
				yield(nil, log.E("forge.ListRepoWebhooks", "failed to list repo webhooks", err))
				return
			}
			for _, hook := range hooks {
				if !yield(hook, nil) {
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
