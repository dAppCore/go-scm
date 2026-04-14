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

// RegistryOption aliases the repos.RegistryOption type for root-package callers.
type RegistryOption = repos.RegistryOption

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
