# CLAUDE.md — go-scm Domain Expert Guide

You are a dedicated domain expert for `forge.lthn.ai/core/go-scm`. Virgil (in core/go) orchestrates your work via TODO.md. Pick up tasks in phase order, mark `[x]` when done, commit and push.

## What This Package Does

SCM integration, AgentCI dispatch, and data collection for the Lethean ecosystem. ~9K LOC across 4 sub-packages:

- **forge/** — Forgejo API client (repos, issues, PRs, labels, webhooks, orgs)
- **gitea/** — Gitea API client (repos, issues, meta) for public mirror at `git.lthn.ai`
- **git/** — Multi-repo git operations (status, commit, push, pull)
- **agentci/** — Clotho Protocol orchestrator for dual-run agent verification
- **jobrunner/** — PR automation pipeline (Forgejo webhook signals → handler dispatch)
- **collect/** — Data collection (BitcoinTalk, GitHub, market, papers, events)

## Architecture

```
forge/           Forgejo SDK wrapper (codeberg.org/mvdkleijn/forgejo-sdk)
  ├── client.go      Config-based auth, SDK wrapper
  ├── repos.go       Create, list, mirror repos
  ├── issues.go      Create, list, assign issues
  ├── prs.go         Create, list, merge PRs
  ├── labels.go      Label management
  ├── webhooks.go    Webhook CRUD
  ├── orgs.go        Organisation management
  ├── meta.go        Instance metadata
  └── config.go      Auth resolution (~/.core/config.yaml, env, flags)

gitea/           Gitea SDK wrapper (code.gitea.io/sdk/gitea)
  ├── client.go      Config-based auth
  ├── repos.go       Repo operations
  ├── issues.go      Issue operations
  ├── meta.go        Instance info
  └── config.go      Auth config

git/             Multi-repo git operations
  ├── git.go         RepoStatus, StatusOptions, exec wrappers
  └── service.go     Bulk status, commit, push, pull across repos

agentci/         Clotho Protocol — dual-run agent verification
  ├── clotho.go      Spinner orchestrator (standard vs dual-run mode)
  ├── config.go      ClothoConfig, AgentConfig
  ├── security.go    Security policy enforcement
  └── config_test.go

jobrunner/       PR automation pipeline
  ├── types.go       PipelineSignal, ActionResult, JobSource, Handler interfaces
  ├── poller.go      Polling loop with tick interval
  ├── journal.go     Execution journal (JSONL audit trail)
  ├── forgejo/
  │   └── source.go  Forgejo webhook signal source
  └── handlers/
      ├── dispatch.go           Agent task dispatch
      ├── completion.go         Handle agent completion signals
      ├── enable_auto_merge.go  Auto-merge when checks pass
      ├── publish_draft.go      Publish draft PRs
      ├── resolve_threads.go    Auto-resolve review threads
      ├── send_fix_command.go   Send fix commands to agents
      └── tick_parent.go        Update parent epic progress

collect/         Data collection pipeline
  ├── collect.go       Collector interface, pipeline orchestrator
  ├── bitcointalk.go   Forum scraping
  ├── github.go        GitHub API data collection
  ├── market.go        Market data
  ├── papers.go        Research paper collection
  ├── events.go        Event tracking
  ├── excavate.go      Deep data extraction
  ├── process.go       Processing pipeline
  ├── ratelimit.go     Per-source rate limiting
  └── state.go         Collection state persistence
```

## Commands

```bash
go test ./...                    # Run all tests
go test -v -run TestName ./...   # Single test
go test -race ./...              # Race detector
go vet ./...                     # Static analysis
```

## Local Dependencies

Resolved via `replace` in go.mod or preferably via `go.work`:

| Module | Local Path | Notes |
|--------|-----------|-------|
| `forge.lthn.ai/core/go` | varies | Framework (log, config, process) |

**Recommended**: Use a `go.work` file in your workspace root to resolve local modules rather than editing go.mod replace directives.

## Key Types

```go
// forge/client.go
type Client struct {
    api   *forgejo.Client
    url   string
    token string
}

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
    // ... dispatch context
}

type ActionResult struct {
    Action string
    Success bool
    Error string
    Duration time.Duration
    Cycle int
}

// agentci/clotho.go
type RunMode string // "standard" or "dual"
type Spinner struct {
    Config ClothoConfig
    Agents map[string]AgentConfig
}
```

## Coding Standards

- **UK English**: colour, organisation, centre
- **Tests**: testify assert/require, table-driven preferred
- **Conventional commits**: `feat(forge):`, `fix(jobrunner):`, `test(collect):`
- **Co-Author**: `Co-Authored-By: Virgil <virgil@lethean.io>`
- **Licence**: EUPL-1.2
- **Imports**: stdlib → forge.lthn.ai → third-party, each group separated by blank line

## Forge

- **Repo**: `forge.lthn.ai/core/go-scm`
- **Push via SSH**: `git push origin main` (remote: `ssh://git@forge.lthn.ai:2223/core/go-scm.git`)

## Task Queue

See `TODO.md` for prioritised work. Phase 1 (test coverage) is the critical path.
See `FINDINGS.md` for research notes.
