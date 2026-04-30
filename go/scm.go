// SPDX-License-Identifier: EUPL-1.2

// Package scm provides the top-level Core service wiring for go-scm.
package scm

import (
	// Note: AX-6 — Core lifecycle hooks use context.Context directly.
	"context"

	core "dappco.re/go"
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
			return core.Fail(core.E("scm.NewCoreService", "core is required", nil))
		}
		if result := registerReposService(c, opts); !result.OK {
			return result
		}
		if result := registerGitService(c, opts); !result.OK {
			return result
		}
		return core.Ok(&Service{ServiceRuntime: core.NewServiceRuntime(c, opts)})
	}
}

func registerReposService(c *core.Core, opts Options) core.Result {
	if opts.RegistryPath == "" && opts.Root == "" && opts.Remote == "" && opts.Branch == "" {
		return core.Ok(nil)
	}
	result := repos.NewService(repos.ServiceOptions{
		Root:         opts.Root,
		RegistryPath: opts.RegistryPath,
		Remote:       opts.Remote,
		Branch:       opts.Branch,
	})(c)
	if !result.OK {
		return result
	}
	return registerServiceIfMissing(c, "repos", result.Value)
}

func registerGitService(c *core.Core, opts Options) core.Result {
	if opts.Root == "" {
		return core.Ok(nil)
	}
	result := git.NewService(git.ServiceOptions{WorkDir: opts.Root})(c)
	if !result.OK {
		return result
	}
	return registerServiceIfMissing(c, "git", result.Value)
}

func registerServiceIfMissing(c *core.Core, name string, service any) core.Result {
	if service == nil || c.Service(name).OK {
		return core.Ok(nil)
	}
	return c.RegisterService(name, service)
}

// OnStartup satisfies the Core lifecycle contract.
func (s *Service) OnStartup(ctx context.Context) core.Result {
	return serviceStartupResult(s, ctx)
}

// OnShutdown satisfies the Core lifecycle contract.
func (s *Service) OnShutdown(ctx context.Context) core.Result {
	return serviceLifecycleResult(s, ctx)
}

func serviceStartupResult(s *Service, ctx context.Context) core.Result {
	return serviceLifecycleResult(s, ctx)
}

func serviceLifecycleResult(s *Service, ctx context.Context) core.Result {
	if s == nil {
		return core.Ok(nil)
	}
	if err := ctx.Err(); err != nil {
		return core.Fail(err)
	}
	return core.Ok(nil)
}
