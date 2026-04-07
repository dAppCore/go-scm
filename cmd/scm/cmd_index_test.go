// SPDX-License-Identifier: EUPL-1.2

package scm

import (
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
	json "dappco.re/go/core/scm/internal/ax/jsonx"
	os "dappco.re/go/core/scm/internal/ax/osx"
	"testing"

	"dappco.re/go/core/io"
	"dappco.re/go/core/scm/manifest"
	"dappco.re/go/core/scm/marketplace"
	"dappco.re/go/core/cli/pkg/cli"
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

func TestRunIndex_Good_PrefersCompiledManifest_Good(t *testing.T) {
	root := t.TempDir()

	modDir := filepath.Join(root, "mod-a")
	require.NoError(t, os.MkdirAll(filepath.Join(modDir, ".core"), 0755))

	cm := &manifest.CompiledManifest{
		Manifest: manifest.Manifest{
			Code:    "compiled-mod",
			Name:    "Compiled Module",
			Version: "2.0.0",
			Sign:    "compiled-key",
		},
		Commit: "abc123",
	}
	data, err := json.Marshal(cm)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(modDir, "core.json"), data, 0644))
	require.NoError(t, os.WriteFile(filepath.Join(modDir, ".core", "manifest.yaml"), []byte(`
code: source-mod
name: Source Module
version: 1.0.0
sign: source-key
`), 0644))

	output := filepath.Join(root, "index.json")
	err = runIndex([]string{root}, output, "https://forge.example.com", "core")
	require.NoError(t, err)

	idx, err := marketplace.LoadIndex(io.Local, output)
	require.NoError(t, err)
	require.Len(t, idx.Modules, 1)
	assert.Equal(t, "compiled-mod", idx.Modules[0].Code)
	assert.Equal(t, "compiled-key", idx.Modules[0].SignKey)
}

func TestAddScmCommands_Good_IndexForgeURLFlagAlias_Good(t *testing.T) {
	root := &cli.Command{Use: "root"}

	AddScmCommands(root)

	var scmCmd *cli.Command
	for _, cmd := range root.Commands() {
		if cmd.Name() == "scm" {
			scmCmd = cmd
			break
		}
	}
	require.NotNil(t, scmCmd)

	var indexCmd *cli.Command
	for _, cmd := range scmCmd.Commands() {
		if cmd.Name() == "index" {
			indexCmd = cmd
			break
		}
	}
	require.NotNil(t, indexCmd)
	assert.NotNil(t, indexCmd.Flags().Lookup("forge-url"))
	assert.NotNil(t, indexCmd.Flags().Lookup("base-url"))
}
