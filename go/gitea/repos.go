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
	return collectGiteaPages(func(page int) ([]*gitea.Repository, *gitea.Response, error) {
		return c.api.ListOrgRepos(org, gitea.ListOrgReposOptions{
			ListOptions: gitea.ListOptions{Page: page, PageSize: 50},
		})
	})
}

func (c *Client) ListOrgReposIter(org string) iter.Seq2[*gitea.Repository, error] {
	return func(yield func(*gitea.Repository, error) bool) {
		yieldGiteaPages(yield, func(page int) ([]*gitea.Repository, *gitea.Response, error) {
			return c.api.ListOrgRepos(org, gitea.ListOrgReposOptions{
				ListOptions: gitea.ListOptions{Page: page, PageSize: 50},
			})
		})
	}
}

func (c *Client) ListUserRepos() ([]*gitea.Repository, error) {
	return collectGiteaPages(func(page int) ([]*gitea.Repository, *gitea.Response, error) {
		return c.api.ListMyRepos(gitea.ListReposOptions{
			ListOptions: gitea.ListOptions{Page: page, PageSize: 50},
		})
	})
}

func (c *Client) ListUserReposIter() iter.Seq2[*gitea.Repository, error] {
	return func(yield func(*gitea.Repository, error) bool) {
		yieldGiteaPages(yield, func(page int) ([]*gitea.Repository, *gitea.Response, error) {
			return c.api.ListMyRepos(gitea.ListReposOptions{
				ListOptions: gitea.ListOptions{Page: page, PageSize: 50},
			})
		})
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
