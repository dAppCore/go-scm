// SPDX-License-Identifier: EUPL-1.2

package scm

import (
	"context"
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
	os "dappco.re/go/core/scm/internal/ax/osx"
	exec "golang.org/x/sys/execabs"
	"testing"

	"dappco.re/go/core/cli/pkg/cli"
	"dappco.re/go/core/io"
	"dappco.re/go/core/scm/marketplace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddScmCommands_Good_PackageAndDevCommandsRegistered_Good(t *testing.T) {
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

	var foundPkg, foundDev bool
	for _, cmd := range scmCmd.Commands() {
		switch cmd.Name() {
		case "pkg":
			foundPkg = true
		case "dev":
			foundDev = true
		}
	}
	assert.True(t, foundPkg)
	assert.True(t, foundDev)
}

func TestPackageInstall_Good_ExplicitVersion_Good(t *testing.T) {
	repo := createTaggedModuleRepo(t, "pkg-a",
		moduleVersion{Version: "1.0.0", Tag: "v1.0.0"},
		moduleVersion{Version: "2.0.0", Tag: "v2.0.0"},
	)
	indexPath := filepath.Join(t.TempDir(), "index.json")
	require.NoError(t, marketplace.WriteIndex(io.Local, indexPath, &marketplace.Index{
		Version: 1,
		Modules: []marketplace.Module{
			{Code: "pkg-a", Name: "Package A", Repo: repo, Version: "2.0.0"},
		},
	}))

	installed, err := packageInstall(context.Background(), indexPath, filepath.Join(t.TempDir(), "modules"), ":memory:", "pkg-a@1.0.0")
	require.NoError(t, err)
	assert.Equal(t, "pkg-a", installed.Code)
	assert.Equal(t, "1.0.0", installed.Version)
}

func TestWorkspaceImpact_Good(t *testing.T) {
	regPath := filepath.Join(t.TempDir(), "repos.yaml")
	require.NoError(t, os.WriteFile(regPath, []byte(`
version: 1
base_path: /tmp/repos
repos:
  core:
    type: foundation
  api:
    type: module
    depends_on: [core]
  ui:
    type: module
    depends_on: [api]
`), 0644))

	impacted, err := workspaceImpact([]string{regPath}, "core")
	require.NoError(t, err)
	require.Len(t, impacted, 2)
	assert.Equal(t, "api", impacted[0].Name)
	assert.Equal(t, "ui", impacted[1].Name)
}

func TestLoadWorkspaceRegistries_Good_ExplicitPaths_Good(t *testing.T) {
	regPath := filepath.Join(t.TempDir(), "repos.yaml")
	require.NoError(t, os.WriteFile(regPath, []byte(`
version: 1
base_path: /tmp/repos
repos:
  core:
    type: foundation
`), 0644))

	regs, err := loadWorkspaceRegistries([]string{regPath})
	require.NoError(t, err)
	require.Len(t, regs, 1)

	repo, ok := regs[0].Get("core")
	require.True(t, ok)
	assert.Equal(t, "/tmp/repos/core", repo.Path)
}

type moduleVersion struct {
	Version string
	Tag     string
}

func createTaggedModuleRepo(t *testing.T, code string, versions ...moduleVersion) string {
	t.Helper()

	dir := filepath.Join(t.TempDir(), code)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".core"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "main.ts"), []byte("export {}\n"), 0644))
	runGitInRepo(t, dir, "init")

	for _, version := range versions {
		manifestYAML := "code: " + code + "\nname: " + code + "\nversion: \"" + version.Version + "\"\n"
		require.NoError(t, os.WriteFile(filepath.Join(dir, ".core", "manifest.yaml"), []byte(manifestYAML), 0644))
		runGitInRepo(t, dir, "add", "--force", ".")
		runGitInRepo(t, dir, "commit", "-m", "version-"+version.Version)
		if version.Tag != "" {
			runGitInRepo(t, dir, "tag", version.Tag)
		}
	}

	return dir
}

func runGitInRepo(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.email=test@test.com", "-c", "user.name=test"}, args...)...)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v: %s", args, string(out))
}
