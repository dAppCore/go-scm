# go-scm Development Guide

---

## Prerequisites

- Go 1.25 or later
- Git (for `git/` package integration tests)
- `gh` CLI (for `collect/github.go` and rate limit checking — not required for unit tests)
- SSH access to agent machines (for `agentci/` integration — not required for unit tests)
- Access to `forge.lthn.ai/core/go` for the framework dependency

---

## Repository Layout

```
go-scm/
├── go.mod              Module definition (forge.lthn.ai/core/go-scm)
├── forge/              Forgejo API client + tests
├── gitea/              Gitea API client + tests
├── git/                Multi-repo git operations + tests
├── agentci/            Clotho Protocol + security + tests
├── jobrunner/          Poller, journal, types + tests
│   ├── forgejo/        Forgejo signal source + tests
│   └── handlers/       Pipeline handlers + tests
├── collect/            Data collection pipeline + tests
└── docs/               Architecture, development, history
```

---

## Building

This module has no binary targets — it is a library. Build validation is via tests and `go vet`:

```bash
go build ./...         # Compile all packages
go vet ./...           # Static analysis
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

Race detection is particularly important for `git/` (parallel status), `jobrunner/` (concurrent poller cycles), and `collect/` (concurrent rate limiter).

### Coverage

```bash
go test -coverprofile=cover.out ./...
go tool cover -html=cover.out
```

Current coverage targets (as of Phase 3 completion):

| Package | Coverage |
|---------|---------|
| forge/ | 91.2% |
| gitea/ | 89.2% |
| git/ | 96.7% |
| agentci/ | 94.5% |
| jobrunner/ | 86.4% |
| jobrunner/forgejo/ | 95.0% |
| jobrunner/handlers/ | 83.8% |
| collect/ | 83.0% |

---

## Local Dependencies

`go-scm` depends on `forge.lthn.ai/core/go` for the framework, logging, config, and IO abstractions. The `go.mod` file includes a `replace` directive pointing to a sibling directory:

```
replace forge.lthn.ai/core/go => ../go
```

**Preferred approach:** use a `go.work` file in your workspace root to avoid editing `go.mod` for local development:

```
// go.work
go 1.25

use (
    ./go
    ./go-scm
)
```

With a workspace file in place, the `replace` directive in `go.mod` is superseded and can be left as a fallback.

---

## Test Patterns

### forge/ and gitea/ — httptest mock server

Both SDK wrappers require a live HTTP server because the Forgejo SDK makes an HTTP GET to `/api/v1/version` during client construction. Use `net/http/httptest` for all tests:

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

SDK route patterns sometimes differ from the public API documentation. Notable divergences discovered during test construction:

- `CreateOrgRepo` uses `/api/v1/org/{name}/repos` (singular `org`)
- `ListOrgRepos` uses `/api/v1/orgs/{name}/repos` (plural `orgs`)

**Config isolation:** always isolate the config file from the real development machine during tests:

```go
t.Setenv("HOME", t.TempDir())
t.Setenv("FORGE_TOKEN", "test-token")
t.Setenv("FORGE_URL", srv.URL)
```

**Gitea mirror validation:** `CreateMirrorRepo` with `Service: GitServiceGithub` requires a non-empty `AuthToken`. The SDK rejects the request locally before sending to the server if the token is absent.

### git/ — real git repositories

`git/` tests use real temporary git repositories rather than mocks. The standard setup pattern:

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

**Service layer:** `git.Service`, `OnStartup`, `handleQuery`, and `handleTask` depend on `framework.Core`. Test these indirectly through `DirtyRepos()`, `AheadRepos()`, and `Status()` by setting `lastStatus` directly, or via integration tests.

### agentci/ — unit tests only

`agentci/` functions are pure (no I/O except SSH exec construction) and test without mocks:

```go
func TestSanitizePath_Good(t *testing.T) {
    result, err := agentci.SanitizePath("myrepo")
    require.NoError(t, err)
    assert.Equal(t, "myrepo", result)
}
```

`SanitizePath("../secret")` returns `"secret"` — it strips the directory component via `filepath.Base` rather than rejecting the input. This is the documented, correct behaviour.

### jobrunner/ — table-driven handler tests

Handler tests use the `JobHandler` interface directly with a mock `forge.Client` constructed via `httptest`. The preferred pattern is table-driven:

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

Integration tests in `handlers/integration_test.go` test the full signal-to-result flow: construct a Poller, register sources and handlers, call `RunOnce`, verify journal entries and Forgejo API calls on the mock server.

### collect/ — mixed unit and HTTP mock

Pure functions (state management, rate limiter logic, event dispatch) test without I/O. HTTP-dependent collectors (`Collect` methods for BitcoinTalk, GitHub, IACR, arXiv) require mock HTTP servers:

```go
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(mockResponse)
}))
t.Cleanup(srv.Close)
```

---

## Coding Standards

### Language

Use UK English throughout: colour, organisation, centre, licence (noun), authorise, behaviour. Never American spellings.

### Go style

- All parameters and return types must have explicit type declarations.
- Error strings follow `"package.Function: context: %w"` for wrapped errors, `"package.Function: message"` for sentinel errors. No bare `fmt.Errorf("something failed")` strings.
- Import groups: stdlib, then `forge.lthn.ai/...`, then third-party, each separated by a blank line.
- Use `testify/require` for fatal assertions, `testify/assert` for non-fatal. Prefer `require.NoError` over `assert.NoError` when subsequent test steps depend on the result.
- Test naming convention: `_Good` (happy path), `_Bad` (expected error), `_Ugly` (panic/edge case).

### Error wrapping

```go
// Correct — contextual prefix with package.Function
return nil, fmt.Errorf("forge.CreateRepo: marshal options: %w", err)

// Correct — using the log.E helper from core/go
return nil, log.E("forge.CreateRepo", "failed to create repository", err)

// Incorrect — bare error with no context
return nil, fmt.Errorf("failed")
```

### Context propagation

- `git/` and `collect/` propagate context correctly via `exec.CommandContext`.
- `forge/` and `gitea/` accept context at the wrapper boundary but cannot pass it to the SDK (SDK limitation, see architecture.md).
- `agentci/` uses `SecureSSHCommandContext` for all SSH operations — never use `SecureSSHCommand` (deprecated).

---

## Commit Conventions

Use conventional commits with a package scope:

```
feat(forge): add GetCombinedStatus wrapper
fix(jobrunner): prevent double-dispatch on in-progress issues
test(git): add ahead/behind with bare remote
docs(agentci): document Clotho dual-run flow
```

Valid types: `feat`, `fix`, `test`, `docs`, `refactor`, `chore`.

Every commit must include the co-author trailer:

```
Co-Authored-By: Virgil <virgil@lethean.io>
```

---

## Licence

EUPL-1.2. All source files must carry the EUPL-1.2 licence header if one is added to the project. The licence is compatible with GPL v2/v3 and AGPL v3.

---

## Remote and Push

The canonical remote is on Forgejo via SSH:

```bash
git push origin main
# Remote: ssh://git@forge.lthn.ai:2223/core/go-scm.git
```

HTTPS authentication to `forge.lthn.ai` is not configured — always use SSH. The SSH port is 2223.

---

## Adding a New Package

1. Create the package directory under the module root.
2. Add `package <name>` with a doc comment describing the package's purpose.
3. Follow the existing `client.go` / `config.go` / `types.go` naming pattern where applicable.
4. Write tests from the start — avoid creating packages without at least a skeleton test file.
5. Add the package to the dependency graph in `docs/architecture.md`.
6. Import groups must be maintained: stdlib, then `forge.lthn.ai/...`, then third-party.

## Adding a New Handler

1. Create `jobrunner/handlers/<name>.go` with a struct implementing `jobrunner.JobHandler`.
2. `Name()` returns a lowercase identifier (e.g. `"tick_parent"`).
3. `Match(signal)` should be narrow — handlers are checked in registration order and the first match wins.
4. `Execute(ctx, signal)` must always return an `*ActionResult`, even on partial failure.
5. Add a corresponding `<name>_test.go` with at minimum one `_Good` and one `_Bad` test.
6. Register the handler in `Poller` configuration alongside existing handlers.
