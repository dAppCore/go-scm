// SPDX-License-Identifier: EUPL-1.2

package plugin

import (
	"context"

	core "dappco.re/go"
	coreio "dappco.re/go/io"
)

func ax7PluginManifestJSON(name string) string {
	return `{"name":"` + name + `","version":"1.0.0","entrypoint":"plugin.so"}`
}

func TestPlugin_BasePlugin_Name_Good(t *core.T) {
	plugin := &BasePlugin{PluginName: "demo"}
	core.AssertEqual(
		t, "demo", plugin.Name(),
	)
}

func TestPlugin_BasePlugin_Name_Bad(t *core.T) {
	plugin := &BasePlugin{}
	core.AssertEqual(
		t, "", plugin.Name(),
	)
}

func TestPlugin_BasePlugin_Name_Ugly(t *core.T) {
	var plugin *BasePlugin
	core.AssertPanics(t, func() {
		_ = plugin.Name()
	})
}

func TestPlugin_BasePlugin_Version_Good(t *core.T) {
	plugin := &BasePlugin{PluginVersion: "1.0.0"}
	core.AssertEqual(
		t, "1.0.0", plugin.Version(),
	)
}

func TestPlugin_BasePlugin_Version_Bad(t *core.T) {
	plugin := &BasePlugin{}
	core.AssertEqual(
		t, "", plugin.Version(),
	)
}

func TestPlugin_BasePlugin_Version_Ugly(t *core.T) {
	var plugin *BasePlugin
	core.AssertPanics(t, func() {
		_ = plugin.Version()
	})
}

func TestPlugin_BasePlugin_Init_Good(t *core.T) {
	err := (&BasePlugin{}).Init(context.Background())
	core.AssertNoError(
		t, err,
	)
}

func TestPlugin_BasePlugin_Init_Bad(t *core.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := (&BasePlugin{}).Init(ctx)
	core.AssertNoError(t, err)
}

func TestPlugin_BasePlugin_Init_Ugly(t *core.T) {
	var plugin *BasePlugin
	err := plugin.Init(nil)
	core.AssertNoError(t, err)
}

func TestPlugin_BasePlugin_Start_Good(t *core.T) {
	err := (&BasePlugin{}).Start(context.Background())
	core.AssertNoError(
		t, err,
	)
}

func TestPlugin_BasePlugin_Start_Bad(t *core.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := (&BasePlugin{}).Start(ctx)
	core.AssertNoError(t, err)
}

func TestPlugin_BasePlugin_Start_Ugly(t *core.T) {
	var plugin *BasePlugin
	err := plugin.Start(nil)
	core.AssertNoError(t, err)
}

func TestPlugin_BasePlugin_Stop_Good(t *core.T) {
	err := (&BasePlugin{}).Stop(context.Background())
	core.AssertNoError(
		t, err,
	)
}

func TestPlugin_BasePlugin_Stop_Bad(t *core.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := (&BasePlugin{}).Stop(ctx)
	core.AssertNoError(t, err)
}

func TestPlugin_BasePlugin_Stop_Ugly(t *core.T) {
	var plugin *BasePlugin
	err := plugin.Stop(nil)
	core.AssertNoError(t, err)
}

func TestPlugin_Manifest_Validate_Good(t *core.T) {
	err := (&Manifest{Name: "demo", Version: "1.0.0", Entrypoint: "plugin.so"}).Validate()
	core.AssertNoError(
		t, err,
	)
}

func TestPlugin_Manifest_Validate_Bad(t *core.T) {
	err := (&Manifest{Name: "demo"}).Validate()
	core.AssertError(
		t, err,
	)
}

func TestPlugin_Manifest_Validate_Ugly(t *core.T) {
	var manifest *Manifest
	err := manifest.Validate()
	core.AssertError(t, err)
}

func TestPlugin_LoadManifest_Good(t *core.T) {
	medium := coreio.NewMockMedium()
	core.RequireNoError(t, medium.Write("plugins/demo/plugin.json", ax7PluginManifestJSON("demo")))
	manifest, err := LoadManifest(medium, "plugins/demo/plugin.json")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "demo", manifest.Name)
}

func TestPlugin_LoadManifest_Bad(t *core.T) {
	_, err := LoadManifest(nil, "plugins/demo/plugin.json")
	core.AssertError(
		t, err,
	)
}

func TestPlugin_LoadManifest_Ugly(t *core.T) {
	medium := coreio.NewMockMedium()
	core.RequireNoError(t, medium.Write("plugins/demo/plugin.json", `{"name":"demo"}`))
	_, err := LoadManifest(medium, "plugins/demo/plugin.json")
	core.AssertError(t, err)
}

func TestPlugin_NewLoader_Good(t *core.T) {
	loader := NewLoader(coreio.NewMockMedium(), "plugins")
	core.AssertNotNil(t, loader)
	core.AssertEqual(t, "plugins", loader.baseDir)
}

func TestPlugin_NewLoader_Bad(t *core.T) {
	loader := NewLoader(nil, "plugins")
	core.AssertNotNil(t, loader)
	core.AssertNil(t, loader.medium)
}

func TestPlugin_NewLoader_Ugly(t *core.T) {
	loader := NewLoader(coreio.NewMockMedium(), "")
	core.AssertEqual(
		t, "", loader.baseDir,
	)
}

func TestPlugin_Loader_Discover_Good(t *core.T) {
	medium := coreio.NewMockMedium()
	core.RequireNoError(t, medium.Write("plugins/demo/plugin.json", ax7PluginManifestJSON("demo")))
	loader := NewLoader(medium, "plugins")
	got, err := loader.Discover()
	core.AssertNoError(t, err)
	core.AssertLen(t, got, 1)
}

func TestPlugin_Loader_Discover_Bad(t *core.T) {
	var loader *Loader
	got, err := loader.Discover()
	core.AssertNoError(t, err)
	core.AssertNil(t, got)
}

func TestPlugin_Loader_Discover_Ugly(t *core.T) {
	loader := NewLoader(coreio.NewMockMedium(), "plugins")
	got, err := loader.Discover()
	core.AssertNoError(t, err)
	core.AssertEmpty(t, got)
}

func TestPlugin_Loader_LoadPlugin_Good(t *core.T) {
	medium := coreio.NewMockMedium()
	core.RequireNoError(t, medium.Write("plugins/demo/plugin.json", ax7PluginManifestJSON("demo")))
	loader := NewLoader(medium, "plugins")
	manifest, err := loader.LoadPlugin("demo")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "demo", manifest.Name)
}

func TestPlugin_Loader_LoadPlugin_Bad(t *core.T) {
	var loader *Loader
	_, err := loader.LoadPlugin("demo")
	core.AssertError(t, err)
}

func TestPlugin_Loader_LoadPlugin_Ugly(t *core.T) {
	loader := NewLoader(coreio.NewMockMedium(), "plugins")
	_, err := loader.LoadPlugin("")
	core.AssertError(t, err)
}

func TestPlugin_NewRegistry_Good(t *core.T) {
	registry := NewRegistry(coreio.NewMockMedium(), "plugins")
	core.AssertNotNil(t, registry)
	core.AssertEqual(t, "plugins", registry.basePath)
}

func TestPlugin_NewRegistry_Bad(t *core.T) {
	registry := NewRegistry(nil, "plugins")
	core.AssertNotNil(t, registry)
	core.AssertNil(t, registry.medium)
}

func TestPlugin_NewRegistry_Ugly(t *core.T) {
	registry := NewRegistry(coreio.NewMockMedium(), "")
	core.AssertNotNil(t, registry.plugins)
	core.AssertEqual(t, "", registry.basePath)
}

func TestPlugin_Registry_Add_Good(t *core.T) {
	registry := NewRegistry(coreio.NewMockMedium(), "plugins")
	err := registry.Add(&PluginConfig{Name: "demo"})
	core.AssertNoError(t, err)
	_, ok := registry.Get("demo")
	core.AssertTrue(t, ok)
}

func TestPlugin_Registry_Add_Bad(t *core.T) {
	err := NewRegistry(coreio.NewMockMedium(), "plugins").Add(nil)
	core.AssertError(
		t, err,
	)
}

func TestPlugin_Registry_Add_Ugly(t *core.T) {
	var registry *Registry
	err := registry.Add(&PluginConfig{Name: "demo"})
	core.AssertError(t, err)
}

func TestPlugin_Registry_Get_Good(t *core.T) {
	registry := NewRegistry(coreio.NewMockMedium(), "plugins")
	core.RequireNoError(t, registry.Add(&PluginConfig{Name: "demo"}))
	cfg, ok := registry.Get("demo")
	core.AssertTrue(t, ok)
	core.AssertEqual(t, "demo", cfg.Name)
}

func TestPlugin_Registry_Get_Bad(t *core.T) {
	registry := NewRegistry(coreio.NewMockMedium(), "plugins")
	_, ok := registry.Get("missing")
	core.AssertFalse(t, ok)
}

func TestPlugin_Registry_Get_Ugly(t *core.T) {
	var registry *Registry
	_, ok := registry.Get("demo")
	core.AssertFalse(t, ok)
}

func TestPlugin_Registry_List_Good(t *core.T) {
	registry := NewRegistry(coreio.NewMockMedium(), "plugins")
	core.RequireNoError(t, registry.Add(&PluginConfig{Name: "b"}))
	core.RequireNoError(t, registry.Add(&PluginConfig{Name: "a"}))
	got := registry.List()
	core.AssertEqual(t, "a", got[0].Name)
	core.AssertEqual(t, "b", got[1].Name)
}

func TestPlugin_Registry_List_Bad(t *core.T) {
	registry := NewRegistry(coreio.NewMockMedium(), "plugins")
	got := registry.List()
	core.AssertEmpty(t, got)
}

func TestPlugin_Registry_List_Ugly(t *core.T) {
	var registry *Registry
	got := registry.List()
	core.AssertNil(t, got)
}

func TestPlugin_Registry_Load_Good(t *core.T) {
	medium := coreio.NewMockMedium()
	core.RequireNoError(t, medium.Write("plugins/registry.json", `{"plugins":{"demo":{"name":"demo"}}}`))
	registry := NewRegistry(medium, "plugins")
	err := registry.Load()
	core.AssertNoError(t, err)
	_, ok := registry.Get("demo")
	core.AssertTrue(t, ok)
}

func TestPlugin_Registry_Load_Bad(t *core.T) {
	registry := NewRegistry(coreio.NewMockMedium(), "plugins")
	err := registry.Load()
	core.AssertNoError(t, err)
	core.AssertEmpty(t, registry.plugins)
}

func TestPlugin_Registry_Load_Ugly(t *core.T) {
	medium := coreio.NewMockMedium()
	core.RequireNoError(t, medium.Write("plugins/registry.json", `{`))
	registry := NewRegistry(medium, "plugins")
	err := registry.Load()
	core.AssertError(t, err)
}

func TestPlugin_Registry_Remove_Good(t *core.T) {
	registry := NewRegistry(coreio.NewMockMedium(), "plugins")
	core.RequireNoError(t, registry.Add(&PluginConfig{Name: "demo"}))
	err := registry.Remove("demo")
	core.AssertNoError(t, err)
	_, ok := registry.Get("demo")
	core.AssertFalse(t, ok)
}

func TestPlugin_Registry_Remove_Bad(t *core.T) {
	var registry *Registry
	err := registry.Remove("demo")
	core.AssertError(t, err)
}

func TestPlugin_Registry_Remove_Ugly(t *core.T) {
	registry := NewRegistry(coreio.NewMockMedium(), "plugins")
	err := registry.Remove("")
	core.AssertNoError(t, err)
}

func TestPlugin_Registry_Save_Good(t *core.T) {
	medium := coreio.NewMockMedium()
	registry := NewRegistry(medium, "plugins")
	core.RequireNoError(t, registry.Add(&PluginConfig{Name: "demo"}))
	err := registry.Save()
	core.AssertNoError(t, err)
	raw, readErr := medium.Read("plugins/registry.json")
	core.RequireNoError(t, readErr)
	core.AssertContains(t, raw, "demo")
}

func TestPlugin_Registry_Save_Bad(t *core.T) {
	var registry *Registry
	err := registry.Save()
	core.AssertNoError(t, err)
}

func TestPlugin_Registry_Save_Ugly(t *core.T) {
	registry := NewRegistry(nil, "plugins")
	err := registry.Save()
	core.AssertNoError(t, err)
}

func TestPlugin_NewInstaller_Good(t *core.T) {
	registry := NewRegistry(coreio.NewMockMedium(), "plugins")
	installer := NewInstaller(coreio.NewMockMedium(), registry)
	core.AssertNotNil(t, installer)
	core.AssertEqual(t, registry, installer.registry)
}

func TestPlugin_NewInstaller_Bad(t *core.T) {
	installer := NewInstaller(nil, nil)
	core.AssertNotNil(t, installer)
	core.AssertNil(t, installer.registry)
}

func TestPlugin_NewInstaller_Ugly(t *core.T) {
	installer := NewInstaller(coreio.NewMockMedium(), nil)
	core.AssertNotNil(
		t, installer.medium,
	)
}

func TestPlugin_ParseSource_Good(t *core.T) {
	org, repo, version, err := ParseSource("core/plugin@v1")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "core", org)
	core.AssertEqual(t, "plugin", repo)
	core.AssertEqual(t, "v1", version)
}

func TestPlugin_ParseSource_Bad(t *core.T) {
	_, _, _, err := ParseSource("")
	core.AssertError(
		t, err,
	)
}

func TestPlugin_ParseSource_Ugly(t *core.T) {
	org, repo, version, err := ParseSource("core/plugin")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "core", org)
	core.AssertEqual(t, "plugin", repo)
	core.AssertEqual(t, "", version)
}

func TestPlugin_Installer_Install_Good(t *core.T) {
	registry := NewRegistry(coreio.NewMockMedium(), "plugins")
	installer := NewInstaller(coreio.NewMockMedium(), registry)
	err := installer.Install(context.Background(), "core/plugin@v1")
	core.AssertNoError(t, err)
	cfg, ok := registry.Get("plugin")
	core.AssertTrue(t, ok)
	core.AssertEqual(t, "v1", cfg.Version)
}

func TestPlugin_Installer_Install_Bad(t *core.T) {
	err := NewInstaller(coreio.NewMockMedium(), nil).Install(context.Background(), "")
	core.AssertError(
		t, err,
	)
}

func TestPlugin_Installer_Install_Ugly(t *core.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := NewInstaller(coreio.NewMockMedium(), nil).Install(ctx, "core/plugin")
	core.AssertErrorIs(t, err, context.Canceled)
}

func TestPlugin_Installer_Remove_Good(t *core.T) {
	registry := NewRegistry(coreio.NewMockMedium(), "plugins")
	installer := NewInstaller(coreio.NewMockMedium(), registry)
	core.RequireNoError(t, installer.Install(context.Background(), "core/plugin@v1"))
	err := installer.Remove("plugin")
	core.AssertNoError(t, err)
	_, ok := registry.Get("plugin")
	core.AssertFalse(t, ok)
}

func TestPlugin_Installer_Remove_Bad(t *core.T) {
	var installer *Installer
	err := installer.Remove("plugin")
	core.AssertError(t, err)
}

func TestPlugin_Installer_Remove_Ugly(t *core.T) {
	err := NewInstaller(coreio.NewMockMedium(), nil).Remove("plugin")
	core.AssertNoError(
		t, err,
	)
}

func TestPlugin_Installer_Update_Good(t *core.T) {
	registry := NewRegistry(coreio.NewMockMedium(), "plugins")
	installer := NewInstaller(coreio.NewMockMedium(), registry)
	core.RequireNoError(t, installer.Install(context.Background(), "core/plugin@v1"))
	err := installer.Update(context.Background(), "plugin")
	core.AssertNoError(t, err)
}

func TestPlugin_Installer_Update_Bad(t *core.T) {
	err := NewInstaller(coreio.NewMockMedium(), NewRegistry(coreio.NewMockMedium(), "plugins")).Update(context.Background(), "missing")
	core.AssertError(
		t, err,
	)
}

func TestPlugin_Installer_Update_Ugly(t *core.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := NewInstaller(coreio.NewMockMedium(), nil).Update(ctx, "plugin")
	core.AssertErrorIs(t, err, context.Canceled)
}
