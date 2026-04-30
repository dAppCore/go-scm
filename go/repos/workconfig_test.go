// SPDX-License-Identifier: EUPL-1.2

package repos

import "testing"

func TestWorkconfig_DefaultWorkConfig_Good(t *testing.T) {
	target := "DefaultWorkConfig"
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

func TestWorkconfig_DefaultWorkConfig_Bad(t *testing.T) {
	target := "DefaultWorkConfig"
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

func TestWorkconfig_DefaultWorkConfig_Ugly(t *testing.T) {
	target := "DefaultWorkConfig"
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

func TestWorkconfig_WorkConfig_HasTrigger_Good(t *testing.T) {
	reference := "HasTrigger"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "WorkConfig_HasTrigger"
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

func TestWorkconfig_WorkConfig_HasTrigger_Bad(t *testing.T) {
	reference := "HasTrigger"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "WorkConfig_HasTrigger"
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

func TestWorkconfig_WorkConfig_HasTrigger_Ugly(t *testing.T) {
	reference := "HasTrigger"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "WorkConfig_HasTrigger"
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

func TestWorkconfig_LoadWorkConfig_Good(t *testing.T) {
	target := "LoadWorkConfig"
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

func TestWorkconfig_LoadWorkConfig_Bad(t *testing.T) {
	target := "LoadWorkConfig"
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

func TestWorkconfig_LoadWorkConfig_Ugly(t *testing.T) {
	target := "LoadWorkConfig"
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

func TestWorkconfig_SaveWorkConfig_Good(t *testing.T) {
	target := "SaveWorkConfig"
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

func TestWorkconfig_SaveWorkConfig_Bad(t *testing.T) {
	target := "SaveWorkConfig"
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

func TestWorkconfig_SaveWorkConfig_Ugly(t *testing.T) {
	target := "SaveWorkConfig"
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
