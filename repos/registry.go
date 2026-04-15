// SPDX-License-Identifier: EUPL-1.2

// Package repos provides functionality for managing multi-repo workspaces.
// It reads a repos.yaml registry file that defines repositories, their types,
// dependencies, and metadata.
package repos

import (
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
	os "dappco.re/go/core/scm/internal/ax/osx"
	strings "dappco.re/go/core/scm/internal/ax/stringsx"
	"net/url"
	"sort"
	stdos "os"
	stdpath "path/filepath"
	stdstrings "strings"

	"dappco.re/go/core/io"
	coreerr "dappco.re/go/core/log"
	"gopkg.in/yaml.v3"
)

// Registry represents a collection of repositories defined in repos.yaml.
type Registry struct {
	Version  int              `yaml:"version"`
	Org      string           `yaml:"org"`
	BasePath string           `yaml:"base_path"`
	Repos    map[string]*Repo `yaml:"repos"`
	Defaults RegistryDefaults `yaml:"defaults"`
	medium   io.Medium        `yaml:"-"`
}

// RegistryOption configures a Registry created with NewRegistry.
type RegistryOption func(*Registry)

// WithMedium sets the filesystem medium used by repo helpers.
// Usage: WithMedium(...)
func WithMedium(m io.Medium) RegistryOption {
	return func(r *Registry) {
		if m != nil {
			r.medium = m
		}
	}
}

// NewRegistry creates an empty registry configured by the supplied options.
// Usage: NewRegistry(...)
func NewRegistry(opts ...RegistryOption) *Registry {
	reg := &Registry{
		Version: 1,
		Repos:   make(map[string]*Repo),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(reg)
		}
	}
	return reg
}

// RegistryDefaults contains default values applied to all repos.
type RegistryDefaults struct {
	CI      string `yaml:"ci"`
	License string `yaml:"license"`
	Branch  string `yaml:"branch"`
}

// UnmarshalYAML accepts both the historical "license" key and the UK-spelled
// "licence" key so existing registries continue to load regardless of which
// spelling they use.
func (d *RegistryDefaults) UnmarshalYAML(value *yaml.Node) error {
	type registryDefaultsYAML struct {
		CI      string `yaml:"ci"`
		License string `yaml:"license"`
		Licence string `yaml:"licence"`
		Branch  string `yaml:"branch"`
	}

	var raw registryDefaultsYAML
	if err := value.Decode(&raw); err != nil {
		return err
	}

	d.CI = raw.CI
	d.Branch = raw.Branch
	d.License = raw.License
	if d.License == "" {
		d.License = raw.Licence
	}
	return nil
}

// RepoType indicates the role of a repository in the ecosystem.
type RepoType string

// Repository type constants for ecosystem classification.
const (
	// RepoTypeFoundation indicates core foundation packages.
	//
	RepoTypeFoundation RepoType = "foundation"
	// RepoTypeModule indicates reusable module packages.
	//
	RepoTypeModule RepoType = "module"
	// RepoTypeProduct indicates end-user product applications.
	//
	RepoTypeProduct RepoType = "product"
	// RepoTypeTemplate indicates starter templates.
	//
	RepoTypeTemplate RepoType = "template"
)

// Repo represents a single repository in the registry.
type Repo struct {
	Name        string   `yaml:"-"` // Set from map key
	Type        string   `yaml:"type"`
	Org         string   `yaml:"org,omitempty"` // Computed from registry metadata
	Remote      string   `yaml:"remote,omitempty"`
	Branch      string   `yaml:"branch,omitempty"`
	DependsOn   []string `yaml:"depends_on"`
	Description string   `yaml:"description"`
	Docs        bool     `yaml:"docs"`
	CI          string   `yaml:"ci"`
	Domain      string   `yaml:"domain,omitempty"`
	Clone       *bool    `yaml:"clone,omitempty"` // nil = true, false = skip cloning

	// Computed fields
	Path     string    `yaml:"path,omitempty"` // Full path to repo directory (optional, defaults to base_path/name)
	registry *Registry `yaml:"-"`
}

// LoadRegistry reads and parses a repos.yaml file from the given medium.
// The path should be a valid path for the provided medium.
//
//	reg, err := repos.LoadRegistry(io.Local, ".core/repos.yaml")
//
// Usage: LoadRegistry(...)
func LoadRegistry(m io.Medium, path string) (*Registry, error) {
	content, err := m.Read(path)
	if err != nil {
		return nil, coreerr.E("repos.LoadRegistry", "failed to read registry file", err)
	}
	data := []byte(content)

	reg, err := decodeRegistry(data)
	if err != nil {
		return nil, coreerr.E("repos.LoadRegistry", "failed to parse registry file", err)
	}

	reg.medium = m

	// Expand base path
	reg.BasePath = expandPath(reg.BasePath)

	// Set computed fields on each repo
	for name, repo := range reg.Repos {
		finaliseRepo(reg, name, repo)
	}

	return reg, nil
}

// FindRegistry searches for repos.yaml in common locations.
// It checks: CORE_REPOS, current directory, parent directories, and home directory.
// This function is primarily intended for use with io.Local or other local-like filesystems.
//
//	path, err := repos.FindRegistry(io.Local)
//
// Usage: FindRegistry(...)
func FindRegistry(m io.Medium) (string, error) {
	for _, candidate := range registryCandidatesFromEnv() {
		if m.Exists(candidate) {
			return candidate, nil
		}
	}

	// Check current directory and parents
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		// Check .core/repos.yaml (new)
		candidate := filepath.Join(dir, ".core", "repos.yaml")
		if m.Exists(candidate) {
			return candidate, nil
		}
		// Check repos.yaml (legacy)
		candidate = filepath.Join(dir, "repos.yaml")
		if m.Exists(candidate) {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Check home directory common locations
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	commonPaths := []string{
		filepath.Join(home, ".core", "repos.yaml"),
		filepath.Join(home, "Code", ".core", "repos.yaml"),
		filepath.Join(home, "Code", "host-uk", ".core", "repos.yaml"),
		filepath.Join(home, "Code", "host-uk", "repos.yaml"),
		filepath.Join(home, ".config", "core", "repos.yaml"),
	}

	for _, p := range commonPaths {
		if m.Exists(p) {
			return p, nil
		}
	}

	return "", coreerr.E("repos.FindRegistry", "repos.yaml not found", nil)
}

// FindRegistries returns every discovered registry file path in priority order.
// CORE_REPOS entries are returned first, followed by conventional discovery.
// Usage: FindRegistries(...)
func FindRegistries(m io.Medium) ([]string, error) {
	seen := make(map[string]bool)
	var paths []string

	add := func(path string) {
		if path == "" || seen[path] {
			return
		}
		seen[path] = true
		paths = append(paths, path)
	}

	for _, candidate := range registryCandidatesFromEnv() {
		if m.Exists(candidate) {
			add(candidate)
		}
	}

	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	for {
		for _, candidate := range []string{
			filepath.Join(dir, ".core", "repos.yaml"),
			filepath.Join(dir, "repos.yaml"),
		} {
			if m.Exists(candidate) {
				add(candidate)
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return paths, nil
	}
	for _, candidate := range []string{
		filepath.Join(home, ".core", "repos.yaml"),
		filepath.Join(home, "Code", ".core", "repos.yaml"),
		filepath.Join(home, "Code", "host-uk", ".core", "repos.yaml"),
		filepath.Join(home, "Code", "host-uk", "repos.yaml"),
		filepath.Join(home, ".config", "core", "repos.yaml"),
	} {
		if m.Exists(candidate) {
			add(candidate)
		}
	}

	if len(paths) == 0 {
		return nil, coreerr.E("repos.FindRegistries", "repos.yaml not found", nil)
	}
	return paths, nil
}

// LoadRegistries discovers and loads all registry files found via FindRegistries.
// Use MergeRegistries when a first-occurrence view across multiple files is needed.
// Usage: LoadRegistries(...)
func LoadRegistries(m io.Medium) ([]*Registry, error) {
	paths, err := FindRegistries(m)
	if err != nil {
		if stdstrings.Contains(err.Error(), "not found") {
			return []*Registry{}, nil
		}
		return nil, err
	}

	regs := make([]*Registry, 0, len(paths))
	for _, path := range paths {
		reg, loadErr := LoadRegistry(m, path)
		if loadErr != nil {
			return nil, loadErr
		}
		regs = append(regs, reg)
	}
	return regs, nil
}

// MergeRegistries combines registries into a single read-only view.
// Repositories are deduplicated by name in first-occurrence order.
// Usage: MergeRegistries(...)
func MergeRegistries(regs ...*Registry) *Registry {
	merged := NewRegistry()
	merged.Repos = make(map[string]*Repo)

	for _, reg := range regs {
		if reg == nil {
			continue
		}
		if merged.medium == nil {
			merged.medium = reg.medium
		}
		if merged.Org == "" {
			merged.Org = reg.Org
		}
		if merged.BasePath == "" {
			merged.BasePath = reg.BasePath
		}

		for _, repo := range reg.List() {
			if repo == nil || strings.TrimSpace(repo.Name) == "" {
				continue
			}
			if _, exists := merged.Repos[repo.Name]; exists {
				continue
			}
			merged.Repos[repo.Name] = repo
		}
	}

	return merged
}

// ScanDirectory creates a Registry by scanning a directory for git repos.
// This is used as a fallback when no repos.yaml is found.
// The dir should be a valid path for the provided medium.
//
//	reg, err := repos.ScanDirectory(io.Local, "/home/user/Code/core")
//
// Usage: ScanDirectory(...)
func ScanDirectory(m io.Medium, dir string) (*Registry, error) {
	entries, err := m.List(dir)
	if err != nil {
		return nil, coreerr.E("repos.ScanDirectory", "failed to read directory", err)
	}

	// Some Medium implementations return an empty slice with no error
	// for nonexistent paths. Surface that as an error for parity with
	// local filesystem semantics.
	if len(entries) == 0 && !m.IsDir(dir) && !m.Exists(dir) {
		return nil, coreerr.E("repos.ScanDirectory", "failed to read directory", nil)
	}

	reg := &Registry{
		Version:  1,
		BasePath: dir,
		Repos:    make(map[string]*Repo),
		medium:   m,
	}

	// Try to detect org from git remote
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		repoPath := filepath.Join(dir, entry.Name())
		gitPath := filepath.Join(repoPath, ".git")

		if !m.IsDir(gitPath) {
			continue // Not a git repo
		}

		repo := &Repo{
			Name:     entry.Name(),
			Org:      reg.Org,
			Path:     repoPath,
			Type:     "module", // Default type
			registry: reg,
		}

		reg.Repos[entry.Name()] = repo

		// Try to detect org from first repo's remote
		if reg.Org == "" {
			if detected := detectOrg(m, repoPath); detected != "" {
				reg.Org = detected
				for _, existing := range reg.Repos {
					if existing != nil && strings.TrimSpace(existing.Org) == "" {
						existing.Org = detected
					}
				}
			}
		}
	}

	return reg, nil
}

// detectOrg tries to extract the organisation from a repo's origin remote.
func detectOrg(m io.Medium, repoPath string) string {
	// Try to read git remote
	configPath := filepath.Join(repoPath, ".git", "config")
	content, err := m.Read(configPath)
	if err != nil {
		return ""
	}
	// Look for patterns like github.com:org/repo or github.com/org/repo
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "url = ") {
			continue
		}
		url := strings.TrimPrefix(line, "url = ")

		if org := remoteOrg(url); org != "" {
			return org
		}
	}

	return ""
}

// List returns all repos in the registry.
//
//	repos := reg.List()
//
// Usage: List(...)
func (r *Registry) List() []*Repo {
	repos := make([]*Repo, 0, len(r.Repos))
	for _, repo := range r.Repos {
		repos = append(repos, repo)
	}
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Name < repos[j].Name
	})
	return repos
}

// Get returns a repo by name.
//
//	repo, ok := reg.Get("go-io")
//
// Usage: Get(...)
func (r *Registry) Get(name string) (*Repo, bool) {
	repo, ok := r.Repos[name]
	return repo, ok
}

// ByType returns repos filtered by type.
//
//	goRepos := reg.ByType("go")
//
// Usage: ByType(...)
func (r *Registry) ByType(t string) []*Repo {
	var repos []*Repo
	for _, repo := range r.Repos {
		if repo.Type == t {
			repos = append(repos, repo)
		}
	}
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Name < repos[j].Name
	})
	return repos
}

// Dependents returns repos that directly depend on the named repo.
// Usage: Dependents(...)
func (r *Registry) Dependents(name string) []*Repo {
	var dependents []*Repo
	for _, repo := range r.Repos {
		for _, dep := range repo.DependsOn {
			if dep == name {
				dependents = append(dependents, repo)
				break
			}
		}
	}
	sort.Slice(dependents, func(i, j int) bool {
		return dependents[i].Name < dependents[j].Name
	})
	return dependents
}

// Impact returns the transitive set of repos affected by changes to the named
// repo, ordered from nearest dependents outward.
// Usage: Impact(...)
func (r *Registry) Impact(name string) ([]*Repo, error) {
	if _, ok := r.Repos[name]; !ok {
		return nil, coreerr.E("repos.Registry.Impact", "unknown repo: "+name, nil)
	}

	visited := map[string]bool{name: true}
	queue := []string{name}
	var impacted []*Repo

	for len(queue) > 0 {
		target := queue[0]
		queue = queue[1:]

		for _, repo := range r.Dependents(target) {
			if repo == nil || visited[repo.Name] {
				continue
			}
			visited[repo.Name] = true
			impacted = append(impacted, repo)
			queue = append(queue, repo.Name)
		}
	}

	return impacted, nil
}

// TopologicalOrder returns repos sorted by dependency order.
// Foundation repos come first, then modules, then products.
//
//	ordered, err := reg.TopologicalOrder()
//
// Usage: TopologicalOrder(...)
func (r *Registry) TopologicalOrder() ([]*Repo, error) {
	// Build dependency graph
	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	var result []*Repo

	var visit func(name string) error
	visit = func(name string) error {
		if visited[name] {
			return nil
		}
		if visiting[name] {
			return coreerr.E("repos.Registry.TopologicalOrder", "circular dependency detected: "+name, nil)
		}

		repo, ok := r.Repos[name]
		if !ok {
			return coreerr.E("repos.Registry.TopologicalOrder", "unknown repo: "+name, nil)
		}

		visiting[name] = true
		for _, dep := range repo.DependsOn {
			if err := visit(dep); err != nil {
				return err
			}
		}
		visiting[name] = false
		visited[name] = true
		result = append(result, repo)
		return nil
	}

	names := make([]string, 0, len(r.Repos))
	for name := range r.Repos {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		if err := visit(name); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// Exists checks if the repo directory exists on disk.
// Usage: Exists(...)
func (repo *Repo) Exists() bool {
	return repo.getMedium().IsDir(repo.Path)
}

// IsGitRepo checks if the repo directory contains a .git folder.
// Usage: IsGitRepo(...)
func (repo *Repo) IsGitRepo() bool {
	gitPath := filepath.Join(repo.Path, ".git")
	return repo.getMedium().IsDir(gitPath)
}

// Registry returns the registry that finalised the repo, if known.
// Usage: Registry(...)
func (repo *Repo) Registry() *Registry {
	if repo == nil {
		return nil
	}
	return repo.registry
}

func (repo *Repo) getMedium() io.Medium {
	if repo.registry != nil && repo.registry.medium != nil {
		return repo.registry.medium
	}
	return io.Local
}

type registryMapFile struct {
	Version  int              `yaml:"version"`
	Org      string           `yaml:"org"`
	BasePath string           `yaml:"base_path"`
	Repos    map[string]*Repo `yaml:"repos"`
	Defaults RegistryDefaults `yaml:"defaults"`
}

type registryListFile struct {
	Version  int                 `yaml:"version"`
	Org      string              `yaml:"org"`
	BasePath string              `yaml:"base_path"`
	Repos    []registryListEntry `yaml:"repos"`
	Defaults RegistryDefaults    `yaml:"defaults"`
}

type registryListEntry struct {
	Name        string   `yaml:"name"`
	Path        string   `yaml:"path"`
	Remote      string   `yaml:"remote,omitempty"`
	Branch      string   `yaml:"branch,omitempty"`
	Type        string   `yaml:"type"`
	DependsOn   []string `yaml:"depends_on"`
	Description string   `yaml:"description"`
	Docs        bool     `yaml:"docs"`
	CI          string   `yaml:"ci"`
	Domain      string   `yaml:"domain,omitempty"`
	Clone       *bool    `yaml:"clone,omitempty"`
}

func decodeRegistry(data []byte) (*Registry, error) {
	var mapFile registryMapFile
	if err := yaml.Unmarshal(data, &mapFile); err == nil {
		return mapRegistryFile(mapFile), nil
	}

	var listFile registryListFile
	if err := yaml.Unmarshal(data, &listFile); err != nil {
		return nil, err
	}

	return listRegistryFile(listFile)
}

func mapRegistryFile(file registryMapFile) *Registry {
	repos := file.Repos
	if repos == nil {
		repos = make(map[string]*Repo)
	}
	return &Registry{
		Version:  file.Version,
		Org:      file.Org,
		BasePath: file.BasePath,
		Repos:    repos,
		Defaults: file.Defaults,
	}
}

func listRegistryFile(file registryListFile) (*Registry, error) {
	reg := &Registry{
		Version:  file.Version,
		Org:      file.Org,
		BasePath: file.BasePath,
		Repos:    make(map[string]*Repo, len(file.Repos)),
		Defaults: file.Defaults,
	}

	for _, entry := range file.Repos {
		repo, name, err := entry.toRepo(reg.BasePath)
		if err != nil {
			return nil, err
		}
		reg.Repos[name] = repo
	}

	return reg, nil
}

func (entry registryListEntry) toRepo(basePath string) (*Repo, string, error) {
	name := strings.TrimSpace(entry.Name)
	if name == "" {
		switch {
		case entry.Path != "":
			name = filepath.Base(entry.Path)
		case entry.Remote != "":
			name = repoNameFromRemote(entry.Remote)
		}
	}
	if name == "" {
		return nil, "", coreerr.E("repos.LoadRegistry", "repo name is required", nil)
	}

	repo := &Repo{
		Name:        name,
		Type:        entry.Type,
		Remote:      entry.Remote,
		Branch:      entry.Branch,
		DependsOn:   entry.DependsOn,
		Description: entry.Description,
		Docs:        entry.Docs,
		CI:          entry.CI,
		Domain:      entry.Domain,
		Clone:       entry.Clone,
	}
	repo.Path = normaliseRepoPath(basePath, entry.Path, name)
	return repo, name, nil
}

func finaliseRepo(reg *Registry, name string, repo *Repo) {
	if repo == nil {
		return
	}
	if repo.Name == "" {
		repo.Name = name
	}
	if repo.Org == "" {
		repo.Org = reg.Org
	}
	repo.Path = normaliseRepoPath(reg.BasePath, repo.Path, repo.Name)
	repo.registry = reg
	if repo.CI == "" {
		repo.CI = reg.Defaults.CI
	}
	if repo.Branch == "" {
		repo.Branch = reg.Defaults.Branch
	}
}

func normaliseRepoPath(basePath, rawPath, fallbackName string) string {
	pathValue := strings.TrimSpace(rawPath)
	if pathValue == "" {
		return filepath.Join(basePath, fallbackName)
	}

	pathValue = expandPath(pathValue)
	if strings.HasPrefix(pathValue, "/") {
		return pathValue
	}
	return filepath.Join(basePath, pathValue)
}

func repoNameFromRemote(remote string) string {
	parts := remotePathParts(remote)
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

// expandPath expands ~ to home directory.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

func registryCandidatesFromEnv() []string {
	raw := os.Getenv("CORE_REPOS")
	if raw == "" {
		return nil
	}

	raw = stdstrings.ReplaceAll(raw, ",", string(stdos.PathListSeparator))
	fields := stdpath.SplitList(raw)
	if len(fields) == 0 {
		return nil
	}

	var candidates []string
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field != "" {
			candidates = append(candidates, expandPath(field))
		}
	}
	return candidates
}

func remoteOrg(remote string) string {
	parts := remotePathParts(remote)
	if len(parts) < 2 {
		return ""
	}
	return parts[0]
}

func remotePathParts(remote string) []string {
	remote = strings.TrimSpace(remote)
	if remote == "" {
		return nil
	}

	remote = strings.TrimSuffix(remote, ".git")

	// SCP-style remotes: git@host:org/repo
	if !strings.Contains(remote, "://") {
		if idx := stdstrings.Index(remote, ":"); idx >= 0 && strings.Contains(remote[:idx], "@") {
			return splitRemotePath(remote[idx+1:])
		}
	}

	if parsed, err := url.Parse(remote); err == nil {
		if path := stdstrings.Trim(parsed.Path, "/"); path != "" {
			return splitRemotePath(path)
		}
	}

	return splitRemotePath(remote)
}

func splitRemotePath(path string) []string {
	path = stdstrings.Trim(path, "/")
	if path == "" {
		return nil
	}

	raw := strings.Split(path, "/")
	parts := make([]string, 0, len(raw))
	for _, part := range raw {
		part = strings.TrimSpace(part)
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}
