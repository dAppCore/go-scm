// SPDX-License-Identifier: EUPL-1.2

package repos

import (
	// Note: AX-6 — Core action handlers and git sync accept context.Context.
	"context"
	// Note: AX-6 — Registry discovery uses fs.ErrNotExist as the filesystem sentinel.
	"io/fs"

	core "dappco.re/go/core"
	"dappco.re/go/scm/git"
	"dappco.re/go/scm/internal/ax/filepathx"
	"dappco.re/go/scm/internal/ax/osx"
	"gopkg.in/yaml.v3"
)

// ServiceOptions configures the repo sync service.
type ServiceOptions struct {
	Root         string
	RegistryPath string
	Remote       string
	Branch       string
}

// WorkspacePushed is an IPC payload that requests a repo sync after a push.
type WorkspacePushed struct {
	Root   string
	Org    string
	Repo   string
	Path   string
	Remote string
	Branch string
}

// Service wires registry-backed sync actions into Core.
type Service struct {
	*core.ServiceRuntime[ServiceOptions]
	registry *Registry
}

// NewService creates a Core-compatible factory for repo sync actions.
func NewService(opts ServiceOptions) func(*core.Core) core.Result {
	return func(c *core.Core) core.Result {
		return core.Result{Value: &Service{ServiceRuntime: core.NewServiceRuntime(c, opts)}, OK: true}
	}
}

// OnStartup loads the registry and registers repo sync actions.
func (s *Service) OnStartup(ctx context.Context) core.Result {
	if s == nil {
		return core.Result{OK: true}
	}
	if err := ctx.Err(); err != nil {
		return core.Result{Value: err, OK: false}
	}
	reg, err := s.loadRegistry()
	if err != nil && !core.Is(err, fs.ErrNotExist) {
		return core.Result{Value: err, OK: false}
	}
	s.registry = reg

	c := s.Core()
	if c == nil {
		return core.Result{Value: core.E("repos.Service.OnStartup", "core is required", nil), OK: false}
	}

	c.Action("repo.sync", s.handleRepoSync)
	c.Action("repo.sync.all", s.handleRepoSyncAll)
	return core.Result{OK: true}
}

// HandleIPCEvents reacts to workspace push broadcasts by resyncing repos.
func (s *Service) HandleIPCEvents(_ *core.Core, msg core.Message) core.Result {
	if s == nil {
		return core.Result{OK: true}
	}
	switch v := msg.(type) {
	case WorkspacePushed:
		return s.syncWorkspace(context.Background(), v)
	case *WorkspacePushed:
		if v == nil {
			return core.Result{OK: true}
		}
		return s.syncWorkspace(context.Background(), *v)
	default:
		return core.Result{OK: true}
	}
}

func (s *Service) handleRepoSync(ctx context.Context, opts core.Options) core.Result {
	result, err := s.syncRepo(ctx, opts)
	if err != nil {
		return core.Result{Value: err, OK: false}
	}
	return core.Result{Value: result, OK: true}
}

func (s *Service) handleRepoSyncAll(ctx context.Context, opts core.Options) core.Result {
	result, err := s.syncAll(ctx, opts)
	if err != nil {
		return core.Result{Value: err, OK: false}
	}
	return core.Result{Value: result, OK: true}
}

func (s *Service) syncWorkspace(ctx context.Context, pushed WorkspacePushed) core.Result {
	opts := core.NewOptions(
		core.Option{Key: "root", Value: pushed.Root},
		core.Option{Key: "org", Value: pushed.Org},
		core.Option{Key: "repo", Value: pushed.Repo},
		core.Option{Key: "path", Value: pushed.Path},
		core.Option{Key: "remote", Value: pushed.Remote},
		core.Option{Key: "branch", Value: pushed.Branch},
	)
	if pushed.Repo != "" || pushed.Path != "" {
		return s.handleRepoSync(ctx, opts)
	}
	return s.handleRepoSyncAll(ctx, opts)
}

func (s *Service) syncRepo(ctx context.Context, opts core.Options) (*git.SyncResult, error) {
	if s == nil {
		return nil, core.E("repos.Service.syncRepo", "service is required", nil)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	remote := optionOrDefault(opts.String("remote"), s.Options().Remote, "origin")
	branch := optionOrDefault(opts.String("branch"), s.Options().Branch, "dev")
	workspacePath, workspaceOK := workspaceRepoPath(opts, s.Options().Root)

	if path := core.Trim(opts.String("path")); path != "" {
		if err := git.SyncWithRemote(ctx, path, remote, branch); err != nil {
			return &git.SyncResult{Name: filepathx.Base(path), Path: path, Success: false, Error: err}, err
		}
		return &git.SyncResult{Name: filepathx.Base(path), Path: path, Success: true}, nil
	}

	if repoName := core.Trim(opts.String("repo")); repoName != "" {
		reg, err := s.registryForPath(opts.String("root"))
		if err != nil && !core.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		if reg != nil {
			if repo, ok := reg.Get(repoName); ok {
				if err := git.SyncWithRemote(ctx, repo.Path, remote, branch); err != nil {
					return &git.SyncResult{Name: repo.Name, Path: repo.Path, Success: false, Error: err}, err
				}
				return &git.SyncResult{Name: repo.Name, Path: repo.Path, Success: true}, nil
			}
		}
		if workspaceOK {
			if err := git.SyncWithRemote(ctx, workspacePath, remote, branch); err != nil {
				return &git.SyncResult{Name: repoName, Path: workspacePath, Success: false, Error: err}, err
			}
			return &git.SyncResult{Name: repoName, Path: workspacePath, Success: true}, nil
		}
		if reg == nil {
			return nil, core.E("repos.Service.syncRepo", "registry not loaded", nil)
		}
		return nil, core.E("repos.Service.syncRepo", core.Sprintf("repo %q not found in registry", repoName), nil)
	}

	return nil, core.E("repos.Service.syncRepo", "repo or path is required", nil)
}

func (s *Service) syncAll(ctx context.Context, opts core.Options) ([]SyncResult, error) {
	if s == nil {
		return nil, core.E("repos.Service.syncAll", "service is required", nil)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	remote := optionOrDefault(opts.String("remote"), s.Options().Remote, "origin")
	branch := optionOrDefault(opts.String("branch"), s.Options().Branch, "dev")

	reg, err := s.registryForPath(opts.String("root"))
	if err != nil {
		return nil, err
	}
	if reg == nil {
		return nil, core.E("repos.Service.syncAll", "registry not loaded", nil)
	}
	return reg.SyncAll(ctx, remote, branch), nil
}

func (s *Service) registryForPath(root string) (*Registry, error) {
	if s == nil {
		return nil, core.E("repos.Service.registryForPath", "service is required", nil)
	}
	if s.registry != nil {
		return s.registry, nil
	}
	reg, err := s.loadRegistryAt(root)
	if err != nil {
		return nil, err
	}
	s.registry = reg
	return reg, nil
}

func (s *Service) loadRegistry() (*Registry, error) {
	return s.loadRegistryAt("")
}

func (s *Service) loadRegistryAt(root string) (*Registry, error) {
	paths, err := s.registryPaths(root)
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, fs.ErrNotExist
	}
	var merged *Registry
	for _, path := range paths {
		reg, err := loadRegistryFile(path)
		if err != nil {
			return nil, err
		}
		if merged == nil {
			merged = reg
			continue
		}
		for name, repo := range reg.Repos {
			if repo == nil {
				continue
			}
			if _, exists := merged.Repos[name]; exists {
				continue
			}
			cp := *repo
			cp.registry = merged
			merged.Repos[name] = &cp
		}
		if merged.BasePath == "" {
			merged.BasePath = reg.BasePath
		}
	}
	if merged == nil {
		return nil, fs.ErrNotExist
	}
	return merged, nil
}

func (s *Service) registryPath(root string) (string, error) {
	paths, err := s.registryPaths(root)
	if err != nil {
		return "", err
	}
	if len(paths) == 0 {
		return "", fs.ErrNotExist
	}
	return paths[0], nil
}

func (s *Service) registryPaths(root string) ([]string, error) {
	if s == nil {
		return nil, core.E("repos.Service.registryPaths", "service is required", nil)
	}
	opts := s.Options()
	candidates := []string{}
	if opts.RegistryPath != "" {
		candidates = append(candidates, opts.RegistryPath)
		return cleanExistingCandidates(candidates), nil
	}
	if root != "" {
		candidates = append(candidates,
			filepathx.Join(root, ".core", "repos.yaml"),
			filepathx.Join(root, "repos.yaml"),
		)
	}
	if opts.Root != "" {
		candidates = append(candidates,
			filepathx.Join(opts.Root, ".core", "repos.yaml"),
			filepathx.Join(opts.Root, "repos.yaml"),
		)
	}
	if cwd, err := osx.Getwd(); err == nil {
		dir := cwd
		for {
			candidates = append(candidates, filepathx.Join(dir, ".core", "repos.yaml"))
			parent := filepathx.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	if home, err := osx.UserHomeDir(); err == nil {
		candidates = append(candidates, filepathx.Join(home, ".core", "repos.yaml"))
	}

	return cleanExistingCandidates(candidates), nil
}

func cleanExistingCandidates(candidates []string) []string {
	seen := map[string]struct{}{}
	paths := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		candidate = filepathx.Clean(candidate)
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		if _, err := osx.Stat(candidate); err == nil {
			paths = append(paths, candidate)
		}
	}
	return paths
}

func loadRegistryFile(path string) (*Registry, error) {
	raw, err := osx.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var reg Registry
	if err := yaml.Unmarshal(raw, &reg); err != nil {
		return nil, err
	}
	if reg.Repos == nil {
		reg.Repos = map[string]*Repo{}
	}
	if reg.BasePath == "" {
		reg.BasePath = inferRegistryBasePath(path)
	}
	for name, repo := range reg.Repos {
		if repo == nil {
			continue
		}
		repo.Name = name
		if repo.Path == "" {
			repo.Path = filepathx.Join(reg.BasePath, name)
		}
		repo.registry = &reg
	}
	return &reg, nil
}

func inferRegistryBasePath(path string) string {
	if path == "" {
		return ""
	}
	dir := filepathx.Dir(path)
	if filepathx.Base(dir) == ".core" {
		return filepathx.Dir(dir)
	}
	return dir
}

func optionOrDefault(values ...string) string {
	for _, value := range values {
		if core.Trim(value) != "" {
			return value
		}
	}
	return ""
}

func workspaceRepoPath(opts core.Options, defaultRoot string) (string, bool) {
	root := core.Trim(opts.String("root"))
	if root == "" {
		root = core.Trim(defaultRoot)
	}
	if root == "" {
		if home, err := osx.UserHomeDir(); err == nil {
			root = filepathx.Join(home, "Code")
		}
	}
	org := core.Trim(opts.String("org"))
	repo := core.Trim(opts.String("repo"))
	if org == "" || repo == "" {
		return "", false
	}
	return filepathx.Join(root, org, repo), true
}
