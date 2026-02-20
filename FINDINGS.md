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
