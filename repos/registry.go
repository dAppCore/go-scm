// SPDX-License-Identifier: EUPL-1.2

package repos

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	coreio "dappco.re/go/core/io"
	"dappco.re/go/scm/git"
	"dappco.re/go/scm/internal/ax/filepathx"
	"dappco.re/go/scm/internal/ax/jsonx"
	"dappco.re/go/scm/internal/ax/osx"
	"gopkg.in/yaml.v3"
)

type RepoType string

type RegistryDefaults struct {
	CI      string `yaml:"ci"`
	License string `yaml:"license"`
	Branch  string `yaml:"branch"`
}

type Repo struct {
	Name        string    `yaml:"-"`
	Type        string    `yaml:"type"`
	DependsOn   []string  `yaml:"depends_on"`
	Description string    `yaml:"description"`
	Docs        bool      `yaml:"docs"`
	CI          string    `yaml:"ci"`
	Domain      string    `yaml:"domain,omitempty"`
	Clone       *bool     `yaml:"clone,omitempty"`
	Path        string    `yaml:"path,omitempty"`
	registry    *Registry `yaml:"-"`
}

type Registry struct {
	Version  int              `yaml:"version"`
	Org      string           `yaml:"org"`
	BasePath string           `yaml:"base_path"`
	Repos    map[string]*Repo `yaml:"repos"`
	Defaults RegistryDefaults `yaml:"defaults"`
	medium   coreio.Medium    `yaml:"-"`
}

// SyncResult reports the outcome of syncing a registry repo clone.
type SyncResult struct {
	Name    string
	Path    string
	Success bool
	Error   error
}

func (repo *Repo) Exists() bool {
	if repo == nil || repo.Path == "" {
		return false
	}
	_, err := osx.Stat(repo.Path)
	return err == nil
}

func (repo *Repo) IsGitRepo() bool {
	if repo == nil || repo.Path == "" {
		return false
	}
	_, err := osx.Stat(filepath.Join(repo.Path, ".git"))
	return err == nil
}

func (r *Registry) List() []*Repo {
	if r == nil || len(r.Repos) == 0 {
		return nil
	}
	names := make([]string, 0, len(r.Repos))
	for name := range r.Repos {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]*Repo, 0, len(names))
	for _, name := range names {
		repo := r.Repos[name]
		if repo == nil {
			continue
		}
		cp := *repo
		cp.Name = name
		cp.registry = r
		if cp.Path == "" {
			cp.Path = filepath.Join(r.BasePath, name)
		}
		out = append(out, &cp)
	}
	return out
}

func (r *Registry) Get(name string) (*Repo, bool) {
	if r == nil {
		return nil, false
	}
	repo, ok := r.Repos[name]
	if !ok || repo == nil {
		return nil, false
	}
	cp := *repo
	cp.Name = name
	cp.registry = r
	if cp.Path == "" {
		cp.Path = filepath.Join(r.BasePath, name)
	}
	return &cp, true
}

func (r *Registry) ByType(t string) []*Repo {
	var out []*Repo
	for _, repo := range r.List() {
		if strings.EqualFold(repo.Type, t) {
			out = append(out, repo)
		}
	}
	return out
}

func (r *Registry) TopologicalOrder() ([]*Repo, error) {
	// Simple dependency ordering with deterministic fallback.
	repos := r.List()
	if len(repos) == 0 {
		return nil, nil
	}
	byName := make(map[string]*Repo, len(repos))
	for _, repo := range repos {
		byName[repo.Name] = repo
	}
	var ordered []*Repo
	seen := map[string]bool{}
	var visit func(string, map[string]bool) error
	visit = func(name string, stack map[string]bool) error {
		if seen[name] {
			return nil
		}
		if stack[name] {
			return errors.New("repos.Registry.TopologicalOrder: dependency cycle")
		}
		repo := byName[name]
		if repo == nil {
			return nil
		}
		if stack == nil {
			stack = make(map[string]bool)
		}
		stack[name] = true
		for _, dep := range repo.DependsOn {
			if err := visit(dep, stack); err != nil {
				return err
			}
		}
		delete(stack, name)
		seen[name] = true
		ordered = append(ordered, repo)
		return nil
	}
	for _, repo := range repos {
		if err := visit(repo.Name, nil); err != nil {
			return nil, err
		}
	}
	return ordered, nil
}

func LoadRegistry(m coreio.Medium, path string) (*Registry, error) {
	if m == nil {
		return nil, errors.New("repos.LoadRegistry: medium is required")
	}
	raw, err := m.Read(path)
	if err != nil {
		return &Registry{Version: 1, Repos: map[string]*Repo{}, medium: m}, nil
	}
	var r Registry
	if err := yaml.Unmarshal([]byte(raw), &r); err != nil {
		return nil, err
	}
	if r.Repos == nil {
		r.Repos = make(map[string]*Repo)
	}
	for name, repo := range r.Repos {
		if repo == nil {
			continue
		}
		repo.Name = name
		repo.registry = &r
		if repo.Path == "" {
			repo.Path = filepath.Join(r.BasePath, name)
		}
	}
	return &r, nil
}

func FindRegistry(m coreio.Medium) (string, error) {
	if m == nil {
		return "", errors.New("repos.FindRegistry: medium is required")
	}
	candidates := []string{"repos.yaml", filepath.Join(".core", "repos.yaml")}
	if env := strings.TrimSpace(osx.Getenv("CORE_REPOS")); env != "" {
		for _, candidate := range strings.Split(env, string(filepath.ListSeparator)) {
			candidate = strings.TrimSpace(candidate)
			if candidate != "" {
				candidates = append([]string{candidate}, candidates...)
			}
		}
	}
	if cwd, err := osx.Getwd(); err == nil {
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
	if home, err := osx.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, ".core", "repos.yaml"))
	}
	for _, candidate := range candidates {
		if m.Exists(candidate) {
			return candidate, nil
		}
	}
	return "", fs.ErrNotExist
}

func ScanDirectory(m coreio.Medium, dir string) (*Registry, error) {
	if m == nil {
		return nil, errors.New("repos.ScanDirectory: medium is required")
	}
	entries, err := m.List(dir)
	if err != nil {
		return nil, err
	}
	reg := &Registry{Version: 1, BasePath: dir, Repos: make(map[string]*Repo), medium: m}
	for _, entry := range entries {
		if entry == nil || !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !m.IsDir(filepathx.Join(dir, name)) {
			continue
		}
		reg.Repos[name] = &Repo{Name: name, Path: filepathx.Join(dir, name), registry: reg}
	}
	return reg, nil
}

func (r *Registry) Save(path string) error {
	if r == nil {
		return errors.New("repos.Registry.Save: registry is required")
	}
	raw, err := jsonx.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	if r.medium != nil {
		return r.medium.Write(path, string(raw))
	}
	return osx.WriteFile(path, raw, 0o600)
}

// SyncRepo fetches and resets a named repo to match its Forge remote branch.
func (r *Registry) SyncRepo(ctx context.Context, name, remote, branch string) error {
	if r == nil {
		return errors.New("repos.Registry.SyncRepo: registry is required")
	}
	repo, ok := r.Get(name)
	if !ok {
		return fmt.Errorf("repos.Registry.SyncRepo: repo %q not found", name)
	}
	if repo.Path == "" {
		return fmt.Errorf("repos.Registry.SyncRepo: repo %q has no path", name)
	}
	return git.SyncWithRemote(ctx, repo.Path, remote, branch)
}

// SyncAll synchronizes every repo in the registry.
func (r *Registry) SyncAll(ctx context.Context, remote, branch string) []SyncResult {
	if r == nil {
		return nil
	}
	repos := r.List()
	out := make([]SyncResult, 0, len(repos))
	for _, repo := range repos {
		if repo == nil {
			continue
		}
		result := SyncResult{Name: repo.Name, Path: repo.Path}
		if err := git.SyncWithRemote(ctx, repo.Path, remote, branch); err != nil {
			result.Error = err
		} else {
			result.Success = true
		}
		out = append(out, result)
	}
	return out
}
