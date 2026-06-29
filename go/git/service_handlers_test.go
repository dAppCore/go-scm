// SPDX-License-Identifier: EUPL-1.2

package git

import (
	"context"

	core "dappco.re/go"
)

func TestGit_parseCount_Good(t *core.T) {
	core.AssertEqual(t, 7, parseCount("7"))
}

func TestGit_parseCount_Bad(t *core.T) {
	core.AssertEqual(t, 0, parseCount("not-a-number"))
}

func TestGit_parseCount_Ugly(t *core.T) {
	core.AssertEqual(t, 12, parseCount("  12 "))
}

func TestGit_parseBranchLine_Good(t *core.T) {
	st := &RepoStatus{}
	parseBranchLine("dev...origin/dev [ahead 2, behind 3]", st)
	core.AssertEqual(t, "dev", st.Branch)
	core.AssertEqual(t, 2, st.Ahead)
	core.AssertEqual(t, 3, st.Behind)
}

func TestGit_parseBranchLine_Bad(t *core.T) {
	st := &RepoStatus{}
	parseBranchLine("dev", st)
	core.AssertEqual(t, "dev", st.Branch)
	core.AssertEqual(t, 0, st.Ahead)
	core.AssertEqual(t, 0, st.Behind)
}

func TestGit_parseBranchLine_Ugly(t *core.T) {
	st := &RepoStatus{}
	parseBranchLine("dev...origin/dev", st)
	core.AssertEqual(t, "dev", st.Branch)
	core.AssertEqual(t, 0, st.Ahead)
	core.AssertEqual(t, 0, st.Behind)
}

func TestGit_parseStatus_Good(t *core.T) {
	// parseStatus trims each line before inspecting it, so the leading
	// porcelain status column collapses: "A " and "M " both surface as a
	// staged entry (line[0] is a non-space, non-'?' rune).
	st := &RepoStatus{}
	parseStatus("## dev...origin/dev [ahead 1]\n?? new.txt\nM edited.txt\nA staged.txt\n", st)
	core.AssertEqual(t, "dev", st.Branch)
	core.AssertEqual(t, 1, st.Ahead)
	core.AssertEqual(t, 1, st.Untracked)
	core.AssertEqual(t, 2, st.Staged)
}

func TestGit_parseStatus_Bad(t *core.T) {
	st := &RepoStatus{}
	parseStatus("", st)
	core.AssertEqual(t, "", st.Branch)
	core.AssertEqual(t, 0, st.Untracked)
}

func TestGit_parseStatus_Ugly(t *core.T) {
	st := &RepoStatus{}
	parseStatus("X\n## main\n", st)
	core.AssertEqual(t, "main", st.Branch)
	core.AssertEqual(t, 0, st.Modified)
}

func TestGit_Service_validatePath_Good(t *core.T) {
	service := &Service{ServiceRuntime: core.NewServiceRuntime(core.New(), ServiceOptions{})}
	err := service.validatePath(core.PathJoin(t.TempDir(), "repo"))
	core.AssertNoError(t, err)
}

func TestGit_Service_validatePath_Bad(t *core.T) {
	service := &Service{ServiceRuntime: core.NewServiceRuntime(core.New(), ServiceOptions{})}
	err := service.validatePath("relative/path")
	core.AssertError(t, err)
}

func TestGit_Service_validatePath_Ugly(t *core.T) {
	work := t.TempDir()
	service := &Service{ServiceRuntime: core.NewServiceRuntime(core.New(), ServiceOptions{WorkDir: work})}
	err := service.validatePath(core.PathJoin(work, "child"))
	core.AssertNoError(t, err)
}

func TestGit_Service_validatePaths_Good(t *core.T) {
	service := &Service{ServiceRuntime: core.NewServiceRuntime(core.New(), ServiceOptions{})}
	err := service.validatePaths([]string{core.PathJoin(t.TempDir(), "a"), core.PathJoin(t.TempDir(), "b")})
	core.AssertNoError(t, err)
}

func TestGit_Service_validatePaths_Bad(t *core.T) {
	service := &Service{ServiceRuntime: core.NewServiceRuntime(core.New(), ServiceOptions{})}
	err := service.validatePaths([]string{core.PathJoin(t.TempDir(), "ok"), "relative"})
	core.AssertError(t, err)
}

func TestGit_Service_validatePaths_Ugly(t *core.T) {
	service := &Service{ServiceRuntime: core.NewServiceRuntime(core.New(), ServiceOptions{})}
	err := service.validatePaths(nil)
	core.AssertNoError(t, err)
}

func TestGit_Service_runPush_Good(t *core.T) {
	repo := testGitRepo(t)
	service := &Service{ServiceRuntime: core.NewServiceRuntime(core.New(), ServiceOptions{})}
	result := service.runPush(context.Background(), repo)
	core.AssertTrue(t, result.OK)
}

func TestGit_Service_runPush_Bad(t *core.T) {
	service := &Service{ServiceRuntime: core.NewServiceRuntime(core.New(), ServiceOptions{})}
	result := service.runPush(context.Background(), "relative")
	core.AssertFalse(t, result.OK)
}

func TestGit_Service_runPush_Ugly(t *core.T) {
	service := &Service{ServiceRuntime: core.NewServiceRuntime(core.New(), ServiceOptions{})}
	result := service.runPush(context.Background(), core.PathJoin(t.TempDir(), "missing"))
	core.AssertFalse(t, result.OK)
}

func TestGit_Service_runPull_Good(t *core.T) {
	repo := testGitRepo(t)
	service := &Service{ServiceRuntime: core.NewServiceRuntime(core.New(), ServiceOptions{})}
	result := service.runPull(context.Background(), repo)
	core.AssertTrue(t, result.OK)
}

func TestGit_Service_runPull_Bad(t *core.T) {
	service := &Service{ServiceRuntime: core.NewServiceRuntime(core.New(), ServiceOptions{})}
	result := service.runPull(context.Background(), "relative")
	core.AssertFalse(t, result.OK)
}

func TestGit_Service_runPull_Ugly(t *core.T) {
	service := &Service{ServiceRuntime: core.NewServiceRuntime(core.New(), ServiceOptions{})}
	result := service.runPull(context.Background(), core.PathJoin(t.TempDir(), "missing"))
	core.AssertFalse(t, result.OK)
}

func TestGit_Service_runPushMultiple_Good(t *core.T) {
	repo := testGitRepo(t)
	service := &Service{ServiceRuntime: core.NewServiceRuntime(core.New(), ServiceOptions{})}
	result := service.runPushMultiple(context.Background(), []string{repo}, map[string]string{repo: "demo"})
	core.AssertTrue(t, result.OK)
}

func TestGit_Service_runPushMultiple_Bad(t *core.T) {
	service := &Service{ServiceRuntime: core.NewServiceRuntime(core.New(), ServiceOptions{})}
	result := service.runPushMultiple(context.Background(), []string{"relative"}, nil)
	core.AssertFalse(t, result.OK)
}

func TestGit_Service_runPushMultiple_Ugly(t *core.T) {
	service := &Service{ServiceRuntime: core.NewServiceRuntime(core.New(), ServiceOptions{})}
	result := service.runPushMultiple(context.Background(), nil, nil)
	core.AssertTrue(t, result.OK)
}

func TestGit_Service_runPullMultiple_Good(t *core.T) {
	repo := testGitRepo(t)
	service := &Service{ServiceRuntime: core.NewServiceRuntime(core.New(), ServiceOptions{})}
	result := service.runPullMultiple(context.Background(), []string{repo}, map[string]string{repo: "demo"})
	core.AssertTrue(t, result.OK)
}

func TestGit_Service_runPullMultiple_Bad(t *core.T) {
	service := &Service{ServiceRuntime: core.NewServiceRuntime(core.New(), ServiceOptions{})}
	result := service.runPullMultiple(context.Background(), []string{"relative"}, nil)
	core.AssertFalse(t, result.OK)
}

func TestGit_Service_runPullMultiple_Ugly(t *core.T) {
	service := &Service{ServiceRuntime: core.NewServiceRuntime(core.New(), ServiceOptions{})}
	result := service.runPullMultiple(context.Background(), nil, nil)
	core.AssertTrue(t, result.OK)
}

func TestGit_Service_handleTaskMessage_Good(t *core.T) {
	repo := testGitRepo(t)
	c := core.New()
	service := &Service{ServiceRuntime: core.NewServiceRuntime(c, ServiceOptions{})}
	result := service.handleTaskMessage(c, TaskPush{Path: repo, Name: "demo"})
	core.AssertTrue(t, result.OK)
}

func TestGit_Service_handleTaskMessage_Bad(t *core.T) {
	c := core.New()
	service := &Service{ServiceRuntime: core.NewServiceRuntime(c, ServiceOptions{})}
	result := service.handleTaskMessage(c, TaskPull{Path: "relative"})
	core.AssertFalse(t, result.OK)
}

func TestGit_Service_handleTaskMessage_Ugly(t *core.T) {
	c := core.New()
	service := &Service{ServiceRuntime: core.NewServiceRuntime(c, ServiceOptions{})}
	result := service.handleTaskMessage(c, "not-a-task")
	core.AssertFalse(t, result.OK)
}

func TestGit_Service_handleQuery_Good(t *core.T) {
	repo := testGitRepo(t)
	c := core.New()
	service := &Service{ServiceRuntime: core.NewServiceRuntime(c, ServiceOptions{})}
	result := service.handleQuery(c, QueryStatus{Paths: []string{repo}, Names: map[string]string{repo: "demo"}})
	core.AssertTrue(t, result.OK)
}

func TestGit_Service_handleQuery_Bad(t *core.T) {
	c := core.New()
	service := &Service{ServiceRuntime: core.NewServiceRuntime(c, ServiceOptions{})}
	result := service.handleQuery(c, QueryStatus{Paths: []string{"relative"}})
	core.AssertFalse(t, result.OK)
}

func TestGit_Service_handleQuery_Ugly(t *core.T) {
	c := core.New()
	service := &Service{ServiceRuntime: core.NewServiceRuntime(c, ServiceOptions{})}
	result := service.handleQuery(c, "not-a-query")
	core.AssertFalse(t, result.OK)
}

func TestGit_Service_handleQuery_DirtyAheadBehind(t *core.T) {
	c := core.New()
	service := &Service{ServiceRuntime: core.NewServiceRuntime(c, ServiceOptions{})}
	core.AssertTrue(t, service.handleQuery(c, QueryDirtyRepos{}).OK)
	core.AssertTrue(t, service.handleQuery(c, QueryAheadRepos{}).OK)
	core.AssertTrue(t, service.handleQuery(c, QueryBehindRepos{}).OK)
}
