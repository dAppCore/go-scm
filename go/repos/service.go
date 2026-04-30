// SPDX-License-Identifier: EUPL-1.2

package repos

import (
	// Note: AX-6 — Core action handlers and git sync accept context.Context.
	"context"
	// Note: AX-6 — Registry discovery uses fs.ErrNotExist as the filesystem sentinel.
	"io/fs"

	core "dappco.re/go"
	"dappco.re/go/scm/git"
	"dappco.re/go/scm/internal/ax/filepathx"
	"dappco.re/go/scm/internal/ax/osx"
	"gopkg.in/yaml.v3"
)

const (
	sonarServiceReposServiceSyncrepo = "repos.Service.syncRepo"
	sonarServiceReposYaml            = "repos.yaml"
	sonarServiceServiceIsRequired    = "service is required"
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
		return core.Ok(&Service{ServiceRuntime: core.NewServiceRuntime(c, opts)})
	}
}

// OnStartup loads the registry and registers repo sync actions.
func (s *Service) OnStartup(ctx context.Context) core.Result {
	if s == nil {
		return core.Ok(nil)
	}
	if err := ctx.Err(); err != nil {
		return core.Fail(err)
	}
	reg, err := s.loadRegistry()
	if err != nil && !core.Is(err, fs.ErrNotExist) {
		return core.Fail(err)
	}
	s.registry = reg

	c := s.Core()
	if c == nil {
		return core.Fail(core.E("repos.Service.OnStartup", "core is required", nil))
	}

	c.Action("repo.sync", s.handleRepoSync)
	c.Action("repo.sync.all", s.handleRepoSyncAll)
	return core.Ok(nil)
}

// HandleIPCEvents reacts to workspace push broadcasts by resyncing repos.
func (s *Service) HandleIPCEvents(_ *core.Core, msg core.Message) core.Result {
	if s == nil {
		return core.Ok(nil)
	}
	switch v := msg.(type) {
	case WorkspacePushed:
		return s.syncWorkspace(context.Background(), v)
	case *WorkspacePushed:
		if v == nil {
			return core.Ok(nil)
		}
		return s.syncWorkspace(context.Background(), *v)
	default:
		return core.Ok(nil)
	}
}

func (s *Service) handleRepoSync(ctx context.Context, opts core.Options) core.Result {
	result, err := s.syncRepo(ctx, opts)
	if err != nil {
		return core.Fail(err)
	}
	return core.Ok(result)
}

func (s *Service) handleRepoSyncAll(ctx context.Context, opts core.Options) core.Result {
	result, err := s.syncAll(ctx, opts)
	if err != nil {
		return core.Fail(err)
	}
	return core.Ok(result)
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
		return nil, core.E(sonarServiceReposServiceSyncrepo, sonarServiceServiceIsRequired, nil)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	remote := optionOrDefault(opts.String("remote"), s.Options().Remote, "origin")
	branch := optionOrDefault(opts.String("branch"), s.Options().Branch, "dev")
	workspacePath, workspaceOK := workspaceRepoPath(opts, s.Options().Root)

	if path := core.Trim(opts.String("path")); path != "" {
		return syncPath(ctx, path, filepathx.Base(path), remote, branch)
	}

	if repoName := core.Trim(opts.String("repo")); repoName != "" {
		return s.syncNamedRepo(ctx, opts, repoName, workspacePath, workspaceOK, remote, branch)
	}

	return nil, core.E(sonarServiceReposServiceSyncrepo, "repo or path is required", nil)
}

func (s *Service) syncNamedRepo(
	ctx context.Context,
	opts core.Options,
	repoName string,
	workspacePath string,
	workspaceOK bool,
	remote string,
	branch string,
) (*git.SyncResult, error) {
	reg, err := s.registryForPath(opts.String("root"))
	if err != nil && !core.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	if result, ok, err := syncRegistryRepo(ctx, reg, repoName, remote, branch); ok || err != nil {
		return result, err
	}
	if workspaceOK {
		return syncPath(ctx, workspacePath, repoName, remote, branch)
	}
	if reg == nil {
		return nil, core.E(sonarServiceReposServiceSyncrepo, "registry not loaded", nil)
	}
	return nil, core.E(sonarServiceReposServiceSyncrepo, core.Sprintf("repo %q not found in registry", repoName), nil)
}

func syncRegistryRepo(ctx context.Context, reg *Registry, repoName, remote, branch string) (*git.SyncResult, bool, error) {
	if reg == nil {
		return nil, false, nil
	}
	repo, ok := reg.Get(repoName)
	if !ok {
		return nil, false, nil
	}
	result, err := syncPath(ctx, repo.Path, repo.Name, remote, branch)
	return result, true, err
}

func syncPath(ctx context.Context, path, name, remote, branch string) (*git.SyncResult, error) {
	result := &git.SyncResult{Name: name, Path: path, Success: true}
	if err := git.SyncWithRemote(ctx, path, remote, branch); err != nil {
		result.Success = false
		result.Error = err
		return result, err
	}
	return result, nil
}

func (s *Service) syncAll(ctx context.Context, opts core.Options) ([]SyncResult, error) {
	if s == nil {
		return nil, core.E("repos.Service.syncAll", sonarServiceServiceIsRequired, nil)
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
		return nil, core.E("repos.Service.registryForPath", sonarServiceServiceIsRequired, nil)
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
		mergeRegistry(merged, reg)
	}
	if merged == nil {
		return nil, fs.ErrNotExist
	}
	return merged, nil
}

func mergeRegistry(merged, reg *Registry) {
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

func (s *Service) registryPaths(root string) ([]string, error) {
	if s == nil {
		return nil, core.E("repos.Service.registryPaths", sonarServiceServiceIsRequired, nil)
	}
	opts := s.Options()
	candidates := []string{}
	if opts.RegistryPath != "" {
		candidates = append(candidates, opts.RegistryPath)
		return cleanExistingCandidates(candidates), nil
	}
	candidates = append(candidates, rootRegistryCandidates(root)...)
	candidates = append(candidates, rootRegistryCandidates(opts.Root)...)
	if cwd, err := osx.Getwd(); err == nil {
		dir := cwd
		for {
			candidates = append(candidates, filepathx.Join(dir, ".core", sonarServiceReposYaml))
			parent := filepathx.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	if home, err := osx.UserHomeDir(); err == nil {
		candidates = append(candidates, filepathx.Join(home, ".core", sonarServiceReposYaml))
	}

	return cleanExistingCandidates(candidates), nil
}

func rootRegistryCandidates(root string) []string {
	if root == "" {
		return nil
	}
	return []string{
		filepathx.Join(root, ".core", sonarServiceReposYaml),
		filepathx.Join(root, sonarServiceReposYaml),
	}
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
