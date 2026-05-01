// SPDX-License-Identifier: EUPL-1.2

package pkg

import (
	"io"
	"os"
	"testing"

	core "dappco.re/go"
	"dappco.re/go/scm/marketplace"
)

func TestRegisterHelp(t *testing.T) {
	app := core.New(core.WithOption("name", "scm"))
	if result := Register(app); !result.OK {
		t.Fatalf("register pkg: %v", result.Value)
	}

	output := captureStdout(t, func() {
		if result := app.Cli().Run("pkg", "--help"); !result.OK {
			t.Fatalf("pkg help failed: %v", result.Value)
		}
	})

	if !core.Contains(output, "usage: scm pkg") {
		t.Fatalf("expected pkg usage, got %q", output)
	}
}

func TestPkgWritesMarketplaceIndex(t *testing.T) {
	root := t.TempDir()
	if r := core.MkdirAll(core.PathJoin(root, ".core"), 0o755); !r.OK {
		t.Fatalf("mkdir manifest dir: %v", r.Error())
	}
	if r := core.WriteFile(core.PathJoin(root, ".core", "manifest.yaml"), []byte(`code: demo
name: Demo
version: 1.0.0
modules: [provider]
`), 0o600); !r.OK {
		t.Fatalf("write manifest: %v", r.Error())
	}

	app := core.New(core.WithOption("name", "scm"))
	if result := Register(app); !result.OK {
		t.Fatalf("register pkg: %v", result.Value)
	}
	if result := app.Cli().Run("pkg", "--root="+root, "--base-url=https://forge.example", "--org=modules"); !result.OK {
		t.Fatalf("pkg failed: %v", result.Value)
	}

	rawR := core.ReadFile(core.PathJoin(root, "marketplace", "index.json"))
	if !rawR.OK {
		t.Fatalf("read index: %v", rawR.Error())
	}
	raw := rawR.Value.([]byte)
	idx, err := marketplace.ParseIndex(raw)
	if err != nil {
		t.Fatalf("parse index: %v", err)
	}
	if len(idx.Modules) != 1 || idx.Modules[0].Code != "demo" || idx.Modules[0].Category != "provider" {
		t.Fatalf("unexpected index: %#v", idx)
	}
	if idx.Modules[0].Repo != "https://forge.example/modules/demo" {
		t.Fatalf("unexpected repo: %q", idx.Modules[0].Repo)
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

func TestCmdPkg_Register_Good(t *core.T) {
	app := core.New(core.WithOption("name", "scm"))
	result := Register(app)
	core.AssertTrue(t, result.OK)
	core.AssertTrue(t, app.Command("pkg").OK)
}

func TestCmdPkg_Register_Bad(t *core.T) {
	result := Register(nil)
	core.AssertFalse(t, result.OK)
	core.AssertContains(t, result.Error(), "core app is required")
}

func TestCmdPkg_Register_Ugly(t *core.T) {
	app := core.New(core.WithOption("name", "scm"))
	first := Register(app)
	second := Register(app)
	core.AssertTrue(t, first.OK)
	core.AssertFalse(t, second.OK)
	core.AssertTrue(t, app.Command("pkg").OK)
}
