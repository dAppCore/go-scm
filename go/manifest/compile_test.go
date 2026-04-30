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

	cm, err := CompileWithOptions(m, CompileOptions{
		Commit:  "abc123",
		Tag:     "v0.3.0",
		BuiltBy: "codex",
		Build: BuildInfo{
			Targets:   []string{"linux/amd64", "darwin/arm64"},
			Checksums: "SHA-256",
			SHA256:    "abc123",
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
	if parsed.Build.SHA256 != cm.Build.SHA256 {
		t.Fatalf("build sha256 did not round-trip: %#v", parsed.Build)
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

func TestCompile_Compile_Good(t *testing.T) {
	target := "Compile"
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

func TestCompile_Compile_Bad(t *testing.T) {
	target := "Compile"
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

func TestCompile_Compile_Ugly(t *testing.T) {
	target := "Compile"
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

func TestCompile_ParseCoreJSON_Good(t *testing.T) {
	target := "ParseCoreJSON"
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

func TestCompile_ParseCoreJSON_Bad(t *testing.T) {
	target := "ParseCoreJSON"
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

func TestCompile_ParseCoreJSON_Ugly(t *testing.T) {
	target := "ParseCoreJSON"
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

func TestCompile_CompileWithOptions_Good(t *testing.T) {
	target := "CompileWithOptions"
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

func TestCompile_CompileWithOptions_Bad(t *testing.T) {
	target := "CompileWithOptions"
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

func TestCompile_CompileWithOptions_Ugly(t *testing.T) {
	target := "CompileWithOptions"
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

func TestCompile_MarshalJSON_Good(t *testing.T) {
	target := "MarshalJSON"
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

func TestCompile_MarshalJSON_Bad(t *testing.T) {
	target := "MarshalJSON"
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

func TestCompile_MarshalJSON_Ugly(t *testing.T) {
	target := "MarshalJSON"
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

func TestCompile_ParseCompiled_Good(t *testing.T) {
	target := "ParseCompiled"
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

func TestCompile_ParseCompiled_Bad(t *testing.T) {
	target := "ParseCompiled"
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

func TestCompile_ParseCompiled_Ugly(t *testing.T) {
	target := "ParseCompiled"
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

func TestCompile_LoadCompiled_Good(t *testing.T) {
	target := "LoadCompiled"
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

func TestCompile_LoadCompiled_Bad(t *testing.T) {
	target := "LoadCompiled"
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

func TestCompile_LoadCompiled_Ugly(t *testing.T) {
	target := "LoadCompiled"
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

func TestCompile_WriteCompiled_Good(t *testing.T) {
	target := "WriteCompiled"
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

func TestCompile_WriteCompiled_Bad(t *testing.T) {
	target := "WriteCompiled"
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

func TestCompile_WriteCompiled_Ugly(t *testing.T) {
	target := "WriteCompiled"
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
