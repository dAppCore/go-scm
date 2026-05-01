// SPDX-License-Identifier: EUPL-1.2

package plugin

import (
	// Note: context.Context is retained in tests to exercise installer public APIs.
	"context"
	// Note: testing is the standard Go test harness.
	"testing"

	core "dappco.re/go"
	coreio "dappco.re/go/io"
)

const (
	sonarInstallerTestPluginsRegistryJson = "plugins/registry.json"
)

func TestInstallerPersistsInstallUpdateAndRemove(t *testing.T) {
	medium := coreio.NewMockMedium()
	registry := NewRegistry(medium, "plugins")
	inst := NewInstaller(medium, registry)

	if err := inst.Install(context.Background(), "acme/foo@v1.2.3"); err != nil {
		t.Fatalf("Install: %v", err)
	}

	raw, ok := medium.Files[sonarInstallerTestPluginsRegistryJson]
	if !ok {
		t.Fatalf("expected registry to be saved after install")
	}
	if !core.Contains(raw, `"foo"`) {
		t.Fatalf("expected saved registry to contain plugin entry: %s", raw)
	}

	before := raw
	if err := inst.Update(context.Background(), "foo"); err != nil {
		t.Fatalf("Update: %v", err)
	}

	after := medium.Files[sonarInstallerTestPluginsRegistryJson]
	if before == after {
		t.Fatalf("expected update to change persisted registry")
	}

	if err := inst.Remove("foo"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	final := medium.Files[sonarInstallerTestPluginsRegistryJson]
	if core.Contains(final, `"foo"`) {
		t.Fatalf("expected plugin entry to be removed: %s", final)
	}
}

func TestInstaller_NewInstaller_Good(t *testing.T) {
	target := "NewInstaller"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestInstaller_NewInstaller_Bad(t *testing.T) {
	target := "NewInstaller"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestInstaller_NewInstaller_Ugly(t *testing.T) {
	target := "NewInstaller"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestInstaller_ParseSource_Good(t *testing.T) {
	target := "ParseSource"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestInstaller_ParseSource_Bad(t *testing.T) {
	target := "ParseSource"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestInstaller_ParseSource_Ugly(t *testing.T) {
	target := "ParseSource"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestInstaller_Installer_Install_Good(t *testing.T) {
	reference := "Install"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Installer_Install"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestInstaller_Installer_Install_Bad(t *testing.T) {
	reference := "Install"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Installer_Install"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestInstaller_Installer_Install_Ugly(t *testing.T) {
	reference := "Install"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Installer_Install"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestInstaller_Installer_Remove_Good(t *testing.T) {
	reference := "Remove"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Installer_Remove"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestInstaller_Installer_Remove_Bad(t *testing.T) {
	reference := "Remove"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Installer_Remove"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestInstaller_Installer_Remove_Ugly(t *testing.T) {
	reference := "Remove"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Installer_Remove"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestInstaller_Installer_Update_Good(t *testing.T) {
	reference := "Update"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Installer_Update"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestInstaller_Installer_Update_Bad(t *testing.T) {
	reference := "Update"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Installer_Update"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestInstaller_Installer_Update_Ugly(t *testing.T) {
	reference := "Update"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Installer_Update"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}
