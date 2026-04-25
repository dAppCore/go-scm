// SPDX-License-Identifier: EUPL-1.2

// Package sign wires package signing into the scm CLI.
package sign

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	core "dappco.re/go/core"
	"dappco.re/go/scm/manifest"
)

const usage = "usage: scm sign [--root=DIR] [--in=core.json] [--out=core.json] [--manifest=FILE] [--key=BASE64_OR_FILE] [--key-file=FILE]"

// Register attaches the sign command to the parent Core command tree.
//
// The command signs the package manifest embedded in core.json. If --manifest
// is supplied, it compiles that manifest first and writes a signed core.json.
func Register(app *core.Core) core.Result {
	if app == nil {
		return core.Result{Value: core.E("cmd.sign.Register", "core app is required", nil), OK: false}
	}
	return app.Command("sign", core.Command{Action: run})
}

func run(opts core.Options) core.Result {
	if wantsHelp(opts) {
		core.Print(nil, usage)
		return core.Result{OK: true}
	}

	priv, err := privateKey(opts)
	if err != nil {
		return failed(err)
	}

	root := option(opts, "root", ".")
	outPath := option(opts, "out", filepath.Join(root, "core.json"))

	cm, err := compiledManifest(opts, root, priv)
	if err != nil {
		return failed(err)
	}

	raw, err := manifest.MarshalJSON(cm)
	if err != nil {
		return failed(err)
	}
	if err := mkdirParent(outPath); err != nil {
		return failed(err)
	}
	if err := os.WriteFile(outPath, raw, 0o600); err != nil {
		return failed(err)
	}

	core.Print(nil, "%s", outPath)
	return core.Result{OK: true}
}

func compiledManifest(opts core.Options, root string, priv ed25519.PrivateKey) (*manifest.CompiledManifest, error) {
	pub := priv.Public().(ed25519.PublicKey)
	signKey := base64.StdEncoding.EncodeToString(pub)

	if path := strings.TrimSpace(opts.String("manifest")); path != "" {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		m, err := manifest.Parse(raw)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(m.SignKey) == "" {
			m.SignKey = signKey
		}
		m.Sign = ""
		return manifest.CompileWithOptions(m, manifest.CompileOptions{SignKey: priv})
	}

	inPath := option(opts, "in", filepath.Join(root, "core.json"))
	raw, err := os.ReadFile(inPath)
	if err != nil {
		return nil, err
	}
	cm, err := manifest.ParseCompiled(raw)
	if err != nil {
		return nil, err
	}
	if _, err := manifest.Compile(&cm.Manifest, cm.Build); err != nil {
		return nil, err
	}
	if strings.TrimSpace(cm.SignKey) == "" {
		cm.SignKey = signKey
	}
	payload, err := canonicalManifestBytes(&cm.Manifest)
	if err != nil {
		return nil, err
	}
	if err := manifest.Sign(&cm.Manifest, payload, priv); err != nil {
		return nil, err
	}
	return cm, nil
}

func privateKey(opts core.Options) (ed25519.PrivateKey, error) {
	value := strings.TrimSpace(opts.String("key"))
	if path := strings.TrimSpace(opts.String("key-file")); path != "" {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		value = strings.TrimSpace(string(raw))
	}
	if value == "" {
		value = strings.TrimSpace(os.Getenv("SCM_SIGN_KEY"))
	}
	if value == "" {
		return nil, errors.New("signing key is required")
	}
	if raw, err := os.ReadFile(value); err == nil {
		value = strings.TrimSpace(string(raw))
	}
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, err
	}
	if len(decoded) != ed25519.PrivateKeySize {
		return nil, errors.New("signing key must be a base64 ed25519 private key")
	}
	return ed25519.PrivateKey(decoded), nil
}

func canonicalManifestBytes(m *manifest.Manifest) ([]byte, error) {
	if m == nil {
		return nil, errors.New("manifest is required")
	}
	cp := *m
	cp.Sign = ""
	cp.SignKey = ""
	return json.Marshal(cp)
}

func option(opts core.Options, key, fallback string) string {
	if value := strings.TrimSpace(opts.String(key)); value != "" {
		return value
	}
	return fallback
}

func mkdirParent(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

func wantsHelp(opts core.Options) bool {
	return opts.Bool("help") || opts.Bool("h")
}

func failed(err error) core.Result {
	return core.Result{Value: err, OK: false}
}
