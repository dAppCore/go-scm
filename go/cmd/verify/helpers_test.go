// SPDX-License-Identifier: EUPL-1.2

package verify

import (
	"crypto/ed25519"
	"encoding/base64"

	core "dappco.re/go"
	"dappco.re/go/scm/manifest"
)

func TestCmdVerify_canonicalManifestBytes_Good(t *core.T) {
	raw, err := canonicalManifestBytes(&manifest.Manifest{Code: "demo", Sign: "sig", SignKey: "key"})
	core.AssertNoError(t, err)
	// Signature fields are stripped from the canonical payload.
	core.AssertFalse(t, core.Contains(string(raw), "sig"))
}

func TestCmdVerify_canonicalManifestBytes_Bad(t *core.T) {
	_, err := canonicalManifestBytes(nil)
	core.AssertError(t, err)
}

func TestCmdVerify_canonicalManifestBytes_Ugly(t *core.T) {
	raw, err := canonicalManifestBytes(&manifest.Manifest{})
	core.AssertNoError(t, err)
	core.AssertTrue(t, len(raw) > 0)
}

func TestCmdVerify_publicKey_Good(t *core.T) {
	pub, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	encoded := base64.StdEncoding.EncodeToString(pub)
	app := core.New(core.WithOption("name", "scm"))
	got, gerr := publicKey(app, core.NewOptions(core.Option{Key: "key", Value: encoded}))
	core.AssertNoError(t, gerr)
	core.AssertEqual(t, encoded, got)
}

func TestCmdVerify_publicKey_Bad(t *core.T) {
	app := core.New(core.WithOption("name", "scm"))
	_, err := publicKey(app, core.NewOptions(core.Option{Key: "key", Value: "not-base64!!"}))
	core.AssertError(t, err)
}

func TestCmdVerify_publicKey_Ugly(t *core.T) {
	// No key configured returns an empty key with no error.
	app := core.New(core.WithOption("name", "scm"))
	got, err := publicKey(app, core.NewOptions())
	core.AssertNoError(t, err)
	core.AssertEqual(t, "", got)
}

func TestCmdVerify_publicKey_WrongSize(t *core.T) {
	app := core.New(core.WithOption("name", "scm"))
	encoded := base64.StdEncoding.EncodeToString([]byte("too short"))
	_, err := publicKey(app, core.NewOptions(core.Option{Key: "key", Value: encoded}))
	core.AssertError(t, err)
}

func TestCmdVerify_failed_Good(t *core.T) {
	core.AssertFalse(t, failed(core.E("cmd.verify", "boom", nil)).OK)
}

func TestCmdVerify_failed_Bad(t *core.T) {
	core.AssertFalse(t, failed(nil).OK)
}

func TestCmdVerify_resultError_Good(t *core.T) {
	err := resultError("cmd.verify", "wrap", core.Fail(core.E("inner", "cause", nil)))
	core.AssertContains(t, err.Error(), "wrap")
}

func TestCmdVerify_resultError_Bad(t *core.T) {
	err := resultError("cmd.verify", "no cause", core.Ok(nil))
	core.AssertContains(t, err.Error(), "no cause")
}

func TestCmdVerify_readFile_Good(t *core.T) {
	app := core.New(core.WithOption("name", "scm"))
	dir := t.TempDir()
	path := core.PathJoin(dir, "data.txt")
	if r := core.WriteFile(path, []byte("hello"), 0o600); !r.OK {
		t.Fatalf("write: %s", r.Error())
	}
	raw, err := readFile(app, path)
	core.AssertNoError(t, err)
	core.AssertEqual(t, "hello", string(raw))
}

func TestCmdVerify_readFile_Bad(t *core.T) {
	_, err := readFile(nil, "x")
	core.AssertError(t, err)
}

func TestCmdVerify_readFile_Ugly(t *core.T) {
	app := core.New(core.WithOption("name", "scm"))
	_, err := readFile(app, core.PathJoin(t.TempDir(), "missing.txt"))
	core.AssertError(t, err)
}
