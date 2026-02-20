# CLAUDE.md — go-scm Agent Instructions

You are a dedicated domain expert for `forge.lthn.ai/core/go-scm`. Virgil orchestrates tasks via Forgejo issues. Pick up tasks in issue order, mark complete, commit and push.

## What This Package Does

SCM integration, AgentCI dispatch, and data collection for the Lethean ecosystem. See `docs/architecture.md` for the full package breakdown.

## Commands

```bash
go test ./...                    # Run all tests
go test -v -run TestName ./...   # Single test
go test -race ./...              # Race detector
go vet ./...                     # Static analysis
```

## Local Dependencies

Resolved via `replace` in go.mod or preferably via `go.work`:

| Module | Local Path |
|--------|-----------|
| `forge.lthn.ai/core/go` | `../go` |

## Key Types

```go
// forge/client.go
type Client struct { api *forgejo.Client; url, token string }

// git/git.go
type RepoStatus struct {
    Name, Path, Branch string
    Modified, Untracked, Staged, Ahead, Behind int
    Error error
}

// jobrunner/types.go
type PipelineSignal struct {
    EpicNumber, ChildNumber, PRNumber int
    RepoOwner, RepoName string
    PRState, Mergeable, CheckStatus string
    NeedsCoding bool
    IssueTitle, IssueBody string
}

// agentci/clotho.go
type RunMode string // "standard" or "dual"
type Spinner struct { Config ClothoConfig; Agents map[string]AgentConfig }
```

## Coding Standards

- **UK English**: colour, organisation, centre
- **Tests**: testify assert/require, table-driven preferred, `_Good`/`_Bad`/`_Ugly` naming
- **Conventional commits**: `feat(forge):`, `fix(jobrunner):`, `test(collect):`
- **Co-Author**: `Co-Authored-By: Virgil <virgil@lethean.io>`
- **Licence**: EUPL-1.2
- **Imports**: stdlib → forge.lthn.ai → third-party, each group separated by blank line
- **Error pattern**: `"package.Func: context: %w"` — no bare `fmt.Errorf`

See `docs/development.md` for full standards and test patterns.

## Forge

- **Repo**: `forge.lthn.ai/core/go-scm`
- **Push via SSH**: `git push origin main` (remote: `ssh://git@forge.lthn.ai:2223/core/go-scm.git`)
