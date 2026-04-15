// SPDX-License-Identifier: EUPL-1.2

package gitea

import (
	"iter"

	"code.gitea.io/sdk/gitea"
)

func (c *Client) CreateOrgRepo(org string, opts gitea.CreateRepoOption) (*gitea.Repository, error) { return nil, nil }
func (c *Client) DeleteRepo(owner, name string) error                                              { return nil }
func (c *Client) GetRepo(owner, name string) (*gitea.Repository, error)                            { return nil, nil }
func (c *Client) ListOrgRepos(org string) ([]*gitea.Repository, error)                             { return nil, nil }
func (c *Client) ListUserRepos() ([]*gitea.Repository, error)                                      { return nil, nil }
func (c *Client) CreateMirror(owner, name, cloneURL, authToken string) (*gitea.Repository, error)  { return nil, nil }
func (c *Client) ListOrgReposIter(org string) iter.Seq2[*gitea.Repository, error] {
	return func(yield func(*gitea.Repository, error) bool) {}
}
func (c *Client) ListUserReposIter() iter.Seq2[*gitea.Repository, error] {
	return func(yield func(*gitea.Repository, error) bool) {}
}
