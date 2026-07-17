// SPDX-License-Identifier: EUPL-1.2

package forgejo

import (
	"context"
	"time"

	gitea "code.gitea.io/sdk/gitea"
	core "dappco.re/go"
	forgejo "dappco.re/go/scm/third_party/forgejo/forgejo"
)

func TestForgejo_splitRepoRef_Good(t *core.T) {
	owner, repo, err := splitRepoRef("core/go-scm")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "core", owner)
	core.AssertEqual(t, "go-scm", repo)
}

func TestForgejo_splitRepoRef_Bad(t *core.T) {
	_, _, err := splitRepoRef("no-slash")
	core.AssertError(t, err)
}

func TestForgejo_splitRepoRef_Ugly(t *core.T) {
	_, _, err := splitRepoRef("core/")
	core.AssertError(t, err)
}

func TestForgejo_parseChildIssueNumbers_Good(t *core.T) {
	body := "- [ ] #12 first\n- [ ] #34 second\n- [x] #56 done\n"
	got := parseChildIssueNumbers(body)
	core.AssertEqual(t, []int64{12, 34}, got)
}

func TestForgejo_parseChildIssueNumbers_Bad(t *core.T) {
	got := parseChildIssueNumbers("no checkboxes here")
	core.AssertNil(t, got)
}

func TestForgejo_parseChildIssueNumbers_Ugly(t *core.T) {
	// Duplicate child references collapse to a single entry.
	got := parseChildIssueNumbers("- [ ] #7\n- [ ] #7\n")
	core.AssertEqual(t, []int64{7}, got)
}

func TestForgejo_pullState_Good(t *core.T) {
	core.AssertEqual(t, "OPEN", pullState(&forgejo.PullRequest{State: forgejo.StateOpen}))
}

func TestForgejo_pullState_Bad(t *core.T) {
	core.AssertEqual(t, "UNKNOWN", pullState(nil))
}

func TestForgejo_pullState_Ugly(t *core.T) {
	core.AssertEqual(t, "MERGED", pullState(&forgejo.PullRequest{HasMerged: true}))
}

func TestForgejo_mergeableState_Good(t *core.T) {
	core.AssertEqual(t, "MERGEABLE", mergeableState(&forgejo.PullRequest{Mergeable: true}))
}

func TestForgejo_mergeableState_Bad(t *core.T) {
	core.AssertEqual(t, "UNKNOWN", mergeableState(nil))
}

func TestForgejo_mergeableState_Ugly(t *core.T) {
	pr := &forgejo.PullRequest{Head: &gitea.PRBranchInfo{}, Base: &gitea.PRBranchInfo{}}
	core.AssertEqual(t, "CONFLICTING", mergeableState(pr))
}

func TestForgejo_issueAssignee_Good(t *core.T) {
	issue := &forgejo.Issue{Assignees: []*forgejo.User{{UserName: "hephaestus"}}}
	core.AssertEqual(t, "hephaestus", issueAssignee(issue))
}

func TestForgejo_issueAssignee_Bad(t *core.T) {
	core.AssertEqual(t, "", issueAssignee(nil))
}

func TestForgejo_issueAssignee_Ugly(t *core.T) {
	core.AssertEqual(t, "", issueAssignee(&forgejo.Issue{}))
}

func TestForgejo_pullCommitSHA_Good(t *core.T) {
	pr := &forgejo.PullRequest{Head: &gitea.PRBranchInfo{Sha: "abc123"}}
	core.AssertEqual(t, "abc123", pullCommitSHA(pr))
}

func TestForgejo_pullCommitSHA_Bad(t *core.T) {
	core.AssertEqual(t, "", pullCommitSHA(nil))
}

func TestForgejo_pullCommitSHA_Ugly(t *core.T) {
	core.AssertEqual(t, "", pullCommitSHA(&forgejo.PullRequest{}))
}

func TestForgejo_pullCommitTime_Good(t *core.T) {
	now := time.Now().UTC()
	pr := &forgejo.PullRequest{Updated: &now}
	core.AssertEqual(t, now, pullCommitTime(pr))
}

func TestForgejo_pullCommitTime_Bad(t *core.T) {
	core.AssertTrue(t, pullCommitTime(nil).IsZero())
}

func TestForgejo_pullCommitTime_Ugly(t *core.T) {
	created := time.Now().UTC().Add(-time.Hour)
	pr := &forgejo.PullRequest{Created: &created}
	core.AssertEqual(t, created, pullCommitTime(pr))
}

func TestForgejo_signalsForEpic_Good(t *core.T) {
	// An epic whose child checkboxes reference numbers but whose forge
	// returns errors for every PR/issue fetch yields no signals (each
	// signalForChild call errors out and is skipped).
	src := New(Config{}, testForgeClient(t))
	got := src.signalsForEpic(context.Background(), "core", "go-scm", &forgejo.Issue{Index: 1, Body: "- [ ] #2 child\n"})
	core.AssertEmpty(t, got)
}

func TestForgejo_signalsForEpic_Bad(t *core.T) {
	src := New(Config{}, nil)
	got := src.signalsForEpic(context.Background(), "core", "go-scm", nil)
	core.AssertNil(t, got)
}

func TestForgejo_signalsForEpic_Ugly(t *core.T) {
	src := New(Config{}, nil)
	got := src.signalsForEpic(context.Background(), "core", "go-scm", &forgejo.Issue{Index: 1, Body: "   "})
	core.AssertNil(t, got)
}
