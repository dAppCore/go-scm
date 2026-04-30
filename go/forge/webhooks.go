// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	// Note: iter.Seq2 is retained because the forge client exposes lazy paginated iterators directly.
	"iter"

	"codeberg.org/forgejo/go-sdk/forgejo"
)

func (c *Client) CreateRepoWebhook(owner, repo string, opts forgejo.CreateHookOption) (*forgejo.Hook, error) {
	hook, _, err := c.api.CreateRepoHook(owner, repo, opts)
	return hook, err
}

func (c *Client) ListRepoWebhooks(owner, repo string) ([]*forgejo.Hook, error) {
	return collectForgePages(func(page int) ([]*forgejo.Hook, *forgeResponse, error) {
		return c.api.ListRepoHooks(owner, repo, forgejo.ListHooksOptions{
			ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
		})
	})
}

func (c *Client) ListRepoWebhooksIter(owner, repo string) iter.Seq2[*forgejo.Hook, error] {
	return func(yield func(*forgejo.Hook, error) bool) {
		yieldForgePages(yield, func(page int) ([]*forgejo.Hook, *forgeResponse, error) {
			return c.api.ListRepoHooks(owner, repo, forgejo.ListHooksOptions{
				ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
			})
		})
	}
}
