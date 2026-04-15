// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	"iter"

	"codeberg.org/forgejo/go-sdk/forgejo"
)

func (c *Client) CreateOrgRepo(org string, opts forgejo.CreateRepoOption) (*forgejo.Repository, error) { return nil, nil }
func (c *Client) DeleteRepo(owner, name string) error                                                { return nil }
func (c *Client) GetRepo(owner, name string) (*forgejo.Repository, error)                            { return nil, nil }
func (c *Client) ListOrgRepos(org string) ([]*forgejo.Repository, error)                              { return nil, nil }
func (c *Client) ListUserRepos() ([]*forgejo.Repository, error)                                      { return nil, nil }
func (c *Client) MigrateRepo(opts forgejo.MigrateRepoOption) (*forgejo.Repository, error)             { return nil, nil }
func (c *Client) ListOrgReposIter(org string) iter.Seq2[*forgejo.Repository, error] {
	return func(yield func(*forgejo.Repository, error) bool) {}
}
func (c *Client) ListUserReposIter() iter.Seq2[*forgejo.Repository, error] {
	return func(yield func(*forgejo.Repository, error) bool) {}
}
