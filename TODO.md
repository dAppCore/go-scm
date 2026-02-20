# TODO.md — go-scm Task Queue

Dispatched from core/go orchestration. Pick up tasks in order.

---

## Phase 1: Test Coverage (Critical)

forge/, gitea/, and git/ have **zero tests**. This is the top priority.

All Phase 1 tasks completed in commit `9db37c6`.

- [x] **forge/ unit tests** — 91.2% coverage. Tested all SDK wrapper functions via httptest mock server: client creation, repos, issues, PRs, labels, webhooks, orgs, meta, config resolution, SetPRDraft raw HTTP. 8 test files.
- [x] **gitea/ unit tests** — 89.2% coverage. Tested all SDK wrapper functions via httptest mock server: client creation, repos, issues, PRs, meta, config resolution. 5 test files.
- [x] **git/ unit tests** — 79.5% coverage. Tested RepoStatus methods, status parsing with real temp git repos, multi-repo parallel status, Push/Pull error paths, ahead/behind with bare remote, context cancellation, GitError, IsNonFastForward. Service DirtyRepos/AheadRepos filtering. 2 test files.
- [x] **jobrunner handler tests** — Audited: 86.4% (jobrunner), 73.3% (forgejo), 61.6% (handlers). All above 60%, no changes needed.
- [x] **collect/ test audit** — 57.3% coverage. Gaps are HTTP-dependent collector functions (fetchPage, Collect methods). Improvement requires mock HTTP servers for external services (BitcoinTalk, GitHub). Deferred to Phase 2.
- [x] **agentci/ bonus** — Improved from 56% to 94.5%. Added tests for Clotho (DeterminePlan, GetVerifierModel, FindByForgejoUser, Weave) and security (SanitizePath, EscapeShellArg, SecureSSHCommand, MaskToken).

## Phase 2: Hardening

All Phase 2 tasks completed in commit `3ba8fbb`.

- [x] **Config resolution audit** — Verified and tested in Phase 1. Both forge/ and gitea/ use identical priority: config file → env vars → flags. Documented in FINDINGS.md.
- [x] **Error wrapping** — All 15 bare `fmt.Errorf` calls converted to `"package.Func: context"` pattern across 5 files (journal.go, config.go, security.go, dispatch.go, labels.go). Existing `log.E()`/`core.E()` calls already followed the pattern.
- [x] **Context propagation** — Verified: Forgejo/Gitea SDK v2 does NOT accept `context.Context` (adding ctx to 66 wrappers = ceremony). Added `SecureSSHCommandContext` and `CheckGitHubRateLimitCtx` for real context propagation in SSH and CLI operations. Updated dispatch handler to pass ctx through. Documented SDK limitation in FINDINGS.md.
- [x] **Rate limiting** — Reviewed: handles all edge cases (context cancellation, unknown sources, concurrent access, adaptive throttling at 75% GitHub usage). GitHub uses `gh` CLI (handles its own headers). Forgejo/Gitea SDKs don't expose rate limit headers. Added context-aware `CheckGitHubRateLimitCtx`.

## Phase 3: AgentCI Pipeline

- [x] **Clotho dual-run validation** — All code paths tested in Phase 1: standard mode, dual-run by agent config, dual-run by critical repo name, non-verified strategy, unknown agent. Also tested GetVerifierModel, FindByForgejoUser, and Weave.
- [ ] **Forgejo signal source tests** — `forgejo/source.go` polls for webhook events. Test signal parsing and filtering.
- [ ] **Journal replay** — `journal.go` writes JSONL audit trail. Add test for write + read-back + filtering by action/repo/time range.
- [ ] **Handler integration** — Test full signal → handler → result flow with mock Forgejo client. Verify `tick_parent` correctly updates epic progress.

## Phase 4: Forge ↔ Gitea Sync

- [ ] **Mirror sync** — Implement repo mirroring from forge.lthn.ai (private) to git.lthn.ai (public). Strip sensitive data (research docs, credentials, internal refs).
- [ ] **Issue sync** — Selective issue sync (public issues only) from Forge to Gitea.
- [ ] **PR status sync** — Update Gitea mirror PRs when Forge PRs merge.

---

## Workflow

1. Virgil in core/go writes tasks here after research
2. This repo's session picks up tasks in phase order
3. Mark `[x]` when done, note commit hash
4. New discoveries → add tasks, flag in FINDINGS.md
