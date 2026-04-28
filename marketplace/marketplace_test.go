// SPDX-License-Identifier: EUPL-1.2

package marketplace

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	core "dappco.re/go"
	coreio "dappco.re/go/io"
	"dappco.re/go/scm/manifest"
)

func TestBuildIndexFromManifestsCarriesSignKey(t *testing.T) {
	idx := BuildIndexFromManifests([]*manifest.Manifest{
		{
			Code:    "go-io",
			Name:    "Core I/O",
			Layout:  "core",
			SignKey: "ed25519:public-key",
		},
	})

	mod, ok := idx.Find("go-io")
	if !ok {
		t.Fatalf("expected module to be indexed")
	}
	if mod.SignKey != "ed25519:public-key" {
		t.Fatalf("unexpected sign key: %q", mod.SignKey)
	}
}

func ax7MarketplaceManifest(code string) *manifest.Manifest {
	return &manifest.Manifest{Code: code, Name: "Demo " + code, Version: "1.0.0", Modules: []string{"provider"}}
}

func ax7SignedModule(t *core.T) Module {
	pub, priv, err := ed25519.GenerateKey(nil)
	core.RequireNoError(t, err)
	mod := Module{Code: "demo", Name: "Demo", Version: "1.0.0", Repo: "https://forge.example/core/demo", SignKey: base64.StdEncoding.EncodeToString(pub), Category: "provider"}
	payload, err := moduleVerificationPayload(mod)
	core.RequireNoError(t, err)
	mod.Sign = base64.StdEncoding.EncodeToString(ed25519.Sign(priv, payload))
	return mod
}

func TestMarketplace_ParseIndex_Good(t *core.T) {
	idx, err := ParseIndex([]byte(`{"version":1,"modules":[{"code":"demo","name":"Demo"}]}`))
	core.AssertNoError(t, err)
	core.AssertEqual(t, "demo", idx.Modules[0].Code)
}

func TestMarketplace_ParseIndex_Bad(t *core.T) {
	_, err := ParseIndex([]byte(`{"version":`))
	core.AssertError(t, err)
}

func TestMarketplace_ParseIndex_Ugly(t *core.T) {
	idx, err := ParseIndex(nil)
	core.AssertError(t, err)
	core.AssertNil(t, idx)
}

func TestMarketplace_Index_Find_Good(t *core.T) {
	idx := &Index{Modules: []Module{{Code: "Demo", Name: "Demo"}}}
	mod, ok := idx.Find("demo")
	core.AssertTrue(t, ok)
	core.AssertEqual(t, "Demo", mod.Code)
}

func TestMarketplace_Index_Find_Bad(t *core.T) {
	idx := &Index{Modules: []Module{{Code: "demo"}}}
	_, ok := idx.Find("missing")
	core.AssertFalse(t, ok)
}

func TestMarketplace_Index_Find_Ugly(t *core.T) {
	var idx *Index
	_, ok := idx.Find("demo")
	core.AssertFalse(t, ok)
}

func TestMarketplace_Index_ByCategory_Good(t *core.T) {
	idx := &Index{Modules: []Module{{Code: "demo", Category: "Provider"}}}
	got := idx.ByCategory("provider")
	core.AssertLen(t, got, 1)
	core.AssertEqual(t, "demo", got[0].Code)
}

func TestMarketplace_Index_ByCategory_Bad(t *core.T) {
	idx := &Index{Modules: []Module{{Code: "demo", Category: "provider"}}}
	got := idx.ByCategory("tool")
	core.AssertEmpty(t, got)
}

func TestMarketplace_Index_ByCategory_Ugly(t *core.T) {
	var idx *Index
	got := idx.ByCategory("provider")
	core.AssertNil(t, got)
}

func TestMarketplace_Index_Search_Good(t *core.T) {
	idx := &Index{Modules: []Module{{Code: "demo", Name: "Demo Provider", Category: "provider"}}}
	got := idx.Search("provider")
	core.AssertLen(t, got, 1)
	core.AssertEqual(t, "demo", got[0].Code)
}

func TestMarketplace_Index_Search_Bad(t *core.T) {
	idx := &Index{Modules: []Module{{Code: "demo", Name: "Demo"}}}
	got := idx.Search("missing")
	core.AssertEmpty(t, got)
}

func TestMarketplace_Index_Search_Ugly(t *core.T) {
	idx := &Index{Modules: []Module{{Code: "demo"}, {Code: "tool"}}}
	got := idx.Search("  ")
	core.AssertLen(t, got, 2)
}

func TestMarketplace_BuildIndexFromManifests_Good(t *core.T) {
	idx := BuildIndexFromManifests([]*manifest.Manifest{ax7MarketplaceManifest("b"), ax7MarketplaceManifest("a")})
	core.AssertEqual(t, 1, idx.Version)
	core.AssertEqual(t, []string{"provider"}, idx.Categories)
	core.AssertEqual(t, "a", idx.Modules[0].Code)
}

func TestMarketplace_BuildIndexFromManifests_Bad(t *core.T) {
	idx := BuildIndexFromManifests(nil)
	core.AssertEqual(t, 1, idx.Version)
	core.AssertEmpty(t, idx.Modules)
}

func TestMarketplace_BuildIndexFromManifests_Ugly(t *core.T) {
	idx := BuildIndexFromManifests([]*manifest.Manifest{nil, &manifest.Manifest{Code: "demo", Name: "Demo"}})
	core.AssertEqual(
		t, "latest", idx.Modules[0].Version,
	)
}

func TestMarketplace_BuildFromManifests_Good(t *core.T) {
	idx := BuildFromManifests([]*manifest.Manifest{ax7MarketplaceManifest("demo")})
	core.AssertEqual(
		t, "demo", idx.Modules[0].Code,
	)
}

func TestMarketplace_BuildFromManifests_Bad(t *core.T) {
	idx := BuildFromManifests(nil)
	core.AssertEqual(t, 1, idx.Version)
	core.AssertEmpty(t, idx.Modules)
}

func TestMarketplace_BuildFromManifests_Ugly(t *core.T) {
	idx := BuildFromManifests([]*manifest.Manifest{{Code: "demo", Name: "Demo", Layout: "tool"}})
	core.AssertEqual(
		t, "tool", idx.Modules[0].Category,
	)
}

func TestMarketplace_Builder_BuildFromDirs_Good(t *core.T) {
	root := t.TempDir()
	pkg := filepath.Join(root, "demo", ".core")
	core.RequireNoError(t, os.MkdirAll(pkg, 0o755))
	core.RequireNoError(t, os.WriteFile(filepath.Join(pkg, "manifest.yaml"), []byte("code: demo\nname: Demo\nversion: 1.0.0\nmodules: [provider]\n"), 0o600))
	idx, err := (&Builder{BaseURL: "https://forge.example", Org: "core"}).BuildFromDirs(root)
	core.AssertNoError(t, err)
	core.AssertEqual(t, "https://forge.example/core/demo", idx.Modules[0].Repo)
}

func TestMarketplace_Builder_BuildFromDirs_Bad(t *core.T) {
	_, err := (&Builder{}).BuildFromDirs(filepath.Join(t.TempDir(), "missing"))
	core.AssertError(
		t, err,
	)
}

func TestMarketplace_Builder_BuildFromDirs_Ugly(t *core.T) {
	root := t.TempDir()
	idx, err := (&Builder{}).BuildFromDirs(root)
	core.AssertNoError(t, err)
	core.AssertEmpty(t, idx.Modules)
}

func TestMarketplace_WriteIndex_Good(t *core.T) {
	path := filepath.Join(t.TempDir(), "index.json")
	err := WriteIndex(path, &Index{Version: 1, Modules: []Module{{Code: "demo"}}})
	core.AssertNoError(t, err)
	raw, readErr := os.ReadFile(path)
	core.RequireNoError(t, readErr)
	core.AssertContains(t, string(raw), "demo")
}

func TestMarketplace_WriteIndex_Bad(t *core.T) {
	err := WriteIndex(filepath.Join(t.TempDir(), "index.json"), nil)
	core.AssertError(
		t, err,
	)
}

func TestMarketplace_WriteIndex_Ugly(t *core.T) {
	err := WriteIndex(filepath.Join(t.TempDir(), "missing", "index.json"), &Index{Version: 1})
	core.AssertError(
		t, err,
	)
}

func TestMarketplace_LoadIndex_Good(t *core.T) {
	medium := coreio.NewMemoryMedium()
	core.RequireNoError(t, medium.Write("marketplace/index.json", `{"version":1,"modules":[{"code":"demo"}]}`))
	idx, err := LoadIndex(medium, "marketplace/index.json")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "demo", idx.Modules[0].Code)
}

func TestMarketplace_LoadIndex_Bad(t *core.T) {
	_, err := LoadIndex(nil, "marketplace/index.json")
	core.AssertError(
		t, err,
	)
}

func TestMarketplace_LoadIndex_Ugly(t *core.T) {
	idx, err := LoadIndex(coreio.NewMemoryMedium(), "missing.json")
	core.AssertNoError(t, err)
	core.AssertEqual(t, 1, idx.Version)
}

func TestMarketplace_WriteIndexToMedium_Good(t *core.T) {
	medium := coreio.NewMemoryMedium()
	err := WriteIndexToMedium(medium, "marketplace/index.json", &Index{Version: 1, Modules: []Module{{Code: "demo"}}})
	core.AssertNoError(t, err)
	raw, readErr := medium.Read("marketplace/index.json")
	core.RequireNoError(t, readErr)
	core.AssertContains(t, raw, "demo")
}

func TestMarketplace_WriteIndexToMedium_Bad(t *core.T) {
	err := WriteIndexToMedium(nil, "marketplace/index.json", &Index{Version: 1})
	core.AssertError(
		t, err,
	)
}

func TestMarketplace_WriteIndexToMedium_Ugly(t *core.T) {
	err := WriteIndexToMedium(coreio.NewMemoryMedium(), "marketplace/index.json", nil)
	core.AssertError(
		t, err,
	)
}

func TestMarketplace_NewInstaller_Good(t *core.T) {
	medium := coreio.NewMemoryMedium()
	installer := NewInstaller(medium, "modules")
	core.AssertNotNil(t, installer)
	core.AssertEqual(t, "modules", installer.modulesDir)
}

func TestMarketplace_NewInstaller_Bad(t *core.T) {
	installer := NewInstaller(nil, "modules")
	core.AssertNotNil(t, installer)
	core.AssertNil(t, installer.medium)
}

func TestMarketplace_NewInstaller_Ugly(t *core.T) {
	installer := NewInstaller(coreio.NewMemoryMedium(), "")
	core.AssertNotNil(t, installer)
	core.AssertEqual(t, "", installer.modulesDir)
}

func TestMarketplace_Installer_Install_Good(t *core.T) {
	medium := coreio.NewMemoryMedium()
	installer := NewInstaller(medium, "modules")
	err := installer.Install(context.Background(), ax7SignedModule(t))
	core.AssertNoError(t, err)
	raw, readErr := medium.Read("modules/demo/module.json")
	core.RequireNoError(t, readErr)
	core.AssertContains(t, raw, "demo")
}

func TestMarketplace_Installer_Install_Bad(t *core.T) {
	err := NewInstaller(coreio.NewMemoryMedium(), "modules").Install(context.Background(), Module{})
	core.AssertError(
		t, err,
	)
}

func TestMarketplace_Installer_Install_Ugly(t *core.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := NewInstaller(coreio.NewMemoryMedium(), "modules").Install(ctx, ax7SignedModule(t))
	core.AssertErrorIs(t, err, context.Canceled)
}

func TestMarketplace_Installer_Installed_Good(t *core.T) {
	medium := coreio.NewMemoryMedium()
	installer := NewInstaller(medium, "modules")
	core.RequireNoError(t, installer.Install(context.Background(), ax7SignedModule(t)))
	got, err := installer.Installed()
	core.AssertNoError(t, err)
	core.AssertLen(t, got, 1)
}

func TestMarketplace_Installer_Installed_Bad(t *core.T) {
	var installer *Installer
	got, err := installer.Installed()
	core.AssertNoError(t, err)
	core.AssertNil(t, got)
}

func TestMarketplace_Installer_Installed_Ugly(t *core.T) {
	got, err := NewInstaller(coreio.NewMockMedium(), "missing").Installed()
	core.AssertNoError(t, err)
	core.AssertEmpty(t, got)
}

func TestMarketplace_Installer_Remove_Good(t *core.T) {
	medium := coreio.NewMemoryMedium()
	installer := NewInstaller(medium, "modules")
	core.RequireNoError(t, installer.Install(context.Background(), ax7SignedModule(t)))
	err := installer.Remove("demo")
	core.AssertNoError(t, err)
	_, readErr := medium.Read("modules/demo/module.json")
	core.AssertError(t, readErr)
}

func TestMarketplace_Installer_Remove_Bad(t *core.T) {
	var installer *Installer
	err := installer.Remove("demo")
	core.AssertError(t, err)
}

func TestMarketplace_Installer_Remove_Ugly(t *core.T) {
	err := NewInstaller(coreio.NewMemoryMedium(), "modules").Remove("")
	core.AssertError(
		t, err,
	)
}

func TestMarketplace_Installer_Update_Good(t *core.T) {
	medium := coreio.NewMemoryMedium()
	installer := NewInstaller(medium, "modules")
	core.RequireNoError(t, installer.Install(context.Background(), ax7SignedModule(t)))
	err := installer.Update(context.Background(), "demo")
	core.AssertNoError(t, err)
}

func TestMarketplace_Installer_Update_Bad(t *core.T) {
	err := NewInstaller(coreio.NewMemoryMedium(), "modules").Update(context.Background(), "missing")
	core.AssertError(
		t, err,
	)
}

func TestMarketplace_Installer_Update_Ugly(t *core.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := NewInstaller(coreio.NewMemoryMedium(), "modules").Update(ctx, "demo")
	core.AssertErrorIs(t, err, context.Canceled)
}

func TestMarketplace_ProviderRegistryFile_Add_Good(t *core.T) {
	reg := &ProviderRegistryFile{}
	reg.Add("demo", ProviderRegistryEntry{Version: "1.0.0"})
	entry, ok := reg.Get("demo")
	core.AssertTrue(t, ok)
	core.AssertEqual(t, "1.0.0", entry.Version)
}

func TestMarketplace_ProviderRegistryFile_Add_Bad(t *core.T) {
	var reg *ProviderRegistryFile
	core.AssertNotPanics(t, func() {
		reg.Add("demo", ProviderRegistryEntry{})
	})
}

func TestMarketplace_ProviderRegistryFile_Add_Ugly(t *core.T) {
	reg := &ProviderRegistryFile{Providers: map[string]ProviderRegistryEntry{"demo": {Version: "1"}}}
	reg.Add("demo", ProviderRegistryEntry{Version: "2"})
	entry, ok := reg.Get("demo")
	core.AssertTrue(t, ok)
	core.AssertEqual(t, "2", entry.Version)
}

func TestMarketplace_ProviderRegistryFile_Get_Good(t *core.T) {
	reg := &ProviderRegistryFile{Providers: map[string]ProviderRegistryEntry{"demo": {Installed: "path"}}}
	entry, ok := reg.Get("demo")
	core.AssertTrue(t, ok)
	core.AssertEqual(t, "path", entry.Installed)
}

func TestMarketplace_ProviderRegistryFile_Get_Bad(t *core.T) {
	reg := &ProviderRegistryFile{Providers: map[string]ProviderRegistryEntry{}}
	_, ok := reg.Get("missing")
	core.AssertFalse(t, ok)
}

func TestMarketplace_ProviderRegistryFile_Get_Ugly(t *core.T) {
	var reg *ProviderRegistryFile
	_, ok := reg.Get("demo")
	core.AssertFalse(t, ok)
}

func TestMarketplace_ProviderRegistryFile_List_Good(t *core.T) {
	reg := &ProviderRegistryFile{Providers: map[string]ProviderRegistryEntry{"b": {}, "a": {}}}
	got := reg.List()
	core.AssertEqual(t, []string{"a", "b"}, got)
}

func TestMarketplace_ProviderRegistryFile_List_Bad(t *core.T) {
	reg := &ProviderRegistryFile{}
	got := reg.List()
	core.AssertEmpty(t, got)
}

func TestMarketplace_ProviderRegistryFile_List_Ugly(t *core.T) {
	var reg *ProviderRegistryFile
	got := reg.List()
	core.AssertNil(t, got)
}

func TestMarketplace_ProviderRegistryFile_AutoStartProviders_Good(t *core.T) {
	reg := &ProviderRegistryFile{Providers: map[string]ProviderRegistryEntry{"b": {AutoStart: true}, "a": {AutoStart: true}, "c": {}}}
	got := reg.AutoStartProviders()
	core.AssertEqual(t, []string{"a", "b"}, got)
}

func TestMarketplace_ProviderRegistryFile_AutoStartProviders_Bad(t *core.T) {
	reg := &ProviderRegistryFile{Providers: map[string]ProviderRegistryEntry{"demo": {}}}
	got := reg.AutoStartProviders()
	core.AssertEmpty(t, got)
}

func TestMarketplace_ProviderRegistryFile_AutoStartProviders_Ugly(t *core.T) {
	var reg *ProviderRegistryFile
	got := reg.AutoStartProviders()
	core.AssertNil(t, got)
}

func TestMarketplace_ProviderRegistryFile_Remove_Good(t *core.T) {
	reg := &ProviderRegistryFile{Providers: map[string]ProviderRegistryEntry{"demo": {}}}
	reg.Remove("demo")
	_, ok := reg.Get("demo")
	core.AssertFalse(t, ok)
}

func TestMarketplace_ProviderRegistryFile_Remove_Bad(t *core.T) {
	var reg *ProviderRegistryFile
	core.AssertNotPanics(t, func() {
		reg.Remove("demo")
	})
}

func TestMarketplace_ProviderRegistryFile_Remove_Ugly(t *core.T) {
	reg := &ProviderRegistryFile{Providers: map[string]ProviderRegistryEntry{}}
	reg.Remove("")
	core.AssertEmpty(t, reg.Providers)
}

func TestMarketplace_DiscoverProviders_Good(t *core.T) {
	root := t.TempDir()
	pkg := filepath.Join(root, "demo", ".core")
	core.RequireNoError(t, os.MkdirAll(pkg, 0o755))
	core.RequireNoError(t, os.WriteFile(filepath.Join(pkg, "manifest.yaml"), []byte("code: demo\nname: Demo\nversion: 1.0.0\nnamespace: demo\nbinary: demo\n"), 0o600))
	got, err := DiscoverProviders(root)
	core.AssertNoError(t, err)
	core.AssertLen(t, got, 1)
}

func TestMarketplace_DiscoverProviders_Bad(t *core.T) {
	_, err := DiscoverProviders(filepath.Join(t.TempDir(), "missing"))
	core.AssertError(
		t, err,
	)
}

func TestMarketplace_DiscoverProviders_Ugly(t *core.T) {
	got, err := DiscoverProviders(t.TempDir())
	core.AssertNoError(t, err)
	core.AssertEmpty(t, got)
}

func TestMarketplace_LoadProviderRegistry_Good(t *core.T) {
	path := filepath.Join(t.TempDir(), "providers.yaml")
	core.RequireNoError(t, os.WriteFile(path, []byte("version: 1\nproviders:\n  demo:\n    version: 1.0.0\n"), 0o600))
	reg, err := LoadProviderRegistry(path)
	core.AssertNoError(t, err)
	entry, ok := reg.Get("demo")
	core.AssertTrue(t, ok)
	core.AssertEqual(t, "1.0.0", entry.Version)
}

func TestMarketplace_LoadProviderRegistry_Bad(t *core.T) {
	reg, err := LoadProviderRegistry(filepath.Join(t.TempDir(), "missing.yaml"))
	core.AssertNoError(t, err)
	core.AssertEqual(t, 1, reg.Version)
}

func TestMarketplace_LoadProviderRegistry_Ugly(t *core.T) {
	path := filepath.Join(t.TempDir(), "providers.yaml")
	core.RequireNoError(t, os.WriteFile(path, []byte("providers: ["), 0o600))
	_, err := LoadProviderRegistry(path)
	core.AssertError(t, err)
}

func TestMarketplace_SaveProviderRegistry_Good(t *core.T) {
	path := filepath.Join(t.TempDir(), "providers.yaml")
	reg := &ProviderRegistryFile{Version: 1, Providers: map[string]ProviderRegistryEntry{"demo": {Version: "1.0.0"}}}
	err := SaveProviderRegistry(path, reg)
	core.AssertNoError(t, err)
	raw, readErr := os.ReadFile(path)
	core.RequireNoError(t, readErr)
	core.AssertContains(t, string(raw), "demo")
}

func TestMarketplace_SaveProviderRegistry_Bad(t *core.T) {
	err := SaveProviderRegistry(filepath.Join(t.TempDir(), "missing", "providers.yaml"), &ProviderRegistryFile{})
	core.AssertError(
		t, err,
	)
}

func TestMarketplace_SaveProviderRegistry_Ugly(t *core.T) {
	path := filepath.Join(t.TempDir(), "providers.yaml")
	err := SaveProviderRegistry(path, nil)
	core.AssertNoError(t, err)
	raw, readErr := os.ReadFile(path)
	core.RequireNoError(t, readErr)
	core.AssertContains(t, string(raw), "version")
}
