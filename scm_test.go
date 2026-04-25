// SPDX-License-Identifier: EUPL-1.2

package scm

import (
	"context"
	"testing"

	core "dappco.re/go/core"
	coreio "dappco.re/go/io"
)

func TestNewCoreServiceRegistersRootAndSubservices(t *testing.T) {
	c := core.New(core.WithService(NewCoreService(Options{Root: t.TempDir()})))
	if r := c.ServiceStartup(context.Background(), nil); !r.OK {
		t.Fatalf("service startup failed: %v", r.Value)
	}

	if !c.Service("scm").OK {
		t.Fatalf("scm service was not registered")
	}
	if !c.Service("repos").OK {
		t.Fatalf("repos service was not registered")
	}
	if !c.Service("git").OK {
		t.Fatalf("git service was not registered")
	}
}

func TestRegistry_WithMedium_Good(t *testing.T) {
	medium := coreio.NewMemoryMedium()

	registry := NewRegistry(WithMedium(medium))
	if registry == nil {
		t.Fatal("expected registry")
	}
	if registry.Medium() != medium {
		t.Fatalf("expected registry medium to be preserved")
	}
}

func TestRegistry_WithMedium_Bad_NilMedium(t *testing.T) {
	registry := NewRegistry(WithMedium(nil))
	if registry == nil {
		t.Fatal("expected registry")
	}
	if registry.Medium() != nil {
		t.Fatalf("expected nil medium to be ignored")
	}
}
