// SPDX-License-Identifier: EUPL-1.2

// Package scm exposes a Core service that synchronises workspace repositories
// and reacts to workspace-pushed IPC messages.
package scm

import (
	"context"
	"path/filepath"
	"strings"

	"dappco.re/go/core"
	"dappco.re/go/core/io"
	coreerr "dappco.re/go/core/log"
	"dappco.re/go/core/scm/agentci"
	"dappco.re/go/core/scm/git"
	os "dappco.re/go/core/scm/internal/ax/osx"
	"dappco.re/go/core/scm/repos"
)

// WorkspacePushed is an IPC message emitted after a workspace push.
// If Repo is empty, the receiver should treat it as a sync-all request.
type WorkspacePushed struct {
	Org    string
	Repo   string
	Branch string
	Root   string
}

// ServiceOptions configures the SCM service.
type ServiceOptions struct {
	Medium        io.Medium
	RegistryPath  string
	WorkspaceRoot string
	DefaultOrg    string
	DefaultBranch string
}

// CoreService provides repo.sync actions and IPC handling for SCM workspaces.
type CoreService struct {
	*core.ServiceRuntime[ServiceOptions]
	medium           io.Medium
	registries       []*repos.Registry
	registryCacheKey string
}

// NewCoreService creates a Core service factory for use with core.WithService.
// Usage: NewCoreService(...)
func NewCoreService(opts ServiceOptions) func(*core.Core) (any, error) {
	return func(c *core.Core) (any, error) {
		if opts.Medium == nil {
			opts.Medium = io.Local
		}
		if opts.WorkspaceRoot == "" {
			opts.WorkspaceRoot = "~/Code"
		}
		return &CoreService{
			ServiceRuntime: core.NewServiceRuntime(c, opts),
			medium:         opts.Medium,
		}, nil
	}
}

// OnStartup registers the repo sync actions and eagerly loads registries.
func (s *CoreService) OnStartup(context.Context) core.Result {
	if _, err := s.loadRegistries(""); err != nil && !strings.Contains(err.Error(), "not found") {
		return core.Result{Value: err, OK: false}
	}

	s.Core().Action("repo.sync", s.handleRepoSync)
	s.Core().Action("repo.sync.all", s.handleRepoSyncAll)
	return core.Result{OK: true}
}

// HandleIPCEvents reacts to workspace push events by refreshing the local clone.
func (s *CoreService) HandleIPCEvents(c *core.Core, msg core.Message) core.Result {
	switch ev := msg.(type) {
	case WorkspacePushed:
		s.invalidateRegistries()
		return s.handleWorkspacePushed(c, ev)
	case *WorkspacePushed:
		if ev == nil {
			return core.Result{OK: true}
		}
		s.invalidateRegistries()
		return s.handleWorkspacePushed(c, *ev)
	default:
		return core.Result{OK: true}
	}
}

func (s *CoreService) handleWorkspacePushed(c *core.Core, ev WorkspacePushed) core.Result {
	opts := core.NewOptions(
		core.Option{Key: "org", Value: ev.Org},
		core.Option{Key: "repo", Value: ev.Repo},
		core.Option{Key: "branch", Value: ev.Branch},
		core.Option{Key: "root", Value: ev.Root},
	)
	if ev.Repo == "" {
		return c.Action("repo.sync.all").Run(c.Context(), opts)
	}
	return c.Action("repo.sync").Run(c.Context(), opts)
}

func (s *CoreService) handleRepoSync(ctx context.Context, opts core.Options) core.Result {
	repoName := firstOption(opts, "repo", "name")
	if repoName == "" {
		return core.Result{Value: coreerr.E("scm.handleRepoSync", "repo is required", nil), OK: false}
	}

	org := firstOption(opts, "org")
	branch := firstOption(opts, "branch")
	root := firstOption(opts, "root")

	repo, reg, repoPath, err := s.resolveRepo(repoName, org, root)
	if err != nil {
		return core.Result{Value: err, OK: false}
	}

	if branch == "" {
		branch = repoBranch(repo, reg, s.Options().DefaultBranch)
	}
	if branch == "" {
		branch = "main"
	}
	if org == "" && reg != nil {
		org = reg.Org
	}

	if err := git.Fetch(ctx, repoPath, branch); err != nil {
		return core.Result{Value: coreerr.E("scm.handleRepoSync", "fetch failed", err), OK: false}
	}
	if err := git.ResetHard(ctx, repoPath, "origin/"+branch); err != nil {
		return core.Result{Value: coreerr.E("scm.handleRepoSync", "reset failed", err), OK: false}
	}

	return core.Result{
		OK: true,
		Value: map[string]any{
			"repo":   repoName,
			"org":    org,
			"branch": branch,
			"path":   repoPath,
		},
	}
}

func (s *CoreService) handleRepoSyncAll(ctx context.Context, opts core.Options) core.Result {
	root := firstOption(opts, "root")
	regs, err := s.loadRegistries(root)
	if err != nil {
		return core.Result{Value: err, OK: false}
	}

	merged := repos.MergeRegistries(regs...)
	var synced, skipped int
	var failures []string
	results := make([]map[string]any, 0)

	order, orderErr := merged.TopologicalOrder()
	if orderErr != nil {
		order = merged.List()
	}

	for _, repo := range order {
		if repo != nil && repo.Clone != nil && !*repo.Clone {
			skipped++
			continue
		}

		// Preserve registry-level defaults for sync-all so repos inherit the
		// same branch resolution rules as single-repo sync.
		branch := repoBranch(repo, repo.Registry(), s.Options().DefaultBranch)
		if branch == "" {
			branch = "main"
		}
		org := repo.Org
		if org == "" {
			org = s.Options().DefaultOrg
		}

		syncOpts := core.NewOptions(
			core.Option{Key: "repo", Value: repo.Name},
			core.Option{Key: "org", Value: org},
			core.Option{Key: "branch", Value: branch},
			core.Option{Key: "root", Value: root},
		)
		r := s.handleRepoSync(ctx, syncOpts)
		if !r.OK {
			skipped++
			if r.Value != nil {
				if syncErr, ok := r.Value.(error); ok {
					failures = append(failures, syncErr.Error())
				}
			}
			continue
		}

		if value, ok := r.Value.(map[string]any); ok {
			results = append(results, value)
		}
		synced++
	}

	return core.Result{
		OK: true,
		Value: map[string]any{
			"synced":   synced,
			"skipped":  skipped,
			"failures": failures,
			"results":  results,
		},
	}
}

func (s *CoreService) loadRegistries(root string) ([]*repos.Registry, error) {
	cacheKey, err := s.registryDiscoveryKey(root)
	if err != nil {
		return nil, err
	}

	if s.registries != nil && s.registryCacheKey == cacheKey {
		return s.registries, nil
	}

	if path := s.Options().RegistryPath; path != "" {
		reg, err := repos.LoadRegistry(s.medium, path)
		if err != nil {
			return nil, err
		}
		s.registries = []*repos.Registry{reg}
		s.registryCacheKey = "path:" + path
		return s.registries, nil
	}

	regs, err := repos.LoadRegistries(s.medium)
	if err != nil {
		return nil, err
	}
	if len(regs) == 0 {
		scanRoot := root
		if scanRoot == "" {
			scanRoot = s.Options().WorkspaceRoot
		}
		scanRoot, rootErr := expandHome(scanRoot)
		if rootErr != nil {
			return nil, rootErr
		}

		scanned, scanErr := repos.ScanDirectory(s.medium, scanRoot)
		if scanErr != nil {
			return nil, scanErr
		}
		regs = []*repos.Registry{scanned}
	}
	s.registries = regs
	s.registryCacheKey = cacheKey
	return regs, nil
}

func (s *CoreService) invalidateRegistries() {
	s.registries = nil
	s.registryCacheKey = ""
}

func (s *CoreService) resolveRepo(name, org, root string) (*repos.Repo, *repos.Registry, string, error) {
	regs, err := s.loadRegistries(root)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return nil, nil, "", err
	}

	requestedOrg := strings.TrimSpace(org)
	for _, reg := range regs {
		if repo, ok := reg.Get(name); ok {
			repoOrg := strings.TrimSpace(repo.Org)
			regOrg := strings.TrimSpace(reg.Org)
			if requestedOrg != "" && requestedOrg != repoOrg && requestedOrg != regOrg {
				continue
			}
			if org == "" {
				org = regOrg
			}
			return repo, reg, repo.Path, nil
		}
	}

	if root == "" {
		root = s.Options().WorkspaceRoot
	}
	if root == "" {
		return nil, nil, "", coreerr.E("scm.resolveRepo", "workspace root is required", nil)
	}
	root, err = expandHome(root)
	if err != nil {
		return nil, nil, "", coreerr.E("scm.resolveRepo", "resolve workspace root", err)
	}

	if org == "" {
		org = s.Options().DefaultOrg
	}
	if org != "" {
		if _, err := agentci.ValidatePathElement(org); err != nil {
			return nil, nil, "", coreerr.E("scm.resolveRepo", "invalid org", err)
		}
	}
	if _, err := agentci.ValidatePathElement(name); err != nil {
		return nil, nil, "", coreerr.E("scm.resolveRepo", "invalid repo", err)
	}

	repoPath := root
	if org != "" {
		repoPath = filepath.Join(repoPath, org)
	}
	repoPath = filepath.Join(repoPath, name)

	return &repos.Repo{Name: name, Path: repoPath}, nil, repoPath, nil
}

func (s *CoreService) registryDiscoveryKey(root string) (string, error) {
	if path := s.Options().RegistryPath; path != "" {
		return "path:" + path, nil
	}

	scanRoot := root
	if scanRoot == "" {
		scanRoot = s.Options().WorkspaceRoot
	}
	if scanRoot == "" {
		return "discover", nil
	}

	scanRoot, err := expandHome(scanRoot)
	if err != nil {
		return "", err
	}
	return "scan:" + scanRoot, nil
}

func repoBranch(repo *repos.Repo, reg *repos.Registry, fallback string) string {
	if repo != nil && repo.Branch != "" {
		return repo.Branch
	}
	if reg != nil && reg.Defaults.Branch != "" {
		return reg.Defaults.Branch
	}
	return fallback
}

func firstOption(opts core.Options, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(opts.String(key)); value != "" {
			return value
		}
	}
	return ""
}

func expandHome(path string) (string, error) {
	if !strings.HasPrefix(path, "~/") {
		return filepath.Clean(path), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Clean(filepath.Join(home, path[2:])), nil
}
