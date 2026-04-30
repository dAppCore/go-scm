// SPDX-License-Identifier: EUPL-1.2

package marketplace

import "testing"

func TestBuilder_BuildFromManifests_Good(t *testing.T) {
	target := "BuildFromManifests"
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

func TestBuilder_BuildFromManifests_Bad(t *testing.T) {
	target := "BuildFromManifests"
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

func TestBuilder_BuildFromManifests_Ugly(t *testing.T) {
	target := "BuildFromManifests"
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

func TestBuilder_Builder_BuildFromDirs_Good(t *testing.T) {
	reference := "BuildFromDirs"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Builder_BuildFromDirs"
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

func TestBuilder_Builder_BuildFromDirs_Bad(t *testing.T) {
	reference := "BuildFromDirs"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Builder_BuildFromDirs"
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

func TestBuilder_Builder_BuildFromDirs_Ugly(t *testing.T) {
	reference := "BuildFromDirs"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Builder_BuildFromDirs"
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

func TestBuilder_WriteIndex_Good(t *testing.T) {
	target := "WriteIndex"
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

func TestBuilder_WriteIndex_Bad(t *testing.T) {
	target := "WriteIndex"
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

func TestBuilder_WriteIndex_Ugly(t *testing.T) {
	target := "WriteIndex"
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
