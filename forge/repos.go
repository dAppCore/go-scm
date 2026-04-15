// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	"iter"

	"codeberg.org/forgejo/go-sdk/forgejo"
)

func (c *Client) CreateOrgRepo(org string, opts forgejo.CreateRepoOption) (*forgejo.Repository, error) {
	repo, _, err := c.api.CreateOrgRepo(org, opts)
	return repo, err
}

func (c *Client) DeleteRepo(owner, name string) error {
	_, err := c.api.DeleteRepo(owner, name)
	return err
}

func (c *Client) GetRepo(owner, name string) (*forgejo.Repository, error) {
	repo, _, err := c.api.GetRepo(owner, name)
	return repo, err
}

func (c *Client) ListOrgRepos(org string) ([]*forgejo.Repository, error) {
	var all []*forgejo.Repository
	page := 1
	for {
		repos, resp, err := c.api.ListOrgRepos(org, forgejo.ListOrgReposOptions{
			ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
		})
		if err != nil {
			return nil, err
		}
		all = append(all, repos...)
		if resp == nil || page >= resp.LastPage {
			break
		}
		page++
	}
	return all, nil
}

func (c *Client) ListOrgReposIter(org string) iter.Seq2[*forgejo.Repository, error] {
	return func(yield func(*forgejo.Repository, error) bool) {
		page := 1
		for {
			repos, resp, err := c.api.ListOrgRepos(org, forgejo.ListOrgReposOptions{
				ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
			})
			if err != nil {
				yield(nil, err)
				return
			}
			for _, repo := range repos {
				if !yield(repo, nil) {
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

func (c *Client) ListUserRepos() ([]*forgejo.Repository, error) {
	var all []*forgejo.Repository
	page := 1
	for {
		repos, resp, err := c.api.ListMyRepos(forgejo.ListReposOptions{
			ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
		})
		if err != nil {
			return nil, err
		}
		all = append(all, repos...)
		if resp == nil || page >= resp.LastPage {
			break
		}
		page++
	}
	return all, nil
}

func (c *Client) ListUserReposIter() iter.Seq2[*forgejo.Repository, error] {
	return func(yield func(*forgejo.Repository, error) bool) {
		page := 1
		for {
			repos, resp, err := c.api.ListMyRepos(forgejo.ListReposOptions{
				ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
			})
			if err != nil {
				yield(nil, err)
				return
			}
			for _, repo := range repos {
				if !yield(repo, nil) {
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

func (c *Client) MigrateRepo(opts forgejo.MigrateRepoOption) (*forgejo.Repository, error) {
	repo, _, err := c.api.MigrateRepo(opts)
	return repo, err
}

