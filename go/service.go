// SPDX-License-Identifier: EUPL-1.2

// Service registration aliases for the scm package — exposes the
// canonical `NewService` + `Register` dual-method shape that #1336
// established across the canonical Go repo set.
//
//	c, _ := core.New(
//	    core.WithService(scm.NewService(scm.Options{
//	        Root: "/Users/snider/Code/core",
//	    })),
//	)
//	// or imperatively:
//	c := core.New()
//	scm.Register(c)
//
// The full service implementation (Service struct, Options, lifecycle
// hooks, sub-service composition) lives in scm.go. This file is the
// thin canonical-name surface so consumers calling `scm.NewService` /
// `scm.Register` get the same canon every other go-* repo provides.

package scm

import (
	core "dappco.re/go"
)

// NewService returns a factory that wires the SCM sub-services (repos,
// git) into a Core and produces a *Service ready for c.Service()
// registration. Use through core.WithService so the framework picks up
// OnStartup / OnShutdown lifecycle hooks.
//
//	core.WithService(scm.NewService(scm.Options{Root: "/code"}))
//
// Identical to NewCoreService — the rename matches the canonical
// `NewService` name used by go-store, go-process, go-i18n, etc.
// NewCoreService is kept as a deprecated alias for one cycle so
// downstream pinned consumers (api, ide, go-ai, lint, go-devops) can
// bump submodule pins on their own schedule.
func NewService(opts Options) func(*core.Core) core.Result {
	return NewCoreService(opts)
}

// Register wires the SCM service into the Core with default Options —
// the imperative-style alternative to NewService for consumers that
// don't use the WithService factory pattern.
//
//	c := core.New()
//	if r := scm.Register(c); !r.OK { return r }
//
// Equivalent to scm.NewService(scm.Options{})(c). For non-default
// configuration use NewService directly.
func Register(c *core.Core) core.Result {
	return NewService(Options{})(c)
}
