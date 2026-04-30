// SPDX-License-Identifier: EUPL-1.2

package git

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	core "dappco.re/go"
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

func ax7GitCommand(t *core.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(out))
	}
}

func ax7GitRepo(t *core.T) string {
	root := t.TempDir()
	remote := filepath.Join(root, "remote.git")
	work := filepath.Join(root, "work")
	ax7GitCommand(t, root, "init", "--bare", remote)
	ax7GitCommand(t, root, "clone", remote, work)
	ax7GitCommand(t, work, "config", "user.name", "AX7")
	ax7GitCommand(t, work, "config", "user.email", "ax7@example.test")
	ax7GitCommand(t, work, "checkout", "-b", "dev")
	core.RequireNoError(t, os.WriteFile(filepath.Join(work, "README.md"), []byte("ready\n"), 0o600))
	ax7GitCommand(t, work, "add", "README.md")
	ax7GitCommand(t, work, "commit", "-m", "initial")
	ax7GitCommand(t, work, "push", "-u", "origin", "dev")
	return work
}

func TestGit_GitError_Error_Good(t *core.T) {
	err := (&GitError{Stderr: " rejected \n"}).Error()
	core.AssertEqual(
		t, "rejected", err,
	)
}

func TestGit_GitError_Error_Bad(t *core.T) {
	err := (&GitError{Err: errors.New("failed")}).Error()
	core.AssertEqual(
		t, "failed", err,
	)
}

func TestGit_GitError_Error_Ugly(t *core.T) {
	var err *GitError
	core.AssertEqual(
		t, "", err.Error(),
	)
}

func TestGit_GitError_Unwrap_Good(t *core.T) {
	cause := errors.New("failed")
	err := (&GitError{Err: cause}).Unwrap()
	core.AssertEqual(t, cause, err)
}

func TestGit_GitError_Unwrap_Bad(t *core.T) {
	err := (&GitError{}).Unwrap()
	core.AssertNil(
		t, err,
	)
}

func TestGit_GitError_Unwrap_Ugly(t *core.T) {
	var gitErr *GitError
	err := gitErr.Unwrap()
	core.AssertNil(t, err)
}

func TestGit_RepoStatus_HasUnpulled_Good(t *core.T) {
	status := &RepoStatus{Behind: 1}
	got := status.HasUnpulled()
	core.AssertTrue(t, got)
}

func TestGit_RepoStatus_HasUnpulled_Bad(t *core.T) {
	status := &RepoStatus{}
	got := status.HasUnpulled()
	core.AssertFalse(t, got)
}

func TestGit_RepoStatus_HasUnpulled_Ugly(t *core.T) {
	var status *RepoStatus
	got := status.HasUnpulled()
	core.AssertFalse(t, got)
}

func TestGit_RepoStatus_HasUnpushed_Good(t *core.T) {
	status := &RepoStatus{Ahead: 1}
	got := status.HasUnpushed()
	core.AssertTrue(t, got)
}

func TestGit_RepoStatus_HasUnpushed_Bad(t *core.T) {
	status := &RepoStatus{}
	got := status.HasUnpushed()
	core.AssertFalse(t, got)
}

func TestGit_RepoStatus_HasUnpushed_Ugly(t *core.T) {
	var status *RepoStatus
	got := status.HasUnpushed()
	core.AssertFalse(t, got)
}

func TestGit_RepoStatus_IsDirty_Good(t *core.T) {
	status := &RepoStatus{Modified: 1}
	got := status.IsDirty()
	core.AssertTrue(t, got)
}

func TestGit_RepoStatus_IsDirty_Bad(t *core.T) {
	status := &RepoStatus{}
	got := status.IsDirty()
	core.AssertFalse(t, got)
}

func TestGit_RepoStatus_IsDirty_Ugly(t *core.T) {
	var status *RepoStatus
	got := status.IsDirty()
	core.AssertFalse(t, got)
}

func TestGit_IsNonFastForward_Good(t *core.T) {
	got := IsNonFastForward(errors.New("push rejected: non-fast-forward"))
	core.AssertTrue(
		t, got,
	)
}

func TestGit_IsNonFastForward_Bad(t *core.T) {
	got := IsNonFastForward(errors.New("permission denied"))
	core.AssertFalse(
		t, got,
	)
}

func TestGit_IsNonFastForward_Ugly(t *core.T) {
	got := IsNonFastForward(nil)
	core.AssertFalse(
		t, got,
	)
}

func TestGit_Pull_Good(t *core.T) {
	repo := ax7GitRepo(t)
	err := Pull(context.Background(), repo)
	core.AssertNoError(t, err)
}

func TestGit_Pull_Bad(t *core.T) {
	err := Pull(context.Background(), "")
	core.AssertError(
		t, err,
	)
}

func TestGit_Pull_Ugly(t *core.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := Pull(ctx, ax7GitRepo(t))
	core.AssertError(t, err)
}

func TestGit_Push_Good(t *core.T) {
	repo := ax7GitRepo(t)
	err := Push(context.Background(), repo)
	core.AssertNoError(t, err)
}

func TestGit_Push_Bad(t *core.T) {
	err := Push(context.Background(), "")
	core.AssertError(
		t, err,
	)
}

func TestGit_Push_Ugly(t *core.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := Push(ctx, ax7GitRepo(t))
	core.AssertError(t, err)
}

func TestGit_Sync_Good(t *core.T) {
	repo := ax7GitRepo(t)
	err := Sync(context.Background(), repo)
	core.AssertNoError(t, err)
}

func TestGit_Sync_Bad(t *core.T) {
	err := Sync(context.Background(), "")
	core.AssertError(
		t, err,
	)
}

func TestGit_Sync_Ugly(t *core.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := Sync(ctx, ax7GitRepo(t))
	core.AssertError(t, err)
}

func TestGit_SyncWithRemote_Good(t *core.T) {
	repo := ax7GitRepo(t)
	err := SyncWithRemote(context.Background(), repo, "origin", "dev")
	core.AssertNoError(t, err)
}

func TestGit_SyncWithRemote_Bad(t *core.T) {
	err := SyncWithRemote(context.Background(), "", "origin", "dev")
	core.AssertError(
		t, err,
	)
}

func TestGit_SyncWithRemote_Ugly(t *core.T) {
	repo := ax7GitRepo(t)
	err := SyncWithRemote(context.Background(), repo, "", "")
	core.AssertNoError(t, err)
}

func TestGit_SyncMultiple_Good(t *core.T) {
	repo := ax7GitRepo(t)
	got := SyncMultiple(context.Background(), []string{repo}, map[string]string{repo: "demo"}, "origin", "dev")
	core.AssertLen(t, got, 1)
	core.AssertTrue(t, got[0].Success)
}

func TestGit_SyncMultiple_Bad(t *core.T) {
	got := SyncMultiple(context.Background(), []string{""}, nil, "origin", "dev")
	core.AssertLen(t, got, 1)
	core.AssertNotNil(t, got[0].Error)
}

func TestGit_SyncMultiple_Ugly(t *core.T) {
	got := SyncMultiple(context.Background(), nil, nil, "", "")
	core.AssertEmpty(
		t, got,
	)
}

func TestGit_PushMultiple_Good(t *core.T) {
	repo := ax7GitRepo(t)
	got := PushMultiple(context.Background(), []string{repo}, map[string]string{repo: "demo"})
	core.AssertLen(t, got, 1)
	core.AssertTrue(t, got[0].Success)
}

func TestGit_PushMultiple_Bad(t *core.T) {
	got := PushMultiple(context.Background(), []string{""}, nil)
	core.AssertLen(t, got, 1)
	core.AssertNotNil(t, got[0].Error)
}

func TestGit_PushMultiple_Ugly(t *core.T) {
	got := PushMultiple(context.Background(), nil, nil)
	core.AssertEmpty(
		t, got,
	)
}

func TestGit_PullMultiple_Good(t *core.T) {
	repo := ax7GitRepo(t)
	got := PullMultiple(context.Background(), []string{repo}, map[string]string{repo: "demo"})
	core.AssertLen(t, got, 1)
	core.AssertTrue(t, got[0].Success)
}

func TestGit_PullMultiple_Bad(t *core.T) {
	got := PullMultiple(context.Background(), []string{""}, nil)
	core.AssertLen(t, got, 1)
	core.AssertNotNil(t, got[0].Error)
}

func TestGit_PullMultiple_Ugly(t *core.T) {
	got := PullMultiple(context.Background(), nil, nil)
	core.AssertEmpty(
		t, got,
	)
}

func TestGit_PushMultipleIter_Good(t *core.T) {
	repo := ax7GitRepo(t)
	var got []PushResult
	for result := range PushMultipleIter(context.Background(), []string{repo}, map[string]string{repo: "demo"}) {
		got = append(got, result)
	}
	core.AssertLen(t, got, 1)
	core.AssertTrue(t, got[0].Success)
}

func TestGit_PushMultipleIter_Bad(t *core.T) {
	var got []PushResult
	for result := range PushMultipleIter(context.Background(), []string{""}, nil) {
		got = append(got, result)
	}
	core.AssertLen(t, got, 1)
	core.AssertNotNil(t, got[0].Error)
}

func TestGit_PushMultipleIter_Ugly(t *core.T) {
	var got []PushResult
	for result := range PushMultipleIter(context.Background(), nil, nil) {
		got = append(got, result)
	}
	core.AssertEmpty(t, got)
}

func TestGit_PullMultipleIter_Good(t *core.T) {
	repo := ax7GitRepo(t)
	var got []PushResult
	for result := range PullMultipleIter(context.Background(), []string{repo}, map[string]string{repo: "demo"}) {
		got = append(got, result)
	}
	core.AssertLen(t, got, 1)
	core.AssertTrue(t, got[0].Success)
}

func TestGit_PullMultipleIter_Bad(t *core.T) {
	var got []PushResult
	for result := range PullMultipleIter(context.Background(), []string{""}, nil) {
		got = append(got, result)
	}
	core.AssertLen(t, got, 1)
	core.AssertNotNil(t, got[0].Error)
}

func TestGit_PullMultipleIter_Ugly(t *core.T) {
	var got []PushResult
	for result := range PullMultipleIter(context.Background(), nil, nil) {
		got = append(got, result)
	}
	core.AssertEmpty(t, got)
}

func TestGit_Status_Good(t *core.T) {
	repo := ax7GitRepo(t)
	core.RequireNoError(t, os.WriteFile(filepath.Join(repo, "new.txt"), []byte("new\n"), 0o600))
	got := Status(context.Background(), StatusOptions{Paths: []string{repo}, Names: map[string]string{repo: "demo"}})
	core.AssertLen(t, got, 1)
	core.AssertEqual(t, "demo", got[0].Name)
	core.AssertEqual(t, 1, got[0].Untracked)
}

func TestGit_Status_Bad(t *core.T) {
	got := Status(context.Background(), StatusOptions{Paths: []string{filepath.Join(t.TempDir(), "missing")}})
	core.AssertLen(t, got, 1)
	core.AssertNotNil(t, got[0].Error)
}

func TestGit_Status_Ugly(t *core.T) {
	got := Status(context.Background(), StatusOptions{})
	core.AssertEmpty(
		t, got,
	)
}

func TestGit_StatusIter_Good(t *core.T) {
	repo := ax7GitRepo(t)
	var got []RepoStatus
	for status := range StatusIter(context.Background(), StatusOptions{Paths: []string{repo}}) {
		got = append(got, status)
	}
	core.AssertLen(t, got, 1)
	core.AssertEqual(t, filepath.Base(repo), got[0].Name)
}

func TestGit_StatusIter_Bad(t *core.T) {
	var got []RepoStatus
	for status := range StatusIter(context.Background(), StatusOptions{Paths: []string{filepath.Join(t.TempDir(), "missing")}}) {
		got = append(got, status)
	}
	core.AssertLen(t, got, 1)
	core.AssertNotNil(t, got[0].Error)
}

func TestGit_StatusIter_Ugly(t *core.T) {
	var got []RepoStatus
	for status := range StatusIter(context.Background(), StatusOptions{}) {
		got = append(got, status)
	}
	core.AssertEmpty(t, got)
}

func TestGit_NewService_Good(t *core.T) {
	c := core.New()
	result := NewService(ServiceOptions{WorkDir: t.TempDir()})(c)
	core.AssertTrue(t, result.OK)
	core.AssertNotNil(t, result.Value)
}

func TestGit_NewService_Bad(t *core.T) {
	result := NewService(ServiceOptions{})(nil)
	core.AssertTrue(t, result.OK)
	core.AssertNotNil(t, result.Value)
}

func TestGit_NewService_Ugly(t *core.T) {
	result := NewService(ServiceOptions{WorkDir: ""})(core.New())
	service := result.Value.(*Service)
	core.AssertEqual(t, "", service.Options().WorkDir)
}

func TestGit_Service_Status_Good(t *core.T) {
	service := &Service{lastStatus: []RepoStatus{{Name: "demo"}}}
	got := service.Status()
	core.AssertEqual(t, []RepoStatus{{Name: "demo"}}, got)
}

func TestGit_Service_Status_Bad(t *core.T) {
	service := &Service{}
	got := service.Status()
	core.AssertEmpty(t, got)
}

func TestGit_Service_Status_Ugly(t *core.T) {
	service := &Service{lastStatus: []RepoStatus{{Name: "demo"}}}
	got := service.Status()
	got[0].Name = "mutated"
	core.AssertEqual(t, "demo", service.lastStatus[0].Name)
}

func TestGit_Service_StatusIter_Good(t *core.T) {
	service := &Service{lastStatus: []RepoStatus{{Name: "demo"}}}
	var got []RepoStatus
	for status := range service.StatusIter() {
		got = append(got, status)
	}
	core.AssertEqual(t, "demo", got[0].Name)
}

func TestGit_Service_StatusIter_Bad(t *core.T) {
	service := &Service{}
	var got []RepoStatus
	for status := range service.StatusIter() {
		got = append(got, status)
	}
	core.AssertEmpty(t, got)
}

func TestGit_Service_StatusIter_Ugly(t *core.T) {
	service := &Service{lastStatus: []RepoStatus{{Name: "a"}, {Name: "b"}}}
	var got []RepoStatus
	for status := range service.StatusIter() {
		got = append(got, status)
		break
	}
	core.AssertEqual(t, []RepoStatus{{Name: "a"}}, got)
}

func TestGit_Service_DirtyRepos_Good(t *core.T) {
	service := &Service{lastStatus: []RepoStatus{{Name: "dirty", Modified: 1}, {Name: "clean"}}}
	got := service.DirtyRepos()
	core.AssertLen(t, got, 1)
	core.AssertEqual(t, "dirty", got[0].Name)
}

func TestGit_Service_DirtyRepos_Bad(t *core.T) {
	service := &Service{lastStatus: []RepoStatus{{Name: "clean"}}}
	got := service.DirtyRepos()
	core.AssertEmpty(t, got)
}

func TestGit_Service_DirtyRepos_Ugly(t *core.T) {
	service := &Service{}
	got := service.DirtyRepos()
	core.AssertEmpty(t, got)
}

func TestGit_Service_DirtyReposIter_Good(t *core.T) {
	service := &Service{lastStatus: []RepoStatus{{Name: "dirty", Untracked: 1}}}
	var got []RepoStatus
	for status := range service.DirtyReposIter() {
		got = append(got, status)
	}
	core.AssertEqual(t, "dirty", got[0].Name)
}

func TestGit_Service_DirtyReposIter_Bad(t *core.T) {
	service := &Service{lastStatus: []RepoStatus{{Name: "clean"}}}
	var got []RepoStatus
	for status := range service.DirtyReposIter() {
		got = append(got, status)
	}
	core.AssertEmpty(t, got)
}

func TestGit_Service_DirtyReposIter_Ugly(t *core.T) {
	service := &Service{}
	var got []RepoStatus
	for status := range service.DirtyReposIter() {
		got = append(got, status)
	}
	core.AssertEmpty(t, got)
}

func TestGit_Service_AheadRepos_Good(t *core.T) {
	service := &Service{lastStatus: []RepoStatus{{Name: "ahead", Ahead: 1}, {Name: "clean"}}}
	got := service.AheadRepos()
	core.AssertLen(t, got, 1)
	core.AssertEqual(t, "ahead", got[0].Name)
}

func TestGit_Service_AheadRepos_Bad(t *core.T) {
	service := &Service{lastStatus: []RepoStatus{{Name: "clean"}}}
	got := service.AheadRepos()
	core.AssertEmpty(t, got)
}

func TestGit_Service_AheadRepos_Ugly(t *core.T) {
	service := &Service{}
	got := service.AheadRepos()
	core.AssertEmpty(t, got)
}

func TestGit_Service_AheadReposIter_Good(t *core.T) {
	service := &Service{lastStatus: []RepoStatus{{Name: "ahead", Ahead: 1}}}
	var got []RepoStatus
	for status := range service.AheadReposIter() {
		got = append(got, status)
	}
	core.AssertEqual(t, "ahead", got[0].Name)
}

func TestGit_Service_AheadReposIter_Bad(t *core.T) {
	service := &Service{lastStatus: []RepoStatus{{Name: "clean"}}}
	var got []RepoStatus
	for status := range service.AheadReposIter() {
		got = append(got, status)
	}
	core.AssertEmpty(t, got)
}

func TestGit_Service_AheadReposIter_Ugly(t *core.T) {
	service := &Service{}
	var got []RepoStatus
	for status := range service.AheadReposIter() {
		got = append(got, status)
	}
	core.AssertEmpty(t, got)
}

func TestGit_Service_BehindRepos_Good(t *core.T) {
	service := &Service{lastStatus: []RepoStatus{{Name: "behind", Behind: 1}, {Name: "clean"}}}
	got := service.BehindRepos()
	core.AssertLen(t, got, 1)
	core.AssertEqual(t, "behind", got[0].Name)
}

func TestGit_Service_BehindRepos_Bad(t *core.T) {
	service := &Service{lastStatus: []RepoStatus{{Name: "clean"}}}
	got := service.BehindRepos()
	core.AssertEmpty(t, got)
}

func TestGit_Service_BehindRepos_Ugly(t *core.T) {
	service := &Service{}
	got := service.BehindRepos()
	core.AssertEmpty(t, got)
}

func TestGit_Service_BehindReposIter_Good(t *core.T) {
	service := &Service{lastStatus: []RepoStatus{{Name: "behind", Behind: 1}}}
	var got []RepoStatus
	for status := range service.BehindReposIter() {
		got = append(got, status)
	}
	core.AssertEqual(t, "behind", got[0].Name)
}

func TestGit_Service_BehindReposIter_Bad(t *core.T) {
	service := &Service{lastStatus: []RepoStatus{{Name: "clean"}}}
	var got []RepoStatus
	for status := range service.BehindReposIter() {
		got = append(got, status)
	}
	core.AssertEmpty(t, got)
}

func TestGit_Service_BehindReposIter_Ugly(t *core.T) {
	service := &Service{}
	var got []RepoStatus
	for status := range service.BehindReposIter() {
		got = append(got, status)
	}
	core.AssertEmpty(t, got)
}

func TestGit_Service_OnStartup_Good(t *core.T) {
	c := core.New(core.WithService(NewService(ServiceOptions{})))
	result := c.ServiceStartup(context.Background(), nil)
	core.AssertTrue(t, result.OK)
	core.AssertTrue(t, c.Action("git.push").Exists())
}

func TestGit_Service_OnStartup_Bad(t *core.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	service := &Service{}
	result := service.OnStartup(ctx)
	core.AssertFalse(t, result.OK)
	core.AssertErrorIs(t, result.Value.(error), context.Canceled)
}

func TestGit_Service_OnStartup_Ugly(t *core.T) {
	var service *Service
	result := service.OnStartup(context.Background())
	core.AssertTrue(t, result.OK)
}
