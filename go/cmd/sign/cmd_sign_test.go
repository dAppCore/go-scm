// SPDX-License-Identifier: EUPL-1.2

package sign

import (
	"crypto/ed25519"
	"encoding/base64"
	"io"
	`os`
	`path/filepath`
	`strings`
	"testing"

	core "dappco.re/go"
	"dappco.re/go/scm/manifest"
)

func TestRegisterHelp(t *testing.T) {
	app := core.New(core.WithOption("name", "scm"))
	if result := Register(app); !result.OK {
		t.Fatalf("register sign: %v", result.Value)
	}

	output := captureStdout(t, func() {
		if result := app.Cli().Run("sign", "--help"); !result.OK {
			t.Fatalf("sign help failed: %v", result.Value)
		}
	})

	if !strings.Contains(output, "usage: scm sign") {
		t.Fatalf("expected sign usage, got %q", output)
	}
}

func TestSignWritesSignature(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	root := t.TempDir()
	cm, err := manifest.CompileWithOptions(&manifest.Manifest{
		Code:    "demo",
		Name:    "Demo",
		Version: "1.0.0",
	}, manifest.CompileOptions{})
	if err != nil {
		t.Fatalf("compile manifest: %v", err)
	}
	raw, err := manifest.MarshalJSON(cm)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "core.json"), raw, 0o600); err != nil {
		t.Fatalf("write core.json: %v", err)
	}

	app := core.New(core.WithOption("name", "scm"))
	if result := Register(app); !result.OK {
		t.Fatalf("register sign: %v", result.Value)
	}
	key := base64.StdEncoding.EncodeToString(priv)
	if result := app.Cli().Run("sign", "--root="+root, "--key="+key); !result.OK {
		t.Fatalf("sign failed: %v", result.Value)
	}

	signedRaw, err := os.ReadFile(filepath.Join(root, "core.json"))
	if err != nil {
		t.Fatalf("read signed core.json: %v", err)
	}
	signed, err := manifest.ParseCompiled(signedRaw)
	if err != nil {
		t.Fatalf("parse signed core.json: %v", err)
	}
	if signed.Sign == "" || signed.SignKey == "" {
		t.Fatalf("expected signature and sign key: %#v", signed.Manifest)
	}
	payload, err := canonicalManifestBytes(&signed.Manifest)
	if err != nil {
		t.Fatalf("canonical manifest: %v", err)
	}
	if err := manifest.Verify(&signed.Manifest, payload); err != nil {
		t.Fatalf("verify signed manifest: %v", err)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = old
	}()

	fn()
	if err := w.Close(); err != nil {
		t.Fatalf("close stdout pipe: %v", err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	return string(out)
}

func TestCmdSign_Register_Good(t *core.T) {
	app := core.New(core.WithOption("name", "scm"))
	result := Register(app)
	core.AssertTrue(t, result.OK)
	core.AssertTrue(t, app.Command("sign").OK)
}

func TestCmdSign_Register_Bad(t *core.T) {
	result := Register(nil)
	core.AssertFalse(t, result.OK)
	core.AssertContains(t, result.Error(), "core app is required")
}

func TestCmdSign_Register_Ugly(t *core.T) {
	app := core.New(core.WithOption("name", "scm"))
	first := Register(app)
	second := Register(app)
	core.AssertTrue(t, first.OK)
	core.AssertFalse(t, second.OK)
	core.AssertTrue(t, app.Command("sign").OK)
}
