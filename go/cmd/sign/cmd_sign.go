// SPDX-License-Identifier: EUPL-1.2

// Package sign wires package signing into the scm CLI.
package sign

import (
	"crypto/ed25519"
	"encoding/base64"

	core "dappco.re/go"
	"dappco.re/go/scm/manifest"
)

const usage = "usage: scm sign [--root=DIR] [--in=core.json] [--out=core.json] [--manifest=FILE] [--key=BASE64_OR_FILE] [--key-file=FILE]"

// Register attaches the sign command to the parent Core command tree.
//
// The command signs the package manifest embedded in core.json. If --manifest
// is supplied, it compiles that manifest first and writes a signed core.json.
func Register(app *core.Core) core.Result {
	if app == nil {
		return core.Fail(core.E("cmd.sign.Register", "core app is required", nil))
	}
	return app.Command("sign", core.Command{Action: run(app)})
}

func run(app *core.Core) core.CommandAction {
	return func(opts core.Options) core.Result {
		if wantsHelp(opts) {
			core.Print(nil, usage)
			return core.Ok(nil)
		}

		priv, err := privateKey(app, opts)
		if err != nil {
			return failed(err)
		}

		root := option(opts, "root", ".")
		outPath := option(opts, "out", core.PathJoin(root, "core.json"))

		cm, err := compiledManifest(app, opts, root, priv)
		if err != nil {
			return failed(err)
		}

		raw, err := manifest.MarshalJSON(cm)
		if err != nil {
			return failed(err)
		}
		if r := app.Fs().WriteMode(outPath, string(raw), 0o600); !r.OK {
			return failed(resultError("cmd.sign.run", "write signed manifest", r))
		}

		core.Print(nil, "%s", outPath)
		return core.Ok(nil)
	}
}

func compiledManifest(app *core.Core, opts core.Options, root string, priv ed25519.PrivateKey) (*manifest.CompiledManifest, error) {
	pub := priv.Public().(ed25519.PublicKey)
	signKey := base64.StdEncoding.EncodeToString(pub)

	if path := core.Trim(opts.String("manifest")); path != "" {
		raw, err := readFile(app, path)
		if err != nil {
			return nil, err
		}
		m, err := manifest.Parse(raw)
		if err != nil {
			return nil, err
		}
		if core.Trim(m.SignKey) == "" {
			m.SignKey = signKey
		}
		m.Sign = ""
		return manifest.CompileWithOptions(m, manifest.CompileOptions{SignKey: priv})
	}

	inPath := option(opts, "in", core.PathJoin(root, "core.json"))
	raw, err := readFile(app, inPath)
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
	if core.Trim(cm.SignKey) == "" {
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

func privateKey(app *core.Core, opts core.Options) (ed25519.PrivateKey, error) {
	value := core.Trim(opts.String("key"))
	if path := core.Trim(opts.String("key-file")); path != "" {
		raw, err := readFile(app, path)
		if err != nil {
			return nil, err
		}
		value = core.Trim(string(raw))
	}
	if value == "" {
		value = core.Trim(app.Env("SCM_SIGN_KEY"))
	}
	if value == "" {
		return nil, core.E("cmd.sign.privateKey", "signing key is required", nil)
	}
	if raw, err := readFile(app, value); err == nil {
		value = core.Trim(string(raw))
	}
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, err
	}
	if len(decoded) != ed25519.PrivateKeySize {
		return nil, core.E("cmd.sign.privateKey", "signing key must be a base64 ed25519 private key", nil)
	}
	return ed25519.PrivateKey(decoded), nil
}

func canonicalManifestBytes(m *manifest.Manifest) ([]byte, error) {
	if m == nil {
		return nil, core.E("cmd.sign.canonicalManifestBytes", "manifest is required", nil)
	}
	cp := *m
	cp.Sign = ""
	cp.SignKey = ""
	r := core.JSONMarshal(cp)
	if !r.OK {
		return nil, resultError("cmd.sign.canonicalManifestBytes", "marshal manifest", r)
	}
	raw, ok := r.Value.([]byte)
	if !ok {
		return nil, core.E("cmd.sign.canonicalManifestBytes", "marshal returned invalid payload", nil)
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
		return nil, core.E("cmd.sign.readFile", "core app is required", nil)
	}
	r := app.Fs().Read(path)
	if !r.OK {
		return nil, resultError("cmd.sign.readFile", "read file", r)
	}
	raw, ok := r.Value.(string)
	if !ok {
		return nil, core.E("cmd.sign.readFile", "read returned invalid payload", nil)
	}
	return []byte(raw), nil
}

func resultError(op, msg string, r core.Result) error {
	if err, ok := r.Value.(error); ok {
		return core.E(op, msg, err)
	}
	return core.E(op, msg, nil)
}
