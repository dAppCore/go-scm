// SPDX-License-Identifier: EUPL-1.2

package git

import (
	"context"
	"iter"

	core "dappco.re/go/core"
)

type ServiceOptions struct {
	WorkDir string
}

type QueryAheadRepos struct{}
type QueryDirtyRepos struct{}
type QueryStatus struct {
	Paths []string
	Names map[string]string
}
type TaskPull struct {
	Path string
	Name string
}
type TaskPush struct {
	Path string
	Name string
}
type TaskPushMultiple struct {
	Paths []string
	Names map[string]string
}

type Service struct {
	*core.ServiceRuntime[ServiceOptions]
	lastStatus []RepoStatus
}

func NewService(opts ServiceOptions) func(*core.Core) (any, error) {
	return func(c *core.Core) (any, error) {
		s := &Service{ServiceRuntime: core.NewServiceRuntime(c, opts)}
		if opts.WorkDir != "" {
			s.lastStatus = Status(context.Background(), StatusOptions{Paths: []string{opts.WorkDir}})
		}
		return s, nil
	}
}

func (s *Service) Status() []RepoStatus { return append([]RepoStatus(nil), s.lastStatus...) }

func (s *Service) StatusIter() iter.Seq[RepoStatus] {
	status := s.Status()
	return func(yield func(RepoStatus) bool) {
		for _, st := range status {
			if !yield(st) {
				return
			}
		}
	}
}

func (s *Service) DirtyRepos() []RepoStatus {
	var out []RepoStatus
	for _, st := range s.Status() {
		if st.IsDirty() {
			out = append(out, st)
		}
	}
	return out
}

func (s *Service) DirtyReposIter() iter.Seq[RepoStatus] {
	return func(yield func(RepoStatus) bool) {
		for _, st := range s.DirtyRepos() {
			if !yield(st) {
				return
			}
		}
	}
}

func (s *Service) AheadRepos() []RepoStatus {
	var out []RepoStatus
	for _, st := range s.Status() {
		if st.HasUnpushed() {
			out = append(out, st)
		}
	}
	return out
}

func (s *Service) AheadReposIter() iter.Seq[RepoStatus] {
	return func(yield func(RepoStatus) bool) {
		for _, st := range s.AheadRepos() {
			if !yield(st) {
				return
			}
		}
	}
}

func (s *Service) OnStartup(ctx context.Context) error {
	if s == nil || s.Options().WorkDir == "" {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.lastStatus = Status(ctx, StatusOptions{Paths: []string{s.Options().WorkDir}})
	return nil
}
