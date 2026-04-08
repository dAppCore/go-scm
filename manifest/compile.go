// SPDX-License-Identifier: EUPL-1.2

package manifest

import (
	"crypto/ed25519"
	"time"

	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
	json "dappco.re/go/core/scm/internal/ax/jsonx"

	"dappco.re/go/core/io"
	coreerr "dappco.re/go/core/log"
)

// CompiledManifest is the distribution-ready form of a manifest, written as
// core.json at the repository root (not inside .core/). It embeds the
// original Manifest and adds build metadata stapled at compile time.
type CompiledManifest struct {
	Manifest `json:",inline" yaml:",inline"`

	// Build metadata — populated by Compile.
	Commit  string `json:"commit,omitempty" yaml:"commit,omitempty"`
	Tag     string `json:"tag,omitempty" yaml:"tag,omitempty"`
	BuiltAt string `json:"built_at,omitempty" yaml:"built_at,omitempty"`
	BuiltBy string `json:"built_by,omitempty" yaml:"built_by,omitempty"`
}

// CompileOptions controls how Compile populates the build metadata.
type CompileOptions struct {
	Version string             // Optional override for the manifest version
	Commit  string             // Git commit hash
	Tag     string             // Git tag (e.g. v1.0.0)
	BuiltBy string             // Builder identity (e.g. "core build")
	SignKey ed25519.PrivateKey // Optional — signs before compiling
}

// Compile produces a CompiledManifest from a source manifest and build
// options. If opts.SignKey is provided the manifest is signed first.
//
// Example:
//   m := &Manifest{Code: "core/api", Version: "1.2.3"}
//   cm, err := manifest.Compile(m, manifest.CompileOptions{Commit: "abc123"})
func Compile(m *Manifest, opts CompileOptions) (*CompiledManifest, error) {
	if m == nil {
		return nil, coreerr.E("manifest.Compile", "nil manifest", nil)
	}
	if m.Code == "" {
		return nil, coreerr.E("manifest.Compile", "missing code", nil)
	}
	// Work on a copy to avoid mutating the caller's manifest.
	mCopy := *m
	if opts.Version != "" {
		mCopy.Version = opts.Version
	}
	m = &mCopy

	if m.Version == "" {
		return nil, coreerr.E("manifest.Compile", "missing version", nil)
	}

	// Sign if a key is supplied.
	if opts.SignKey != nil {
		if err := Sign(m, opts.SignKey); err != nil {
			return nil, coreerr.E("manifest.Compile", "sign failed", err)
		}
	}

	return &CompiledManifest{
		Manifest: *m,
		Commit:   opts.Commit,
		Tag:      opts.Tag,
		BuiltAt:  time.Now().UTC().Format(time.RFC3339),
		BuiltBy:  opts.BuiltBy,
	}, nil
}

// MarshalJSON serialises a CompiledManifest to JSON bytes.
//
// Example:
//   data, err := manifest.MarshalJSON(cm)
func MarshalJSON(cm *CompiledManifest) ([]byte, error) {
	return json.MarshalIndent(cm, "", "  ")
}

// ParseCompiled decodes a core.json into a CompiledManifest.
//
// Example:
//   cm, err := manifest.ParseCompiled([]byte(`{"code":"core/api"}`))
func ParseCompiled(data []byte) (*CompiledManifest, error) {
	var cm CompiledManifest
	if err := json.Unmarshal(data, &cm); err != nil {
		return nil, coreerr.E("manifest.ParseCompiled", "unmarshal failed", err)
	}
	return &cm, nil
}

const compiledPath = "core.json"

// WriteCompiled writes a CompiledManifest as core.json to the given root
// directory. The file lives at the distribution root, not inside .core/.
//
// Example:
//   err := manifest.WriteCompiled(io.Local, "/tmp/repo", cm)
func WriteCompiled(medium io.Medium, root string, cm *CompiledManifest) error {
	data, err := MarshalJSON(cm)
	if err != nil {
		return coreerr.E("manifest.WriteCompiled", "marshal failed", err)
	}
	path := filepath.Join(root, compiledPath)
	return medium.Write(path, string(data))
}

// LoadCompiled reads and parses a core.json from the given root directory.
//
// Example:
//   cm, err := manifest.LoadCompiled(io.Local, "/tmp/repo")
func LoadCompiled(medium io.Medium, root string) (*CompiledManifest, error) {
	path := filepath.Join(root, compiledPath)
	data, err := medium.Read(path)
	if err != nil {
		return nil, coreerr.E("manifest.LoadCompiled", "read failed", err)
	}
	return ParseCompiled([]byte(data))
}
