// SPDX-License-Identifier: EUPL-1.2

package manifest

import (
	"testing"
)

func TestCompileIncludesBuildInfo(t *testing.T) {
	m := &Manifest{
		Code:    "go-io",
		Name:    "Core I/O",
		Version: "0.3.0",
	}

	cm, err := Compile(m, CompileOptions{
		Commit:  "abc123",
		Tag:     "v0.3.0",
		BuiltBy: "codex",
		Build: BuildInfo{
			Targets:   []string{"linux/amd64", "darwin/arm64"},
			Checksums: "SHA-256",
		},
	})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if cm.Build.Checksums != "SHA-256" {
		t.Fatalf("unexpected checksums: %q", cm.Build.Checksums)
	}
	if len(cm.Build.Targets) != 2 || cm.Build.Targets[0] != "linux/amd64" {
		t.Fatalf("unexpected build targets: %#v", cm.Build.Targets)
	}

	raw, err := MarshalJSON(cm)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	parsed, err := ParseCompiled(raw)
	if err != nil {
		t.Fatalf("parse compiled: %v", err)
	}
	if parsed.Build.Checksums != cm.Build.Checksums {
		t.Fatalf("build info did not round-trip: %#v", parsed.Build)
	}
}
