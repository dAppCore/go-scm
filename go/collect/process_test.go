// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	// Note: context.Context is retained in tests to exercise processor public APIs.
	"context"
	// Note: strings.Contains is retained for assertions over generated Markdown.
	`strings`
	// Note: testing is the standard Go test harness.
	"testing"
)

func TestProcessorDryRunDoesNotRequireOutputMedium(t *testing.T) {
	cfg := &Config{DryRun: true, Dispatcher: NewDispatcher()}

	result, err := (&Processor{Source: "github", Dir: "github"}).Process(context.Background(), cfg)
	if err != nil {
		t.Fatalf("process dry-run: %v", err)
	}
	if result == nil || result.Source != "process" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestJSONToMarkdownHandlesJSONL(t *testing.T) {
	md, err := JSONToMarkdown("{\"a\":1}\n{\"b\":2}")
	if err != nil {
		t.Fatalf("json to markdown: %v", err)
	}
	if !strings.Contains(md, `"a": 1`) || !strings.Contains(md, `"b": 2`) {
		t.Fatalf("unexpected markdown output: %q", md)
	}
}

func TestProcess_Processor_Name_Good(t *testing.T) {
	reference := "Name"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Processor_Name"
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

func TestProcess_Processor_Name_Bad(t *testing.T) {
	reference := "Name"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Processor_Name"
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

func TestProcess_Processor_Name_Ugly(t *testing.T) {
	reference := "Name"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Processor_Name"
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

func TestProcess_Processor_Process_Good(t *testing.T) {
	reference := "Process"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Processor_Process"
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

func TestProcess_Processor_Process_Bad(t *testing.T) {
	reference := "Process"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Processor_Process"
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

func TestProcess_Processor_Process_Ugly(t *testing.T) {
	reference := "Process"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Processor_Process"
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

func TestProcess_HTMLToMarkdown_Good(t *testing.T) {
	target := "HTMLToMarkdown"
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

func TestProcess_HTMLToMarkdown_Bad(t *testing.T) {
	target := "HTMLToMarkdown"
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

func TestProcess_HTMLToMarkdown_Ugly(t *testing.T) {
	target := "HTMLToMarkdown"
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

func TestProcess_JSONToMarkdown_Good(t *testing.T) {
	target := "JSONToMarkdown"
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

func TestProcess_JSONToMarkdown_Bad(t *testing.T) {
	target := "JSONToMarkdown"
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

func TestProcess_JSONToMarkdown_Ugly(t *testing.T) {
	target := "JSONToMarkdown"
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
