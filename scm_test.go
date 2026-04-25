// SPDX-License-Identifier: EUPL-1.2

package scm

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	core "dappco.re/go/core"
	coreio "dappco.re/go/io"
)

func TestScm_NewCoreService_Good(t *testing.T) {
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

func TestScm_NewCoreService_Bad(t *testing.T) {
	result := NewCoreService(Options{})(nil)
	if result.OK {
		t.Fatalf("expected nil core to fail")
	}
	if result.Value == nil {
		t.Fatalf("expected nil core failure to include an error")
	}
}

func TestScm_NewCoreService_Ugly(t *testing.T) {
	registryPath := filepath.Join(t.TempDir(), "repos.yaml")
	if err := os.WriteFile(registryPath, []byte("repos: ["), 0o600); err != nil {
		t.Fatalf("write malformed registry: %v", err)
	}

	c := core.New(core.WithService(NewCoreService(Options{RegistryPath: registryPath})))
	if r := c.ServiceStartup(context.Background(), nil); r.OK {
		t.Fatalf("expected malformed registry to fail startup")
	}

	if !c.Service("scm").OK {
		t.Fatalf("scm service was not registered before startup failure")
	}
	if !c.Service("repos").OK {
		t.Fatalf("repos service was not registered before startup failure")
	}
	if c.Service("git").OK {
		t.Fatalf("git service should not be registered without a root")
	}
}

func TestScm_WithMedium_Good(t *testing.T) {
	medium := coreio.NewMemoryMedium()

	registry := NewRegistry(WithMedium(medium))
	if registry == nil {
		t.Fatal("expected registry")
	}
	if registry.Medium() != medium {
		t.Fatalf("expected registry medium to be preserved")
	}
}

func TestScm_WithMedium_Bad(t *testing.T) {
	registry := NewRegistry(WithMedium(nil))
	if registry == nil {
		t.Fatal("expected registry")
	}
	if registry.Medium() != nil {
		t.Fatalf("expected nil medium to be ignored")
	}
}

func TestScm_WithMedium_Ugly(t *testing.T) {
	medium := coreio.NewMemoryMedium()

	registry := NewRegistry(WithMedium(medium), WithMedium(nil))
	if registry == nil {
		t.Fatal("expected registry")
	}
	if registry.Medium() != medium {
		t.Fatalf("expected nil medium option not to clear an existing medium")
	}
}
