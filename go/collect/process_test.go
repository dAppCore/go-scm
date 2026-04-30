// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	// Note: context.Context is retained in tests to exercise processor public APIs.
	"context"
	// Note: strings.Contains is retained for assertions over generated Markdown.
	"strings"
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
