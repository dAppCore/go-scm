// SPDX-License-Identifier: EUPL-1.2

package marketplace

import (
	"context"
	"path/filepath"
	"testing"

	coreio "dappco.re/go/io"
)

func TestInstallerPersistsInstalledModulesToMedium(t *testing.T) {
	medium := coreio.NewMockMedium()

	installer := NewInstaller(medium, "modules")
	mod := Module{
		Code:    "go-io",
		Name:    "Core I/O",
		Version: "0.3.0",
		Repo:    "ssh://example.org/core/go-io.git",
		SignKey: "ed25519:public-key",
	}

	if err := installer.Install(context.Background(), mod); err != nil {
		t.Fatalf("install: %v", err)
	}

	raw, ok := medium.Files[filepath.Join("modules", "go-io", "module.json")]
	if !ok || raw == "" {
		t.Fatalf("expected module.json to be written")
	}

	installed, err := installer.Installed()
	if err != nil {
		t.Fatalf("installed: %v", err)
	}
	if len(installed) != 1 || installed[0].Code != "go-io" {
		t.Fatalf("unexpected installed modules: %#v", installed)
	}
	if installed[0].Version != "0.3.0" {
		t.Fatalf("expected installed version to be preserved, got %#v", installed[0].Version)
	}

	if err := installer.Update(context.Background(), "go-io"); err != nil {
		t.Fatalf("update: %v", err)
	}

	updated, err := installer.Installed()
	if err != nil {
		t.Fatalf("installed after update: %v", err)
	}
	if len(updated) != 1 || updated[0].Code != "go-io" {
		t.Fatalf("unexpected updated modules: %#v", updated)
	}
	if updated[0].Version != "0.3.0" {
		t.Fatalf("expected updated version to be preserved, got %#v", updated[0].Version)
	}

	if err := installer.Remove("go-io"); err != nil {
		t.Fatalf("remove: %v", err)
	}
	installed, err = installer.Installed()
	if err != nil {
		t.Fatalf("installed after remove: %v", err)
	}
	if len(installed) != 0 {
		t.Fatalf("expected no installed modules after remove, got %#v", installed)
	}
}
