// SPDX-License-Identifier: EUPL-1.2

package scm

import (
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
	os "dappco.re/go/core/scm/internal/ax/osx"
	"testing"

	"dappco.re/go/core/io"
	"dappco.re/go/core/scm/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCompile_Good_DefaultOutput_Good(t *testing.T) {
	dir := t.TempDir()
	coreDir := filepath.Join(dir, ".core")
	require.NoError(t, os.MkdirAll(coreDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(coreDir, "manifest.yaml"), []byte(`
code: compile-default
name: Compile Default
version: 1.0.0
`), 0644))

	err := runCompile(dir, "", "core scm compile", "core.json")
	require.NoError(t, err)

	raw, err := io.Local.Read(filepath.Join(dir, "core.json"))
	require.NoError(t, err)

	cm, err := manifest.ParseCompiled([]byte(raw))
	require.NoError(t, err)
	assert.Equal(t, "compile-default", cm.Code)
	assert.Equal(t, "core scm compile", cm.BuiltBy)
}

func TestRunCompile_Good_CustomOutput_Good(t *testing.T) {
	dir := t.TempDir()
	coreDir := filepath.Join(dir, ".core")
	require.NoError(t, os.MkdirAll(coreDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(coreDir, "manifest.yaml"), []byte(`
code: compile-custom
name: Compile Custom
version: 2.0.0
`), 0644))

	output := filepath.Join("dist", "core.json")
	err := runCompile(dir, "", "custom builder", output)
	require.NoError(t, err)

	raw, err := io.Local.Read(filepath.Join(dir, output))
	require.NoError(t, err)

	cm, err := manifest.ParseCompiled([]byte(raw))
	require.NoError(t, err)
	assert.Equal(t, "compile-custom", cm.Code)
	assert.Equal(t, "custom builder", cm.BuiltBy)
}
