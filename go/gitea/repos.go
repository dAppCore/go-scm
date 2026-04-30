// SPDX-License-Identifier: EUPL-1.2

package gitea

import (
	// Note: iter.Seq2 is retained because the Gitea client exposes lazy paginated iterators directly.
	"iter"

	"code.gitea.io/sdk/gitea"
)

func (c *Client) CreateOrgRepo(org string, opts gitea.CreateRepoOption) (*gitea.Repository, error) {
	repo, _, err := c.api.CreateOrgRepo(org, opts)
	return repo, err
}

func (c *Client) DeleteRepo(owner, name string) error {
	_, err := c.api.DeleteRepo(owner, name)
	return err
}

func (c *Client) GetRepo(owner, name string) (*gitea.Repository, error) {
	repo, _, err := c.api.GetRepo(owner, name)
	return repo, err
}

func (c *Client) ListOrgRepos(org string) ([]*gitea.Repository, error) {
	var all []*gitea.Repository
	page := 1
	for {
		repos, resp, err := c.api.ListOrgRepos(org, gitea.ListOrgReposOptions{
			ListOptions: gitea.ListOptions{Page: page, PageSize: 50},
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

func (c *Client) ListOrgReposIter(org string) iter.Seq2[*gitea.Repository, error] {
	return func(yield func(*gitea.Repository, error) bool) {
		page := 1
		for {
			repos, resp, err := c.api.ListOrgRepos(org, gitea.ListOrgReposOptions{
				ListOptions: gitea.ListOptions{Page: page, PageSize: 50},
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

func (c *Client) ListUserRepos() ([]*gitea.Repository, error) {
	var all []*gitea.Repository
	page := 1
	for {
		repos, resp, err := c.api.ListMyRepos(gitea.ListReposOptions{
			ListOptions: gitea.ListOptions{Page: page, PageSize: 50},
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

func (c *Client) ListUserReposIter() iter.Seq2[*gitea.Repository, error] {
	return func(yield func(*gitea.Repository, error) bool) {
		page := 1
		for {
			repos, resp, err := c.api.ListMyRepos(gitea.ListReposOptions{
				ListOptions: gitea.ListOptions{Page: page, PageSize: 50},
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

func (c *Client) CreateMirror(owner, name, cloneURL, authToken string) (*gitea.Repository, error) {
	opts := gitea.MigrateRepoOption{
		RepoName:    name,
		RepoOwner:   owner,
		CloneAddr:   cloneURL,
		Mirror:      true,
		Description: "Mirror of " + cloneURL,
	}
	if authToken != "" {
		opts.AuthToken = authToken
	}
	repo, _, err := c.api.MigrateRepo(opts)
	return repo, err
}

func (c *Client) CreateMirrorFromService(owner, name, cloneURL string, service gitea.GitServiceType, authToken string) (*gitea.Repository, error) {
	opts := gitea.MigrateRepoOption{
		RepoName:    name,
		RepoOwner:   owner,
		CloneAddr:   cloneURL,
		Service:     service,
		Mirror:      true,
		Description: "Mirror of " + cloneURL,
	}
	if authToken != "" {
		opts.AuthToken = authToken
	}
	repo, _, err := c.api.MigrateRepo(opts)
	return repo, err
}
