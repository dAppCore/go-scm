// SPDX-License-Identifier: EUPL-1.2

// Package verify wires package signature verification into the scm CLI.
package verify

import (
	"crypto/ed25519"
	"encoding/base64"

	core "dappco.re/go"
	"dappco.re/go/scm/manifest"
)

const usage = "usage: scm verify [--root=DIR] [--in=core.json] [--manifest=FILE] [--key=BASE64_OR_FILE] [--key-file=FILE]"

// Register attaches the verify command to the parent Core command tree.
//
// The command verifies the signature stored in core.json or manifest.yaml
// against the canonical package manifest payload.
func Register(app *core.Core) core.Result {
	if app == nil {
		return core.Fail(core.E("cmd.verify.Register", "core app is required", nil))
	}
	return app.Command("verify", core.Command{Action: run(app)})
}

func run(app *core.Core) core.CommandAction {
	return func(opts core.Options) core.Result {
		if wantsHelp(opts) {
			core.Print(nil, usage)
			return core.Ok(nil)
		}

		root := option(opts, "root", ".")
		input := option(opts, "in", core.PathJoin(root, "core.json"))

		m, source, err := loadManifest(app, opts, input)
		if err != nil {
			return failed(err)
		}
		if key, err := publicKey(app, opts); err != nil {
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
		return core.Ok(nil)
	}
}

func loadManifest(app *core.Core, opts core.Options, defaultInput string) (*manifest.Manifest, string, error) {
	if path := core.Trim(opts.String("manifest")); path != "" {
		raw, err := readFile(app, path)
		if err != nil {
			return nil, path, err
		}
		m, err := manifest.Parse(raw)
		return m, path, err
	}

	raw, err := readFile(app, defaultInput)
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

func publicKey(app *core.Core, opts core.Options) (string, error) {
	value := core.Trim(opts.String("key"))
	if path := core.Trim(opts.String("key-file")); path != "" {
		raw, err := readFile(app, path)
		if err != nil {
			return "", err
		}
		value = core.Trim(string(raw))
	}
	if value == "" {
		return "", nil
	}
	if raw, err := readFile(app, value); err == nil {
		value = core.Trim(string(raw))
	}
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}
	if len(decoded) != ed25519.PublicKeySize {
		return "", core.E("cmd.verify.publicKey", "verification key must be a base64 ed25519 public key", nil)
	}
	return base64.StdEncoding.EncodeToString(decoded), nil
}

func canonicalManifestBytes(m *manifest.Manifest) ([]byte, error) {
	if m == nil {
		return nil, core.E("cmd.verify.canonicalManifestBytes", "manifest is required", nil)
	}
	cp := *m
	cp.Sign = ""
	cp.SignKey = ""
	r := core.JSONMarshal(cp)
	if !r.OK {
		return nil, resultError("cmd.verify.canonicalManifestBytes", "marshal manifest", r)
	}
	raw, ok := r.Value.([]byte)
	if !ok {
		return nil, core.E("cmd.verify.canonicalManifestBytes", "marshal returned invalid payload", nil)
	}
	return raw, nil
}

func option(opts core.Options, key, fallback string) string {
	if value := core.Trim(opts.String(key)); value != "" {
		return value
	}
	return fallback
}

func wantsHelp(opts core.Options) bool {
	return opts.Bool("help") || opts.Bool("h")
}

func failed(err error) core.Result {
	return core.Fail(err)
}

func readFile(app *core.Core, path string) ([]byte, error) {
	if app == nil {
		return nil, core.E("cmd.verify.readFile", "core app is required", nil)
	}
	r := app.Fs().Read(path)
	if !r.OK {
		return nil, resultError("cmd.verify.readFile", "read file", r)
	}
	raw, ok := r.Value.(string)
	if !ok {
		return nil, core.E("cmd.verify.readFile", "read returned invalid payload", nil)
	}
	return []byte(raw), nil
}

func resultError(op, msg string, r core.Result) error {
	if err, ok := r.Value.(error); ok {
		return core.E(op, msg, err)
	}
	return core.E(op, msg, nil)
}
