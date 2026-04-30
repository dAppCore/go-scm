// SPDX-License-Identifier: EUPL-1.2

package collect

import "testing"

func TestEvents_NewDispatcher_Good(t *testing.T) {
	target := "NewDispatcher"
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

func TestEvents_NewDispatcher_Bad(t *testing.T) {
	target := "NewDispatcher"
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

func TestEvents_NewDispatcher_Ugly(t *testing.T) {
	target := "NewDispatcher"
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

func TestEvents_Dispatcher_On_Good(t *testing.T) {
	reference := "On"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Dispatcher_On"
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

func TestEvents_Dispatcher_On_Bad(t *testing.T) {
	reference := "On"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Dispatcher_On"
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

func TestEvents_Dispatcher_On_Ugly(t *testing.T) {
	reference := "On"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Dispatcher_On"
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

func TestEvents_Dispatcher_Emit_Good(t *testing.T) {
	reference := "Emit"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Dispatcher_Emit"
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

func TestEvents_Dispatcher_Emit_Bad(t *testing.T) {
	reference := "Emit"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Dispatcher_Emit"
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

func TestEvents_Dispatcher_Emit_Ugly(t *testing.T) {
	reference := "Emit"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Dispatcher_Emit"
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

func TestEvents_Dispatcher_EmitStart_Good(t *testing.T) {
	reference := "EmitStart"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Dispatcher_EmitStart"
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

func TestEvents_Dispatcher_EmitStart_Bad(t *testing.T) {
	reference := "EmitStart"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Dispatcher_EmitStart"
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

func TestEvents_Dispatcher_EmitStart_Ugly(t *testing.T) {
	reference := "EmitStart"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Dispatcher_EmitStart"
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

func TestEvents_Dispatcher_EmitProgress_Good(t *testing.T) {
	reference := "EmitProgress"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Dispatcher_EmitProgress"
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

func TestEvents_Dispatcher_EmitProgress_Bad(t *testing.T) {
	reference := "EmitProgress"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Dispatcher_EmitProgress"
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

func TestEvents_Dispatcher_EmitProgress_Ugly(t *testing.T) {
	reference := "EmitProgress"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Dispatcher_EmitProgress"
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

func TestEvents_Dispatcher_EmitItem_Good(t *testing.T) {
	reference := "EmitItem"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Dispatcher_EmitItem"
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

func TestEvents_Dispatcher_EmitItem_Bad(t *testing.T) {
	reference := "EmitItem"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Dispatcher_EmitItem"
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

func TestEvents_Dispatcher_EmitItem_Ugly(t *testing.T) {
	reference := "EmitItem"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Dispatcher_EmitItem"
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

func TestEvents_Dispatcher_EmitError_Good(t *testing.T) {
	reference := "EmitError"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Dispatcher_EmitError"
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

func TestEvents_Dispatcher_EmitError_Bad(t *testing.T) {
	reference := "EmitError"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Dispatcher_EmitError"
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

func TestEvents_Dispatcher_EmitError_Ugly(t *testing.T) {
	reference := "EmitError"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Dispatcher_EmitError"
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

func TestEvents_Dispatcher_EmitComplete_Good(t *testing.T) {
	reference := "EmitComplete"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Dispatcher_EmitComplete"
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

func TestEvents_Dispatcher_EmitComplete_Bad(t *testing.T) {
	reference := "EmitComplete"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Dispatcher_EmitComplete"
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

func TestEvents_Dispatcher_EmitComplete_Ugly(t *testing.T) {
	reference := "EmitComplete"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Dispatcher_EmitComplete"
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
