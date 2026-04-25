// SPDX-License-Identifier: EUPL-1.2

package manifest

import (
	"crypto/ed25519" // intrinsic
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestManifest_Compile_Good(t *testing.T) {
	m := &Manifest{
		Code:        "go-io",
		Name:        "Core I/O",
		Description: "I/O provider",
		Version:     "0.3.0",
		Modules:     []string{"scm"},
	}
	info := BuildInfo{
		Targets:   []string{"linux/amd64", "darwin/arm64"},
		Checksums: "checksums.txt",
		SHA256:    "7b1f",
	}

	raw, err := Compile(m, info)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	var got Manifest
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal core json: %v", err)
	}
	if got.Code != m.Code || got.Name != m.Name || got.Version != m.Version {
		t.Fatalf("manifest fields did not round-trip: %#v", got)
	}
	if got.Build.Checksums != info.Checksums || got.Build.SHA256 != info.SHA256 {
		t.Fatalf("build info did not compile into core.json: %#v", got.Build)
	}
	if len(got.Build.Targets) != len(info.Targets) || got.Build.Targets[0] != info.Targets[0] {
		t.Fatalf("build targets did not compile into core.json: %#v", got.Build.Targets)
	}
	if m.Build.Checksums != "" || len(m.Build.Targets) != 0 {
		t.Fatalf("compile mutated source manifest build info: %#v", m.Build)
	}
}

func TestManifest_Compile_Bad_InvalidManifest(t *testing.T) {
	_, err := Compile(&Manifest{Name: "Core I/O", Version: "0.3.0"}, BuildInfo{})
	if err == nil {
		t.Fatal("expected invalid manifest error")
	}
}

func TestManifest_ParseCoreJSON_Good(t *testing.T) {
	raw, err := Compile(&Manifest{
		Code:    "go-io",
		Name:    "Core I/O",
		Version: "0.3.0",
	}, BuildInfo{
		Targets:   []string{"linux/amd64"},
		Checksums: "checksums.txt",
		SHA256:    "7b1f",
	})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	got, err := ParseCoreJSON(raw)
	if err != nil {
		t.Fatalf("parse core json: %v", err)
	}
	if got.Code != "go-io" || got.Build.SHA256 != "7b1f" {
		t.Fatalf("unexpected parsed manifest: %#v", got)
	}
	if len(got.Build.Targets) != 1 || got.Build.Targets[0] != "linux/amd64" {
		t.Fatalf("unexpected parsed build targets: %#v", got.Build.Targets)
	}
}

func TestManifest_ParseCoreJSON_Bad_Malformed(t *testing.T) {
	_, err := ParseCoreJSON([]byte(`{"code":`))
	if err == nil {
		t.Fatal("expected malformed JSON error")
	}
}

func TestManifest_Verify_Good(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	m := &Manifest{SignKey: base64.StdEncoding.EncodeToString(pub)}
	payload := []byte(`{"code":"go-io","version":"0.3.0"}`)

	if err := Sign(m, payload, priv); err != nil {
		t.Fatalf("sign: %v", err)
	}
	if m.Sign == "" {
		t.Fatal("expected signature to be populated")
	}
	if err := Verify(m, payload); err != nil {
		t.Fatalf("verify: %v", err)
	}
}

func TestManifest_Verify_Bad_WrongKey(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate signing key: %v", err)
	}
	wrongPub, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate wrong key: %v", err)
	}
	m := &Manifest{SignKey: base64.StdEncoding.EncodeToString(wrongPub)}
	payload := []byte(`{"code":"go-io","version":"0.3.0"}`)

	if err := Sign(m, payload, priv); err != nil {
		t.Fatalf("sign: %v", err)
	}
	if err := Verify(m, payload); err == nil {
		t.Fatal("expected wrong key verification to fail")
	}
}

func TestManifest_Verify_Bad_TamperedPayload(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	m := &Manifest{SignKey: base64.StdEncoding.EncodeToString(pub)}
	payload := []byte(`{"code":"go-io","version":"0.3.0"}`)

	if err := Sign(m, payload, priv); err != nil {
		t.Fatalf("sign: %v", err)
	}
	if err := Verify(m, []byte(`{"code":"go-io","version":"0.3.1"}`)); err == nil {
		t.Fatal("expected tampered payload verification to fail")
	}
}
