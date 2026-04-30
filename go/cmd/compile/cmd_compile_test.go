// SPDX-License-Identifier: EUPL-1.2

package compile

import (
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
		t.Fatalf("register compile: %v", result.Value)
	}

	output := captureStdout(t, func() {
		if result := app.Cli().Run("compile", "--help"); !result.OK {
			t.Fatalf("compile help failed: %v", result.Value)
		}
	})

	if !strings.Contains(output, "usage: scm compile") {
		t.Fatalf("expected compile usage, got %q", output)
	}
}

func TestCompileWritesCoreJSON(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".core"), 0o755); err != nil {
		t.Fatalf("mkdir manifest dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".core", "manifest.yaml"), []byte(`code: demo
name: Demo
version: 1.0.0
`), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	app := core.New(core.WithOption("name", "scm"))
	if result := Register(app); !result.OK {
		t.Fatalf("register compile: %v", result.Value)
	}
	if result := app.Cli().Run("compile", "--root="+root, "--commit=abc123", "--targets=linux/amd64,darwin/arm64"); !result.OK {
		t.Fatalf("compile failed: %v", result.Value)
	}

	raw, err := os.ReadFile(filepath.Join(root, "core.json"))
	if err != nil {
		t.Fatalf("read core.json: %v", err)
	}
	cm, err := manifest.ParseCompiled(raw)
	if err != nil {
		t.Fatalf("parse core.json: %v", err)
	}
	if cm.Code != "demo" || cm.Commit != "abc123" {
		t.Fatalf("unexpected compiled manifest: %#v", cm)
	}
	if len(cm.Build.Targets) != 2 || cm.Build.Targets[1] != "darwin/arm64" {
		t.Fatalf("unexpected targets: %#v", cm.Build.Targets)
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

func TestCmdCompile_Register_Good(t *core.T) {
	app := core.New(core.WithOption("name", "scm"))
	result := Register(app)
	core.AssertTrue(t, result.OK)
	core.AssertTrue(t, app.Command("compile").OK)
}

func TestCmdCompile_Register_Bad(t *core.T) {
	result := Register(nil)
	core.AssertFalse(t, result.OK)
	core.AssertContains(t, result.Error(), "core app is required")
}

func TestCmdCompile_Register_Ugly(t *core.T) {
	app := core.New(core.WithOption("name", "scm"))
	first := Register(app)
	second := Register(app)
	core.AssertTrue(t, first.OK)
	core.AssertFalse(t, second.OK)
	core.AssertTrue(t, app.Command("compile").OK)
}
