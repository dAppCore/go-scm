---
title: Architecture
description: Internal design of go-scm -- key types, data flow, and subsystem interaction.
---

# Architecture

go-scm is organised into five major subsystems, each with a clear responsibility boundary:

1. **Forge Clients** (`forge/`, `gitea/`) -- API wrappers for Forgejo and Gitea
2. **Git Operations** (`git/`) -- multi-repo status, push, pull
3. **AgentCI Pipeline** (`jobrunner/`, `agentci/`) -- automated PR lifecycle for AI agents
4. **Data Collection** (`collect/`) -- pluggable scrapers with rate limiting and state
5. **Workspace Management** (`repos/`, `manifest/`, `marketplace/`, `plugin/`) -- multi-repo registry, manifests, extensibility

---

## 1. Forge Clients

Both `forge/` (Forgejo) and `gitea/` (Gitea) follow an identical pattern: a thin `Client` struct wrapping the upstream SDK client with config-based authentication and contextual error handling.

### Client Lifecycle

```
NewFromConfig(flagURL, flagToken)
    |
    v
ResolveConfig()        <- config file -> env vars -> flags
    |
    v
New(url, token)        <- creates SDK client
    |
    v
Client{api, url, token}
```

### Key Types -- forge package

```go
// Client wraps the Forgejo SDK with config-based auth.
type Client struct {
    api   *forgejo.Client
    url   string
    token string
}

// PRMeta holds structural signals from a pull request.
type PRMeta struct {
    Number       int64
    Title        string
    State        string
    Author       string
    Branch       string
    BaseBranch   string
    Labels       []string
    Assignees    []string
    IsMerged     bool
    CreatedAt    time.Time
    UpdatedAt    time.Time
    CommentCount int
}

// ListIssuesOpts configures issue listing.
type ListIssuesOpts struct {
    State  string   // "open", "closed", "all"
    Labels []string
    Page   int
    Limit  int
}
```

### Auth Resolution

Authentication follows a fixed priority order (lowest to highest):

1. `~/.core/config.yaml` keys `forge.url` and `forge.token`
2. `FORGE_URL` and `FORGE_TOKEN` environment variables
3. Flag overrides passed at call time
4. Default URL `http://localhost:4000` if nothing is configured

The `gitea/` package mirrors this using `GITEA_URL`/`GITEA_TOKEN` and `gitea.*` config keys, with a default of `https://gitea.snider.dev`.

### Available Operations

**forge/**

| File | Operations |
|------|-----------|
| `client.go` | `New`, `NewFromConfig`, `GetCurrentUser`, `ForkRepo`, `CreatePullRequest` |
| `repos.go` | `ListOrgRepos`, `ListOrgReposIter`, `ListUserRepos`, `ListUserReposIter`, `GetRepo`, `CreateOrgRepo`, `DeleteRepo`, `MigrateRepo` |
| `issues.go` | `ListIssues`, `ListIssuesIter`, `GetIssue`, `CreateIssue`, `EditIssue`, `AssignIssue`, `ListPullRequests`, `ListPullRequestsIter`, `GetPullRequest`, `CreateIssueComment`, `ListIssueComments`, `ListIssueCommentsIter`, `CloseIssue` |
| `labels.go` | `ListOrgLabels`, `ListOrgLabelsIter`, `ListRepoLabels`, `ListRepoLabelsIter`, `CreateRepoLabel`, `GetLabelByName`, `EnsureLabel`, `AddIssueLabels`, `RemoveIssueLabel` |
| `prs.go` | `MergePullRequest`, `SetPRDraft`, `ListPRReviews`, `GetCombinedStatus`, `DismissReview`, `UndismissReview` |
| `webhooks.go` | `CreateRepoWebhook`, `ListRepoWebhooks` |
| `orgs.go` | `ListMyOrgs`, `GetOrg`, `CreateOrg` |
| `meta.go` | `GetPRMeta`, `GetCommentBodies`, `GetIssueBody` |

### Pagination

All list methods handle pagination internally. Slice-returning methods exhaust all pages and return the full collection. Iterator-returning methods (suffixed `Iter`) yield items lazily via Go `iter.Seq2`:

```go
// Collects everything into a slice
repos, err := client.ListOrgRepos("core")

// Lazy iteration -- stops early if the consumer breaks
for repo, err := range client.ListOrgReposIter("core") {
    if repo.Name == "go-scm" { break }
}
```

### forge vs gitea

The two packages are structurally parallel but intentionally not unified behind an interface. They wrap different SDK libraries (`forgejo-sdk/v2` vs `gitea-sdk`), and the Forgejo client has additional capabilities not present in the Gitea client:

- Labels management (create, ensure, add, remove)
- Organisation creation
- Webhooks
- PR merge, draft status, reviews, combined status, review dismissal
- Repository migration (full import with issues/labels/PRs)

The Gitea client has a `CreateMirror` method for setting up pull mirrors from GitHub -- a capability specific to the public mirror workflow.

**SDK limitation:** The Forgejo SDK v2 does not accept `context.Context` on API methods. All SDK calls are synchronous. Context propagation through the wrapper layer is nominal -- contexts are accepted at the boundary but cannot be forwarded.

---

## 2. Git Operations

The `git/` package provides two layers: stateless functions and a DI-integrated service.

### Functions (Stateless)

```go
// Parallel status check across many repos
statuses := git.Status(ctx, git.StatusOptions{Paths: paths, Names: names})

// Push/pull a single repo (interactive -- attaches to terminal for SSH prompts)
git.Push(ctx, "/path/to/repo")
git.Pull(ctx, "/path/to/repo")

// Sequential multi-push with iterator
for result := range git.PushMultipleIter(ctx, paths, names) {
    fmt.Println(result.Name, result.Success)
}
```

Status checks run in parallel via goroutines, one per repository. Each goroutine shells out to `git status --porcelain` and `git rev-list --count` via `exec.CommandContext`. Push and pull operations are sequential because SSH passphrase prompts require terminal interaction -- `Stdin`, `Stdout`, and `Stderr` are connected to the process terminal.

### RepoStatus

```go
type RepoStatus struct {
    Name      string
    Path      string
    Modified  int    // Working tree modifications
    Untracked int    // Untracked files
    Staged    int    // Index changes
    Ahead     int    // Commits ahead of upstream
    Behind    int    // Commits behind upstream
    Branch    string
    Error     error
}

func (s *RepoStatus) IsDirty() bool     // Modified > 0 || Untracked > 0 || Staged > 0
func (s *RepoStatus) HasUnpushed() bool  // Ahead > 0
func (s *RepoStatus) HasUnpulled() bool  // Behind > 0
```

### GitError

```go
type GitError struct {
    Err    error
    Stderr string
}
```

All git command errors wrap stderr output for diagnostics. The `IsNonFastForward` helper checks error text for common rejection patterns.

### Service (DI-Integrated)

The `Service` struct integrates with the Core DI framework via `ServiceRuntime[ServiceOptions]`. On startup it registers query and task handlers:

| Message Type | Struct | Behaviour |
|-------------|--------|-----------|
| Query | `QueryStatus` | Runs parallel status check, caches result |
| Query | `QueryDirtyRepos` | Filters cached status for dirty repos |
| Query | `QueryAheadRepos` | Filters cached status for repos with unpushed commits |
| Task | `TaskPush` | Pushes a single repo |
| Task | `TaskPull` | Pulls a single repo |
| Task | `TaskPushMultiple` | Pushes multiple repos sequentially |

---

## 3. AgentCI Pipeline

The AgentCI subsystem automates the lifecycle of AI-agent-generated pull requests. It follows a poll-dispatch-journal architecture.

### Data Flow

```
[Forgejo API]
      |
      v
  ForgejoSource.Poll()      <- Finds epic issues, parses checklists, resolves linked PRs
      |
      v
  []PipelineSignal           <- One signal per unchecked child issue
      |
      v
  Poller.RunOnce()           <- For each signal, find first matching handler
      |
      v
  Handler.Execute()          <- Performs the action (merge, comment, dispatch, etc.)
      |
      v
  Journal.Append()           <- JSONL audit log, date-partitioned by repo
      |
      v
  Source.Report()            <- Posts result as comment on the epic issue
```

### PipelineSignal

The central data carrier. It captures the structural state of a child issue and its linked PR at poll time:

```go
type PipelineSignal struct {
    EpicNumber      int
    ChildNumber     int
    PRNumber        int
    RepoOwner       string
    RepoName        string
    PRState         string    // OPEN, MERGED, CLOSED
    IsDraft         bool
    Mergeable       string    // MERGEABLE, CONFLICTING, UNKNOWN
    CheckStatus     string    // SUCCESS, FAILURE, PENDING
    ThreadsTotal    int
    ThreadsResolved int
    LastCommitSHA   string
    LastCommitAt    time.Time
    LastReviewAt    time.Time
    NeedsCoding     bool      // true if no PR exists yet
    Assignee        string    // Forgejo username
    IssueTitle      string
    IssueBody       string
    Type            string    // e.g. "agent_completion"
    Success         bool
    Error           string
    Message         string
}
```

### Epic Issue Structure

The `ForgejoSource` expects epic issues labelled `epic` with a Markdown checklist body:

```markdown
- [ ] #42   <- unchecked = work needed
- [x] #43   <- checked = completed
- [ ] #44
```

Each unchecked child is polled. If the child has a linked PR (body references `#42`), a signal with PR metadata is emitted. If no PR exists but the issue is assigned to a known agent, a `NeedsCoding` signal is emitted instead.

### Interfaces

```go
type JobSource interface {
    Name() string
    Poll(ctx context.Context) ([]*PipelineSignal, error)
    Report(ctx context.Context, result *ActionResult) error
}

type JobHandler interface {
    Name() string
    Match(signal *PipelineSignal) bool
    Execute(ctx context.Context, signal *PipelineSignal) (*ActionResult, error)
}
```

### Poller

The `Poller` runs a blocking poll-dispatch loop. On each tick it snapshots sources and handlers (under a mutex), calls each source's `Poll`, matches the first applicable handler per signal, executes it, appends to the journal, and calls `Report` on the source. Dry-run mode logs what would execute without running handlers.

```go
poller := jobrunner.NewPoller(jobrunner.PollerConfig{
    Sources:      []jobrunner.JobSource{forgejoSrc},
    Handlers:     []jobrunner.JobHandler{dispatch, tickParent, autoMerge},
    Journal:      journal,
    PollInterval: 60 * time.Second,
})
poller.Run(ctx)  // blocks until ctx cancelled
```

Sources and handlers can be added dynamically via `AddSource` and `AddHandler`.

### Handlers

Handlers are checked in registration order. The first match wins.

| Handler | Match Condition | Action |
|---------|----------------|--------|
| `DispatchHandler` | `NeedsCoding=true`, assignee is a known agent | Build `DispatchTicket` JSON, transfer via SSH to agent queue, add `in-progress` label |
| `CompletionHandler` | `Type="agent_completion"` | Update labels (`agent-completed` or `agent-failed`), post status comment |
| `PublishDraftHandler` | Draft PR, checks passing | Remove draft status via raw HTTP PATCH |
| `EnableAutoMergeHandler` | Open, mergeable, checks passing, no unresolved threads | Squash-merge the PR |
| `DismissReviewsHandler` | Open, has unresolved threads | Dismiss stale "request changes" reviews |
| `SendFixCommandHandler` | Open, conflicting or failing with unresolved threads | Post comment asking for fixes |
| `TickParentHandler` | `PRState=MERGED` | Tick checkbox in epic body (`- [ ] #N` to `- [x] #N`), close child issue |

### Journal

`Journal` writes append-only JSONL files partitioned by date and repository:

```
{baseDir}/{owner}/{repo}/2026-03-11.jsonl
```

Each line is a `JournalEntry` with a signal snapshot (PR state, check status, mergeability) and a result snapshot (success, error, duration in milliseconds). Path components are validated against `^[a-zA-Z0-9][a-zA-Z0-9._-]*$` and resolved to absolute paths to prevent traversal. Writes are mutex-protected.

### Clotho Protocol

The `agentci.Spinner` orchestrator determines whether a dispatch should use standard or dual-run verification mode.

**Agent configuration** lives in `~/.core/config.yaml`:

```yaml
agentci:
  clotho:
    strategy: clotho-verified   # or: direct
    validation_threshold: 0.85
  agents:
    charon:
      host: build-server.leth.in
      queue_dir: /home/claude/ai-work/queue
      forgejo_user: charon
      model: sonnet
      runner: claude
      verify_model: gemini-1.5-pro
      dual_run: false
      active: true
```

`DeterminePlan` decides between `ModeStandard` and `ModeDual`:

1. If the global strategy is not `clotho-verified`, always standard.
2. If the agent's `dual_run` flag is set, dual.
3. If the repository name is `core` or contains `security`, dual (Axiom 1: critical repos always verified).
4. Otherwise, standard.

In dual-run mode, `DispatchHandler` populates `DispatchTicket.VerifyModel` and `DispatchTicket.DualRun=true`. The `Weave` method compares primary and verifier outputs for convergence using a deterministic token-overlap score against `validation_threshold`; richer semantic diffing remains a future phase.

### Dispatch Ticket Transfer

```
DispatchHandler.Execute()
  +-- SanitizePath(owner), SanitizePath(repo)
  +-- EnsureLabel(in-progress), check not already dispatched
  +-- AssignIssue, AddIssueLabels(in-progress), RemoveIssueLabel(agent-ready)
  +-- DeterminePlan(signal, agentName) -> runMode
  +-- Marshal DispatchTicket to JSON
  +-- ticketExists() via SSH (dedup check across queue/active/done)
  +-- secureTransfer(ticket JSON, mode 0644) via SSH stdin
  +-- secureTransfer(.env with FORGE_TOKEN, mode 0600) via SSH stdin
  +-- CreateIssueComment (dispatch confirmation)
```

The Forge token is transferred as a separate `.env.$ticketID` file with `0600` permissions, never embedded in the ticket JSON.

### Security Functions

| Function | Purpose |
|----------|---------|
| `SanitizePath(input)` | Returns `filepath.Base(input)` after validating against `^[a-zA-Z0-9\-\_\.]+$` |
| `EscapeShellArg(arg)` | Wraps in single quotes with internal quote escaping |
| `SecureSSHCommand(host, cmd)` | SSH with `StrictHostKeyChecking=yes`, `BatchMode=yes`, `ConnectTimeout=10` |
| `MaskToken(token)` | Returns first 4 + `****` + last 4 characters |

---

## 4. Data Collection

The `collect/` package provides a pluggable framework for gathering data from external sources.

### Collector Interface

```go
type Collector interface {
    Name() string
    Collect(ctx context.Context, cfg *Config) (*Result, error)
}
```

### Built-in Collectors

| Collector | Source | Method | Rate Limit |
|-----------|--------|--------|-----------|
| `GitHubCollector` | GitHub issues and PRs | `gh` CLI | 500ms, auto-pauses at 75% API usage |
| `BitcoinTalkCollector` | Forum topic pages | HTTP scraping + HTML parse | 2s |
| `MarketCollector` | CoinGecko current + historical data | HTTP JSON API | 1.5s |
| `PapersCollector` | IACR ePrint + arXiv | HTTP (HTML scrape + Atom XML) | 1s |
| `Processor` | Local HTML/JSON/Markdown files | Filesystem | None |

All collectors write Markdown output files, organised by source under the configured output directory:

```
{outputDir}/github/{org}/{repo}/issues/42.md
{outputDir}/bitcointalk/{topicID}/posts/1.md
{outputDir}/market/{coinID}/current.json
{outputDir}/market/{coinID}/summary.md
{outputDir}/papers/iacr/{id}.md
{outputDir}/papers/arxiv/{id}.md
{outputDir}/processed/{source}/{file}.md
```

### Excavator

The `Excavator` orchestrates multiple collectors sequentially:

```go
excavator := &collect.Excavator{
    Collectors: []collect.Collector{github, market, papers},
    Resume:     true,   // skip previously completed collectors
    ScanOnly:   false,  // true = report what would run without executing
}
result, err := excavator.Run(ctx, cfg)
```

Features:

- Rate limit respect between API calls
- Incremental state tracking (skip previously completed collectors on resume)
- Context cancellation between collectors
- Aggregated results via `MergeResults`

### Config

```go
type Config struct {
    Output     io.Medium      // Storage backend (filesystem abstraction)
    OutputDir  string         // Base directory for all output
    Limiter    *RateLimiter   // Per-source rate limits
    State      *State         // Incremental run tracking
    Dispatcher *Dispatcher    // Event dispatch for progress reporting
    Verbose    bool
    DryRun     bool           // Simulate without writing
}
```

### Rate Limiting

The `RateLimiter` tracks per-source last-request timestamps. `Wait(ctx, source)` blocks for the configured delay minus elapsed time. The mutex is released during the wait to avoid holding it across a timer.

Default delays:

| Source | Delay |
|--------|-------|
| `github` | 500ms |
| `bitcointalk` | 2s |
| `coingecko` | 1.5s |
| `iacr` | 1s |
| `arxiv` | 1s |

The `CheckGitHubRateLimitCtx` method queries `gh api rate_limit` and automatically increases the GitHub delay to 5 seconds when usage exceeds 75%.

### Events

The `Dispatcher` provides synchronous event dispatch with five event types:

| Constant | Meaning |
|----------|---------|
| `EventStart` | Collector begins its run |
| `EventProgress` | Incremental progress update |
| `EventItem` | Single item collected |
| `EventError` | Error during collection |
| `EventComplete` | Collector finished |

Register handlers with `dispatcher.On(eventType, handler)`. Convenience methods `EmitStart`, `EmitProgress`, `EmitItem`, `EmitError`, `EmitComplete` are provided.

### State Persistence

The `State` tracker serialises per-source progress to `.collect-state.json` via an `io.Medium` backend. Each `StateEntry` records:

- Source name
- Last run timestamp
- Last item ID (opaque)
- Total items collected
- Pagination cursor (opaque)

Thread-safe via mutex. Returns copies from `Get` to prevent callers mutating internal state.

---

## 5. Workspace Management

### repos.yaml Registry

The `repos/` package reads a `repos.yaml` file defining a multi-repo workspace:

```yaml
version: 1
org: core
base_path: ~/Code/core
defaults:
  ci: forgejo
  license: EUPL-1.2
  branch: main
repos:
  go-scm:
    type: module
    depends_on: [go-io, go-log, config]
    description: SCM integration
  go-ai:
    type: module
    depends_on: [go-ml, go-rag]
```

**Repository types:** `foundation`, `module`, `product`, `template`.

The `Registry` provides:

- **Lookups:** `List()`, `Get(name)`, `ByType(t)`
- **Dependency sorting:** `TopologicalOrder()` -- returns repos in dependency order, detects cycles
- **Discovery:** `FindRegistry(medium)` searches cwd, parent directories, and well-known home paths
- **Fallback:** `ScanDirectory(medium, dir)` scans for `.git` directories when no `repos.yaml` exists

Each `Repo` struct has computed fields (`Path`, `Name`) and methods (`Exists()`, `IsGitRepo()`). The `Clone` field (pointer to bool) allows excluding repos from cloning operations (nil defaults to true).

### WorkConfig and GitState

Workspace sync behaviour is split into two files:

| File | Scope | Git-tracked? |
|------|-------|-------------|
| `.core/work.yaml` | Sync policy (intervals, auto-pull/push, agent heartbeats) | Yes |
| `.core/git.yaml` | Per-machine state (last pull/push times, agent presence) | No (.gitignored) |

**WorkConfig** controls:

```go
type SyncConfig struct {
    Interval     time.Duration  // How often to sync
    AutoPull     bool           // Pull automatically
    AutoPush     bool           // Push automatically
    CloneMissing bool           // Clone repos not yet present
}

type AgentPolicy struct {
    Heartbeat     time.Duration  // How often agents check in
    StaleAfter    time.Duration  // When to consider an agent stale
    WarnOnOverlap bool           // Warn if multiple agents touch same repo
}
```

**GitState** tracks:

- Per-repo: last pull/push timestamps, branch, remote, ahead/behind counts
- Per-agent: last seen timestamp, list of active repos
- Methods: `TouchPull`, `TouchPush`, `UpdateRepo`, `Heartbeat`, `StaleAgents`, `ActiveAgentsFor`, `NeedsPull`

### KBConfig

The `.core/kb.yaml` file configures a knowledge base layer:

```go
type KBConfig struct {
    Wiki   WikiConfig  // Local wiki mirroring from Forgejo
    Search KBSearch    // Vector search via Qdrant + Ollama embeddings
}
```

The `WikiRepoURL` and `WikiLocalPath` methods compute clone URLs and local paths for wiki repos.

### Manifest

The `manifest/` package handles `.core/manifest.yaml` files describing application modules:

```yaml
code: my-module
name: My Module
version: 1.0.0
permissions:
  read: ["/data"]
  write: ["/output"]
  net: ["api.example.com"]
  run: ["./worker"]
modules: [dep-a, dep-b]
daemons:
  worker:
    binary: ./worker
    args: ["--port", "8080"]
    health: http://localhost:8080/health
    default: true
```

**Key operations:**

| Function | Purpose |
|----------|---------|
| `Parse(data)` | Decode YAML bytes into a `Manifest` |
| `Load(medium, root)` | Read `.core/manifest.yaml` from a directory |
| `LoadVerified(medium, root, pubKey)` | Load and verify ed25519 signature |
| `Sign(manifest, privKey)` | Compute ed25519 signature, store as base64 in `Sign` field |
| `Verify(manifest, pubKey)` | Check the `Sign` field against the public key |
| `SlotNames()` | Deduplicated component names from the slots map |
| `DefaultDaemon()` | Resolve the default daemon (explicit `Default: true` or sole daemon) |

Signing works by zeroing the `Sign` field, marshalling to YAML, and computing `ed25519.Sign` over the canonical bytes. The base64-encoded signature is stored back in `Sign`.

### Marketplace

The `marketplace/` package provides a module catalogue and installer:

```go
// Catalogue
index, _ := marketplace.ParseIndex(jsonData)
results := index.Search("analytics")
byCategory := index.ByCategory("monitoring")
mod, found := index.Find("my-module")

// Installation
installer := marketplace.NewInstaller(medium, "/path/to/modules", store)
installer.Install(ctx, mod)     // Clone, verify manifest, register
installer.Update(ctx, "code")   // Pull, re-verify, update metadata
installer.Remove("code")        // Delete files and store entry
installed, _ := installer.Installed()  // List all installed modules
```

The installer:

1. Clones the module repo with `--depth=1`
2. Loads the manifest via a sandboxed `io.Medium`
3. If a `SignKey` is present on the catalogue entry, verifies the ed25519 signature
4. Registers metadata (code, name, version, permissions, entry point) in a `store.Store`
5. Cleans up the cloned directory on any failure after clone

### Plugin System

The `plugin/` package provides a CLI extension mechanism:

```go
type Plugin interface {
    Name() string
    Version() string
    Init(ctx context.Context) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

`BasePlugin` provides a default (no-op) implementation for embedding.

**Components:**

| Type | Purpose |
|------|---------|
| `Manifest` | `plugin.json` with name, version, description, author, entrypoint, dependencies |
| `Registry` | JSON-backed store of installed plugins (`registry.json`) |
| `Loader` | Discovers plugins by scanning directories for `plugin.json` |
| `Installer` | Clones from GitHub via `gh`, validates manifest, registers |

Source format: `org/repo` or `org/repo@v1.0`. The `ParseSource` function splits these into organisation, repository, and version components.

---

## Dependency Graph

```
                                     forge.lthn.ai/core/go  (DI, log, config, io)
                                           |
                    +----------------------+----------------------+
                    |                      |                      |
                forge/                  gitea/                  git/
                    |                      |                      |
            +-------+-------+              |                      |
            |               |              |                      |
        agentci/        jobrunner/         |                      |
            |           |       |          |                      |
            |   forgejo/source  |          |                      |
            |           |       |          |                      |
            +-----------+-------+          |                      |
                    |                      |                      |
              handlers/                    |                      |
                                           |                      |
                collect/  -----------------+                      |
                                                                  |
                repos/  ------------------------------------------+
                manifest/
                marketplace/ (depends on manifest/, io/)
                plugin/ (depends on io/)
```

External SDK dependencies:

- `codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2` -- Forgejo API
- `code.gitea.io/sdk/gitea` -- Gitea API
- `github.com/stretchr/testify` -- test assertions
- `golang.org/x/net` -- HTML parsing
- `gopkg.in/yaml.v3` -- YAML parsing
