---
title: Development Guide
description: How to build, test, and contribute to go-scm.
---

# Development Guide

---

## Prerequisites

- **Go 1.26** or later
- **Git** (for `git/` package tests)
- **`gh` CLI** (for `collect/github.go` and rate limit checking -- not required for unit tests)
- SSH access to agent machines (for `agentci/` integration -- not required for unit tests)
- Access to `forge.lthn.ai/core/go` and sibling modules for the framework dependency

---

## Repository Layout

```
go-scm/
+-- go.mod              Module definition (dappco.re/go/core/scm)
+-- forge/              Forgejo API client + tests
+-- gitea/              Gitea API client + tests
+-- git/                Multi-repo git operations + tests
+-- agentci/            Clotho Protocol, agent config, security + tests
+-- jobrunner/          Poller, journal, types + tests
|   +-- forgejo/        Forgejo signal source + tests
|   +-- handlers/       Pipeline handlers + tests
+-- collect/            Data collection pipeline + tests
+-- manifest/           Application manifests, ed25519 signing + tests
+-- marketplace/        Module catalogue and installer + tests
+-- plugin/             CLI plugin system + tests
+-- repos/              Workspace registry, work config, git state + tests
+-- cmd/
|   +-- forge/          CLI commands for `core forge`
|   +-- gitea/          CLI commands for `core gitea`
|   +-- collect/        CLI commands for data collection
+-- docs/               Documentation
+-- .core/              Build and release configuration
```

---

## Building

This module is primarily a library. Build validation:

```bash
go build ./...         # Compile all packages
go vet ./...           # Static analysis
```

If using the `core` CLI with a `.core/build.yaml` present:

```bash
core go qa             # fmt + vet + lint + test
core go qa full        # + race, vuln, security
```

---

## Testing

### Run all tests

```bash
go test ./...
```

### Run a specific test or package

```bash
go test -v -run TestName ./forge/
go test -v ./agentci/...
```

### Run with race detector

```bash
go test -race ./...
```

Race detection is particularly important for `git/` (parallel status), `jobrunner/` (concurrent poller cycles), and `collect/` (concurrent rate limiter access).

### Coverage

```bash
go test -coverprofile=cover.out ./...
go tool cover -html=cover.out
```

---

## Local Dependencies

`go-scm` depends on several `forge.lthn.ai/core/*` modules. The recommended approach is to use a Go workspace file:

```go
// ~/Code/go.work
go 1.26

use (
    ./core/go
    ./core/go-io
    ./core/go-log
    ./core/config
    ./core/go-scm
    ./core/go-i18n
    ./core/go-crypt
)
```

With a workspace file in place, `replace` directives in `go.mod` are superseded and local edits across modules work seamlessly.

---

## Test Patterns

### forge/ and gitea/ -- httptest Mock Server

Both SDK wrappers require a live HTTP server because the Forgejo/Gitea SDKs make an HTTP GET to `/api/v1/version` during client construction. Use `net/http/httptest`:

```go
func setupServer(t *testing.T) (*forge.Client, *httptest.Server) {
    mux := http.NewServeMux()
    mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{"version": "1.20.0"})
    })
    // ... register other handlers ...
    srv := httptest.NewServer(mux)
    t.Cleanup(srv.Close)
    client, err := forge.New(srv.URL, "test-token")
    require.NoError(t, err)
    return client, srv
}
```

**Config isolation** -- always isolate the config file from the real machine:

```go
t.Setenv("HOME", t.TempDir())
t.Setenv("FORGE_TOKEN", "test-token")
t.Setenv("FORGE_URL", srv.URL)
```

**SDK route divergences** discovered during testing:

- `CreateOrgRepo` uses `/api/v1/org/{name}/repos` (singular `org`)
- `ListOrgRepos` uses `/api/v1/orgs/{name}/repos` (plural `orgs`)

### git/ -- Real Git Repositories

`git/` tests use real temporary git repos rather than mocks:

```go
func setupRepo(t *testing.T) string {
    dir := t.TempDir()
    run := func(args ...string) {
        cmd := exec.Command("git", args...)
        cmd.Dir = dir
        require.NoError(t, cmd.Run())
    }
    run("init")
    run("config", "user.email", "test@example.com")
    run("config", "user.name", "Test")
    // write a file, stage, commit...
    return dir
}
```

Testing `getAheadBehind` requires a bare remote and a clone:

```go
bare := t.TempDir()
exec.Command("git", "init", "--bare", bare).Run()
clone := t.TempDir()
exec.Command("git", "clone", bare, clone).Run()
```

### agentci/ -- Unit Tests Only

`agentci/` functions are pure (no I/O except SSH exec construction) and test without mocks:

```go
func TestSanitizePath_Good(t *testing.T) {
    result, err := agentci.SanitizePath("myrepo")
    require.NoError(t, err)
    assert.Equal(t, "myrepo", result)
}
```

### jobrunner/ -- Table-Driven Handler Tests

Handler tests use the `JobHandler` interface directly with a mock `forge.Client`:

```go
tests := []struct {
    name   string
    signal *jobrunner.PipelineSignal
    want   bool
}{
    {"merged PR", &jobrunner.PipelineSignal{PRState: "MERGED"}, true},
    {"open PR", &jobrunner.PipelineSignal{PRState: "OPEN"}, false},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        assert.Equal(t, tt.want, handler.Match(tt.signal))
    })
}
```

### collect/ -- Mixed Unit and HTTP Mock

Pure functions (state, rate limiter, events) test without I/O. HTTP-dependent collectors use mock servers:

```go
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(mockResponse)
}))
t.Cleanup(srv.Close)
```

The `SetHTTPClient` function allows injecting a custom HTTP client for tests.

### manifest/, marketplace/, plugin/ -- io.Medium Mocks

These packages use the `io.Medium` abstraction. Tests use `io.NewMockMedium()` to avoid filesystem interaction:

```go
m := io.NewMockMedium()
m.Write(".core/manifest.yaml", yamlContent)
manifest, err := manifest.Load(m, ".")
```

### repos/ -- io.Medium with Seed Data

```go
m := io.NewMockMedium()
m.Write("repos.yaml", registryYAML)
reg, err := repos.LoadRegistry(m, "repos.yaml")
```

---

## Test Naming Convention

Tests use the `_Good` / `_Bad` / `_Ugly` suffix pattern:

| Suffix | Meaning |
|--------|---------|
| `_Good` | Happy path -- expected success |
| `_Bad` | Expected error conditions |
| `_Ugly` | Panic, edge cases, malformed input |

---

## Coding Standards

### Language

Use **UK English** throughout: colour, organisation, centre, licence (noun), authorise, behaviour. Never American spellings.

### Go Style

- All parameters and return types must have explicit type declarations.
- Import groups: stdlib, then `forge.lthn.ai/...`, then third-party, each separated by a blank line.
- Use `testify/require` for fatal assertions, `testify/assert` for non-fatal. Prefer `require.NoError` when subsequent steps depend on the result.

### Error Wrapping

```go
// Correct -- using the log.E helper from core/go-log
return nil, log.E("forge.CreateRepo", "failed to create repository", err)

// Correct -- contextual prefix with package.Function
return nil, fmt.Errorf("forge.CreateRepo: marshal options: %w", err)

// Incorrect -- bare error with no context
return nil, fmt.Errorf("failed")
```

### Context Propagation

- `git/` and `collect/` propagate context correctly via `exec.CommandContext`.
- `forge/` and `gitea/` accept context at the wrapper boundary but cannot pass it to the SDK (SDK limitation).
- `agentci/` uses `SecureSSHCommand` for all SSH operations.

---

## Commit Conventions

Use conventional commits with a package scope:

```
feat(forge): add GetCombinedStatus wrapper
fix(jobrunner): prevent double-dispatch on in-progress issues
test(git): add ahead/behind with bare remote
docs(agentci): document Clotho dual-run flow
refactor(collect): extract common HTTP fetch into generic function
```

Valid types: `feat`, `fix`, `test`, `docs`, `refactor`, `chore`.

Every commit must include the co-author trailer:

```
Co-Authored-By: Virgil <virgil@lethean.io>
```

---

## Adding a New Package

1. Create the package directory under the module root.
2. Add `package <name>` with a doc comment describing the package's purpose.
3. Follow the existing `client.go` / `config.go` / `types.go` naming pattern where applicable.
4. Write tests from the start -- avoid creating packages without at least a skeleton test file.
5. Add the package to the architecture documentation.
6. Maintain import group ordering: stdlib, then `forge.lthn.ai/...`, then third-party.

## Adding a New Handler

1. Create `jobrunner/handlers/<name>.go` with a struct implementing `jobrunner.JobHandler`.
2. `Name()` returns a lowercase identifier (e.g. `"tick_parent"`).
3. `Match(signal)` should be narrow -- handlers are checked in registration order and the first match wins.
4. `Execute(ctx, signal)` must always return an `*ActionResult`, even on partial failure.
5. Add a corresponding `<name>_test.go` with at minimum one `_Good` and one `_Bad` test.
6. Register the handler in `Poller` configuration alongside existing handlers.

## Adding a New Collector

1. Create a new file in `collect/` (e.g. `collect/mynewsource.go`).
2. Implement the `Collector` interface (`Name()` and `Collect(ctx, cfg)`).
3. Use `cfg.Limiter.Wait(ctx, "source-name")` before each HTTP request.
4. Emit events via `cfg.Dispatcher` for progress reporting.
5. Write output via `cfg.Output` (the `io.Medium`), not directly to the filesystem.
6. Honour `cfg.DryRun` -- log what would be done without writing.
7. Return a `*Result` with accurate `Items`, `Errors`, `Skipped`, and `Files` counts.

---

## Remote and Push

The canonical remote is on Forgejo via SSH:

```bash
git push origin main
# Remote: ssh://git@forge.lthn.ai:2223/core/go-scm.git
```

HTTPS authentication to `forge.lthn.ai` is not configured -- always use SSH on port 2223.

---

## Licence

EUPL-1.2. The licence is compatible with GPL v2/v3 and AGPL v3.
