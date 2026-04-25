// SPDX-License-Identifier: EUPL-1.2

// Package scm provides the top-level Core service wiring for go-scm.
package scm

import (
	// Note: AX-6 — Core lifecycle hooks use context.Context directly.
	"context"
	// Note: AX-6 — Constructor failures return standard error values through core.Result.
	"errors"

	core "dappco.re/go/core"
	coreio "dappco.re/go/io"
	"dappco.re/go/scm/git"
	"dappco.re/go/scm/repos"
)

// Options configures the top-level SCM service.
//
// Root enables workspace-local git operations.
// RegistryPath lets callers point at a specific repos.yaml file.
type Options struct {
	Root         string
	RegistryPath string
	Remote       string
	Branch       string
}

// Service is the root SCM service. It provides a single place to hang
// repo registry, git status, and sync wiring under the Core runtime.
type Service struct {
	*core.ServiceRuntime[Options]
}

// RegistryOption configures a package registry.
type RegistryOption func(*Registry)

// Registry holds the package registry storage backend.
type Registry struct {
	medium coreio.Medium
}

// NewRegistry creates a package registry using the supplied options.
func NewRegistry(opts ...RegistryOption) *Registry {
	r := &Registry{}
	for _, opt := range opts {
		if opt != nil {
			opt(r)
		}
	}
	return r
}

// WithMedium configures the registry storage backend.
func WithMedium(m coreio.Medium) RegistryOption {
	return func(r *Registry) {
		if r == nil || m == nil {
			return
		}
		r.medium = m
	}
}

// Medium returns the registry storage backend.
func (r *Registry) Medium() coreio.Medium {
	if r == nil {
		return nil
	}
	return r.medium
}

// NewCoreService creates a Core service factory that registers the SCM
// sub-services needed by the package-level workflow.
func NewCoreService(opts Options) func(*core.Core) core.Result {
	return func(c *core.Core) core.Result {
		if c == nil {
			return core.Result{Value: errors.New("scm.NewCoreService: core is required"), OK: false}
		}

		if opts.RegistryPath != "" || opts.Root != "" || opts.Remote != "" || opts.Branch != "" {
			repoResult := repos.NewService(repos.ServiceOptions{
				Root:         opts.Root,
				RegistryPath: opts.RegistryPath,
				Remote:       opts.Remote,
				Branch:       opts.Branch,
			})(c)
			if !repoResult.OK {
				return repoResult
			}
			if repoResult.Value != nil && !c.Service("repos").OK {
				if r := c.RegisterService("repos", repoResult.Value); !r.OK {
					return r
				}
			}
		}

		if opts.Root != "" {
			gitResult := git.NewService(git.ServiceOptions{WorkDir: opts.Root})(c)
			if !gitResult.OK {
				return gitResult
			}
			if gitResult.Value != nil && !c.Service("git").OK {
				if r := c.RegisterService("git", gitResult.Value); !r.OK {
					return r
				}
			}
		}

		return core.Result{
			Value: &Service{ServiceRuntime: core.NewServiceRuntime(c, opts)},
			OK:    true,
		}
	}
}

// OnStartup satisfies the Core lifecycle contract.
func (s *Service) OnStartup(ctx context.Context) core.Result {
	if s == nil {
		return core.Result{OK: true}
	}
	if err := ctx.Err(); err != nil {
		return core.Result{Value: err, OK: false}
	}
	return core.Result{OK: true}
}

// OnShutdown satisfies the Core lifecycle contract.
func (s *Service) OnShutdown(ctx context.Context) core.Result {
	if s == nil {
		return core.Result{OK: true}
	}
	if err := ctx.Err(); err != nil {
		return core.Result{Value: err, OK: false}
	}
	return core.Result{OK: true}
}
