// SPDX-License-Identifier: EUPL-1.2

package repos

import (
	"context"
	"testing"

	core "dappco.re/go"
)

// seedSyncedRepo creates a bare remote plus a working clone that is one
// commit ahead, mirroring the fixture shape used elsewhere in this
// package. It returns the working repo path and its tracked file path.
func seedSyncedRepo(t *testing.T, root, name string) (repoPath, filePath string) {
	t.Helper()
	repoPath = core.PathJoin(root, name)
	remotePath := core.PathJoin(root, name+"-"+sonarServiceTestRemoteGit)
	filePath = core.PathJoin(repoPath, sonarServiceTestStateTxt)

	runGitCmd(t, root, "git", "init", sonarServiceTestBare, remotePath)
	runGitCmd(t, root, "git", "clone", remotePath, repoPath)
	runGitCmd(t, repoPath, "git", "-C", repoPath, "config", sonarServiceTestUserName, sonarServiceTestTestUser)
	runGitCmd(t, repoPath, "git", "-C", repoPath, "config", sonarServiceTestUserEmail, sonarServiceTestTestExampleCom)
	runGitCmd(t, repoPath, "git", "-C", repoPath, "checkout", "-b", "dev")

	if r := core.WriteFile(filePath, []byte(sonarServiceTestRemoteState), 0o600); !r.OK {
		t.Fatalf(sonarServiceTestWriteSeedFileV, r.Error())
	}
	runGitCmd(t, repoPath, "git", "-C", repoPath, "add", sonarServiceTestStateTxt)
	runGitCmd(t, repoPath, "git", "-C", repoPath, "commit", "-m", "initial")
	runGitCmd(t, repoPath, "git", "-C", repoPath, "push", "-u", "origin", "dev")

	if r := core.WriteFile(filePath, []byte(sonarServiceTestLocalChanges), 0o600); !r.OK {
		t.Fatalf(sonarServiceTestWriteLocalChangeV, r.Error())
	}
	runGitCmd(t, repoPath, "git", "-C", repoPath, "commit", "-am", sonarServiceTestLocalChange)
	return repoPath, filePath
}

func TestRepos_Service_syncAll_Good(t *testing.T) {
	root := t.TempDir()
	repoPath, _ := seedSyncedRepo(t, root, "repo1")

	reg := &Registry{Version: 1, BasePath: root, Repos: map[string]*Repo{"repo1": {Name: "repo1", Path: repoPath}}}
	svc := &Service{ServiceRuntime: core.NewServiceRuntime(core.New(), ServiceOptions{Root: root, Branch: "dev", Remote: "origin"}), registry: reg}

	results, err := svc.syncAll(context.Background(), core.NewOptions())
	if err != nil {
		t.Fatalf("syncAll: %v", err)
	}
	core.AssertLen(t, results, 1)
}

func TestRepos_Service_syncAll_Bad(t *testing.T) {
	svc := &Service{ServiceRuntime: core.NewServiceRuntime(core.New(), ServiceOptions{
		RegistryPath: core.PathJoin(t.TempDir(), ".core", "missing.yaml"),
	})}
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CORE_REPOS", "")
	_, err := svc.syncAll(context.Background(), core.NewOptions())
	core.AssertError(t, err)
}

func TestRepos_Service_syncAll_Ugly(t *testing.T) {
	var svc *Service
	_, err := svc.syncAll(context.Background(), core.NewOptions())
	core.AssertError(t, err)
}

func TestRepos_Service_handleRepoSyncAll_Good(t *testing.T) {
	root := t.TempDir()
	repoPath, _ := seedSyncedRepo(t, root, "repo1")

	reg := &Registry{Version: 1, BasePath: root, Repos: map[string]*Repo{"repo1": {Name: "repo1", Path: repoPath}}}
	svc := &Service{ServiceRuntime: core.NewServiceRuntime(core.New(), ServiceOptions{Root: root, Branch: "dev", Remote: "origin"}), registry: reg}

	result := svc.handleRepoSyncAll(context.Background(), core.NewOptions())
	core.AssertTrue(t, result.OK)
}

func TestRepos_Service_handleRepoSyncAll_Bad(t *testing.T) {
	svc := &Service{ServiceRuntime: core.NewServiceRuntime(core.New(), ServiceOptions{
		RegistryPath: core.PathJoin(t.TempDir(), ".core", "missing.yaml"),
	})}
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CORE_REPOS", "")
	result := svc.handleRepoSyncAll(context.Background(), core.NewOptions())
	core.AssertFalse(t, result.OK)
}

func TestRepos_Service_handleRepoSyncAll_Ugly(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	svc := &Service{ServiceRuntime: core.NewServiceRuntime(core.New(), ServiceOptions{})}
	result := svc.handleRepoSyncAll(ctx, core.NewOptions())
	core.AssertFalse(t, result.OK)
}

func TestRepos_Service_syncRepo_Good(t *testing.T) {
	root := t.TempDir()
	repoPath, _ := seedSyncedRepo(t, root, "repo1")
	svc := &Service{ServiceRuntime: core.NewServiceRuntime(core.New(), ServiceOptions{Branch: "dev", Remote: "origin"})}
	result, err := svc.syncRepo(context.Background(), core.NewOptions(core.Option{Key: `path`, Value: repoPath}))
	if err != nil {
		t.Fatalf("syncRepo by path: %v", err)
	}
	core.AssertTrue(t, result.Success)
}

func TestRepos_Service_syncRepo_Bad(t *testing.T) {
	svc := &Service{ServiceRuntime: core.NewServiceRuntime(core.New(), ServiceOptions{})}
	_, err := svc.syncRepo(context.Background(), core.NewOptions())
	core.AssertError(t, err)
}

func TestRepos_Service_syncRepo_Ugly(t *testing.T) {
	var svc *Service
	_, err := svc.syncRepo(context.Background(), core.NewOptions(core.Option{Key: "repo", Value: "x"}))
	core.AssertError(t, err)
}

func TestRepos_Service_syncNamedRepo_RegistryNotLoaded(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CORE_REPOS", "")
	svc := &Service{ServiceRuntime: core.NewServiceRuntime(core.New(), ServiceOptions{
		Root:         root,
		RegistryPath: core.PathJoin(root, ".core", "missing.yaml"),
	})}
	_, err := svc.syncNamedRepo(context.Background(), core.NewOptions(), "ghost", "", false, "origin", "dev")
	core.AssertError(t, err)
}
