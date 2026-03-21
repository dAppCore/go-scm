# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

SCM integration and data collection library for the Lethean ecosystem (`dappco.re/go/core/scm`). Provides Forgejo/Gitea API clients, an AgentCI pipeline for automated PR lifecycle, pluggable data collectors, and workspace management (repos registry, manifests with ed25519 signing, marketplace, plugin system).

Virgil orchestrates tasks via Forgejo issues. Pick up tasks in issue order, mark complete, commit and push.

## Commands

```bash
go build ./...                     # Compile all packages
go test ./...                      # Run all tests
go test -v -run TestName ./forge/  # Single test in a package
go test -race ./...                # Race detector (important for git/, jobrunner/, collect/)
go vet ./...                       # Static analysis
golangci-lint run ./...            # Lint (govet, errcheck, staticcheck, unused, gocritic, gofmt)
```

If using the `core` CLI:

```bash
core go qa             # fmt + vet + lint + test
core go qa full        # + race, vuln, security
```

## Architecture

Five subsystems — see `docs/architecture.md` for full detail:

1. **Forge Clients** (`forge/`, `gitea/`) — thin wrappers over Forgejo/Gitea SDKs. Auth resolves: config file → env vars (`FORGE_URL`/`FORGE_TOKEN`) → flags → default localhost. All list methods handle pagination; `*Iter` variants yield lazily via `iter.Seq2`. SDK limitation: Forgejo SDK v2 does not accept `context.Context`.
2. **AgentCI Pipeline** (`agentci/`, `jobrunner/`, `jobrunner/forgejo/`, `jobrunner/handlers/`) — poll-dispatch-journal loop. `ForgejoSource` finds epic issues with checklists, emits `PipelineSignal` per unchecked child, handlers match and execute in registration order (first match wins). Includes Clotho Protocol dual-run verification.
3. **Data Collection** (`collect/`) — `Collector` interface with built-in scrapers (GitHub via `gh` CLI, BitcoinTalk, CoinGecko, IACR/arXiv). `Excavator` orchestrates sequential runs with rate limiting and incremental state.
4. **Workspace Management** (`repos/`, `manifest/`, `marketplace/`, `plugin/`) — `repos.yaml` multi-repo registry with topological sorting, `.core/manifest.yaml` with ed25519 signing, module marketplace with install/update/remove, CLI plugin system.
5. **Git Operations** (`git/`) — parallel multi-repo status, push/pull. DI-integrated `Service` with query/task handlers.

CLI commands live in `cmd/forge/`, `cmd/gitea/`, `cmd/collect/`.

**Note:** `agentci/` + `jobrunner/` are being extracted to `core/go-agent`, `git/` to `core/go-git`. `forge/` is still used by `core/go-agent` as a dependency.

## Local Dependencies

Use a Go workspace file (preferred over `replace` directives):

```go
// ~/Code/go.work — includes all forge.lthn.ai/core/* modules
```

Key dependencies: `core/go` (DI framework), `core/go-io` (filesystem abstraction), `core/go-log`, `core/config`, `core/go-i18n`, `core/go-crypt`.

## Test Patterns

Each subsystem has different test infrastructure — see `docs/development.md` for examples:

- **forge/, gitea/**: `httptest` mock server required (SDK does GET `/api/v1/version` on client construction). Always isolate config: `t.Setenv("HOME", t.TempDir())`.
- **manifest/, marketplace/, plugin/, repos/**: Use `io.NewMockMedium()` to avoid filesystem interaction.
- **git/**: Real temporary git repos via `t.TempDir()` + `exec.Command("git", ...)`.
- **agentci/**: Pure unit tests, no mocks needed.
- **jobrunner/**: Table-driven tests against `JobHandler` interface.
- **collect/**: Mixed — pure functions test directly, HTTP-dependent use mock servers. `SetHTTPClient` for injection.

## Coding Standards

- **UK English**: colour, organisation, centre, licence (noun), authorise, behaviour
- **Tests**: testify assert/require, table-driven preferred, `_Good`/`_Bad`/`_Ugly` suffix naming
- **Imports**: stdlib → `dappco.re/...` → third-party, each group separated by blank line
- **Errors**: `coreerr.E("package.Func", "context", err)` via `coreerr "dappco.re/go/core/log"` — no bare `fmt.Errorf` or `errors.New`
- **Conventional commits**: `feat(forge):`, `fix(gitea):`, `test(collect):`, `docs(agentci):`, `refactor(collect):`, `chore:`
- **Co-Author trailer**: `Co-Authored-By: Virgil <virgil@lethean.io>`
- **Licence**: EUPL-1.2

## Forge

- **Repo**: `dappco.re/go/core/scm`
- **Push via SSH**: `git push origin main` (remote: `ssh://git@forge.lthn.ai:2223/core/go-scm.git`)
- **CI**: Forgejo Actions — runs tests with race detector and coverage on push to main/dev and PRs to main
