// SPDX-License-Identifier: EUPL-1.2

package manifest

import (
	"crypto/ed25519"
	"crypto/rand"
	json "dappco.re/go/core/scm/internal/ax/jsonx"
	"testing"

	"dappco.re/go/core/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompile_Good(t *testing.T) {
	m := &Manifest{
		Code:    "my-widget",
		Name:    "My Widget",
		Version: "1.2.3",
		Author:  "tester",
	}

	cm, err := Compile(m, CompileOptions{
		Version: "2.0.0",
		Commit:  "abc1234",
		Tag:     "v1.2.3",
		BuiltBy: "core build",
	})
	require.NoError(t, err)

	assert.Equal(t, "my-widget", cm.Code)
	assert.Equal(t, "My Widget", cm.Name)
	assert.Equal(t, "2.0.0", cm.Version)
	assert.Equal(t, "abc1234", cm.Commit)
	assert.Equal(t, "v1.2.3", cm.Tag)
	assert.Equal(t, "core build", cm.BuiltBy)
	assert.NotEmpty(t, cm.BuiltAt)
}

func TestCompile_Good_WithSigning_Good(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	m := &Manifest{
		Code:    "signed-mod",
		Name:    "Signed Module",
		Version: "0.1.0",
	}

	cm, err := Compile(m, CompileOptions{
		Commit:  "def5678",
		SignKey: priv,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, cm.Sign)

	// Verify signature is valid.
	ok, vErr := Verify(&cm.Manifest, pub)
	require.NoError(t, vErr)
	assert.True(t, ok)
}

func TestCompile_Bad_NilManifest_Good(t *testing.T) {
	_, err := Compile(nil, CompileOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil manifest")
}

func TestCompile_Bad_MissingCode_Good(t *testing.T) {
	m := &Manifest{Version: "1.0.0"}
	_, err := Compile(m, CompileOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing code")
}

func TestCompile_Bad_MissingVersion_Good(t *testing.T) {
	m := &Manifest{Code: "test"}
	_, err := Compile(m, CompileOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing version")
}

func TestMarshalJSON_Good(t *testing.T) {
	cm := &CompiledManifest{
		Manifest: Manifest{
			Code:    "test-mod",
			Name:    "Test Module",
			Version: "1.0.0",
		},
		Commit:  "abc123",
		Tag:     "v1.0.0",
		BuiltAt: "2026-03-15T10:00:00Z",
		BuiltBy: "test",
	}

	data, err := MarshalJSON(cm)
	require.NoError(t, err)

	// Round-trip: parse back.
	parsed, err := ParseCompiled(data)
	require.NoError(t, err)
	assert.Equal(t, "test-mod", parsed.Code)
	assert.Equal(t, "abc123", parsed.Commit)
	assert.Equal(t, "v1.0.0", parsed.Tag)
	assert.Equal(t, "2026-03-15T10:00:00Z", parsed.BuiltAt)
}

func TestParseCompiled_Good(t *testing.T) {
	raw := `{
		"code": "demo",
		"name": "Demo",
		"version": "0.5.0",
		"commit": "aaa111",
		"tag": "v0.5.0",
		"built_at": "2026-03-15T12:00:00Z",
		"built_by": "ci"
	}`
	cm, err := ParseCompiled([]byte(raw))
	require.NoError(t, err)
	assert.Equal(t, "demo", cm.Code)
	assert.Equal(t, "aaa111", cm.Commit)
	assert.Equal(t, "ci", cm.BuiltBy)
}

func TestParseCompiled_Bad(t *testing.T) {
	_, err := ParseCompiled([]byte("not json"))
	assert.Error(t, err)
}

func TestWriteCompiled_Good(t *testing.T) {
	medium := io.NewMockMedium()
	cm := &CompiledManifest{
		Manifest: Manifest{
			Code:    "write-test",
			Name:    "Write Test",
			Version: "1.0.0",
		},
		Commit: "ccc333",
	}

	err := WriteCompiled(medium, "/project", cm)
	require.NoError(t, err)

	// Verify the file was written.
	content, err := medium.Read("/project/core.json")
	require.NoError(t, err)

	var parsed CompiledManifest
	require.NoError(t, json.Unmarshal([]byte(content), &parsed))
	assert.Equal(t, "write-test", parsed.Code)
	assert.Equal(t, "ccc333", parsed.Commit)
}

func TestLoadCompiled_Good(t *testing.T) {
	medium := io.NewMockMedium()
	raw := `{"code":"load-test","name":"Load Test","version":"2.0.0","commit":"ddd444"}`
	medium.Files["/project/core.json"] = raw

	cm, err := LoadCompiled(medium, "/project")
	require.NoError(t, err)
	assert.Equal(t, "load-test", cm.Code)
	assert.Equal(t, "ddd444", cm.Commit)
}

func TestLoadCompiled_Bad_NotFound_Good(t *testing.T) {
	medium := io.NewMockMedium()
	_, err := LoadCompiled(medium, "/missing")
	assert.Error(t, err)
}

func TestCompile_Good_MinimalOptions_Good(t *testing.T) {
	m := &Manifest{
		Code:    "minimal",
		Name:    "Minimal",
		Version: "0.0.1",
	}
	cm, err := Compile(m, CompileOptions{})
	require.NoError(t, err)
	assert.Empty(t, cm.Commit)
	assert.Empty(t, cm.Tag)
	assert.Empty(t, cm.BuiltBy)
	assert.NotEmpty(t, cm.BuiltAt)
}

func TestCompile_Good_WithVersionOverride_Good(t *testing.T) {
	m := &Manifest{
		Code:    "override",
		Name:    "Override",
		Version: "0.0.1",
	}

	cm, err := Compile(m, CompileOptions{
		Version: "9.9.9",
	})
	require.NoError(t, err)
	// Compiled manifest carries the override.
	assert.Equal(t, "9.9.9", cm.Version)
	// Caller's manifest is not mutated by Compile.
	assert.Equal(t, "0.0.1", m.Version)
}
