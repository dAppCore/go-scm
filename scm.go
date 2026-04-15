// SPDX-License-Identifier: EUPL-1.2

// Package scm provides the top-level Core service wiring for go-scm.
package scm

import (
	"context"
	"errors"

	core "dappco.re/go/core"
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
