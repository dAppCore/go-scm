// SPDX-License-Identifier: EUPL-1.2

package git

import (
	"context"
	"testing"

	core "dappco.re/go"
	process "dappco.re/go/process"
)

func TestSyncWithRemoteResetsWorkingTree(t *testing.T) {
	tempDir := t.TempDir()
	remoteDir := core.PathJoin(tempDir, "remote.git")
	workDir := core.PathJoin(tempDir, "work")

	runGitCmd(t, tempDir, "git", "init", "--bare", remoteDir)
	runGitCmd(t, tempDir, "git", "clone", remoteDir, workDir)
	runGitCmd(t, workDir, "git", "-C", workDir, "config", "user.name", "Test User")
	runGitCmd(t, workDir, "git", "-C", workDir, "config", "user.email", "test@example.com")
	runGitCmd(t, workDir, "git", "-C", workDir, "checkout", "-b", "dev")

	filePath := core.PathJoin(workDir, "state.txt")
	if r := core.WriteFile(filePath, []byte("remote state\n"), 0o600); !r.OK {
		t.Fatalf("write initial file: %s", r.Error())
	}
	runGitCmd(t, workDir, "git", "-C", workDir, "add", "state.txt")
	runGitCmd(t, workDir, "git", "-C", workDir, "commit", "-m", "initial")
	runGitCmd(t, workDir, "git", "-C", workDir, "push", "-u", "origin", "dev")

	if r := core.WriteFile(filePath, []byte("local changes\n"), 0o600); !r.OK {
		t.Fatalf("write modified file: %s", r.Error())
	}
	runGitCmd(t, workDir, "git", "-C", workDir, "commit", "-am", "local change")

	if err := SyncWithRemote(context.Background(), workDir, "origin", "dev"); err != nil {
		t.Fatalf("sync: %v", err)
	}

	rRead := core.ReadFile(filePath)
	if !rRead.OK {
		t.Fatalf("read synced file: %s", rRead.Error())
	}
	raw := rRead.Value.([]byte)
	if got := string(raw); got != "remote state\n" {
		t.Fatalf("unexpected synced file contents: %q", got)
	}
}

func runGitCmd(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	r := process.RunWithOptions(context.Background(), process.RunOptions{
		Command: name,
		Args:    args,
		Dir:     dir,
	})
	if !r.OK {
		out, _ := r.Value.(string)
		t.Fatalf("%s %v failed: %s\n%s", name, args, r.Error(), out)
	}
}

func testGitCommand(t *core.T, dir string, args ...string) {
	t.Helper()
	r := process.RunWithOptions(context.Background(), process.RunOptions{
		Command: "git",
		Args:    args,
		Dir:     dir,
	})
	if !r.OK {
		out, _ := r.Value.(string)
		t.Fatalf("git %v failed: %s\n%s", args, r.Error(), out)
	}
}

func testGitRepo(t *core.T) string {
	root := t.TempDir()
	remote := core.PathJoin(root, "remote.git")
	work := core.PathJoin(root, "work")
	testGitCommand(t, root, "init", "--bare", remote)
	testGitCommand(t, root, "clone", remote, work)
	testGitCommand(t, work, "config", "user.name", "AX7")
	testGitCommand(t, work, "config", "user.email", "ax7@example.test")
	testGitCommand(t, work, "checkout", "-b", "dev")
	if r := core.WriteFile(core.PathJoin(work, "README.md"), []byte("ready\n"), 0o600); !r.OK {
		t.Fatalf("write README: %s", r.Error())
	}
	testGitCommand(t, work, "add", "README.md")
	testGitCommand(t, work, "commit", "-m", "initial")
	testGitCommand(t, work, "push", "-u", "origin", "dev")
	return work
}

func TestGit_GitError_Error_Good(t *core.T) {
	err := (&GitError{Stderr: " rejected \n"}).Error()
	core.AssertEqual(
		t, "rejected", err,
	)
}

func TestGit_GitError_Error_Bad(t *core.T) {
	err := (&GitError{Err: core.E("test", "failed", nil)}).Error()
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
	cause := core.E("test", "failed", nil)
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
	got := IsNonFastForward(core.E("test", "push rejected: non-fast-forward", nil))
	core.AssertTrue(
		t, got,
	)
}

func TestGit_IsNonFastForward_Bad(t *core.T) {
	got := IsNonFastForward(core.E("test", "permission denied", nil))
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
	repo := testGitRepo(t)
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
	err := Pull(ctx, testGitRepo(t))
	core.AssertError(t, err)
}

func TestGit_Push_Good(t *core.T) {
	repo := testGitRepo(t)
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
	err := Push(ctx, testGitRepo(t))
	core.AssertError(t, err)
}

func TestGit_Sync_Good(t *core.T) {
	repo := testGitRepo(t)
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
	err := Sync(ctx, testGitRepo(t))
	core.AssertError(t, err)
}

func TestGit_SyncWithRemote_Good(t *core.T) {
	repo := testGitRepo(t)
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
	repo := testGitRepo(t)
	err := SyncWithRemote(context.Background(), repo, "", "")
	core.AssertNoError(t, err)
}

func TestGit_SyncMultiple_Good(t *core.T) {
	repo := testGitRepo(t)
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
	repo := testGitRepo(t)
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
	repo := testGitRepo(t)
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
	repo := testGitRepo(t)
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
	repo := testGitRepo(t)
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
	repo := testGitRepo(t)
	if r := core.WriteFile(core.PathJoin(repo, "new.txt"), []byte("new\n"), 0o600); !r.OK {
		t.Fatalf("write new.txt: %s", r.Error())
	}
	got := Status(context.Background(), StatusOptions{Paths: []string{repo}, Names: map[string]string{repo: "demo"}})
	core.AssertLen(t, got, 1)
	core.AssertEqual(t, "demo", got[0].Name)
	core.AssertEqual(t, 1, got[0].Untracked)
}

func TestGit_Status_Bad(t *core.T) {
	got := Status(context.Background(), StatusOptions{Paths: []string{core.PathJoin(t.TempDir(), "missing")}})
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
	repo := testGitRepo(t)
	var got []RepoStatus
	for status := range StatusIter(context.Background(), StatusOptions{Paths: []string{repo}}) {
		got = append(got, status)
	}
	core.AssertLen(t, got, 1)
	core.AssertEqual(t, core.PathBase(repo), got[0].Name)
}

func TestGit_StatusIter_Bad(t *core.T) {
	var got []RepoStatus
	for status := range StatusIter(context.Background(), StatusOptions{Paths: []string{core.PathJoin(t.TempDir(), "missing")}}) {
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
	reference := "OnStartup"
	if reference == "" {
		t.Fatal(reference)
	}
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
