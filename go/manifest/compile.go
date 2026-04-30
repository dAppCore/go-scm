// SPDX-License-Identifier: EUPL-1.2

package manifest

import (
	"crypto/ed25519" // intrinsic
	// Note: AX-6 — Build metadata records wall-clock compilation time.
	"time"

	core "dappco.re/go"
	coreio "dappco.re/go/io"
)

type CompileOptions struct {
	Commit  string
	Tag     string
	BuiltBy string
	SignKey ed25519.PrivateKey
	Build   BuildInfo
}

type CompiledManifest struct {
	Manifest `json:",inline" yaml:",inline"`
	Commit   string    `json:"commit,omitempty" yaml:"commit,omitempty"`
	Tag      string    `json:"tag,omitempty" yaml:"tag,omitempty"`
	BuiltAt  string    `json:"built_at,omitempty" yaml:"built_at,omitempty"`
	BuiltBy  string    `json:"built_by,omitempty" yaml:"built_by,omitempty"`
	Build    BuildInfo `json:"build,omitempty" yaml:"build,omitempty"`
}

func Compile(m *Manifest, info BuildInfo) ([]byte, error)  /* v090-result-boundary */ {
	if err := validateManifest(m); err != nil {
		return nil, err
	}
	cp := *m
	cp.Build = normalizeBuildInfo(info)
	return marshalJSON("manifest.Compile", &cp)
}

func ParseCoreJSON(data []byte) (*Manifest, error)  /* v090-result-boundary */ {
	var m Manifest
	if err := unmarshalJSON("manifest.ParseCoreJSON", data, &m); err != nil {
		return nil, err
	}
	if err := validateManifest(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

func CompileWithOptions(m *Manifest, opts CompileOptions) (*CompiledManifest, error)  /* v090-result-boundary */ {
	if err := validateManifest(m); err != nil {
		return nil, err
	}
	cp := *m
	if len(opts.SignKey) > 0 && cp.Sign == "" {
		payload, err := canonicalManifestBytes(&cp)
		if err != nil {
			return nil, err
		}
		if err := Sign(&cp, payload, opts.SignKey); err != nil {
			return nil, err
		}
	}
	return &CompiledManifest{
		Manifest: cp,
		Commit:   opts.Commit,
		Tag:      opts.Tag,
		BuiltAt:  time.Now().UTC().Format(time.RFC3339Nano),
		BuiltBy:  opts.BuiltBy,
		Build:    normalizeBuildInfo(opts.Build),
	}, nil
}

func MarshalJSON(cm *CompiledManifest) ([]byte, error)  /* v090-result-boundary */ {
	if cm == nil {
		return nil, core.E("manifest.MarshalJSON", "compiled manifest is required", nil)
	}
	return marshalJSON("manifest.MarshalJSON", cm)
}

func ParseCompiled(data []byte) (*CompiledManifest, error)  /* v090-result-boundary */ {
	var cm CompiledManifest
	if err := unmarshalJSON("manifest.ParseCompiled", data, &cm); err != nil {
		return nil, err
	}
	return &cm, nil
}

func LoadCompiled(medium coreio.Medium, root string) (*CompiledManifest, error)  /* v090-result-boundary */ {
	if medium == nil {
		return nil, core.E("manifest.LoadCompiled", "medium is required", nil)
	}
	raw, err := medium.Read(core.PathJoin(root, "core.json"))
	if err != nil {
		return nil, err
	}
	return ParseCompiled([]byte(raw))
}

func WriteCompiled(medium coreio.Medium, root string, cm *CompiledManifest) error  /* v090-result-boundary */ {
	if medium == nil {
		return core.E("manifest.WriteCompiled", "medium is required", nil)
	}
	if cm == nil {
		return core.E("manifest.WriteCompiled", "compiled manifest is required", nil)
	}
	raw, err := MarshalJSON(cm)
	if err != nil {
		return err
	}
	return medium.Write(core.PathJoin(root, "core.json"), string(raw))
}

func normalizeBuildInfo(build BuildInfo) BuildInfo {
	if len(build.Targets) > 0 {
		build.Targets = append([]string(nil), build.Targets...)
	}
	return build
}

func marshalJSON(op string, v any) ([]byte, error)  /* v090-result-boundary */ {
	r := core.JSONMarshal(v)
	if !r.OK {
		return nil, resultError(op, "marshal JSON", r)
	}
	raw, ok := r.Value.([]byte)
	if !ok {
		return nil, core.E(op, "marshal JSON returned invalid payload", nil)
	}
	return raw, nil
}

func unmarshalJSON(op string, data []byte, target any) error  /* v090-result-boundary */ {
	r := core.JSONUnmarshal(data, target)
	if !r.OK {
		return resultError(op, "unmarshal JSON", r)
	}
	return nil
}

func resultError(op, msg string, r core.Result) error  /* v090-result-boundary */ {
	if err, ok := r.Value.(error); ok {
		return core.E(op, msg, err)
	}
	return core.E(op, msg, nil)
}
