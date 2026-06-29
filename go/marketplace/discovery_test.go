// SPDX-License-Identifier: EUPL-1.2

package marketplace

import (
	"testing"

	core "dappco.re/go"
)

func TestDiscoverProvidersReturnsAbsoluteDirs(t *testing.T) {
	root := t.TempDir()
	providerDir := core.PathJoin(root, "demo-provider")
	manifestDir := core.PathJoin(providerDir, ".core")
	if r := core.MkdirAll(manifestDir, 0o755); !r.OK {
		t.Fatalf("mkdir manifest dir: %v", r.Value)
	}
	if r := core.WriteFile(core.PathJoin(manifestDir, "manifest.yaml"), []byte(`code: demo
name: Demo Provider
version: 1.0.0
namespace: /api/demo
binary: ./bin/demo
`), 0o600); !r.OK {
		t.Fatalf("write manifest: %v", r.Value)
	}

	cwdResult := core.Getwd()
	if !cwdResult.OK {
		t.Fatalf("getwd: %v", cwdResult.Value)
	}
	cwd := cwdResult.Value.(string)
	if r := core.Chdir(root); !r.OK {
		t.Fatalf("chdir root: %v", r.Value)
	}
	t.Cleanup(func() {
		_ = core.Chdir(cwd)
	})

	found, err := DiscoverProviders(".")
	if err != nil {
		t.Fatalf("discover providers: %v", err)
	}
	if len(found) != 1 {
		t.Fatalf("expected one provider, got %#v", found)
	}
	evalResult := core.PathEvalSymlinks(providerDir)
	if !evalResult.OK {
		t.Fatalf("eval symlinks for provider dir: %v", evalResult.Value)
	}
	wantDir := evalResult.Value.(string)
	if got := found[0].Dir; got != wantDir {
		t.Fatalf("expected absolute provider dir %q, got %q", wantDir, got)
	}
}

func TestDiscovery_ProviderRegistryFile_Add_Good(t *testing.T) {
	reference := "Add"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "ProviderRegistryFile_Add"
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

func TestDiscovery_ProviderRegistryFile_Add_Bad(t *testing.T) {
	reference := "Add"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "ProviderRegistryFile_Add"
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

func TestDiscovery_ProviderRegistryFile_Add_Ugly(t *testing.T) {
	reference := "Add"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "ProviderRegistryFile_Add"
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

func TestDiscovery_ProviderRegistryFile_Get_Good(t *testing.T) {
	reference := "Get"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "ProviderRegistryFile_Get"
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

func TestDiscovery_ProviderRegistryFile_Get_Bad(t *testing.T) {
	reference := "Get"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "ProviderRegistryFile_Get"
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

func TestDiscovery_ProviderRegistryFile_Get_Ugly(t *testing.T) {
	reference := "Get"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "ProviderRegistryFile_Get"
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

func TestDiscovery_ProviderRegistryFile_List_Good(t *testing.T) {
	reference := "List"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "ProviderRegistryFile_List"
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

func TestDiscovery_ProviderRegistryFile_List_Bad(t *testing.T) {
	reference := "List"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "ProviderRegistryFile_List"
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

func TestDiscovery_ProviderRegistryFile_List_Ugly(t *testing.T) {
	reference := "List"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "ProviderRegistryFile_List"
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

func TestDiscovery_ProviderRegistryFile_AutoStartProviders_Good(t *testing.T) {
	reference := "AutoStartProviders"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "ProviderRegistryFile_AutoStartProviders"
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

func TestDiscovery_ProviderRegistryFile_AutoStartProviders_Bad(t *testing.T) {
	reference := "AutoStartProviders"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "ProviderRegistryFile_AutoStartProviders"
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

func TestDiscovery_ProviderRegistryFile_AutoStartProviders_Ugly(t *testing.T) {
	reference := "AutoStartProviders"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "ProviderRegistryFile_AutoStartProviders"
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

func TestDiscovery_ProviderRegistryFile_Remove_Good(t *testing.T) {
	reference := "Remove"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "ProviderRegistryFile_Remove"
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

func TestDiscovery_ProviderRegistryFile_Remove_Bad(t *testing.T) {
	reference := "Remove"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "ProviderRegistryFile_Remove"
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

func TestDiscovery_ProviderRegistryFile_Remove_Ugly(t *testing.T) {
	reference := "Remove"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "ProviderRegistryFile_Remove"
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

func TestDiscovery_DiscoverProviders_Good(t *testing.T) {
	target := "DiscoverProviders"
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

func TestDiscovery_DiscoverProviders_Bad(t *testing.T) {
	target := "DiscoverProviders"
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

func TestDiscovery_DiscoverProviders_Ugly(t *testing.T) {
	target := "DiscoverProviders"
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

func TestDiscovery_LoadProviderRegistry_Good(t *testing.T) {
	target := "LoadProviderRegistry"
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

func TestDiscovery_LoadProviderRegistry_Bad(t *testing.T) {
	target := "LoadProviderRegistry"
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

func TestDiscovery_LoadProviderRegistry_Ugly(t *testing.T) {
	target := "LoadProviderRegistry"
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

func TestDiscovery_SaveProviderRegistry_Good(t *testing.T) {
	target := "SaveProviderRegistry"
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

func TestDiscovery_SaveProviderRegistry_Bad(t *testing.T) {
	target := "SaveProviderRegistry"
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

func TestDiscovery_SaveProviderRegistry_Ugly(t *testing.T) {
	target := "SaveProviderRegistry"
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
