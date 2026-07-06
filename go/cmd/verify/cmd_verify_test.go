// SPDX-License-Identifier: EUPL-1.2

package verify

import (
	"crypto/ed25519"
	"encoding/base64"
	"io"
	"os"
	"testing"

	core "dappco.re/go"
	"dappco.re/go/scm/manifest"
)

func TestRegisterHelp(t *testing.T) {
	app := core.New(core.WithOption("name", "scm"), core.WithCli())
	if result := Register(app); !result.OK {
		t.Fatalf("register verify: %v", result.Value)
	}

	output := captureStdout(t, func() {
		if result := app.Cli().Run("verify", "--help"); !result.OK {
			t.Fatalf("verify help failed: %v", result.Value)
		}
	})

	if !core.Contains(output, "usage: scm verify") {
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
	if r := core.WriteFile(core.PathJoin(root, "core.json"), raw, 0o600); !r.OK {
		t.Fatalf("write core.json: %v", r.Error())
	}

	app := core.New(core.WithOption("name", "scm"), core.WithCli())
	if result := Register(app); !result.OK {
		t.Fatalf("register verify: %v", result.Value)
	}
	output := captureStdout(t, func() {
		if result := app.Cli().Run("verify", "--root="+root); !result.OK {
			t.Fatalf("verify failed: %v", result.Value)
		}
	})

	if !core.Contains(output, "verified") {
		t.Fatalf("expected verification output, got %q", output)
	}
}

// captureStdout redirects the process stdout for the duration of fn and
// returns what was written. The verify command emits via core.Print(nil,
// ...), which targets core.Stdout() (os.Stdout) directly, so the capture
// has to swap the OS-level stream rather than a cli writer.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	saved := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = saved }()

	fn()

	_ = w.Close()
	os.Stdout = saved
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read captured stdout: %v", err)
	}
	return string(data)
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
