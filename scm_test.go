// SPDX-License-Identifier: EUPL-1.2

package scm

import (
	"context"
	"testing"

	core "dappco.re/go/core"
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
