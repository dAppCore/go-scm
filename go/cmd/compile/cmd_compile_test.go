// SPDX-License-Identifier: EUPL-1.2

package compile

import (
	cli "dappco.re/go/cli/pkg/cli"
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

	if !core.Contains(output, "usage: scm compile") {
		t.Fatalf("expected compile usage, got %q", output)
	}
}

func TestCompileWritesCoreJSON(t *testing.T) {
	root := t.TempDir()
	if r := core.MkdirAll(core.PathJoin(root, ".core"), 0o755); !r.OK {
		t.Fatalf("mkdir manifest dir: %v", r.Error())
	}
	if r := core.WriteFile(core.PathJoin(root, ".core", "manifest.yaml"), []byte(`code: demo
name: Demo
version: 1.0.0
`), 0o600); !r.OK {
		t.Fatalf("write manifest: %v", r.Error())
	}

	app := core.New(core.WithOption("name", "scm"))
	if result := Register(app); !result.OK {
		t.Fatalf("register compile: %v", result.Value)
	}
	if result := app.Cli().Run("compile", "--root="+root, "--commit=abc123", "--targets=linux/amd64,darwin/arm64"); !result.OK {
		t.Fatalf("compile failed: %v", result.Value)
	}

	rawR := core.ReadFile(core.PathJoin(root, "core.json"))
	if !rawR.OK {
		t.Fatalf("read core.json: %v", rawR.Error())
	}
	raw := rawR.Value.([]byte)
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
	out := core.NewBuilder()
	cli.SetStdout(out)
	defer cli.SetStdout(nil)
	fn()
	return out.String()
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
