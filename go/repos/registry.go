// SPDX-License-Identifier: EUPL-1.2

package repos

import (
	// Note: AX-6 — Git sync operations propagate cancellation through context.Context.
	"context"
	// Note: AX-6 — Registry discovery uses fs.ErrNotExist as the filesystem sentinel.
	"io/fs"
	// Note: AX-6 — Registry listing must be deterministic across map iteration (no core sort primitive).
	"sort"

	core "dappco.re/go"
	coreio "dappco.re/go/io"
	"dappco.re/go/scm/git"
	"gopkg.in/yaml.v3"
)

const (
	sonarRegistryMediumIsRequired      = "medium is required"
	sonarRegistryReposRegistrySyncrepo = "repos.Registry.SyncRepo"
	sonarRegistryReposYaml             = "repos.yaml"
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
	return core.Stat(repo.Path).OK
}

func (repo *Repo) IsGitRepo() bool {
	if repo == nil || repo.Path == "" {
		return false
	}
	return core.Stat(core.PathJoin(repo.Path, ".git")).OK
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
			cp.Path = core.PathJoin(r.BasePath, name)
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
		cp.Path = core.PathJoin(r.BasePath, name)
	}
	return &cp, true
}

func (r *Registry) ByType(t string) []*Repo {
	var out []*Repo
	for _, repo := range r.List() {
		if core.Lower(repo.Type) == core.Lower(t) {
			out = append(out, repo)
		}
	}
	return out
}

func (r *Registry) TopologicalOrder() ([]*Repo, error)  /* v090-result-boundary */ {
	// Simple dependency ordering with deterministic fallback.
	repos := r.List()
	if len(repos) == 0 {
		return nil, nil
	}
	byName := reposByName(repos)
	var ordered []*Repo
	seen := map[string]bool{}
	for _, repo := range repos {
		if err := visitRepo(repo.Name, byName, seen, nil, &ordered); err != nil {
			return nil, err
		}
	}
	return ordered, nil
}

func reposByName(repos []*Repo) map[string]*Repo {
	byName := make(map[string]*Repo, len(repos))
	for _, repo := range repos {
		byName[repo.Name] = repo
	}
	return byName
}

func visitRepo(name string, byName map[string]*Repo, seen map[string]bool, stack map[string]bool, ordered *[]*Repo) error  /* v090-result-boundary */ {
	if seen[name] {
		return nil
	}
	if stack[name] {
		return core.E("repos.Registry.TopologicalOrder", "dependency cycle", nil)
	}
	repo := byName[name]
	if repo == nil {
		return nil
	}
	nextStack := repoVisitStack(stack, name)
	for _, dep := range repo.DependsOn {
		if err := visitRepo(dep, byName, seen, nextStack, ordered); err != nil {
			return err
		}
	}
	delete(nextStack, name)
	seen[name] = true
	*ordered = append(*ordered, repo)
	return nil
}

func repoVisitStack(stack map[string]bool, name string) map[string]bool {
	if stack == nil {
		stack = make(map[string]bool)
	}
	stack[name] = true
	return stack
}

func LoadRegistry(m coreio.Medium, path string) (*Registry, error)  /* v090-result-boundary */ {
	if m == nil {
		return nil, core.E("repos.LoadRegistry", sonarRegistryMediumIsRequired, nil)
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
			repo.Path = core.PathJoin(r.BasePath, name)
		}
	}
	r.medium = m
	return &r, nil
}

func FindRegistry(m coreio.Medium) (string, error)  /* v090-result-boundary */ {
	if m == nil {
		return "", core.E("repos.FindRegistry", sonarRegistryMediumIsRequired, nil)
	}
	for _, candidate := range registryCandidates() {
		if m.Exists(candidate) {
			return candidate, nil
		}
	}
	return "", fs.ErrNotExist
}

func registryCandidates() []string {
	candidates := []string{sonarRegistryReposYaml, core.PathJoin(".core", sonarRegistryReposYaml)}
	candidates = prependEnvRegistryCandidates(candidates)
	candidates = append(candidates, cwdRegistryCandidates()...)
	if homeResult := core.UserHomeDir(); homeResult.OK {
		candidates = append(candidates, core.PathJoin(homeResult.Value.(string), ".core", sonarRegistryReposYaml))
	}
	return candidates
}

func prependEnvRegistryCandidates(candidates []string) []string {
	env := core.Trim(core.Getenv("CORE_REPOS"))
	if env == "" {
		return candidates
	}
	for _, candidate := range core.Split(env, core.Env("PS")) {
		candidate = core.Trim(candidate)
		if candidate != "" {
			candidates = append([]string{candidate}, candidates...)
		}
	}
	return candidates
}

func cwdRegistryCandidates() []string {
	cwdResult := core.Getwd()
	if !cwdResult.OK {
		return nil
	}
	cwd := cwdResult.Value.(string)
	var candidates []string
	for dir := cwd; ; dir = core.PathDir(dir) {
		candidates = append(candidates, core.PathJoin(dir, ".core", sonarRegistryReposYaml))
		if parent := core.PathDir(dir); parent == dir {
			return candidates
		}
	}
}

func ScanDirectory(m coreio.Medium, dir string) (*Registry, error)  /* v090-result-boundary */ {
	if m == nil {
		return nil, core.E("repos.ScanDirectory", sonarRegistryMediumIsRequired, nil)
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
		if !m.IsDir(core.PathJoin(dir, name)) {
			continue
		}
		reg.Repos[name] = &Repo{Name: name, Path: core.PathJoin(dir, name), registry: reg}
	}
	return reg, nil
}

func (r *Registry) Save(path string) error  /* v090-result-boundary */ {
	if r == nil {
		return core.E("repos.Registry.Save", "registry is required", nil)
	}
	raw, err := yaml.Marshal(r)
	if err != nil {
		return err
	}
	if r.medium != nil {
		return r.medium.Write(path, string(raw))
	}
	writeResult := core.WriteFile(path, raw, 0o600)
	if !writeResult.OK {
		return core.E("repos.Registry.Save", "write registry", nil)
	}
	return nil
}

// SyncRepo fetches and resets a named repo to match its Forge remote branch.
func (r *Registry) SyncRepo(ctx context.Context, name, remote, branch string) error  /* v090-result-boundary */ {
	if r == nil {
		return core.E(sonarRegistryReposRegistrySyncrepo, "registry is required", nil)
	}
	repo, ok := r.Get(name)
	if !ok {
		return core.E(sonarRegistryReposRegistrySyncrepo, core.Sprintf("repo %q not found", name), nil)
	}
	if repo.Path == "" {
		return core.E(sonarRegistryReposRegistrySyncrepo, core.Sprintf("repo %q has no path", name), nil)
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
