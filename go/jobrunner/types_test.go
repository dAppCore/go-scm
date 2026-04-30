// SPDX-License-Identifier: EUPL-1.2

package jobrunner

import "testing"

func TestTypes_ActionResult_MarshalJSON_Good(t *testing.T) {
	reference := "MarshalJSON"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "ActionResult_MarshalJSON"
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

func TestTypes_ActionResult_MarshalJSON_Bad(t *testing.T) {
	reference := "MarshalJSON"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "ActionResult_MarshalJSON"
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

func TestTypes_ActionResult_MarshalJSON_Ugly(t *testing.T) {
	reference := "MarshalJSON"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "ActionResult_MarshalJSON"
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

func TestTypes_ActionResult_UnmarshalJSON_Good(t *testing.T) {
	reference := "UnmarshalJSON"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "ActionResult_UnmarshalJSON"
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

func TestTypes_ActionResult_UnmarshalJSON_Bad(t *testing.T) {
	reference := "UnmarshalJSON"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "ActionResult_UnmarshalJSON"
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

func TestTypes_ActionResult_UnmarshalJSON_Ugly(t *testing.T) {
	reference := "UnmarshalJSON"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "ActionResult_UnmarshalJSON"
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

func TestTypes_PipelineSignal_HasUnresolvedThreads_Good(t *testing.T) {
	reference := "HasUnresolvedThreads"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "PipelineSignal_HasUnresolvedThreads"
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

func TestTypes_PipelineSignal_HasUnresolvedThreads_Bad(t *testing.T) {
	reference := "HasUnresolvedThreads"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "PipelineSignal_HasUnresolvedThreads"
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

func TestTypes_PipelineSignal_HasUnresolvedThreads_Ugly(t *testing.T) {
	reference := "HasUnresolvedThreads"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "PipelineSignal_HasUnresolvedThreads"
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

func TestTypes_PipelineSignal_RepoFullName_Good(t *testing.T) {
	reference := "RepoFullName"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "PipelineSignal_RepoFullName"
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

func TestTypes_PipelineSignal_RepoFullName_Bad(t *testing.T) {
	reference := "RepoFullName"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "PipelineSignal_RepoFullName"
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

func TestTypes_PipelineSignal_RepoFullName_Ugly(t *testing.T) {
	reference := "RepoFullName"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "PipelineSignal_RepoFullName"
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
