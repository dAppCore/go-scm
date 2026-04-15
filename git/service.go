// SPDX-License-Identifier: EUPL-1.2

package git

import (
	"context"
	"errors"
	"iter"
	"path/filepath"
	"strings"
	"sync"

	core "dappco.re/go/core"
)

type ServiceOptions struct {
	WorkDir string
}

type QueryAheadRepos struct{}
type QueryDirtyRepos struct{}
type QueryBehindRepos struct{}
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
type TaskPullMultiple struct {
	Paths []string
	Names map[string]string
}

type Service struct {
	*core.ServiceRuntime[ServiceOptions]
	mu         sync.RWMutex
	lastStatus []RepoStatus
}

func NewService(opts ServiceOptions) func(*core.Core) core.Result {
	return func(c *core.Core) core.Result {
		return core.Result{Value: &Service{ServiceRuntime: core.NewServiceRuntime(c, opts)}, OK: true}
	}
}

func (s *Service) Status() []RepoStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]RepoStatus(nil), s.lastStatus...)
}

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

func (s *Service) BehindRepos() []RepoStatus {
	var out []RepoStatus
	for _, st := range s.Status() {
		if st.HasUnpulled() {
			out = append(out, st)
		}
	}
	return out
}

func (s *Service) BehindReposIter() iter.Seq[RepoStatus] {
	return func(yield func(RepoStatus) bool) {
		for _, st := range s.BehindRepos() {
			if !yield(st) {
				return
			}
		}
	}
}

func (s *Service) OnStartup(ctx context.Context) core.Result {
	if s == nil {
		return core.Result{OK: true}
	}
	if err := ctx.Err(); err != nil {
		return core.Result{Value: err, OK: false}
	}

	c := s.Core()
	if c == nil {
		return core.Result{Value: errors.New("git.Service.OnStartup: core is required"), OK: false}
	}

	c.RegisterQuery(s.handleQuery)
	c.RegisterAction(s.handleTaskMessage)

	c.Action("git.push", func(ctx context.Context, opts core.Options) core.Result {
		return s.runPush(ctx, opts.String("path"))
	})
	c.Action("git.pull", func(ctx context.Context, opts core.Options) core.Result {
		return s.runPull(ctx, opts.String("path"))
	})
	c.Action("git.push-multiple", func(ctx context.Context, opts core.Options) core.Result {
		paths, _ := opts.Get("paths").Value.([]string)
		names, _ := opts.Get("names").Value.(map[string]string)
		return s.runPushMultiple(ctx, paths, names)
	})
	c.Action("git.pull-multiple", func(ctx context.Context, opts core.Options) core.Result {
		paths, _ := opts.Get("paths").Value.([]string)
		names, _ := opts.Get("names").Value.(map[string]string)
		return s.runPullMultiple(ctx, paths, names)
	})

	if workDir := s.Options().WorkDir; workDir != "" {
		s.mu.Lock()
		s.lastStatus = Status(ctx, StatusOptions{Paths: []string{workDir}})
		s.mu.Unlock()
	}
	return core.Result{OK: true}
}

func (s *Service) handleQuery(c *core.Core, q core.Query) core.Result {
	if s == nil {
		return core.Result{}
	}
	ctx := c.Context()
	switch m := q.(type) {
	case QueryStatus:
		if err := s.validatePaths(m.Paths); err != nil {
			return c.LogError(err, "git.handleQuery", "path validation failed")
		}
		statuses := Status(ctx, StatusOptions(m))
		s.mu.Lock()
		s.lastStatus = statuses
		s.mu.Unlock()
		return core.Result{Value: statuses, OK: true}
	case QueryDirtyRepos:
		return core.Result{Value: s.DirtyRepos(), OK: true}
	case QueryAheadRepos:
		return core.Result{Value: s.AheadRepos(), OK: true}
	case QueryBehindRepos:
		return core.Result{Value: s.BehindRepos(), OK: true}
	default:
		return core.Result{}
	}
}

func (s *Service) handleTaskMessage(c *core.Core, msg core.Message) core.Result {
	switch m := msg.(type) {
	case TaskPush:
		return s.runPush(c.Context(), m.Path)
	case TaskPull:
		return s.runPull(c.Context(), m.Path)
	case TaskPushMultiple:
		return s.runPushMultiple(c.Context(), m.Paths, m.Names)
	case TaskPullMultiple:
		return s.runPullMultiple(c.Context(), m.Paths, m.Names)
	default:
		return core.Result{}
	}
}

func (s *Service) runPush(ctx context.Context, path string) core.Result {
	if err := s.validatePath(path); err != nil {
		return s.Core().LogError(err, "git.push", "path validation failed")
	}
	if err := Push(ctx, path); err != nil {
		return s.Core().LogError(err, "git.push", "push failed")
	}
	return core.Result{OK: true}
}

func (s *Service) runPull(ctx context.Context, path string) core.Result {
	if err := s.validatePath(path); err != nil {
		return s.Core().LogError(err, "git.pull", "path validation failed")
	}
	if err := Pull(ctx, path); err != nil {
		return s.Core().LogError(err, "git.pull", "pull failed")
	}
	return core.Result{OK: true}
}

func (s *Service) runPushMultiple(ctx context.Context, paths []string, names map[string]string) core.Result {
	if err := s.validatePaths(paths); err != nil {
		return s.Core().LogError(err, "git.push-multiple", "path validation failed")
	}
	results := PushMultiple(ctx, paths, names)
	for _, result := range results {
		if result.Error != nil {
			return core.Result{Value: results, OK: false}
		}
	}
	return core.Result{Value: results, OK: true}
}

func (s *Service) runPullMultiple(ctx context.Context, paths []string, names map[string]string) core.Result {
	if err := s.validatePaths(paths); err != nil {
		return s.Core().LogError(err, "git.pull-multiple", "path validation failed")
	}
	results := PullMultiple(ctx, paths, names)
	for _, result := range results {
		if result.Error != nil {
			return core.Result{Value: results, OK: false}
		}
	}
	return core.Result{Value: results, OK: true}
}

func (s *Service) validatePath(path string) error {
	if filepath.IsAbs(path) {
		return nil
	}
	workDir := s.Options().WorkDir
	if workDir == "" {
		return errors.New("git.validatePath: path must be absolute")
	}
	workDir = filepath.Clean(workDir)
	if !filepath.IsAbs(workDir) {
		return errors.New("git.validatePath: WorkDir must be absolute")
	}
	rel, err := filepath.Rel(workDir, filepath.Clean(path))
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return errors.New("git.validatePath: path is outside of allowed WorkDir")
	}
	return nil
}

func (s *Service) validatePaths(paths []string) error {
	for _, path := range paths {
		if err := s.validatePath(path); err != nil {
			return err
		}
	}
	return nil
}
