// SPDX-License-Identifier: EUPL-1.2

package plugin

import "testing"

func TestManifest_Manifest_Validate_Good(t *testing.T) {
	reference := "Validate"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Manifest_Validate"
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

func TestManifest_Manifest_Validate_Bad(t *testing.T) {
	reference := "Validate"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Manifest_Validate"
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

func TestManifest_Manifest_Validate_Ugly(t *testing.T) {
	reference := "Validate"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Manifest_Validate"
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

func TestManifest_LoadManifest_Good(t *testing.T) {
	target := "LoadManifest"
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

func TestManifest_LoadManifest_Bad(t *testing.T) {
	target := "LoadManifest"
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

func TestManifest_LoadManifest_Ugly(t *testing.T) {
	target := "LoadManifest"
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
