// SPDX-License-Identifier: EUPL-1.2

// Package scm exposes top-level conveniences for working with workspace
// registries from the root module path.
package scm

import (
	"dappco.re/go/core/io"
	"dappco.re/go/core/scm/repos"
)

// Registry aliases the repos.Registry type for root-package callers.
type Registry = repos.Registry

// Repo aliases the repos.Repo type for root-package callers.
type Repo = repos.Repo

// RegistryOption aliases the repos.RegistryOption type for root-package callers.
type RegistryOption = repos.RegistryOption

// RegistryDefaults aliases the repos.RegistryDefaults type for root-package callers.
type RegistryDefaults = repos.RegistryDefaults

// RepoType aliases the repos.RepoType type for root-package callers.
type RepoType = repos.RepoType

const (
	// RepoTypeFoundation indicates core foundation packages.
	RepoTypeFoundation = repos.RepoTypeFoundation
	// RepoTypeModule indicates reusable module packages.
	RepoTypeModule = repos.RepoTypeModule
	// RepoTypeProduct indicates end-user product applications.
	RepoTypeProduct = repos.RepoTypeProduct
	// RepoTypeTemplate indicates starter templates.
	RepoTypeTemplate = repos.RepoTypeTemplate
)

// NewRegistry creates an empty registry with the supplied options.
// Usage: NewRegistry(...)
func NewRegistry(opts ...RegistryOption) *Registry {
	return repos.NewRegistry(opts...)
}

// WithMedium configures the filesystem medium used by the registry.
// Usage: WithMedium(...)
func WithMedium(m io.Medium) RegistryOption {
	return repos.WithMedium(m)
}

// LoadRegistry reads and parses a repos.yaml file from the given medium.
// Usage: LoadRegistry(...)
func LoadRegistry(m io.Medium, path string) (*Registry, error) {
	return repos.LoadRegistry(m, path)
}

// FindRegistry discovers a single repos.yaml path using the standard search order.
// Usage: FindRegistry(...)
func FindRegistry(m io.Medium) (string, error) {
	return repos.FindRegistry(m)
}

// FindRegistries discovers all repos.yaml paths using the standard search order.
// Usage: FindRegistries(...)
func FindRegistries(m io.Medium) ([]string, error) {
	return repos.FindRegistries(m)
}

// LoadRegistries discovers and loads all registry files.
// Usage: LoadRegistries(...)
func LoadRegistries(m io.Medium) ([]*Registry, error) {
	return repos.LoadRegistries(m)
}

// MergeRegistries combines registry views into a single deduplicated registry.
// Usage: MergeRegistries(...)
func MergeRegistries(regs ...*Registry) *Registry {
	return repos.MergeRegistries(regs...)
}

// ScanDirectory discovers git repositories under a directory and builds a registry view.
// Usage: ScanDirectory(...)
func ScanDirectory(m io.Medium, dir string) (*Registry, error) {
	return repos.ScanDirectory(m, dir)
}
