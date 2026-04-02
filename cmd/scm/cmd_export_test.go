// SPDX-License-Identifier: EUPL-1.2

package scm

import (
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
	os "dappco.re/go/core/scm/internal/ax/osx"
	"testing"

	"dappco.re/go/core/scm/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunExport_Good_CompiledManifest_Good(t *testing.T) {
	dir := t.TempDir()

	cm := &manifest.CompiledManifest{
		Manifest: manifest.Manifest{
			Code:    "compiled-mod",
			Name:    "Compiled Module",
			Version: "1.0.0",
		},
		Commit: "abc123",
	}
	data, err := manifest.MarshalJSON(cm)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "core.json"), data, 0644))

	err = runExport(dir)
	require.NoError(t, err)
}

func TestRunExport_Good_FallsBackToSource_Good(t *testing.T) {
	dir := t.TempDir()

	coreDir := filepath.Join(dir, ".core")
	require.NoError(t, os.MkdirAll(coreDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(coreDir, "manifest.yaml"), []byte(`
code: source-mod
name: Source Module
version: 1.0.0
`), 0644))

	err := runExport(dir)
	require.NoError(t, err)
}

func TestRunExport_Bad_InvalidCompiledManifest_Good(t *testing.T) {
	dir := t.TempDir()

	coreDir := filepath.Join(dir, ".core")
	require.NoError(t, os.MkdirAll(coreDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(coreDir, "manifest.yaml"), []byte(`
code: source-mod
name: Source Module
version: 1.0.0
`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "core.json"), []byte("{not-json"), 0644))

	err := runExport(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "manifest.ParseCompiled")
}
