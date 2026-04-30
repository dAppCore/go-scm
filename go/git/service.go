// SPDX-License-Identifier: EUPL-1.2

package git

import (
	"context"
	"iter"
	"sync"

	core "dappco.re/go"
)

const (
	sonarServiceGitPull              = "git.pull"
	sonarServiceGitPush              = "git.push"
	sonarServiceGitValidatepath      = "git.validatePath"
	sonarServicePathValidationFailed = "path validation failed"
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
		return core.Ok(&Service{ServiceRuntime: core.NewServiceRuntime(c, opts)})
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
		return core.Ok(nil)
	}
	if err := ctx.Err(); err != nil {
		return core.Fail(err)
	}

	c := s.Core()
	if c == nil {
		return core.Fail(core.E("git.Service.OnStartup", "core is required", nil))
	}

	c.RegisterQuery(s.handleQuery)
	c.RegisterAction(s.handleTaskMessage)

	c.Action(sonarServiceGitPush, func(ctx context.Context, opts core.Options) core.Result {
		return s.runPush(ctx, opts.String(`path`))
	})
	c.Action(sonarServiceGitPull, func(ctx context.Context, opts core.Options) core.Result {
		return s.runPull(ctx, opts.String(`path`))
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
	return core.Ok(nil)
}

func (s *Service) handleQuery(c *core.Core, q core.Query) core.Result {
	if s == nil {
		return core.Fail(nil)
	}
	ctx := c.Context()
	switch m := q.(type) {
	case QueryStatus:
		if err := s.validatePaths(m.Paths); err != nil {
			return c.LogError(err, "git.handleQuery", sonarServicePathValidationFailed)
		}
		statuses := Status(ctx, StatusOptions(m))
		s.mu.Lock()
		s.lastStatus = statuses
		s.mu.Unlock()
		return core.Ok(statuses)
	case QueryDirtyRepos:
		return core.Ok(s.DirtyRepos())
	case QueryAheadRepos:
		return core.Ok(s.AheadRepos())
	case QueryBehindRepos:
		return core.Ok(s.BehindRepos())
	default:
		return core.Fail(nil)
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
		return core.Fail(nil)
	}
}

func (s *Service) runPush(ctx context.Context, path string) core.Result {
	if err := s.validatePath(path); err != nil {
		return s.Core().LogError(err, sonarServiceGitPush, sonarServicePathValidationFailed)
	}
	if err := Push(ctx, path); err != nil {
		return s.Core().LogError(err, sonarServiceGitPush, "push failed")
	}
	return core.Ok(nil)
}

func (s *Service) runPull(ctx context.Context, path string) core.Result {
	if err := s.validatePath(path); err != nil {
		return s.Core().LogError(err, sonarServiceGitPull, sonarServicePathValidationFailed)
	}
	if err := Pull(ctx, path); err != nil {
		return s.Core().LogError(err, sonarServiceGitPull, "pull failed")
	}
	return core.Ok(nil)
}

func (s *Service) runPushMultiple(ctx context.Context, paths []string, names map[string]string) core.Result {
	if err := s.validatePaths(paths); err != nil {
		return s.Core().LogError(err, "git.push-multiple", sonarServicePathValidationFailed)
	}
	results := PushMultiple(ctx, paths, names)
	for _, result := range results {
		if result.Error != nil {
			return core.Fail(result.Error)
		}
	}
	return core.Ok(results)
}

func (s *Service) runPullMultiple(ctx context.Context, paths []string, names map[string]string) core.Result {
	if err := s.validatePaths(paths); err != nil {
		return s.Core().LogError(err, "git.pull-multiple", sonarServicePathValidationFailed)
	}
	results := PullMultiple(ctx, paths, names)
	for _, result := range results {
		if result.Error != nil {
			return core.Fail(result.Error)
		}
	}
	return core.Ok(results)
}

func (s *Service) validatePath(path string) error  /* v090-result-boundary */ {
	ds := core.Env("DS")
	if core.PathIsAbs(path) {
		return nil
	}
	workDir := s.Options().WorkDir
	if workDir == "" {
		return core.E(sonarServiceGitValidatepath, "path must be absolute", nil)
	}
	workDir = core.CleanPath(workDir, ds)
	if !core.PathIsAbs(workDir) {
		return core.E(sonarServiceGitValidatepath, "WorkDir must be absolute", nil)
	}
	relResult := core.PathRel(workDir, core.CleanPath(path, ds))
	rel, _ := relResult.Value.(string)
	if !relResult.OK {
		err, _ := relResult.Value.(error)
		return core.E(sonarServiceGitValidatepath, "path is outside of allowed WorkDir", err)
	}
	if rel == ".." || core.HasPrefix(rel, ".."+ds) {
		return core.E(sonarServiceGitValidatepath, "path is outside of allowed WorkDir", nil)
	}
	return nil
}

func (s *Service) validatePaths(paths []string) error  /* v090-result-boundary */ {
	for _, path := range paths {
		if err := s.validatePath(path); err != nil {
			return err
		}
	}
	return nil
}
