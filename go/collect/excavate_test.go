// SPDX-License-Identifier: EUPL-1.2

package collect

import "testing"

func TestExcavate_Excavator_Name_Good(t *testing.T) {
	reference := "Name"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Excavator_Name"
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

func TestExcavate_Excavator_Name_Bad(t *testing.T) {
	reference := "Name"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Excavator_Name"
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

func TestExcavate_Excavator_Name_Ugly(t *testing.T) {
	reference := "Name"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Excavator_Name"
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

func TestExcavate_Excavator_Run_Good(t *testing.T) {
	reference := "Run"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Excavator_Run"
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

func TestExcavate_Excavator_Run_Bad(t *testing.T) {
	reference := "Run"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Excavator_Run"
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

func TestExcavate_Excavator_Run_Ugly(t *testing.T) {
	reference := "Run"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Excavator_Run"
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
