// SPDX-License-Identifier: EUPL-1.2

package git

import (
	"context"
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
	os "dappco.re/go/core/scm/internal/ax/osx"
	exec "golang.org/x/sys/execabs"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListRemoteTags_Good(t *testing.T) {
	repo := createTaggedRepo(t, "tagged-repo",
		repoVersion{Version: "1.0.0", Tag: "v1.0.0"},
		repoVersion{Version: "2.0.0", Tag: "v2.0.0"},
	)

	tags, err := ListRemoteTags(context.Background(), repo)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"v1.0.0", "v2.0.0"}, tags)
}

func TestClone_Good_WithTag_Good(t *testing.T) {
	repo := createTaggedRepo(t, "clone-repo",
		repoVersion{Version: "1.0.0", Tag: "v1.0.0"},
		repoVersion{Version: "2.0.0", Tag: "v2.0.0"},
	)
	dest := filepath.Join(t.TempDir(), "clone")

	require.NoError(t, Clone(context.Background(), repo, dest, "v1.0.0"))

	tag, err := CurrentTag(context.Background(), dest)
	require.NoError(t, err)
	assert.Equal(t, "v1.0.0", tag)
}

func TestCheckout_Good_SwitchTag_Good(t *testing.T) {
	repo := createTaggedRepo(t, "checkout-repo",
		repoVersion{Version: "1.0.0", Tag: "v1.0.0"},
		repoVersion{Version: "2.0.0", Tag: "v2.0.0"},
	)
	dest := filepath.Join(t.TempDir(), "clone")

	require.NoError(t, Clone(context.Background(), repo, dest, "v1.0.0"))
	require.NoError(t, FetchTags(context.Background(), dest))
	require.NoError(t, Checkout(context.Background(), dest, "v2.0.0"))

	tag, err := CurrentTag(context.Background(), dest)
	require.NoError(t, err)
	assert.Equal(t, "v2.0.0", tag)
}

func TestCreateBranch_Good(t *testing.T) {
	repo := createTaggedRepo(t, "branch-repo",
		repoVersion{Version: "1.0.0"},
	)

	require.NoError(t, CreateBranch(context.Background(), repo, "feature/rfc", ""))

	current, err := exec.Command("git", "-C", repo, "rev-parse", "--abbrev-ref", "HEAD").CombinedOutput()
	require.NoError(t, err, string(current))
	assert.Equal(t, "feature/rfc\n", string(current))
}

func TestSwitchBranch_Good(t *testing.T) {
	repo := createTaggedRepo(t, "switch-branch-repo",
		repoVersion{Version: "1.0.0"},
	)
	runGit(t, repo, "checkout", "-b", "feature/switch")
	runGit(t, repo, "checkout", "-")

	require.NoError(t, SwitchBranch(context.Background(), repo, "feature/switch"))

	current, err := exec.Command("git", "-C", repo, "rev-parse", "--abbrev-ref", "HEAD").CombinedOutput()
	require.NoError(t, err, string(current))
	assert.Equal(t, "feature/switch\n", string(current))
}

func TestAddAllCommit_Good(t *testing.T) {
	repo := createTaggedRepo(t, "commit-repo",
		repoVersion{Version: "1.0.0"},
	)

	require.NoError(t, os.WriteFile(filepath.Join(repo, "notes.txt"), []byte("hello\n"), 0644))
	require.NoError(t, AddAll(context.Background(), repo))
	require.NoError(t, Commit(context.Background(), repo, "add notes"))

	statuses := Status(context.Background(), StatusOptions{
		Paths: []string{repo},
		Names: map[string]string{repo: "commit-repo"},
	})
	require.Len(t, statuses, 1)
	assert.False(t, statuses[0].IsDirty())
}

func TestVerifyCommitSignature_Good_UnsignedCommit_Good(t *testing.T) {
	repo := createTaggedRepo(t, "unsigned-commit-repo",
		repoVersion{Version: "1.0.0"},
	)

	valid, err := VerifyCommitSignature(context.Background(), repo, "HEAD")
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestVerifyTagSignature_Good_UnsignedTag_Good(t *testing.T) {
	repo := createTaggedRepo(t, "unsigned-tag-repo",
		repoVersion{Version: "1.0.0", Tag: "v1.0.0"},
	)

	valid, err := VerifyTagSignature(context.Background(), repo, "v1.0.0")
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestSSHCommand_Good(t *testing.T) {
	cmd := SSHCommand(SSHOptions{
		KeyPath:        "/tmp/test key",
		KnownHostsPath: "/tmp/known_hosts",
	})

	assert.Contains(t, cmd, "BatchMode=yes")
	assert.Contains(t, cmd, "IdentitiesOnly=yes")
	assert.Contains(t, cmd, "StrictHostKeyChecking=yes")
	assert.Contains(t, cmd, "'/tmp/test key'")
	assert.Contains(t, cmd, "UserKnownHostsFile=/tmp/known_hosts")
}

func TestConfigureSSH_Good(t *testing.T) {
	cmd := exec.Command("git", "status")

	ConfigureSSH(cmd, SSHOptions{KeyPath: "/tmp/id_ed25519"})

	found := false
	for _, entry := range cmd.Env {
		if strings.HasPrefix(entry, "GIT_SSH_COMMAND=") {
			found = true
			assert.Contains(t, entry, "/tmp/id_ed25519")
		}
	}
	assert.True(t, found)
}

type repoVersion struct {
	Version string
	Tag     string
}

func createTaggedRepo(t *testing.T, name string, versions ...repoVersion) string {
	t.Helper()

	dir := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".core"), 0755))
	runGit(t, dir, "init")

	for _, version := range versions {
		manifestYAML := "code: " + name + "\nname: " + name + "\nversion: \"" + version.Version + "\"\n"
		require.NoError(t, os.WriteFile(filepath.Join(dir, ".core", "manifest.yaml"), []byte(manifestYAML), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "version.txt"), []byte(version.Version+"\n"), 0644))
		runGit(t, dir, "add", "--force", ".")
		runGit(t, dir, "commit", "-m", "version-"+version.Version)
		if version.Tag != "" {
			runGit(t, dir, "tag", version.Tag)
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
