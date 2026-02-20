# FINDINGS.md — go-scm Research & Discovery

## 2026-02-20: Initial Assessment (Virgil)

### Origin

Extracted from `forge.lthn.ai/core/go` on 19 Feb 2026 as part of the go-ai split. Contains all SCM integration, CI dispatch, and data collection code.

### Package Inventory

| Package | Files | LOC (approx) | Tests | Coverage |
|---------|-------|--------------|-------|----------|
| forge/ | 9 | ~900 | 0 | 0% |
| gitea/ | 5 | ~500 | 0 | 0% |
| git/ | 2 | ~400 | 0 | 0% |
| agentci/ | 4 | ~300 | 1 | partial |
| jobrunner/ | 4 + handlers/ | ~1,500 | several | partial |
| collect/ | 12 | ~5,000 | 8 | unknown |

### Dependencies

- `codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2` — Forgejo API (v2)
- `code.gitea.io/sdk/gitea` — Gitea API
- `forge.lthn.ai/core/go` — Framework (log, config, viper)
- `github.com/stretchr/testify` — Testing

### Key Observations

1. **forge/ and gitea/ are structurally identical** — Same pattern: client.go, config.go, repos.go, issues.go, meta.go. Could share an interface.
2. **Zero tests in forge/, gitea/, git/** — Most critical gap. These are the foundational packages.
3. **collect/ has the most tests** — 8 test files covering all collectors. Coverage unknown.
4. **jobrunner/handlers/ has test files** — dispatch_test.go, enable_auto_merge_test.go, publish_draft_test.go, resolve_threads_test.go, send_fix_command_test.go, tick_parent_test.go. Quality unknown.
5. **agentci/ Clotho Protocol** — Dual-run verification for critical repos. Currently basic (string match on repo name). Needs more sophisticated policy engine.

### Auth Resolution

Both forge/ and gitea/ resolve auth from:
1. `~/.core/config.yaml` (forge.token/forge.url or gitea.token/gitea.url)
2. Environment variables (FORGE_TOKEN/FORGE_URL or GITEA_TOKEN/GITEA_URL)
3. CLI flag overrides (highest priority)

This is handled via core/go's viper integration.

### Infrastructure Context

- **Forge** (`forge.lthn.ai`) — Production Forgejo instance on de2. Full IP/intel/research.
- **Gitea** (`git.lthn.ai`) — Public mirror with reduced data. Breach-safe.
- **Split policy**: Forge = source of truth, Gitea = public-facing mirror with sensitive data stripped.

---

## 2026-02-20: Phase 1 Test Coverage (Charon)

### Coverage Results After Phase 1

| Package | Before | After | Target | Notes |
|---------|--------|-------|--------|-------|
| forge/ | 0% | 91.2% | 50% | Exceeded target |
| gitea/ | 0% | 89.2% | 50% | Exceeded target |
| git/ | 0% | 79.5% | 80% | Remaining ~0.5% is framework service integration |
| agentci/ | 56% | 94.5% | 60% | Added clotho.go + security.go tests |
| collect/ | 57.3% | 57.3% | — | Audited, no changes (HTTP-dependent collectors) |
| jobrunner/ | 86.4% | 86.4% | — | Already above 60%, no changes needed |
| jobrunner/forgejo | 73.3% | 73.3% | — | Already above 60% |
| jobrunner/handlers | 61.6% | 61.6% | — | Already above 60% |

### SDK Testability Findings

1. **Forgejo SDK (`forgejo/v2`) validation on client creation**: `NewClient()` makes an HTTP GET to `/api/v1/version` during construction. All tests that create a client need an `httptest.Server` with at least a `/api/v1/version` handler. This means no zero-cost unit test instantiation — every forge/ test needs a mock server.

2. **Forgejo SDK route patterns differ from the public API docs**: The SDK uses `/api/v1/org/{name}/repos` (singular `org`) for `CreateOrgRepo`, but `/api/v1/orgs/{name}/repos` (plural `orgs`) for `ListOrgRepos`. This was discovered during test construction and would bite anyone writing integration tests.

3. **Gitea SDK mirror validation**: `CreateMirror` with `Service: GitServiceGithub` requires a non-empty `AuthToken`. Without it, the SDK rejects the request locally before sending to the server. Tests must always provide a token.

4. **forge/ and gitea/ are testable with httptest**: Despite being SDK wrappers, coverage above 89% was achievable for both packages using `net/http/httptest`. The mock server approach covers: client creation, error handling, state mapping logic (issue states, PR merge styles), pagination termination, config resolution, and raw HTTP endpoints (SetPRDraft).

5. **git/ requires real git repos for integration testing**: `t.TempDir()` + `git init` + `git commit` provides clean, isolated test environments. The `getAheadBehind` function requires a bare remote + clone setup to test properly.

6. **git/service.go framework dependency**: `NewService`, `OnStartup`, `handleQuery`, and `handleTask` depend on `framework.Core` from `core/go`. These are better tested in integration tests (Phase 3). The `DirtyRepos()`, `AheadRepos()`, and `Status()` helper methods are tested by directly setting `lastStatus`.

7. **Thin SDK wrappers**: Most forge/ and gitea/ functions are 3-5 line SDK pass-throughs (call SDK, check error, return). Despite being thin, they were all testable via mock server because the SDK sends real HTTP requests. No function was skipped as "untestable".

8. **agentci/security.go `SanitizePath`**: `filepath.Base("../secret")` returns `"secret"`, which passes validation. This means `SanitizePath` protects against path traversal by stripping the directory component, not by rejecting the input. This is correct behaviour — documented in test.

### Config Resolution Verified

Both forge/ and gitea/ follow the same priority order:
1. Config file (`~/.core/config.yaml`) — lowest priority
2. Environment variables (`FORGE_URL`/`FORGE_TOKEN` or `GITEA_URL`/`GITEA_TOKEN`)
3. Flag overrides — highest priority
4. Default URL when nothing configured (`http://localhost:4000` for forge, `https://gitea.snider.dev` for gitea)

Tests must use `t.Setenv("HOME", t.TempDir())` to isolate from the real config file on the development machine.

---

## 2026-02-20: Phase 2 Hardening (Virgil)

### Error Wrapping Audit

All `fmt.Errorf` calls now follow the `"package.Func: context: %w"` pattern (for wrapped errors) or `"package.Func: message"` pattern (for sentinel/validation errors). Files fixed:

| File | Bare errors | Pattern applied |
|------|------------|-----------------|
| `jobrunner/journal.go` | 8 | `journal.NewJournal:`, `journal.sanitizePathComponent:`, `journal.Append:` |
| `agentci/config.go` | 3 | `agentci.LoadAgents:`, `agentci.RemoveAgent:` |
| `agentci/security.go` | 2 | `agentci.SanitizePath:` |
| `jobrunner/handlers/dispatch.go` | 1 | `handlers.Dispatch.Execute:` |
| `forge/labels.go` | 1 | `forge.GetLabelByName:` |

Existing `log.E()` and `core.E()` calls already followed the pattern — no changes needed for those.

### Context Propagation Audit

**SDK limitation (verified):** The Forgejo SDK v2 (`codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2`) does NOT accept `context.Context` parameters on any API method. All SDK calls are synchronous and blocking. Adding `ctx` to the 66 forge/ and gitea/ wrapper signatures would be ceremony without real propagation — the context would be accepted but never passed to the SDK. The same applies to the Gitea SDK (`code.gitea.io/sdk/gitea`).

**Where context IS propagated:**

1. **git/** — All git operations use `exec.CommandContext(ctx, ...)` via `gitCommand()` and `gitInteractive()`. Context cancellation works correctly for all git operations (Status, Push, Pull, PushMultiple). Verified in existing tests.

2. **collect/** — All HTTP requests use `http.NewRequestWithContext(ctx, ...)` for BitcoinTalk, CoinGecko, IACR, and arXiv collectors. The `RateLimiter.Wait()` method respects context cancellation. The `Excavator.Run()` checks `ctx.Err()` between collectors. Added `CheckGitHubRateLimitCtx(ctx)` for context-aware rate limit checking via `exec.CommandContext`.

3. **agentci/security.go** — Added `SecureSSHCommandContext(ctx, host, cmd)` which uses `exec.CommandContext` for cancellable SSH operations. The original `SecureSSHCommand` is preserved (deprecated) and delegates to the context-aware version.

4. **jobrunner/handlers/dispatch.go** — Updated `secureTransfer()`, `runRemote()`, and `ticketExists()` to pass their existing `ctx` parameter through to `SecureSSHCommandContext()`. Previously these methods accepted context but did not propagate it.

**Recommendation:** When the Forgejo SDK v3 adds context support, a follow-up task should thread `ctx` through all forge/ and gitea/ wrapper signatures.

### Rate Limiting Review

**Current implementation (`collect/ratelimit.go`):**

- Per-source delay tracking with configurable durations per source
- Default delays: GitHub 500ms, BitcoinTalk 2s, CoinGecko 1.5s, IACR 1s, arXiv 1s
- Unknown sources default to 500ms
- `Wait(ctx, source)` blocks until the rate limit allows, respects context cancellation
- `CheckGitHubRateLimit()` uses `gh api rate_limit` CLI to check GitHub API usage
- Auto-pauses at 75% usage by increasing GitHub delay to 5s

**Assessment:**

1. The rate limiter handles all current edge cases: context cancellation during wait, unknown sources, concurrent access (mutex-protected), and adaptive throttling.
2. GitHub rate limiting relies on the `gh` CLI rather than parsing HTTP response headers. This is appropriate because GitHub collection also uses the `gh` CLI (not direct HTTP). The `gh` CLI handles its own authentication and rate limit headers internally.
3. Forgejo/Gitea API calls go through their respective SDKs, which do not expose rate limit response headers. The SDKs handle HTTP internally. Adding header parsing would require forking or wrapping the SDK HTTP transport — not worth the complexity for the current usage pattern.
4. The `CheckGitHubRateLimitCtx` variant was added for context-aware cancellation of the `gh api` subprocess.
