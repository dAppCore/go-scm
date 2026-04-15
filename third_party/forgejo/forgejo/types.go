// SPDX-License-Identifier: EUPL-1.2

package forgejo

import "time"

type Client struct{}

type User struct{}
type Repository struct{}
type PullRequest struct{}
type Issue struct{}
type Comment struct{}
type Label struct{}
type Organization struct{}
type Hook struct{}
type CombinedStatus struct{}
type PullReview struct{}

type CreateIssueOption struct{}
type EditIssueOption struct{}
type CreateOrgOption struct{}
type CreateRepoOption struct{}
type CreatePullRequestOption struct{}
type CreateLabelOption struct{}
type CreateHookOption struct{}
type MigrateRepoOption struct{}

type PullRequestMeta struct {
	UpdatedAt time.Time
}
