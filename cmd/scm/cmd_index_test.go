// SPDX-License-Identifier: EUPL-1.2

package scm

import (
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
	os "dappco.re/go/core/scm/internal/ax/osx"
	"testing"

	"dappco.re/go/core/io"
	"dappco.re/go/core/scm/marketplace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunIndex_Good_WritesIndex_Good(t *testing.T) {
	root := t.TempDir()

	modDir := filepath.Join(root, "mod-a")
	require.NoError(t, os.MkdirAll(filepath.Join(modDir, ".core"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(modDir, ".core", "manifest.yaml"), []byte(`
code: mod-a
name: Module A
version: 1.0.0
sign: key-a
`), 0644))

	output := filepath.Join(root, "index.json")
	err := runIndex([]string{root}, output, "https://forge.example.com", "core")
	require.NoError(t, err)

	idx, err := marketplace.LoadIndex(io.Local, output)
	require.NoError(t, err)
	require.Len(t, idx.Modules, 1)
	assert.Equal(t, "mod-a", idx.Modules[0].Code)
	assert.Equal(t, "https://forge.example.com/core/mod-a.git", idx.Modules[0].Repo)
}
