// SPDX-License-Identifier: EUPL-1.2

package verify

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
		t.Fatalf("register verify: %v", result.Value)
	}

	output := captureStdout(t, func() {
		if result := app.Cli().Run("verify", "--help"); !result.OK {
			t.Fatalf("verify help failed: %v", result.Value)
		}
	})

	if !strings.Contains(output, "usage: scm verify") {
		t.Fatalf("expected verify usage, got %q", output)
	}
}

func TestVerifySignedCoreJSON(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	root := t.TempDir()
	cm, err := manifest.CompileWithOptions(&manifest.Manifest{
		Code:    "demo",
		Name:    "Demo",
		Version: "1.0.0",
		SignKey: base64.StdEncoding.EncodeToString(pub),
	}, manifest.CompileOptions{SignKey: priv})
	if err != nil {
		t.Fatalf("compile signed manifest: %v", err)
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
		t.Fatalf("register verify: %v", result.Value)
	}
	output := captureStdout(t, func() {
		if result := app.Cli().Run("verify", "--root="+root); !result.OK {
			t.Fatalf("verify failed: %v", result.Value)
		}
	})

	if !strings.Contains(output, "verified") {
		t.Fatalf("expected verification output, got %q", output)
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

func TestCmdVerify_Register_Good(t *core.T) {
	app := core.New(core.WithOption("name", "scm"))
	result := Register(app)
	core.AssertTrue(t, result.OK)
	core.AssertTrue(t, app.Command("verify").OK)
}

func TestCmdVerify_Register_Bad(t *core.T) {
	result := Register(nil)
	core.AssertFalse(t, result.OK)
	core.AssertContains(t, result.Error(), "core app is required")
}

func TestCmdVerify_Register_Ugly(t *core.T) {
	app := core.New(core.WithOption("name", "scm"))
	first := Register(app)
	second := Register(app)
	core.AssertTrue(t, first.OK)
	core.AssertFalse(t, second.OK)
	core.AssertTrue(t, app.Command("verify").OK)
}
