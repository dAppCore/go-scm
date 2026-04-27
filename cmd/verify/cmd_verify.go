// SPDX-License-Identifier: EUPL-1.2

// Package verify wires package signature verification into the scm CLI.
package verify

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

const usage = "usage: scm verify [--root=DIR] [--in=core.json] [--manifest=FILE] [--key=BASE64_OR_FILE] [--key-file=FILE]"

// Register attaches the verify command to the parent Core command tree.
//
// The command verifies the signature stored in core.json or manifest.yaml
// against the canonical package manifest payload.
func Register(app *core.Core) core.Result {
	if app == nil {
		return core.Result{Value: core.E("cmd.verify.Register", "core app is required", nil), OK: false}
	}
	return app.Command("verify", core.Command{Action: run})
}

func run(opts core.Options) core.Result {
	if wantsHelp(opts) {
		core.Print(nil, usage)
		return core.Result{OK: true}
	}

	root := option(opts, "root", ".")
	input := option(opts, "in", filepath.Join(root, "core.json"))

	m, source, err := loadManifest(opts, input)
	if err != nil {
		return failed(err)
	}
	if key, err := publicKey(opts); err != nil {
		return failed(err)
	} else if key != "" {
		cp := *m
		cp.SignKey = key
		m = &cp
	}

	payload, err := canonicalManifestBytes(m)
	if err != nil {
		return failed(err)
	}
	if err := manifest.Verify(m, payload); err != nil {
		return failed(err)
	}

	core.Print(nil, "verified %s", source)
	return core.Result{OK: true}
}

func loadManifest(opts core.Options, defaultInput string) (*manifest.Manifest, string, error) {
	if path := strings.TrimSpace(opts.String("manifest")); path != "" {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, path, err
		}
		m, err := manifest.Parse(raw)
		return m, path, err
	}

	raw, err := os.ReadFile(defaultInput)
	if err != nil {
		return nil, defaultInput, err
	}
	cm, err := manifest.ParseCompiled(raw)
	if err != nil {
		return nil, defaultInput, err
	}
	if _, err := manifest.Compile(&cm.Manifest, cm.Build); err != nil {
		return nil, defaultInput, err
	}
	return &cm.Manifest, defaultInput, nil
}

func publicKey(opts core.Options) (string, error) {
	value := strings.TrimSpace(opts.String("key"))
	if path := strings.TrimSpace(opts.String("key-file")); path != "" {
		raw, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		value = strings.TrimSpace(string(raw))
	}
	if value == "" {
		return "", nil
	}
	if raw, err := os.ReadFile(value); err == nil {
		value = strings.TrimSpace(string(raw))
	}
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}
	if len(decoded) != ed25519.PublicKeySize {
		return "", errors.New("verification key must be a base64 ed25519 public key")
	}
	return base64.StdEncoding.EncodeToString(decoded), nil
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

func wantsHelp(opts core.Options) bool {
	return opts.Bool("help") || opts.Bool("h")
}

func failed(err error) core.Result {
	return core.Result{Value: err, OK: false}
}
