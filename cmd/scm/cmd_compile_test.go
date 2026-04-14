// SPDX-License-Identifier: EUPL-1.2

package scm

import (
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
	os "dappco.re/go/core/scm/internal/ax/osx"
	"encoding/hex"
	"testing"

	"dappco.re/go/core/cli/pkg/cli"
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

	err := runCompile(dir, "", "", "core scm compile", "core.json", nil, "")
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
	err := runCompile(dir, "", "", "custom builder", output, nil, "")
	require.NoError(t, err)

	raw, err := io.Local.Read(filepath.Join(dir, output))
	require.NoError(t, err)

	cm, err := manifest.ParseCompiled([]byte(raw))
	require.NoError(t, err)
	assert.Equal(t, "compile-custom", cm.Code)
	assert.Equal(t, "custom builder", cm.BuiltBy)
}

func TestRunCompile_Bad_InvalidSignKey_Good(t *testing.T) {
	dir := t.TempDir()
	coreDir := filepath.Join(dir, ".core")
	require.NoError(t, os.MkdirAll(coreDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(coreDir, "manifest.yaml"), []byte(`
code: compile-invalid-key
name: Compile Invalid Key
version: 1.0.0
`), 0644))

	err := runCompile(dir, "", hex.EncodeToString([]byte("short")), "core scm compile", "core.json", nil, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid private key length")
}

func TestRunCompile_Good_VersionOverride_Good(t *testing.T) {
	dir := t.TempDir()
	coreDir := filepath.Join(dir, ".core")
	require.NoError(t, os.MkdirAll(coreDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(coreDir, "manifest.yaml"), []byte(`
code: compile-version
name: Compile Version
version: 1.0.0
`), 0644))

	err := runCompile(dir, "3.2.1", "", "core scm compile", "core.json", nil, "")
	require.NoError(t, err)

	raw, err := io.Local.Read(filepath.Join(dir, "core.json"))
	require.NoError(t, err)

	cm, err := manifest.ParseCompiled([]byte(raw))
	require.NoError(t, err)
	assert.Equal(t, "3.2.1", cm.Version)
}

func TestRunCompile_Good_BuildMetadata_Good(t *testing.T) {
	dir := t.TempDir()
	coreDir := filepath.Join(dir, ".core")
	require.NoError(t, os.MkdirAll(coreDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(coreDir, "manifest.yaml"), []byte(`
code: compile-build
name: Compile Build
version: 1.0.0
`), 0644))

	err := runCompile(dir, "", "", "core scm compile", "core.json", []string{"linux/amd64", "darwin/arm64"}, "SHA-256")
	require.NoError(t, err)

	raw, err := io.Local.Read(filepath.Join(dir, "core.json"))
	require.NoError(t, err)

	cm, err := manifest.ParseCompiled([]byte(raw))
	require.NoError(t, err)
	require.NotNil(t, cm.Build)
	assert.Equal(t, []string{"linux/amd64", "darwin/arm64"}, cm.Build.Targets)
	assert.Equal(t, "SHA-256", cm.Build.Checksums)
}

func TestAddScmCommands_Good_CompileVersionFlagRegistered_Good(t *testing.T) {
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

	var compileCmd *cli.Command
	for _, cmd := range scmCmd.Commands() {
		if cmd.Name() == "compile" {
			compileCmd = cmd
			break
		}
	}
	require.NotNil(t, compileCmd)
	assert.NotNil(t, compileCmd.Flags().Lookup("version"))
	assert.NotNil(t, compileCmd.Flags().Lookup("target"))
	assert.NotNil(t, compileCmd.Flags().Lookup("checksums"))
}
