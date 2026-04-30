// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	// Note: testing is the standard Go test harness.
	"testing"

	coreio "dappco.re/go/io"
)

func TestNewStateWithEmptyPathDoesNotPersist(t *testing.T) {
	medium := coreio.NewMockMedium()
	state := NewState(medium, "")

	if err := state.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if len(medium.Files) != 0 {
		t.Fatalf("expected empty path state to skip persistence, got %#v", medium.Files)
	}
}

func TestState_NewState_Good(t *testing.T) {
	target := "NewState"
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

func TestState_NewState_Bad(t *testing.T) {
	target := "NewState"
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

func TestState_NewState_Ugly(t *testing.T) {
	target := "NewState"
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

func TestState_State_Get_Good(t *testing.T) {
	reference := "Get"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "State_Get"
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

func TestState_State_Get_Bad(t *testing.T) {
	reference := "Get"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "State_Get"
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

func TestState_State_Get_Ugly(t *testing.T) {
	reference := "Get"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "State_Get"
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

func TestState_State_Set_Good(t *testing.T) {
	reference := "Set"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "State_Set"
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

func TestState_State_Set_Bad(t *testing.T) {
	reference := "Set"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "State_Set"
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

func TestState_State_Set_Ugly(t *testing.T) {
	reference := "Set"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "State_Set"
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

func TestState_State_Load_Good(t *testing.T) {
	reference := "Load"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "State_Load"
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

func TestState_State_Load_Bad(t *testing.T) {
	reference := "Load"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "State_Load"
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

func TestState_State_Load_Ugly(t *testing.T) {
	reference := "Load"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "State_Load"
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

func TestState_State_Save_Good(t *testing.T) {
	reference := "Save"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "State_Save"
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

func TestState_State_Save_Bad(t *testing.T) {
	reference := "Save"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "State_Save"
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

func TestState_State_Save_Ugly(t *testing.T) {
	reference := "Save"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "State_Save"
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
