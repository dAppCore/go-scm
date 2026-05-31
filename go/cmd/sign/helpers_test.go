// SPDX-License-Identifier: EUPL-1.2

package sign

import (
	"crypto/ed25519"
	"encoding/base64"

	core "dappco.re/go"
	"dappco.re/go/scm/manifest"
)

func TestCmdSign_privateKey_Good(t *core.T) {
	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	encoded := base64.StdEncoding.EncodeToString(priv)
	app := core.New(core.WithOption("name", "scm"))
	got, gerr := privateKey(app, core.NewOptions(core.Option{Key: "key", Value: encoded}))
	core.AssertNoError(t, gerr)
	core.AssertEqual(t, ed25519.PrivateKeySize, len(got))
}

func TestCmdSign_privateKey_Bad(t *core.T) {
	app := core.New(core.WithOption("name", "scm"))
	_, err := privateKey(app, core.NewOptions())
	core.AssertError(t, err)
}

func TestCmdSign_privateKey_Ugly(t *core.T) {
	app := core.New(core.WithOption("name", "scm"))
	encoded := base64.StdEncoding.EncodeToString([]byte("too short"))
	_, err := privateKey(app, core.NewOptions(core.Option{Key: "key", Value: encoded}))
	core.AssertError(t, err)
}

func TestCmdSign_canonicalManifestBytes_Good(t *core.T) {
	raw, err := canonicalManifestBytes(&manifest.Manifest{Code: "demo", Sign: "sig", SignKey: "key"})
	core.AssertNoError(t, err)
	core.AssertFalse(t, core.Contains(string(raw), "sig"))
}

func TestCmdSign_canonicalManifestBytes_Bad(t *core.T) {
	_, err := canonicalManifestBytes(nil)
	core.AssertError(t, err)
}

func TestCmdSign_canonicalManifestBytes_Ugly(t *core.T) {
	raw, err := canonicalManifestBytes(&manifest.Manifest{})
	core.AssertNoError(t, err)
	core.AssertTrue(t, len(raw) > 0)
}

func TestCmdSign_compiledManifest_Good(t *core.T) {
	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	root := t.TempDir()
	manifestPath := core.PathJoin(root, "manifest.yaml")
	if r := core.WriteFile(manifestPath, []byte("code: demo\nname: Demo\nversion: 1.0.0\n"), 0o600); !r.OK {
		t.Fatalf("write manifest: %s", r.Error())
	}
	app := core.New(core.WithOption("name", "scm"))
	cm, gerr := compiledManifest(app, core.NewOptions(core.Option{Key: "manifest", Value: manifestPath}), root, priv)
	core.AssertNoError(t, gerr)
	core.AssertEqual(t, "demo", cm.Code)
	core.AssertTrue(t, cm.SignKey != "")
}

func TestCmdSign_compiledManifest_Bad(t *core.T) {
	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	app := core.New(core.WithOption("name", "scm"))
	_, gerr := compiledManifest(app, core.NewOptions(core.Option{Key: "manifest", Value: core.PathJoin(t.TempDir(), "missing.yaml")}), t.TempDir(), priv)
	core.AssertError(t, gerr)
}

func TestCmdSign_compiledManifest_Ugly(t *core.T) {
	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	app := core.New(core.WithOption("name", "scm"))
	// No manifest flag and no core.json under root: falls back to the
	// in-path read, which errors out for the missing file.
	_, gerr := compiledManifest(app, core.NewOptions(), t.TempDir(), priv)
	core.AssertError(t, gerr)
}

func TestCmdSign_failed_Good(t *core.T) {
	core.AssertFalse(t, failed(core.E("cmd.sign", "boom", nil)).OK)
}

func TestCmdSign_failed_Bad(t *core.T) {
	core.AssertFalse(t, failed(nil).OK)
}

func TestCmdSign_resultError_Good(t *core.T) {
	err := resultError("cmd.sign", "wrap", core.Fail(core.E("inner", "cause", nil)))
	core.AssertContains(t, err.Error(), "wrap")
}

func TestCmdSign_resultError_Bad(t *core.T) {
	err := resultError("cmd.sign", "no cause", core.Ok(nil))
	core.AssertContains(t, err.Error(), "no cause")
}

func TestCmdSign_readFile_Good(t *core.T) {
	app := core.New(core.WithOption("name", "scm"))
	path := core.PathJoin(t.TempDir(), "data.txt")
	if r := core.WriteFile(path, []byte("hello"), 0o600); !r.OK {
		t.Fatalf("write: %s", r.Error())
	}
	raw, err := readFile(app, path)
	core.AssertNoError(t, err)
	core.AssertEqual(t, "hello", string(raw))
}

func TestCmdSign_readFile_Bad(t *core.T) {
	_, err := readFile(nil, "x")
	core.AssertError(t, err)
}

func TestCmdSign_readFile_Ugly(t *core.T) {
	app := core.New(core.WithOption("name", "scm"))
	_, err := readFile(app, core.PathJoin(t.TempDir(), "missing.txt"))
	core.AssertError(t, err)
}
