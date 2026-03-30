// SPDX-License-Identifier: EUPL-1.2

package git

import (
	"context"
	"iter"
	"slices"

	"dappco.re/go/core"
)

// Queries for git service

// QueryStatus requests git status for paths.
type QueryStatus struct {
	Paths []string
	Names map[string]string
}

// QueryDirtyRepos requests repos with uncommitted changes.
type QueryDirtyRepos struct{}

// QueryAheadRepos requests repos with unpushed commits.
type QueryAheadRepos struct{}

// Tasks for git service

// TaskPush requests git push for a path.
type TaskPush struct {
	Path string
	Name string
}

// TaskPull requests git pull for a path.
type TaskPull struct {
	Path string
	Name string
}

// TaskPushMultiple requests git push for multiple paths.
type TaskPushMultiple struct {
	Paths []string
	Names map[string]string
}

// ServiceOptions for configuring the git service.
type ServiceOptions struct {
	WorkDir string
}

// Service provides git operations as a Core service.
type Service struct {
	*core.ServiceRuntime[ServiceOptions]
	lastStatus []RepoStatus
}

// NewService creates a git service factory.
// Usage: NewService(...)
func NewService(opts ServiceOptions) func(*core.Core) (any, error) {
	return func(c *core.Core) (any, error) {
		return &Service{
			ServiceRuntime: core.NewServiceRuntime(c, opts),
		}, nil
	}
}

// OnStartup registers query and task handlers.
// Usage: OnStartup(...)
func (s *Service) OnStartup(ctx context.Context) error {
	s.Core().RegisterQuery(s.handleQuery)
	s.Core().RegisterTask(s.handleTask)
	return nil
}

func (s *Service) handleQuery(c *core.Core, q core.Query) core.Result {
	switch m := q.(type) {
	case QueryStatus:
		statuses := Status(context.Background(), StatusOptions(m))
		s.lastStatus = statuses
		return core.Result{Value: statuses, OK: true}

	case QueryDirtyRepos:
		return core.Result{Value: s.DirtyRepos(), OK: true}

	case QueryAheadRepos:
		return core.Result{Value: s.AheadRepos(), OK: true}
	}
	return core.Result{}
}

func (s *Service) handleTask(c *core.Core, t core.Task) core.Result {
	switch m := t.(type) {
	case TaskPush:
		return core.Result{}.Result(nil, Push(context.Background(), m.Path))

	case TaskPull:
		return core.Result{}.Result(nil, Pull(context.Background(), m.Path))

	case TaskPushMultiple:
		results := PushMultiple(context.Background(), m.Paths, m.Names)
		return core.Result{Value: results, OK: true}
	}
	return core.Result{}
}

// Status returns last status result.
// Usage: Status(...)
func (s *Service) Status() []RepoStatus { return s.lastStatus }

// StatusIter returns an iterator over last status result.
// Usage: StatusIter(...)
func (s *Service) StatusIter() iter.Seq[RepoStatus] {
	return slices.Values(s.lastStatus)
}

// DirtyRepos returns repos with uncommitted changes.
// Usage: DirtyRepos(...)
func (s *Service) DirtyRepos() []RepoStatus {
	var dirty []RepoStatus
	for _, st := range s.lastStatus {
		if st.Error == nil && st.IsDirty() {
			dirty = append(dirty, st)
		}
	}
	return dirty
}

// DirtyReposIter returns an iterator over repos with uncommitted changes.
// Usage: DirtyReposIter(...)
func (s *Service) DirtyReposIter() iter.Seq[RepoStatus] {
	return func(yield func(RepoStatus) bool) {
		for _, st := range s.lastStatus {
			if st.Error == nil && st.IsDirty() {
				if !yield(st) {
					return
				}
			}
		}
	}
}

// AheadRepos returns repos with unpushed commits.
// Usage: AheadRepos(...)
func (s *Service) AheadRepos() []RepoStatus {
	var ahead []RepoStatus
	for _, st := range s.lastStatus {
		if st.Error == nil && st.HasUnpushed() {
			ahead = append(ahead, st)
		}
	}
	return ahead
}

// AheadReposIter returns an iterator over repos with unpushed commits.
// Usage: AheadReposIter(...)
func (s *Service) AheadReposIter() iter.Seq[RepoStatus] {
	return func(yield func(RepoStatus) bool) {
		for _, st := range s.lastStatus {
			if st.Error == nil && st.HasUnpushed() {
				if !yield(st) {
					return
				}
			}
		}
	}
}
