// SPDX-License-Identifier: EUPL-1.2

package forgejo

import (
	"net/http"

	"code.gitea.io/sdk/gitea"
)

type Client struct {
	api *gitea.Client
}

type ClientOption = gitea.ClientOption
type StateType = gitea.StateType
type IssueType = gitea.IssueType
type MergeStyle = gitea.MergeStyle

const (
	StateOpen   = gitea.StateOpen
	StateClosed = gitea.StateClosed
	StateAll    = gitea.StateAll

	IssueTypeIssue = gitea.IssueTypeIssue

	MergeStyleMerge  = gitea.MergeStyleMerge
	MergeStyleRebase = gitea.MergeStyleRebase
	MergeStyleSquash = gitea.MergeStyleSquash
)

type ListOptions = gitea.ListOptions
type ListReposOptions = gitea.ListReposOptions
type ListOrgsOptions = gitea.ListOrgsOptions
type ListOrgReposOptions = gitea.ListOrgReposOptions
type ListLabelsOptions = gitea.ListLabelsOptions
type ListHooksOptions = gitea.ListHooksOptions
type ListIssueOption = gitea.ListIssueOption
type ListIssueCommentOptions = gitea.ListIssueCommentOptions
type ListPullRequestsOptions = gitea.ListPullRequestsOptions
type ListPullReviewsOptions = gitea.ListPullReviewsOptions
type CreateIssueOption = gitea.CreateIssueOption
type EditIssueOption = gitea.EditIssueOption
type CreateOrgOption = gitea.CreateOrgOption
type CreateRepoOption = gitea.CreateRepoOption
type CreatePullRequestOption = gitea.CreatePullRequestOption
type CreateLabelOption = gitea.CreateLabelOption
type CreateHookOption = gitea.CreateHookOption
type CreateIssueCommentOption = gitea.CreateIssueCommentOption
type IssueLabelsOption = gitea.IssueLabelsOption
type CreateForkOption = gitea.CreateForkOption
type MigrateRepoOption = gitea.MigrateRepoOption
type DismissPullReviewOptions = gitea.DismissPullReviewOptions
type MergePullRequestOption = gitea.MergePullRequestOption
type User = gitea.User
type Repository = gitea.Repository
type PullRequest = gitea.PullRequest
type Issue = gitea.Issue
type Comment = gitea.Comment
type Label = gitea.Label
type Organization = gitea.Organization
type Hook = gitea.Hook
type CombinedStatus = gitea.CombinedStatus
type PullReview = gitea.PullReview

func SetToken(token string) ClientOption { return gitea.SetToken(token) }

func SetHTTPClient(client *http.Client) ClientOption { return gitea.SetHTTPClient(client) }

func NewClient(url string, options ...ClientOption) (*Client, error) {
	api, err := gitea.NewClient(url, options...)
	if err != nil {
		return nil, err
	}
	return &Client{api: api}, nil
}

func (c *Client) API() *gitea.Client {
	if c == nil {
		return nil
	}
	return c.api
}

func (c *Client) GetMyUserInfo() (*gitea.User, *gitea.Response, error) {
	return c.api.GetMyUserInfo()
}

func (c *Client) CreateFork(owner, repo string, opts gitea.CreateForkOption) (*gitea.Repository, *gitea.Response, error) {
	return c.api.CreateFork(owner, repo, opts)
}

func (c *Client) CreatePullRequest(owner, repo string, opts gitea.CreatePullRequestOption) (*gitea.PullRequest, *gitea.Response, error) {
	return c.api.CreatePullRequest(owner, repo, opts)
}

func (c *Client) GetIssue(owner, repo string, number int64) (*gitea.Issue, *gitea.Response, error) {
	return c.api.GetIssue(owner, repo, number)
}

func (c *Client) EditIssue(owner, repo string, number int64, opts gitea.EditIssueOption) (*gitea.Issue, *gitea.Response, error) {
	return c.api.EditIssue(owner, repo, number, opts)
}

func (c *Client) CreateIssue(owner, repo string, opts gitea.CreateIssueOption) (*gitea.Issue, *gitea.Response, error) {
	return c.api.CreateIssue(owner, repo, opts)
}

func (c *Client) CreateIssueComment(owner, repo string, issue int64, opts gitea.CreateIssueCommentOption) (*gitea.Comment, *gitea.Response, error) {
	return c.api.CreateIssueComment(owner, repo, issue, opts)
}

func (c *Client) ListRepoIssues(owner, repo string, opts gitea.ListIssueOption) ([]*gitea.Issue, *gitea.Response, error) {
	return c.api.ListRepoIssues(owner, repo, opts)
}

func (c *Client) ListRepoPullRequests(owner, repo string, opts gitea.ListPullRequestsOptions) ([]*gitea.PullRequest, *gitea.Response, error) {
	return c.api.ListRepoPullRequests(owner, repo, opts)
}

func (c *Client) ListIssueComments(owner, repo string, number int64, opts gitea.ListIssueCommentOptions) ([]*gitea.Comment, *gitea.Response, error) {
	return c.api.ListIssueComments(owner, repo, number, opts)
}

func (c *Client) GetIssueLabels(owner, repo string, number int64, opts gitea.ListLabelsOptions) ([]*gitea.Label, *gitea.Response, error) {
	return c.api.GetIssueLabels(owner, repo, number, opts)
}

func (c *Client) AddIssueLabels(owner, repo string, number int64, opts gitea.IssueLabelsOption) ([]*gitea.Label, *gitea.Response, error) {
	return c.api.AddIssueLabels(owner, repo, number, opts)
}

func (c *Client) DeleteIssueLabel(owner, repo string, number, labelID int64) (*gitea.Response, error) {
	return c.api.DeleteIssueLabel(owner, repo, number, labelID)
}

func (c *Client) GetPullRequest(owner, repo string, number int64) (*gitea.PullRequest, *gitea.Response, error) {
	return c.api.GetPullRequest(owner, repo, number)
}

func (c *Client) ListPullReviews(owner, repo string, number int64, opts gitea.ListPullReviewsOptions) ([]*gitea.PullReview, *gitea.Response, error) {
	return c.api.ListPullReviews(owner, repo, number, opts)
}

func (c *Client) GetCombinedStatus(owner, repo, ref string) (*gitea.CombinedStatus, *gitea.Response, error) {
	return c.api.GetCombinedStatus(owner, repo, ref)
}

func (c *Client) DismissPullReview(owner, repo string, number, reviewID int64, opts gitea.DismissPullReviewOptions) (*gitea.Response, error) {
	return c.api.DismissPullReview(owner, repo, number, reviewID, opts)
}

func (c *Client) UnDismissPullReview(owner, repo string, number, reviewID int64) (*gitea.Response, error) {
	return c.api.UnDismissPullReview(owner, repo, number, reviewID)
}

func (c *Client) CreateOrgRepo(org string, opts gitea.CreateRepoOption) (*gitea.Repository, *gitea.Response, error) {
	return c.api.CreateOrgRepo(org, opts)
}

func (c *Client) DeleteRepo(owner, name string) (*gitea.Response, error) {
	return c.api.DeleteRepo(owner, name)
}

func (c *Client) GetRepo(owner, name string) (*gitea.Repository, *gitea.Response, error) {
	return c.api.GetRepo(owner, name)
}

func (c *Client) ListOrgRepos(org string, opts gitea.ListOrgReposOptions) ([]*gitea.Repository, *gitea.Response, error) {
	return c.api.ListOrgRepos(org, opts)
}

func (c *Client) ListMyRepos(opts gitea.ListReposOptions) ([]*gitea.Repository, *gitea.Response, error) {
	return c.api.ListMyRepos(opts)
}

func (c *Client) MigrateRepo(opts gitea.MigrateRepoOption) (*gitea.Repository, *gitea.Response, error) {
	return c.api.MigrateRepo(opts)
}

func (c *Client) ListOrgReposIter(org string) func(func(*gitea.Repository, error) bool) {
	return func(yield func(*gitea.Repository, error) bool) {
		page := 1
		for {
			repos, resp, err := c.ListOrgRepos(org, gitea.ListOrgReposOptions{
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

func (c *Client) ListMyReposIter() func(func(*gitea.Repository, error) bool) {
	return func(yield func(*gitea.Repository, error) bool) {
		page := 1
		for {
			repos, resp, err := c.ListMyRepos(gitea.ListReposOptions{
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

func (c *Client) ListRepoLabels(owner, repo string, opts gitea.ListLabelsOptions) ([]*gitea.Label, *gitea.Response, error) {
	return c.api.ListRepoLabels(owner, repo, opts)
}

func (c *Client) CreateLabel(owner, repo string, opts gitea.CreateLabelOption) (*gitea.Label, *gitea.Response, error) {
	return c.api.CreateLabel(owner, repo, opts)
}

func (c *Client) ListMyOrgs(opts gitea.ListOrgsOptions) ([]*gitea.Organization, *gitea.Response, error) {
	return c.api.ListMyOrgs(opts)
}

func (c *Client) GetOrg(name string) (*gitea.Organization, *gitea.Response, error) {
	return c.api.GetOrg(name)
}

func (c *Client) CreateOrg(opts gitea.CreateOrgOption) (*gitea.Organization, *gitea.Response, error) {
	return c.api.CreateOrg(opts)
}

func (c *Client) CreateRepoHook(owner, repo string, opts gitea.CreateHookOption) (*gitea.Hook, *gitea.Response, error) {
	return c.api.CreateRepoHook(owner, repo, opts)
}

func (c *Client) ListRepoHooks(owner, repo string, opts gitea.ListHooksOptions) ([]*gitea.Hook, *gitea.Response, error) {
	return c.api.ListRepoHooks(owner, repo, opts)
}

func (c *Client) MergePullRequest(owner, repo string, index int64, opts gitea.MergePullRequestOption) (bool, *gitea.Response, error) {
	return c.api.MergePullRequest(owner, repo, index, opts)
}
