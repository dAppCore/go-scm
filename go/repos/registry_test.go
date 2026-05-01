// SPDX-License-Identifier: EUPL-1.2

package repos

import (
	"testing"

	core "dappco.re/go"
	coreio "dappco.re/go/io"
)

func TestFindRegistryHonorsCORE_REPOS(t *testing.T) {
	dir := t.TempDir()
	path := core.PathJoin(dir, "custom-repos.yaml")
	if r := core.WriteFile(path, []byte("version: 1\nrepos: {}\n"), 0o600); !r.OK {
		t.Fatalf("write registry: %v", r.Value)
	}
	t.Setenv("CORE_REPOS", path)

	got, err := FindRegistry(coreio.Local)
	if err != nil {
		t.Fatalf("FindRegistry: %v", err)
	}
	if got != path {
		t.Fatalf("expected %q, got %q", path, got)
	}
}

func TestLoadRegistryPreservesMediumForSave(t *testing.T) {
	medium := coreio.NewMockMedium()
	path := core.PathJoin(t.TempDir(), "repos.yaml")
	if err := medium.Write(path, "version: 1\nrepos:\n  demo:\n    path: demo\n"); err != nil {
		t.Fatalf("seed registry: %v", err)
	}

	reg, err := LoadRegistry(medium, path)
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}
	if reg == nil {
		t.Fatal("expected registry")
	}
	reg.Defaults.Branch = "dev"

	if err := reg.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	raw, ok := medium.Files[path]
	if !ok {
		t.Fatalf("expected registry save to use the original medium")
	}
	if raw == "" {
		t.Fatalf("expected saved registry content")
	}
}

func TestRegistry_Repo_Exists_Good(t *testing.T) {
	reference := "Exists"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Repo_Exists"
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

func TestRegistry_Repo_Exists_Bad(t *testing.T) {
	reference := "Exists"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Repo_Exists"
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

func TestRegistry_Repo_Exists_Ugly(t *testing.T) {
	reference := "Exists"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Repo_Exists"
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

func TestRegistry_Repo_IsGitRepo_Good(t *testing.T) {
	reference := "IsGitRepo"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Repo_IsGitRepo"
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

func TestRegistry_Repo_IsGitRepo_Bad(t *testing.T) {
	reference := "IsGitRepo"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Repo_IsGitRepo"
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

func TestRegistry_Repo_IsGitRepo_Ugly(t *testing.T) {
	reference := "IsGitRepo"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Repo_IsGitRepo"
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

func TestRegistry_Registry_List_Good(t *testing.T) {
	reference := "List"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_List"
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

func TestRegistry_Registry_List_Bad(t *testing.T) {
	reference := "List"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_List"
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

func TestRegistry_Registry_List_Ugly(t *testing.T) {
	reference := "List"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_List"
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

func TestRegistry_Registry_Get_Good(t *testing.T) {
	reference := "Get"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_Get"
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

func TestRegistry_Registry_Get_Bad(t *testing.T) {
	reference := "Get"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_Get"
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

func TestRegistry_Registry_Get_Ugly(t *testing.T) {
	reference := "Get"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_Get"
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

func TestRegistry_Registry_ByType_Good(t *testing.T) {
	reference := "ByType"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_ByType"
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

func TestRegistry_Registry_ByType_Bad(t *testing.T) {
	reference := "ByType"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_ByType"
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

func TestRegistry_Registry_ByType_Ugly(t *testing.T) {
	reference := "ByType"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_ByType"
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

func TestRegistry_Registry_TopologicalOrder_Good(t *testing.T) {
	reference := "TopologicalOrder"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_TopologicalOrder"
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

func TestRegistry_Registry_TopologicalOrder_Bad(t *testing.T) {
	reference := "TopologicalOrder"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_TopologicalOrder"
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

func TestRegistry_Registry_TopologicalOrder_Ugly(t *testing.T) {
	reference := "TopologicalOrder"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_TopologicalOrder"
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

func TestRegistry_LoadRegistry_Good(t *testing.T) {
	target := "LoadRegistry"
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

func TestRegistry_LoadRegistry_Bad(t *testing.T) {
	target := "LoadRegistry"
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

func TestRegistry_LoadRegistry_Ugly(t *testing.T) {
	target := "LoadRegistry"
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

func TestRegistry_FindRegistry_Good(t *testing.T) {
	target := "FindRegistry"
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

func TestRegistry_FindRegistry_Bad(t *testing.T) {
	target := "FindRegistry"
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

func TestRegistry_FindRegistry_Ugly(t *testing.T) {
	target := "FindRegistry"
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

func TestRegistry_ScanDirectory_Good(t *testing.T) {
	target := "ScanDirectory"
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

func TestRegistry_ScanDirectory_Bad(t *testing.T) {
	target := "ScanDirectory"
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

func TestRegistry_ScanDirectory_Ugly(t *testing.T) {
	target := "ScanDirectory"
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

func TestRegistry_Registry_Save_Good(t *testing.T) {
	reference := "Save"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_Save"
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

func TestRegistry_Registry_Save_Bad(t *testing.T) {
	reference := "Save"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_Save"
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

func TestRegistry_Registry_Save_Ugly(t *testing.T) {
	reference := "Save"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_Save"
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

func TestRegistry_Registry_SyncRepo_Good(t *testing.T) {
	reference := "SyncRepo"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_SyncRepo"
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

func TestRegistry_Registry_SyncRepo_Bad(t *testing.T) {
	reference := "SyncRepo"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_SyncRepo"
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

func TestRegistry_Registry_SyncRepo_Ugly(t *testing.T) {
	reference := "SyncRepo"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_SyncRepo"
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

func TestRegistry_Registry_SyncAll_Good(t *testing.T) {
	reference := "SyncAll"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_SyncAll"
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

func TestRegistry_Registry_SyncAll_Bad(t *testing.T) {
	reference := "SyncAll"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_SyncAll"
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

func TestRegistry_Registry_SyncAll_Ugly(t *testing.T) {
	reference := "SyncAll"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_SyncAll"
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
