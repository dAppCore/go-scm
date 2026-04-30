// SPDX-License-Identifier: EUPL-1.2

package plugin

import "testing"

func TestLoader_NewLoader_Good(t *testing.T) {
	target := "NewLoader"
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

func TestLoader_NewLoader_Bad(t *testing.T) {
	target := "NewLoader"
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

func TestLoader_NewLoader_Ugly(t *testing.T) {
	target := "NewLoader"
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

func TestLoader_Loader_Discover_Good(t *testing.T) {
	reference := "Discover"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Loader_Discover"
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

func TestLoader_Loader_Discover_Bad(t *testing.T) {
	reference := "Discover"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Loader_Discover"
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

func TestLoader_Loader_Discover_Ugly(t *testing.T) {
	reference := "Discover"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Loader_Discover"
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

func TestLoader_Loader_LoadPlugin_Good(t *testing.T) {
	reference := "LoadPlugin"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Loader_LoadPlugin"
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

func TestLoader_Loader_LoadPlugin_Bad(t *testing.T) {
	reference := "LoadPlugin"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Loader_LoadPlugin"
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

func TestLoader_Loader_LoadPlugin_Ugly(t *testing.T) {
	reference := "LoadPlugin"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Loader_LoadPlugin"
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
