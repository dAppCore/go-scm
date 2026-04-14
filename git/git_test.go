// SPDX-License-Identifier: EUPL-1.2

package git

import (
	"context"
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
	os "dappco.re/go/core/scm/internal/ax/osx"
	exec "golang.org/x/sys/execabs"
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
