// SPDX-License-Identifier: EUPL-1.2

package plugin

import (
	// Note: context.Context is retained in tests to exercise installer public APIs.
	"context"
	// Note: strings.Contains is retained for assertions against persisted registry YAML.
	"strings"
	// Note: testing is the standard Go test harness.
	"testing"

	coreio "dappco.re/go/core/io"
)

func TestInstallerPersistsInstallUpdateAndRemove(t *testing.T) {
	medium := coreio.NewMockMedium()
	registry := NewRegistry(medium, "plugins")
	inst := NewInstaller(medium, registry)

	if err := inst.Install(context.Background(), "acme/foo@v1.2.3"); err != nil {
		t.Fatalf("Install: %v", err)
	}

	raw, ok := medium.Files["plugins/registry.json"]
	if !ok {
		t.Fatalf("expected registry to be saved after install")
	}
	if !strings.Contains(raw, `"foo"`) {
		t.Fatalf("expected saved registry to contain plugin entry: %s", raw)
	}

	before := raw
	if err := inst.Update(context.Background(), "foo"); err != nil {
		t.Fatalf("Update: %v", err)
	}

	after := medium.Files["plugins/registry.json"]
	if before == after {
		t.Fatalf("expected update to change persisted registry")
	}

	if err := inst.Remove("foo"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	final := medium.Files["plugins/registry.json"]
	if strings.Contains(final, `"foo"`) {
		t.Fatalf("expected plugin entry to be removed: %s", final)
	}
}
