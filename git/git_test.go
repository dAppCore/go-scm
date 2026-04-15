// SPDX-License-Identifier: EUPL-1.2

package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestSyncWithRemoteResetsWorkingTree(t *testing.T) {
	tempDir := t.TempDir()
	remoteDir := filepath.Join(tempDir, "remote.git")
	workDir := filepath.Join(tempDir, "work")

	runGitCmd(t, tempDir, "git", "init", "--bare", remoteDir)
	runGitCmd(t, tempDir, "git", "clone", remoteDir, workDir)
	runGitCmd(t, workDir, "git", "-C", workDir, "config", "user.name", "Test User")
	runGitCmd(t, workDir, "git", "-C", workDir, "config", "user.email", "test@example.com")
	runGitCmd(t, workDir, "git", "-C", workDir, "checkout", "-b", "dev")

	filePath := filepath.Join(workDir, "state.txt")
	if err := os.WriteFile(filePath, []byte("remote state\n"), 0o600); err != nil {
		t.Fatalf("write initial file: %v", err)
	}
	runGitCmd(t, workDir, "git", "-C", workDir, "add", "state.txt")
	runGitCmd(t, workDir, "git", "-C", workDir, "commit", "-m", "initial")
	runGitCmd(t, workDir, "git", "-C", workDir, "push", "-u", "origin", "dev")

	if err := os.WriteFile(filePath, []byte("local changes\n"), 0o600); err != nil {
		t.Fatalf("write modified file: %v", err)
	}
	runGitCmd(t, workDir, "git", "-C", workDir, "commit", "-am", "local change")

	if err := SyncWithRemote(context.Background(), workDir, "origin", "dev"); err != nil {
		t.Fatalf("sync: %v", err)
	}

	raw, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read synced file: %v", err)
	}
	if got := string(raw); got != "remote state\n" {
		t.Fatalf("unexpected synced file contents: %q", got)
	}
}

func runGitCmd(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, string(out))
	}
}
