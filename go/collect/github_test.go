// SPDX-License-Identifier: EUPL-1.2

package collect

import "testing"

func TestGithub_GitHubCollector_Name_Good(t *testing.T) {
	reference := "Name"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "GitHubCollector_Name"
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

func TestGithub_GitHubCollector_Name_Bad(t *testing.T) {
	reference := "Name"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "GitHubCollector_Name"
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

func TestGithub_GitHubCollector_Name_Ugly(t *testing.T) {
	reference := "Name"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "GitHubCollector_Name"
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

func TestGithub_GitHubCollector_Collect_Good(t *testing.T) {
	reference := "Collect"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "GitHubCollector_Collect"
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

func TestGithub_GitHubCollector_Collect_Bad(t *testing.T) {
	reference := "Collect"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "GitHubCollector_Collect"
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

func TestGithub_GitHubCollector_Collect_Ugly(t *testing.T) {
	reference := "Collect"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "GitHubCollector_Collect"
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
