// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	"iter"

	"codeberg.org/forgejo/go-sdk/forgejo"
)

func (c *Client) CreateRepoWebhook(owner, repo string, opts forgejo.CreateHookOption) (*forgejo.Hook, error) {
	hook, _, err := c.api.CreateRepoHook(owner, repo, opts)
	return hook, err
}

func (c *Client) ListRepoWebhooks(owner, repo string) ([]*forgejo.Hook, error) {
	var all []*forgejo.Hook
	page := 1
	for {
		hooks, resp, err := c.api.ListRepoHooks(owner, repo, forgejo.ListHooksOptions{
			ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
		})
		if err != nil {
			return nil, err
		}
		all = append(all, hooks...)
		if resp == nil || page >= resp.LastPage {
			break
		}
		page++
	}
	return all, nil
}

func (c *Client) ListRepoWebhooksIter(owner, repo string) iter.Seq2[*forgejo.Hook, error] {
	return func(yield func(*forgejo.Hook, error) bool) {
		page := 1
		for {
			hooks, resp, err := c.api.ListRepoHooks(owner, repo, forgejo.ListHooksOptions{
				ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
			})
			if err != nil {
				yield(nil, err)
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

