// SPDX-License-Identifier: EUPL-1.2

package manifest

import (
	"crypto/ed25519"
	"encoding/base64"
	"testing"

	core "dappco.re/go"
	coreio "dappco.re/go/io"
)

const (
	sonarManifestTestCodeGoScmNameCore = "code: go-scm\nname: Core SCM\nversion: 0.9.0\n"
	sonarManifestTestCoreIO            = "Core I/O"
	sonarManifestTestCoreManifestYaml  = ".core/manifest.yaml"
	sonarManifestTestGoScm             = "go-scm"
	sonarManifestTestLinuxAmd64        = "linux/amd64"
	sonarManifestTestSignV             = "sign: %v"
)

func TestManifest_Compile_Good(t *testing.T) {
	m := &Manifest{
		Code:        "go-io",
		Name:        sonarManifestTestCoreIO,
		Description: "I/O provider",
		Version:     "0.3.0",
		Modules:     []string{"scm"},
	}
	info := BuildInfo{
		Targets:   []string{sonarManifestTestLinuxAmd64, "darwin/arm64"},
		Checksums: "checksums.txt",
		SHA256:    "7b1f",
	}

	raw, err := Compile(m, info)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	var got Manifest
	if r := core.JSONUnmarshal(raw, &got); !r.OK {
		t.Fatalf("unmarshal core json: %v", r.Value)
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
	_, err := Compile(&Manifest{Name: sonarManifestTestCoreIO, Version: "0.3.0"}, BuildInfo{})
	if err == nil {
		t.Fatal("expected invalid manifest error")
	}
}

func TestManifest_ParseCoreJSON_Good(t *testing.T) {
	raw, err := Compile(&Manifest{
		Code:    "go-io",
		Name:    sonarManifestTestCoreIO,
		Version: "0.3.0",
	}, BuildInfo{
		Targets:   []string{sonarManifestTestLinuxAmd64},
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
	if len(got.Build.Targets) != 1 || got.Build.Targets[0] != sonarManifestTestLinuxAmd64 {
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
		t.Fatalf(sonarManifestTestSignV, err)
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
		t.Fatalf(sonarManifestTestSignV, err)
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
		t.Fatalf(sonarManifestTestSignV, err)
	}
	if err := Verify(m, []byte(`{"code":"go-io","version":"0.3.1"}`)); err == nil {
		t.Fatal("expected tampered payload verification to fail")
	}
}

func testManifestFixture() *Manifest {
	return &Manifest{Code: sonarManifestTestGoScm, Name: "Core SCM", Version: "0.9.0"}
}

func TestManifestV090_Parse_Good(t *core.T) {
	m, err := Parse([]byte(sonarManifestTestCodeGoScmNameCore))
	core.AssertNoError(t, err)
	core.AssertEqual(t, sonarManifestTestGoScm, m.Code)
}

func TestManifestV090_Parse_Bad(t *core.T) {
	_, err := Parse([]byte("code: ["))
	core.AssertError(
		t, err,
	)
}

func TestManifestV090_Parse_Ugly(t *core.T) {
	m, err := Parse(nil)
	core.AssertNoError(t, err)
	core.AssertEqual(t, "", m.Code)
}

func TestManifestV090_Manifest_IsProvider_Good(t *core.T) {
	m := &Manifest{Namespace: "scm", Binary: "scm-provider"}
	core.AssertTrue(
		t, m.IsProvider(),
	)
}

func TestManifestV090_Manifest_IsProvider_Bad(t *core.T) {
	m := &Manifest{Namespace: "scm"}
	core.AssertFalse(
		t, m.IsProvider(),
	)
}

func TestManifestV090_Manifest_IsProvider_Ugly(t *core.T) {
	var m *Manifest
	core.AssertFalse(
		t, m.IsProvider(),
	)
}

func TestManifestV090_Manifest_SlotNames_Good(t *core.T) {
	m := &Manifest{Slots: map[string]string{"b": "write", "a": "read"}}
	got := m.SlotNames()
	core.AssertEqual(t, []string{"read", "write"}, got)
}

func TestManifestV090_Manifest_SlotNames_Bad(t *core.T) {
	m := &Manifest{}
	got := m.SlotNames()
	core.AssertNil(t, got)
}

func TestManifestV090_Manifest_SlotNames_Ugly(t *core.T) {
	m := &Manifest{Slots: map[string]string{"a": "read", "b": " read ", "c": ""}}
	got := m.SlotNames()
	core.AssertEqual(t, []string{"read"}, got)
}

func TestManifestV090_Manifest_DefaultDaemon_Good(t *core.T) {
	m := &Manifest{Daemons: map[string]DaemonSpec{"api": {Binary: "scm-api"}}}
	name, spec, ok := m.DefaultDaemon()
	core.AssertTrue(t, ok)
	core.AssertEqual(t, "api", name)
	core.AssertEqual(t, "scm-api", spec.Binary)
}

func TestManifestV090_Manifest_DefaultDaemon_Bad(t *core.T) {
	m := &Manifest{}
	name, _, ok := m.DefaultDaemon()
	core.AssertFalse(t, ok)
	core.AssertEqual(t, "", name)
}

func TestManifestV090_Manifest_DefaultDaemon_Ugly(t *core.T) {
	m := &Manifest{Daemons: map[string]DaemonSpec{
		"a": {Default: true},
		"b": {Default: true},
	}}
	name, _, ok := m.DefaultDaemon()
	core.AssertFalse(t, ok)
	core.AssertEqual(t, "", name)
}

func TestManifestV090_Compile_Good(t *core.T) {
	raw, err := Compile(testManifestFixture(), BuildInfo{SHA256: "abc123"})
	core.AssertNoError(t, err)
	core.AssertContains(t, string(raw), "abc123")
}

func TestManifestV090_Compile_Bad(t *core.T) {
	_, err := Compile(&Manifest{Name: "Core SCM", Version: "0.9.0"}, BuildInfo{})
	core.AssertError(
		t, err,
	)
}

func TestManifestV090_Compile_Ugly(t *core.T) {
	_, err := Compile(nil, BuildInfo{})
	core.AssertError(
		t, err,
	)
}

func TestManifestV090_ParseCoreJSON_Good(t *core.T) {
	raw, err := Compile(testManifestFixture(), BuildInfo{Targets: []string{sonarManifestTestLinuxAmd64}})
	core.RequireNoError(t, err)
	m, err := ParseCoreJSON(raw)
	core.AssertNoError(t, err)
	core.AssertEqual(t, sonarManifestTestGoScm, m.Code)
}

func TestManifestV090_ParseCoreJSON_Bad(t *core.T) {
	_, err := ParseCoreJSON([]byte(`{"code":`))
	core.AssertError(t, err)
}

func TestManifestV090_ParseCoreJSON_Ugly(t *core.T) {
	_, err := ParseCoreJSON([]byte(`{}`))
	core.AssertError(
		t, err,
	)
}

func TestManifestV090_CompileWithOptions_Good(t *core.T) {
	cm, err := CompileWithOptions(testManifestFixture(), CompileOptions{Commit: "abc123", Tag: "v0.9.0"})
	core.AssertNoError(t, err)
	core.AssertEqual(t, "abc123", cm.Commit)
	core.AssertEqual(t, "v0.9.0", cm.Tag)
}

func TestManifestV090_CompileWithOptions_Bad(t *core.T) {
	_, err := CompileWithOptions(&Manifest{Code: sonarManifestTestGoScm}, CompileOptions{})
	core.AssertError(
		t, err,
	)
}

func TestManifestV090_CompileWithOptions_Ugly(t *core.T) {
	_, priv, err := ed25519.GenerateKey(nil)
	core.RequireNoError(t, err)
	cm, err := CompileWithOptions(testManifestFixture(), CompileOptions{SignKey: priv})
	core.AssertNoError(t, err)
	core.AssertTrue(t, cm.Sign != "")
}

func TestManifestV090_MarshalJSON_Good(t *core.T) {
	cm, err := CompileWithOptions(testManifestFixture(), CompileOptions{Commit: "abc123"})
	core.RequireNoError(t, err)
	raw, err := MarshalJSON(cm)
	core.AssertNoError(t, err)
	core.AssertContains(t, string(raw), "abc123")
}

func TestManifestV090_MarshalJSON_Bad(t *core.T) {
	_, err := MarshalJSON(nil)
	core.AssertError(
		t, err,
	)
}

func TestManifestV090_MarshalJSON_Ugly(t *core.T) {
	raw, err := MarshalJSON(&CompiledManifest{})
	core.AssertNoError(t, err)
	core.AssertContains(t, string(raw), "code")
}

func TestManifestV090_ParseCompiled_Good(t *core.T) {
	cm, err := CompileWithOptions(testManifestFixture(), CompileOptions{Commit: "abc123"})
	core.RequireNoError(t, err)
	raw, err := MarshalJSON(cm)
	core.RequireNoError(t, err)
	got, err := ParseCompiled(raw)
	core.AssertNoError(t, err)
	core.AssertEqual(t, "abc123", got.Commit)
}

func TestManifestV090_ParseCompiled_Bad(t *core.T) {
	_, err := ParseCompiled([]byte(`{"code":`))
	core.AssertError(t, err)
}

func TestManifestV090_ParseCompiled_Ugly(t *core.T) {
	got, err := ParseCompiled([]byte(`{}`))
	core.AssertNoError(t, err)
	core.AssertEqual(t, "", got.Code)
}

func TestManifestV090_LoadCompiled_Good(t *core.T) {
	medium := coreio.NewMemoryMedium()
	cm, err := CompileWithOptions(testManifestFixture(), CompileOptions{Commit: "abc123"})
	core.RequireNoError(t, err)
	core.RequireNoError(t, WriteCompiled(medium, ".", cm))
	got, err := LoadCompiled(medium, ".")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "abc123", got.Commit)
}

func TestManifestV090_LoadCompiled_Bad(t *core.T) {
	_, err := LoadCompiled(nil, ".")
	core.AssertError(
		t, err,
	)
}

func TestManifestV090_LoadCompiled_Ugly(t *core.T) {
	medium := coreio.NewMemoryMedium()
	_, err := LoadCompiled(medium, "missing")
	core.AssertError(t, err)
}

func TestManifestV090_WriteCompiled_Good(t *core.T) {
	medium := coreio.NewMemoryMedium()
	cm, err := CompileWithOptions(testManifestFixture(), CompileOptions{})
	core.RequireNoError(t, err)
	err = WriteCompiled(medium, "pkg", cm)
	core.AssertNoError(t, err)
	raw, readErr := medium.Read("pkg/core.json")
	core.AssertNoError(t, readErr)
	core.AssertContains(t, raw, sonarManifestTestGoScm)
}

func TestManifestV090_WriteCompiled_Bad(t *core.T) {
	err := WriteCompiled(nil, ".", &CompiledManifest{})
	core.AssertError(
		t, err,
	)
}

func TestManifestV090_WriteCompiled_Ugly(t *core.T) {
	medium := coreio.NewMemoryMedium()
	err := WriteCompiled(medium, ".", nil)
	core.AssertError(t, err)
}

func TestManifestV090_Load_Good(t *core.T) {
	medium := coreio.NewMemoryMedium()
	core.RequireNoError(t, medium.Write(sonarManifestTestCoreManifestYaml, sonarManifestTestCodeGoScmNameCore))
	m, err := Load(medium, ".")
	core.AssertNoError(t, err)
	core.AssertEqual(t, sonarManifestTestGoScm, m.Code)
}

func TestManifestV090_Load_Bad(t *core.T) {
	_, err := Load(nil, ".")
	core.AssertError(
		t, err,
	)
}

func TestManifestV090_Load_Ugly(t *core.T) {
	medium := coreio.NewMemoryMedium()
	_, err := Load(medium, "missing")
	core.AssertError(t, err)
}

func TestManifestV090_LoadVerified_Good(t *core.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	core.RequireNoError(t, err)
	m := testManifestFixture()
	payload, err := canonicalManifestBytes(m)
	core.RequireNoError(t, err)
	m.SignKey = base64.StdEncoding.EncodeToString(pub)
	core.RequireNoError(t, Sign(m, payload, priv))
	raw, err := MarshalYAML(m)
	core.RequireNoError(t, err)
	medium := coreio.NewMemoryMedium()
	core.RequireNoError(t, medium.Write(sonarManifestTestCoreManifestYaml, string(raw)))
	got, err := LoadVerified(medium, ".", pub)
	core.AssertNoError(t, err)
	core.AssertEqual(t, sonarManifestTestGoScm, got.Code)
}

func TestManifestV090_LoadVerified_Bad(t *core.T) {
	medium := coreio.NewMemoryMedium()
	core.RequireNoError(t, medium.Write(sonarManifestTestCoreManifestYaml, sonarManifestTestCodeGoScmNameCore))
	_, err := LoadVerified(medium, ".", nil)
	core.AssertError(t, err)
}

func TestManifestV090_LoadVerified_Ugly(t *core.T) {
	_, err := LoadVerified(nil, ".", nil)
	core.AssertError(
		t, err,
	)
}

func TestManifestV090_MarshalYAML_Good(t *core.T) {
	raw, err := MarshalYAML(testManifestFixture())
	core.AssertNoError(t, err)
	core.AssertContains(t, string(raw), sonarManifestTestGoScm)
}

func TestManifestV090_MarshalYAML_Bad(t *core.T) {
	_, err := MarshalYAML(nil)
	core.AssertError(
		t, err,
	)
}

func TestManifestV090_MarshalYAML_Ugly(t *core.T) {
	raw, err := MarshalYAML(&Manifest{Slots: map[string]string{"default": "read"}})
	core.AssertNoError(t, err)
	core.AssertContains(t, string(raw), "slots")
}

func TestManifestV090_Sign_Good(t *core.T) {
	_, priv, err := ed25519.GenerateKey(nil)
	core.RequireNoError(t, err)
	m := testManifestFixture()
	err = Sign(m, []byte("payload"), priv)
	core.AssertNoError(t, err)
	core.AssertTrue(t, m.Sign != "")
}

func TestManifestV090_Sign_Bad(t *core.T) {
	err := Sign(nil, []byte("payload"), nil)
	core.AssertError(
		t, err,
	)
}

func TestManifestV090_Sign_Ugly(t *core.T) {
	m := testManifestFixture()
	err := Sign(m, []byte("payload"), ed25519.PrivateKey([]byte("short")))
	core.AssertError(t, err)
}

func TestManifestV090_Verify_Good(t *core.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	core.RequireNoError(t, err)
	m := testManifestFixture()
	m.SignKey = base64.StdEncoding.EncodeToString(pub)
	payload := []byte("payload")
	core.RequireNoError(t, Sign(m, payload, priv))
	err = Verify(m, payload)
	core.AssertNoError(t, err)
}

func TestManifestV090_Verify_Bad(t *core.T) {
	err := Verify(nil, []byte("payload"))
	core.AssertError(
		t, err,
	)
}

func TestManifestV090_Verify_Ugly(t *core.T) {
	m := testManifestFixture()
	m.SignKey = "not-base64"
	m.Sign = "not-base64"
	err := Verify(m, []byte("payload"))
	core.AssertError(t, err)
}

func TestManifest_Parse_Good(t *testing.T) {
	target := "Parse"
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

func TestManifest_Parse_Bad(t *testing.T) {
	target := "Parse"
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

func TestManifest_Parse_Ugly(t *testing.T) {
	target := "Parse"
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

func TestManifest_Manifest_IsProvider_Good(t *testing.T) {
	target := "Manifest_IsProvider"
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

func TestManifest_Manifest_IsProvider_Bad(t *testing.T) {
	target := "Manifest_IsProvider"
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

func TestManifest_Manifest_IsProvider_Ugly(t *testing.T) {
	target := "Manifest_IsProvider"
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

func TestManifest_Manifest_SlotNames_Good(t *testing.T) {
	target := "Manifest_SlotNames"
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

func TestManifest_Manifest_SlotNames_Bad(t *testing.T) {
	target := "Manifest_SlotNames"
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

func TestManifest_Manifest_SlotNames_Ugly(t *testing.T) {
	target := "Manifest_SlotNames"
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

func TestManifest_Manifest_DefaultDaemon_Good(t *testing.T) {
	target := "Manifest_DefaultDaemon"
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

func TestManifest_Manifest_DefaultDaemon_Bad(t *testing.T) {
	target := "Manifest_DefaultDaemon"
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

func TestManifest_Manifest_DefaultDaemon_Ugly(t *testing.T) {
	target := "Manifest_DefaultDaemon"
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
