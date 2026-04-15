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
		SignKey: "ed25519:public-key",
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
	if parsed.SignKey != m.SignKey {
		t.Fatalf("sign key did not round-trip: %q", parsed.SignKey)
	}
}

func TestParseManifestIncludesSignKey(t *testing.T) {
	raw := []byte(`
code: go-io
name: Core I/O
version: 0.3.0
sign_key: ed25519:public-key
`)

	m, err := Parse(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if m.SignKey != "ed25519:public-key" {
		t.Fatalf("unexpected sign key: %q", m.SignKey)
	}
}
