// SPDX-License-Identifier: EUPL-1.2

package repos

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	core "dappco.re/go/core"
	"dappco.re/go/scm/git"
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
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return core.Result{Value: err, OK: false}
	}
	s.registry = reg

	c := s.Core()
	if c == nil {
		return core.Result{Value: errors.New("repos.Service.OnStartup: core is required"), OK: false}
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
		return nil, errors.New("repos.Service.syncRepo: service is required")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	remote := optionOrDefault(opts.String("remote"), s.Options().Remote, "origin")
	branch := optionOrDefault(opts.String("branch"), s.Options().Branch, "dev")

	if path := strings.TrimSpace(opts.String("path")); path != "" {
		if err := git.SyncWithRemote(ctx, path, remote, branch); err != nil {
			return &git.SyncResult{Name: filepath.Base(path), Path: path, Success: false, Error: err}, err
		}
		return &git.SyncResult{Name: filepath.Base(path), Path: path, Success: true}, nil
	}

	if repoName := strings.TrimSpace(opts.String("repo")); repoName != "" {
		reg, err := s.registryForPath(opts.String("root"))
		if err != nil {
			return nil, err
		}
		if reg == nil {
			return nil, fmt.Errorf("repos.Service.syncRepo: registry not loaded")
		}
		if err := reg.SyncRepo(ctx, repoName, remote, branch); err != nil {
			return &git.SyncResult{Name: repoName, Success: false, Error: err}, err
		}
		repo, ok := reg.Get(repoName)
		if !ok {
			return &git.SyncResult{Name: repoName, Success: true}, nil
		}
		return &git.SyncResult{Name: repo.Name, Path: repo.Path, Success: true}, nil
	}

	return nil, errors.New("repos.Service.syncRepo: repo or path is required")
}

func (s *Service) syncAll(ctx context.Context, opts core.Options) ([]SyncResult, error) {
	if s == nil {
		return nil, errors.New("repos.Service.syncAll: service is required")
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
		return nil, fmt.Errorf("repos.Service.syncAll: registry not loaded")
	}
	return reg.SyncAll(ctx, remote, branch), nil
}

func (s *Service) registryForPath(root string) (*Registry, error) {
	if s == nil {
		return nil, errors.New("repos.Service.registryForPath: service is required")
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
	path, err := s.registryPath(root)
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(path)
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
			repo.Path = filepath.Join(reg.BasePath, name)
		}
		repo.registry = &reg
	}
	return &reg, nil
}

func (s *Service) registryPath(root string) (string, error) {
	if s == nil {
		return "", errors.New("repos.Service.registryPath: service is required")
	}
	opts := s.Options()
	candidates := []string{}
	if opts.RegistryPath != "" {
		candidates = append(candidates, opts.RegistryPath)
	}
	if root != "" {
		candidates = append(candidates,
			filepath.Join(root, ".core", "repos.yaml"),
			filepath.Join(root, "repos.yaml"),
		)
	}
	if opts.Root != "" {
		candidates = append(candidates,
			filepath.Join(opts.Root, ".core", "repos.yaml"),
			filepath.Join(opts.Root, "repos.yaml"),
		)
	}
	if cwd, err := os.Getwd(); err == nil {
		dir := cwd
		for {
			candidates = append(candidates, filepath.Join(dir, ".core", "repos.yaml"))
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, ".core", "repos.yaml"))
	}

	seen := map[string]struct{}{}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		candidate = filepath.Clean(candidate)
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", os.ErrNotExist
}

func inferRegistryBasePath(path string) string {
	if path == "" {
		return ""
	}
	dir := filepath.Dir(path)
	if filepath.Base(dir) == ".core" {
		return filepath.Dir(dir)
	}
	return dir
}

func optionOrDefault(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
