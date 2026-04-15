# go-scm Project History

---

## Origin

`go-scm` was extracted from `forge.lthn.ai/core/go` on 19 February 2026 as part of a broader split of the monolithic `core/go` module into focused, independently versionable packages. The extraction commit is `3e883f69769c4be7401819b2ae3e2122324a666c`.

At extraction, the module contained six packages totalling approximately 9,000 LOC with the following test coverage state:

| Package | LOC (approx) | Tests | Coverage |
|---------|-------------|-------|---------|
| forge/ | 900 | 0 | 0% |
| gitea/ | 500 | 0 | 0% |
| git/ | 400 | 0 | 0% |
| agentci/ | 300 | 1 | partial (~56%) |
| jobrunner/ | 1,500 | several | partial |
| collect/ | 5,000 | 8 | unknown (~57%) |

The delegation model for development uses a fleet of AI agents (Virgil for architecture/orchestration, Charon for implementation). Initial task queues were written to `TODO.md` and `FINDINGS.md` at the time of extraction (`bedcb4e65292fb3e2436a7254eca8012ccd5dd7b`).

---

## Phase 1: Test Coverage

Commit: `9db37c6fb3628a94727dae61c2fb9704f1924eb4`
Additional coverage pushes: `b4e3d055`, `2505aff4`, `f50fc405`, `b3e3ef2e`, `8d1d7fce`

The critical gap at extraction was zero tests in the three foundational packages. Phase 1 addressed this by bringing all packages to meaningful coverage thresholds.

**forge/ — 0% to 91.2%** (8 test files)

All SDK wrapper functions were tested via `net/http/httptest` mock servers. Key findings during test construction:

- The Forgejo SDK v2 makes an HTTP GET to `/api/v1/version` during `NewClient()`. Every test requires a mock server — there is no zero-cost unit test instantiation.
- SDK route patterns differ from the public API documentation in at least one case: `CreateOrgRepo` uses the singular `/api/v1/org/{name}/repos` while `ListOrgRepos` uses the plural `/api/v1/orgs/{name}/repos`.
- `SetPRDraft` is implemented as a raw HTTP PATCH (not an SDK method) and was testable via the mock server by registering the endpoint directly.

**gitea/ — 0% to 89.2%** (5 test files)

Mirror of the forge/ approach. Additional finding: `CreateMirrorRepo` with `Service: GitServiceGithub` requires a non-empty `AuthToken` or the SDK rejects the request before sending to the server.

**git/ — 0% to 96.7%** (4 test files across two rounds)

Required real temporary git repositories (`t.TempDir()` + `git init` + `git commit`). The `getAheadBehind` function required a bare remote plus a clone to test correctly. Context cancellation was verified by constructing a pre-cancelled context. `IsNonFastForward` was verified against the specific error string patterns returned by git.

The `git.Service` struct and its framework integration methods (`OnStartup`, `handleQuery`, `handleTask`) depend on `framework.Core` and were not unit-tested; they are covered by integration context.

**agentci/ — 56% to 94.5%**

Tests added for `Spinner` methods (`DeterminePlan`, `GetVerifierModel`, `FindByForgejoUser`, `Weave`) and all security functions (`SanitizePath`, `EscapeShellArg`, `SecureSSHCommand`, `MaskToken`). `SanitizePath` now rejects traversal and separator characters instead of normalising them away, which matches the stricter agent dispatch contract.

**collect/ — 57.3% to 83.0%** (additional test files)

HTTP-dependent `Collect` methods were tested using `httptest` mock servers for BitcoinTalk, CoinGecko, IACR, and arXiv. The gap at extraction was that all HTTP-dependent paths were untested; these were filled in Phase 1 and supplemented in later coverage pushes.

**jobrunner/ — 86.4% to 95.0%** (forgejo/ source)

The forgejo signal source was already above the 60% floor at extraction. Phase 1 audited but did not change it. Phase 3 added 20 new tests covering signal parsing, epic body edge cases, and `Poll` error paths.

---

## Phase 2: Hardening

Commit: `3ba8fbb8fba5785fd64b08d5e3673cd502f4ef9f`

**Error wrapping audit**

All 15 bare `fmt.Errorf` calls (those without a `package.Function:` prefix) were converted to the project standard. Affected files:

| File | Errors fixed |
|------|-------------|
| `jobrunner/journal.go` | 8 |
| `agentci/config.go` | 3 |
| `agentci/security.go` | 2 |
| `jobrunner/handlers/dispatch.go` | 1 |
| `forge/labels.go` | 1 |

Existing `log.E()` and `core.E()` call sites already followed the pattern and required no changes.

**Context propagation audit**

The audit confirmed a fundamental SDK limitation: neither the Forgejo SDK v2 nor the Gitea SDK accept `context.Context` on any API method. Adding context parameters to the 66 forge/ and gitea/ wrapper functions would be ceremony without real propagation. The wrappers accept contexts at their boundaries but cannot pass them through.

Where context propagation was added:

- `agentci/security.go` — `SecureSSHCommandContext(ctx, host, cmd)` added; `SecureSSHCommand` deprecated and delegates to it.
- `jobrunner/handlers/dispatch.go` — `secureTransfer`, `runRemote`, and `ticketExists` updated to pass their existing `ctx` argument through to `SecureSSHCommandContext`.
- `collect/ratelimit.go` — `CheckGitHubRateLimitCtx(ctx)` added alongside the deprecated `CheckGitHubRateLimit`.

Both `git/` and the existing `collect/` HTTP collectors already propagated context correctly (`exec.CommandContext` and `http.NewRequestWithContext` respectively).

**Rate limiting review**

The existing `collect/ratelimit.go` implementation was reviewed and found complete. It handles all current edge cases: context cancellation during `Wait`, unknown sources (default 500 ms), concurrent access (mutex-protected), and adaptive throttling (auto-pause at 75% GitHub API usage). No changes were required to the rate limiter logic.

---

## Phase 3: AgentCI Pipeline

Commit: `0fe2978b9d1c300a44ac8cf4cbb4ad51e101f6f1`

Phase 3 added integration-level test coverage across the three components that form the dispatch pipeline: the Forgejo signal source, the journal, and the handlers.

**Forgejo signal source** (`source_phase3_test.go` — 20 tests)

Covered signal parsing with empty bodies, mixed content, and large issue numbers; `findLinkedPR` edge cases (nil head, empty SHA, matching by issue reference); `Poll` with combined status failures; unassigned child issues; child fetch failures; multiple epics in one repo; mixed label sets; and the `Report` method format.

**Journal replay** (`journal_replay_test.go` — 8 tests)

Write-then-read-back round-trip verified JSON marshalling and file partitioning. Filter tests covered action name, repo full name, time range, and combined filters. Concurrent write safety was verified with 20 goroutines writing simultaneously. Empty journal reads were verified to return zero entries without error.

**Handler integration** (`integration_test.go` — 14 tests)

Full signal-to-result flow tested for all five handlers via a mock Forgejo server and real Poller. Tests verified: `tick_parent` epic checkbox update and child issue closure; `enable_auto_merge` on passing/mergeable PRs; `publish_draft` on resolved threads; `send_fix_command` on failing checks; `completion` handler recording. Handler priority matching (first match wins) was also verified.

---

## Known Limitations

**SDK context support**

The Forgejo SDK v2 and Gitea SDK do not accept `context.Context`. All Forgejo/Gitea API calls are blocking with no cancellation path. When the SDK is updated to support context (v3 or later), a follow-up task should thread `ctx` through all forge/ and gitea/ wrapper signatures.

**Clotho Weave — thresholded token overlap**

`Spinner.Weave(ctx, primary, signed)` now uses the configured `validation_threshold` to decide convergence from a deterministic token-overlap score. This is still a lightweight approximation rather than full semantic diffing, but it now honours the config knob already exposed by `ClothoConfig`.

**collect/ HTTP collectors — no retry**

None of the HTTP-dependent collectors (`bitcointalk.go`, `github.go`, `market.go`, `papers.go`) implement retry on transient failures. A single HTTP error causes the collector to return an error and increment the `Errors` count in the result. The `Excavator` continues to the next collector. For long-running collection runs, transient network errors cause silent data gaps.

**Journal replay**

The journal now exposes `Journal.Query(...)` for replay and filtering over the JSONL archive. It supports repo, action, and time-range filters while preserving the date-partitioned storage layout used by `Append(...)`.

**git.Service framework integration**

`git.Service.OnStartup`, `handleQuery`, and `handleTask` are not covered by unit tests because they depend on `framework.Core`. These code paths are exercised in integration contexts only.

**ForgejoSource — no webhook support**

`ForgejoSource` polls on a configurable interval. It does not support webhook push delivery for lower-latency signalling. Webhook delivery would require a separate HTTP server and is tracked as a future phase.

---

## Phase 4: Forge-to-Gitea Sync (Planned)

The following tasks have been identified but not yet implemented:

- **Mirror sync** — automated repository mirroring from `forge.lthn.ai` (private, source of truth) to `git.lthn.ai` (public mirror). Requires stripping sensitive data (research documents, internal references, credentials).
- **Issue sync** — selective issue sync from Forge to Gitea (public issues only).
- **PR status sync** — update Gitea mirror PRs when corresponding Forge PRs are merged.

These tasks were deferred from Phase 3 to allow the dispatch pipeline to stabilise first.
