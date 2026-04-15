// SPDX-License-Identifier: EUPL-1.2

package forge

import "codeberg.org/forgejo/go-sdk/forgejo"

func (c *Client) CreateRepoWebhook(owner, repo string, opts forgejo.CreateHookOption) (*forgejo.Hook, error) {
	return nil, nil
}
func (c *Client) ListRepoWebhooks(owner, repo string) ([]*forgejo.Hook, error) { return nil, nil }
