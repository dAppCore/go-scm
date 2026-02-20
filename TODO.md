# TODO.md — go-scm Task Queue

Dispatched from core/go orchestration. Pick up tasks in order.

---

## Phase 1: Test Coverage (Critical)

forge/, gitea/, and git/ have **zero tests**. This is the top priority.

- [ ] **forge/ unit tests** — Test `New()` client creation, `GetCurrentUser()`, error handling. Mock the Forgejo SDK client. Cover: `repos.go` (create, list, mirror), `issues.go` (create, list, assign), `prs.go` (create, list, merge), `labels.go`, `webhooks.go`, `orgs.go`. Target: 70% coverage.
- [ ] **gitea/ unit tests** — Test `New()` client creation, repo/issue operations. Mock the Gitea SDK client. Cover: `repos.go`, `issues.go`, `meta.go`. Target: 70% coverage.
- [ ] **git/ unit tests** — Test `RepoStatus` methods (`IsDirty`, `HasUnpushed`, `HasUnpulled`). Test status parsing with mock git output. Test bulk operations with temp repos. Cover: `git.go`, `service.go`. Target: 80% coverage.
- [ ] **jobrunner handler tests** — handlers/ has test files but verify coverage. Add table-driven tests for `dispatch.go`, `completion.go`, `enable_auto_merge.go`. Test `PipelineSignal` state transitions.
- [ ] **collect/ test audit** — collect/ has test files for each collector. Run `go test -cover ./collect/...` and identify gaps below 60%.

## Phase 2: Hardening

- [ ] **Config resolution audit** — forge/ and gitea/ both resolve auth from `~/.core/config.yaml` → env vars → flags. Ensure consistent priority order. Document in FINDINGS.md.
- [ ] **Error wrapping** — Ensure all errors use `fmt.Errorf("package.Func: ...: %w", err)` or `log.E()` consistently. Some files may use bare `fmt.Errorf` without wrapping.
- [ ] **Context propagation** — Verify all Forgejo/Gitea API calls pass `context.Context` for cancellation. Add context to any blocking operations missing it.
- [ ] **Rate limiting** — collect/ has its own `ratelimit.go`. Verify it handles API rate limit headers from GitHub, Forgejo, Gitea.

## Phase 3: AgentCI Pipeline

- [ ] **Clotho dual-run validation** — `DeterminePlan()` logic is simple (check strategy + agent config + repo name). Add tests for all code paths: standard mode, dual-run by agent config, dual-run by critical repo.
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
