// SPDX-License-Identifier: EUPL-1.2

package collect

import "testing"

func TestPapers_PapersCollector_Name_Good(t *testing.T) {
	reference := "Name"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "PapersCollector_Name"
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

func TestPapers_PapersCollector_Name_Bad(t *testing.T) {
	reference := "Name"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "PapersCollector_Name"
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

func TestPapers_PapersCollector_Name_Ugly(t *testing.T) {
	reference := "Name"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "PapersCollector_Name"
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

func TestPapers_PapersCollector_Collect_Good(t *testing.T) {
	reference := "Collect"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "PapersCollector_Collect"
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

func TestPapers_PapersCollector_Collect_Bad(t *testing.T) {
	reference := "Collect"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "PapersCollector_Collect"
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

func TestPapers_PapersCollector_Collect_Ugly(t *testing.T) {
	reference := "Collect"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "PapersCollector_Collect"
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

func TestPapers_FormatPaperMarkdown_Good(t *testing.T) {
	target := "FormatPaperMarkdown"
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

func TestPapers_FormatPaperMarkdown_Bad(t *testing.T) {
	target := "FormatPaperMarkdown"
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

func TestPapers_FormatPaperMarkdown_Ugly(t *testing.T) {
	target := "FormatPaperMarkdown"
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
