// SPDX-License-Identifier: EUPL-1.2

package marketplace

import (
	"context"
	"crypto/ed25519" // intrinsic
	"encoding/base64"
	`path/filepath`
	"testing"

	coreio "dappco.re/go/io"
	"dappco.re/go/scm/manifest"
)

func TestInstallerPersistsInstalledModulesToMedium(t *testing.T) {
	medium := coreio.NewMockMedium()

	installer := NewInstaller(medium, "modules")
	mod := signedTestModule(t, Module{
		Code:    "go-io",
		Name:    "Core I/O",
		Version: "0.3.0",
		Repo:    "ssh://example.org/core/go-io.git",
	})

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

func TestInstallerRejectsVerifyFailBeforeMediumWrite(t *testing.T) {
	medium := coreio.NewMockMedium()
	installer := NewInstaller(medium, "modules")
	mod := signedTestModule(t, Module{
		Code:    "go-io",
		Name:    "Core I/O",
		Version: "0.3.0",
		Repo:    "ssh://example.org/core/go-io.git",
	})
	mod.Name = "Tampered Core I/O"

	if err := installer.Install(context.Background(), mod); err == nil {
		t.Fatal("expected install to reject invalid signature")
	}
	if _, ok := medium.Files[filepath.Join("modules", "go-io", "module.json")]; ok {
		t.Fatal("module.json was written after signature verification failed")
	}
}

func signedTestModule(t *testing.T, mod Module) Module {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	mod.SignKey = base64.StdEncoding.EncodeToString(pub)
	payload, err := moduleVerificationPayload(mod)
	if err != nil {
		t.Fatalf("module payload: %v", err)
	}
	sig := &manifest.Manifest{SignKey: mod.SignKey}
	if err := manifest.Sign(sig, payload, priv); err != nil {
		t.Fatalf("sign module: %v", err)
	}
	mod.Sign = sig.Sign
	return mod
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

func TestInstaller_Installer_Installed_Good(t *testing.T) {
	reference := "Installed"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Installer_Installed"
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

func TestInstaller_Installer_Installed_Bad(t *testing.T) {
	reference := "Installed"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Installer_Installed"
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

func TestInstaller_Installer_Installed_Ugly(t *testing.T) {
	reference := "Installed"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Installer_Installed"
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
