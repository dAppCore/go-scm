// SPDX-License-Identifier: EUPL-1.2

package marketplace

import (
	"context"
	"crypto/ed25519"
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
	os "dappco.re/go/core/scm/internal/ax/osx"
	"encoding/hex"
	exec "golang.org/x/sys/execabs"
	"testing"

	"dappco.re/go/core/io"
	"dappco.re/go/core/io/store"
	"dappco.re/go/core/scm/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestRepo creates a bare-bones git repo with a manifest and main.ts.
// Returns the repo path (usable as Module.Repo for local clone).
func createTestRepo(t *testing.T, code, version string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), code)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".core"), 0755))

	manifestYAML := "code: " + code + "\nname: Test " + code + "\nversion: \"" + version + "\"\n"
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, ".core", "manifest.yaml"),
		[]byte(manifestYAML), 0644,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "main.ts"),
		[]byte("export async function init(core: any) {}\n"), 0644,
	))

	runGit(t, dir, "init")
	runGit(t, dir, "add", "--force", ".")
	runGit(t, dir, "commit", "-m", "init")
	return dir
}

// createSignedTestRepo creates a git repo with a signed manifest.
// Returns (repo path, hex-encoded public key).
func createSignedTestRepo(t *testing.T, code, version string) (string, string) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	dir := filepath.Join(t.TempDir(), code)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".core"), 0755))

	m := &manifest.Manifest{
		Code:    code,
		Name:    "Test " + code,
		Version: version,
	}
	require.NoError(t, manifest.Sign(m, priv))

	data, err := manifest.MarshalYAML(m)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".core", "manifest.yaml"), data, 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "main.ts"), []byte("export async function init(core: any) {}\n"), 0644))

	runGit(t, dir, "init")
	runGit(t, dir, "add", "--force", ".")
	runGit(t, dir, "commit", "-m", "init")

	return dir, hex.EncodeToString(pub)
}

type taggedRepoVersion struct {
	Version string
	Name    string
	Tag     string
}

func createTaggedTestRepo(t *testing.T, code string, versions ...taggedRepoVersion) string {
	t.Helper()

	dir := filepath.Join(t.TempDir(), code)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".core"), 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "main.ts"),
		[]byte("export async function init(core: any) {}\n"), 0644,
	))

	runGit(t, dir, "init")

	for idx, version := range versions {
		name := version.Name
		if name == "" {
			name = "Test " + code
		}

		manifestYAML := "code: " + code + "\nname: " + name + "\nversion: \"" + version.Version + "\"\n"
		require.NoError(t, os.WriteFile(
			filepath.Join(dir, ".core", "manifest.yaml"),
			[]byte(manifestYAML), 0644,
		))
		require.NoError(t, os.WriteFile(
			filepath.Join(dir, "version.txt"),
			[]byte(version.Version+"\n"), 0644,
		))

		runGit(t, dir, "add", "--force", ".")
		runGit(t, dir, "commit", "-m", "version-"+version.Version)
		if version.Tag != "" {
			runGit(t, dir, "tag", version.Tag)
		}

		if idx == len(versions)-1 {
			break
		}
	}

	return dir
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.email=test@test.com", "-c", "user.name=test"}, args...)...)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v: %s", args, string(out))
}

func TestInstall_Good(t *testing.T) {
	repo := createTestRepo(t, "hello-mod", "1.0")
	modulesDir := filepath.Join(t.TempDir(), "modules")

	st, err := store.New(store.Options{Path: ":memory:"})
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(io.Local, modulesDir, st)
	err = inst.Install(context.Background(), Module{
		Code: "hello-mod",
		Repo: repo,
	})
	require.NoError(t, err)

	// Verify directory exists
	_, err = os.Stat(filepath.Join(modulesDir, "hello-mod", "main.ts"))
	assert.NoError(t, err, "main.ts should exist in installed module")

	// Verify store entry
	raw, err := st.Get("_modules", "hello-mod")
	require.NoError(t, err)
	assert.Contains(t, raw, `"code":"hello-mod"`)
	assert.Contains(t, raw, `"version":"1.0"`)
}

func TestInstall_Good_UsesLatestTaggedVersion_Good(t *testing.T) {
	repo := createTaggedTestRepo(t, "tagged-mod",
		taggedRepoVersion{Version: "1.0.0", Tag: "v1.0.0"},
		taggedRepoVersion{Version: "2.0.0", Tag: "v2.0.0"},
		taggedRepoVersion{Version: "9.9.9"},
	)
	modulesDir := filepath.Join(t.TempDir(), "modules")

	st, err := store.New(store.Options{Path: ":memory:"})
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(io.Local, modulesDir, st)
	require.NoError(t, inst.Install(context.Background(), Module{
		Code: "tagged-mod",
		Repo: repo,
	}))

	installed, err := inst.Installed()
	require.NoError(t, err)
	require.Len(t, installed, 1)
	assert.Equal(t, "2.0.0", installed[0].Version)
	assert.Equal(t, "Test tagged-mod", installed[0].Name)
}

func TestInstall_Good_ExplicitTaggedVersion_Good(t *testing.T) {
	repo := createTaggedTestRepo(t, "pinned-mod",
		taggedRepoVersion{Version: "1.0.0", Tag: "v1.0.0"},
		taggedRepoVersion{Version: "2.0.0", Tag: "v2.0.0"},
	)
	modulesDir := filepath.Join(t.TempDir(), "modules")

	st, err := store.New(store.Options{Path: ":memory:"})
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(io.Local, modulesDir, st)
	require.NoError(t, inst.Install(context.Background(), Module{
		Code:    "pinned-mod",
		Repo:    repo,
		Version: "1.0.0",
	}))

	installed, err := inst.Installed()
	require.NoError(t, err)
	require.Len(t, installed, 1)
	assert.Equal(t, "1.0.0", installed[0].Version)
}

func TestInstall_Bad_ExplicitVersionMissingTag_Good(t *testing.T) {
	repo := createTaggedTestRepo(t, "missing-tag-mod",
		taggedRepoVersion{Version: "1.0.0", Tag: "v1.0.0"},
	)
	modulesDir := filepath.Join(t.TempDir(), "modules")

	st, err := store.New(store.Options{Path: ":memory:"})
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(io.Local, modulesDir, st)
	err = inst.Install(context.Background(), Module{
		Code:    "missing-tag-mod",
		Repo:    repo,
		Version: "2.0.0",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tag not found")

	_, statErr := os.Stat(filepath.Join(modulesDir, "missing-tag-mod"))
	assert.True(t, os.IsNotExist(statErr))
}

func TestInstall_Good_TagIsSourceOfTruth_Good(t *testing.T) {
	repo := createTaggedTestRepo(t, "tag-source-mod",
		taggedRepoVersion{Version: "9.9.9", Tag: "v2.3.4"},
	)
	modulesDir := filepath.Join(t.TempDir(), "modules")

	st, err := store.New(store.Options{Path: ":memory:"})
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(io.Local, modulesDir, st)
	require.NoError(t, inst.Install(context.Background(), Module{
		Code: "tag-source-mod",
		Repo: repo,
	}))

	installed, err := inst.Installed()
	require.NoError(t, err)
	require.Len(t, installed, 1)
	assert.Equal(t, "2.3.4", installed[0].Version)
}

func TestInstall_Good_Signed_Good(t *testing.T) {
	repo, signKey := createSignedTestRepo(t, "signed-mod", "2.0")
	modulesDir := filepath.Join(t.TempDir(), "modules")

	st, err := store.New(store.Options{Path: ":memory:"})
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(io.Local, modulesDir, st)
	err = inst.Install(context.Background(), Module{
		Code:    "signed-mod",
		Repo:    repo,
		SignKey: signKey,
	})
	require.NoError(t, err)

	raw, err := st.Get("_modules", "signed-mod")
	require.NoError(t, err)
	assert.Contains(t, raw, `"version":"2.0"`)
}

func TestInstall_Good_UsesEmbeddedManifestSignKey_Good(t *testing.T) {
	repo, _ := createSignedTestRepo(t, "embedded-signed-mod", "2.1")
	modulesDir := filepath.Join(t.TempDir(), "modules")

	st, err := store.New(store.Options{Path: ":memory:"})
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(io.Local, modulesDir, st)
	err = inst.Install(context.Background(), Module{
		Code: "embedded-signed-mod",
		Repo: repo,
	})
	require.NoError(t, err)

	raw, err := st.Get("_modules", "embedded-signed-mod")
	require.NoError(t, err)
	assert.Contains(t, raw, `"version":"2.1"`)
	assert.Contains(t, raw, `"sign_key":"`)
}

func TestInstall_Bad_SignedManifestMissingSignKey_Good(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	repo := filepath.Join(t.TempDir(), "missing-sign-key-mod")
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".core"), 0755))

	m := &manifest.Manifest{
		Code:    "missing-sign-key-mod",
		Name:    "Missing Sign Key Module",
		Version: "1.0.0",
	}
	require.NoError(t, manifest.Sign(m, priv))
	m.SignKey = ""

	data, err := manifest.MarshalYAML(m)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(repo, ".core", "manifest.yaml"), data, 0644))
	require.NoError(t, os.WriteFile(filepath.Join(repo, "main.ts"), []byte("export async function init(core: any) {}\n"), 0644))

	runGit(t, repo, "init")
	runGit(t, repo, "add", "--force", ".")
	runGit(t, repo, "commit", "-m", "init")

	modulesDir := filepath.Join(t.TempDir(), "modules")
	st, err := store.New(store.Options{Path: ":memory:"})
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(io.Local, modulesDir, st)
	err = inst.Install(context.Background(), Module{
		Code: "missing-sign-key-mod",
		Repo: repo,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing sign_key")

	_, statErr := os.Stat(filepath.Join(modulesDir, "missing-sign-key-mod"))
	assert.True(t, os.IsNotExist(statErr))
}

func TestInstall_Bad_AlreadyInstalled_Good(t *testing.T) {
	repo := createTestRepo(t, "dup-mod", "1.0")
	modulesDir := filepath.Join(t.TempDir(), "modules")

	st, err := store.New(store.Options{Path: ":memory:"})
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(io.Local, modulesDir, st)
	mod := Module{Code: "dup-mod", Repo: repo}

	require.NoError(t, inst.Install(context.Background(), mod))
	err = inst.Install(context.Background(), mod)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already installed")
}

func TestInstall_Bad_InvalidSignature_Good(t *testing.T) {
	// Sign with key A, verify with key B
	repo, _ := createSignedTestRepo(t, "bad-sig", "1.0")
	_, wrongKey := createSignedTestRepo(t, "dummy", "1.0") // different key

	modulesDir := filepath.Join(t.TempDir(), "modules")

	st, err := store.New(store.Options{Path: ":memory:"})
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(io.Local, modulesDir, st)
	err = inst.Install(context.Background(), Module{
		Code:    "bad-sig",
		Repo:    repo,
		SignKey: wrongKey,
	})
	assert.Error(t, err)

	// Verify directory was cleaned up
	_, statErr := os.Stat(filepath.Join(modulesDir, "bad-sig"))
	assert.True(t, os.IsNotExist(statErr), "directory should be cleaned up on failure")
}

func TestInstall_Bad_PathTraversalCode_Good(t *testing.T) {
	repo := createTestRepo(t, "safe-mod", "1.0")
	modulesDir := filepath.Join(t.TempDir(), "modules")

	st, err := store.New(store.Options{Path: ":memory:"})
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(io.Local, modulesDir, st)
	err = inst.Install(context.Background(), Module{
		Code: "../escape",
		Repo: repo,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid module code")

	_, err = st.Get("_modules", "escape")
	assert.Error(t, err)

	_, err = os.Stat(filepath.Join(filepath.Dir(modulesDir), "escape"))
	assert.True(t, os.IsNotExist(err))
}

func TestRemove_Good(t *testing.T) {
	repo := createTestRepo(t, "rm-mod", "1.0")
	modulesDir := filepath.Join(t.TempDir(), "modules")

	st, err := store.New(store.Options{Path: ":memory:"})
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(io.Local, modulesDir, st)
	require.NoError(t, inst.Install(context.Background(), Module{Code: "rm-mod", Repo: repo}))

	err = inst.Remove("rm-mod")
	require.NoError(t, err)

	// Directory gone
	_, statErr := os.Stat(filepath.Join(modulesDir, "rm-mod"))
	assert.True(t, os.IsNotExist(statErr))

	// Store entry gone
	_, err = st.Get("_modules", "rm-mod")
	assert.Error(t, err)
}

func TestRemove_Bad_NotInstalled_Good(t *testing.T) {
	st, err := store.New(store.Options{Path: ":memory:"})
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(io.Local, t.TempDir(), st)
	err = inst.Remove("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not installed")
}

func TestRemove_Bad_PathTraversalCode_Good(t *testing.T) {
	baseDir := t.TempDir()
	modulesDir := filepath.Join(baseDir, "modules")
	escapeDir := filepath.Join(baseDir, "escape")
	require.NoError(t, os.MkdirAll(escapeDir, 0755))

	st, err := store.New(store.Options{Path: ":memory:"})
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(io.Local, modulesDir, st)
	err = inst.Remove("../escape")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid module code")

	info, statErr := os.Stat(escapeDir)
	require.NoError(t, statErr)
	assert.True(t, info.IsDir())
}

func TestInstalled_Good(t *testing.T) {
	modulesDir := filepath.Join(t.TempDir(), "modules")

	st, err := store.New(store.Options{Path: ":memory:"})
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(io.Local, modulesDir, st)

	repo1 := createTestRepo(t, "mod-a", "1.0")
	repo2 := createTestRepo(t, "mod-b", "2.0")

	require.NoError(t, inst.Install(context.Background(), Module{Code: "mod-a", Repo: repo1}))
	require.NoError(t, inst.Install(context.Background(), Module{Code: "mod-b", Repo: repo2}))

	installed, err := inst.Installed()
	require.NoError(t, err)
	assert.Len(t, installed, 2)

	codes := map[string]bool{}
	for _, m := range installed {
		codes[m.Code] = true
	}
	assert.True(t, codes["mod-a"])
	assert.True(t, codes["mod-b"])
}

func TestInstalled_Good_Empty_Good(t *testing.T) {
	st, err := store.New(store.Options{Path: ":memory:"})
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(io.Local, t.TempDir(), st)
	installed, err := inst.Installed()
	require.NoError(t, err)
	assert.Empty(t, installed)
}

func TestUpdate_Good(t *testing.T) {
	repo := createTestRepo(t, "upd-mod", "1.0")
	modulesDir := filepath.Join(t.TempDir(), "modules")

	st, err := store.New(store.Options{Path: ":memory:"})
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(io.Local, modulesDir, st)
	require.NoError(t, inst.Install(context.Background(), Module{Code: "upd-mod", Repo: repo}))

	// Update the origin repo
	newManifest := "code: upd-mod\nname: Updated Module\nversion: \"2.0\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(repo, ".core", "manifest.yaml"), []byte(newManifest), 0644))
	runGit(t, repo, "add", ".")
	runGit(t, repo, "commit", "-m", "bump version")

	err = inst.Update(context.Background(), "upd-mod")
	require.NoError(t, err)

	// Verify updated metadata
	installed, err := inst.Installed()
	require.NoError(t, err)
	require.Len(t, installed, 1)
	assert.Equal(t, "2.0", installed[0].Version)
	assert.Equal(t, "Updated Module", installed[0].Name)
}

func TestUpdate_Good_AdvancesToLatestTag_Good(t *testing.T) {
	repo := createTaggedTestRepo(t, "upd-tagged-mod",
		taggedRepoVersion{Version: "1.0.0", Name: "Tagged Module", Tag: "v1.0.0"},
	)
	modulesDir := filepath.Join(t.TempDir(), "modules")

	st, err := store.New(store.Options{Path: ":memory:"})
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(io.Local, modulesDir, st)
	require.NoError(t, inst.Install(context.Background(), Module{
		Code: "upd-tagged-mod",
		Repo: repo,
	}))

	manifestYAML := "code: upd-tagged-mod\nname: Tagged Module Updated\nversion: \"2.0.0\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(repo, ".core", "manifest.yaml"), []byte(manifestYAML), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(repo, "version.txt"), []byte("2.0.0\n"), 0644))
	runGit(t, repo, "add", "--force", ".")
	runGit(t, repo, "commit", "-m", "version-2.0.0")
	runGit(t, repo, "tag", "v2.0.0")

	require.NoError(t, inst.Update(context.Background(), "upd-tagged-mod"))

	installed, err := inst.Installed()
	require.NoError(t, err)
	require.Len(t, installed, 1)
	assert.Equal(t, "2.0.0", installed[0].Version)
	assert.Equal(t, "Tagged Module Updated", installed[0].Name)
}

func TestUpdate_Good_TagRemainsSourceOfTruth_Good(t *testing.T) {
	repo := createTaggedTestRepo(t, "upd-tag-source-mod",
		taggedRepoVersion{Version: "1.0.0", Name: "Tagged Source Module", Tag: "v1.0.0"},
	)
	modulesDir := filepath.Join(t.TempDir(), "modules")

	st, err := store.New(store.Options{Path: ":memory:"})
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(io.Local, modulesDir, st)
	require.NoError(t, inst.Install(context.Background(), Module{
		Code: "upd-tag-source-mod",
		Repo: repo,
	}))

	manifestYAML := "code: upd-tag-source-mod\nname: Tagged Source Module Updated\nversion: \"9.9.9\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(repo, ".core", "manifest.yaml"), []byte(manifestYAML), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(repo, "version.txt"), []byte("9.9.9\n"), 0644))
	runGit(t, repo, "add", "--force", ".")
	runGit(t, repo, "commit", "-m", "version-9.9.9")
	runGit(t, repo, "tag", "v2.0.0")

	require.NoError(t, inst.Update(context.Background(), "upd-tag-source-mod"))

	installed, err := inst.Installed()
	require.NoError(t, err)
	require.Len(t, installed, 1)
	assert.Equal(t, "2.0.0", installed[0].Version)
	assert.Equal(t, "Tagged Source Module Updated", installed[0].Name)
}

func TestUpdate_Bad_PathTraversalCode_Good(t *testing.T) {
	modulesDir := filepath.Join(t.TempDir(), "modules")

	st, err := store.New(store.Options{Path: ":memory:"})
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(io.Local, modulesDir, st)
	err = inst.Update(context.Background(), "../escape")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid module code")
}

// A marketplace entry keyed as "claimed-mod" must not be allowed to redirect
// to a repo whose manifest declares a different code — that would let an
// attacker borrow another module's signature for their own payload.
func TestInstall_Bad_ManifestCodeMismatch_Good(t *testing.T) {
	repo := createTestRepo(t, "real-code", "1.0")
	modulesDir := filepath.Join(t.TempDir(), "modules")

	st, err := store.New(store.Options{Path: ":memory:"})
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(io.Local, modulesDir, st)
	err = inst.Install(context.Background(), Module{
		Code: "claimed-code",
		Repo: repo,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "manifest code mismatch")

	// Directory must be cleaned up — the install was rejected.
	_, statErr := os.Stat(filepath.Join(modulesDir, "claimed-code"))
	assert.True(t, os.IsNotExist(statErr))

	_, err = st.Get("_modules", "claimed-code")
	assert.Error(t, err)
}

// An update that re-homes the manifest to a different code (e.g. an attacker
// pushes a new tag that changes the code) must be rejected rather than
// silently replacing the entry.
func TestUpdate_Bad_ManifestCodeChanged_Good(t *testing.T) {
	repo := createTestRepo(t, "stable-code", "1.0")
	modulesDir := filepath.Join(t.TempDir(), "modules")

	st, err := store.New(store.Options{Path: ":memory:"})
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(io.Local, modulesDir, st)
	require.NoError(t, inst.Install(context.Background(), Module{
		Code: "stable-code",
		Repo: repo,
	}))

	// Upstream flips the manifest code mid-flight — reject.
	swapped := "code: evil-code\nname: Stable Module\nversion: \"1.1\"\n"
	require.NoError(t, os.WriteFile(
		filepath.Join(repo, ".core", "manifest.yaml"),
		[]byte(swapped), 0644,
	))
	runGit(t, repo, "add", ".")
	runGit(t, repo, "commit", "-m", "swap code")

	err = inst.Update(context.Background(), "stable-code")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "manifest code changed")
}
