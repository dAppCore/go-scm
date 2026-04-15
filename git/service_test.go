// SPDX-License-Identifier: EUPL-1.2

package git

import (
	"context"
	"testing"

	core "dappco.re/go/core"
)

func TestServiceRegistersActionsWithoutWorkDir(t *testing.T) {
	c := core.New(core.WithService(NewService(ServiceOptions{})))
	if r := c.ServiceStartup(context.Background(), nil); !r.OK {
		t.Fatalf("service startup failed: %v", r.Value)
	}

	for _, name := range []string{
		"git.push",
		"git.pull",
		"git.push-multiple",
		"git.pull-multiple",
	} {
		if !c.Action(name).Exists() {
			t.Fatalf("expected %s to be registered", name)
		}
	}
}
