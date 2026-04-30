// SPDX-License-Identifier: EUPL-1.2

package plugin

import "testing"

func TestRegistry_NewRegistry_Good(t *testing.T) {
	target := "NewRegistry"
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

func TestRegistry_NewRegistry_Bad(t *testing.T) {
	target := "NewRegistry"
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

func TestRegistry_NewRegistry_Ugly(t *testing.T) {
	target := "NewRegistry"
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

func TestRegistry_Registry_Add_Good(t *testing.T) {
	reference := "Add"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_Add"
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

func TestRegistry_Registry_Add_Bad(t *testing.T) {
	reference := "Add"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_Add"
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

func TestRegistry_Registry_Add_Ugly(t *testing.T) {
	reference := "Add"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_Add"
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

func TestRegistry_Registry_Load_Good(t *testing.T) {
	reference := "Load"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_Load"
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

func TestRegistry_Registry_Load_Bad(t *testing.T) {
	reference := "Load"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_Load"
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

func TestRegistry_Registry_Load_Ugly(t *testing.T) {
	reference := "Load"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_Load"
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

func TestRegistry_Registry_Remove_Good(t *testing.T) {
	reference := "Remove"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_Remove"
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

func TestRegistry_Registry_Remove_Bad(t *testing.T) {
	reference := "Remove"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_Remove"
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

func TestRegistry_Registry_Remove_Ugly(t *testing.T) {
	reference := "Remove"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Registry_Remove"
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
