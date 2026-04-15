// SPDX-License-Identifier: EUPL-1.2

package manifest

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"time"

	coreio "dappco.re/go/core/io"
	"dappco.re/go/scm/internal/ax/filepathx"
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

func Compile(m *Manifest, opts CompileOptions) (*CompiledManifest, error) {
	if err := validateManifest(m); err != nil {
		return nil, err
	}
	cp := *m
	if len(opts.SignKey) > 0 && cp.Sign == "" {
		if err := Sign(&cp, opts.SignKey); err != nil {
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

func MarshalJSON(cm *CompiledManifest) ([]byte, error) {
	if cm == nil {
		return nil, errors.New("manifest.MarshalJSON: compiled manifest is required")
	}
	return json.Marshal(cm)
}

func ParseCompiled(data []byte) (*CompiledManifest, error) {
	var cm CompiledManifest
	if err := json.Unmarshal(data, &cm); err != nil {
		return nil, err
	}
	return &cm, nil
}

func LoadCompiled(medium coreio.Medium, root string) (*CompiledManifest, error) {
	if medium == nil {
		return nil, errors.New("manifest.LoadCompiled: medium is required")
	}
	raw, err := medium.Read(filepathx.Join(root, "core.json"))
	if err != nil {
		return nil, err
	}
	return ParseCompiled([]byte(raw))
}

func WriteCompiled(medium coreio.Medium, root string, cm *CompiledManifest) error {
	if medium == nil {
		return errors.New("manifest.WriteCompiled: medium is required")
	}
	if cm == nil {
		return errors.New("manifest.WriteCompiled: compiled manifest is required")
	}
	raw, err := MarshalJSON(cm)
	if err != nil {
		return err
	}
	return medium.Write(filepathx.Join(root, "core.json"), string(raw))
}

func normalizeBuildInfo(build BuildInfo) BuildInfo {
	if len(build.Targets) > 0 {
		build.Targets = append([]string(nil), build.Targets...)
	}
	return build
}
