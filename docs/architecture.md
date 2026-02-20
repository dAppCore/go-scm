# go-scm Architecture

Module path: `forge.lthn.ai/core/go-scm`

go-scm provides SCM integration, CI dispatch automation, and data collection for the Lethean ecosystem. It is composed of six packages, each with a distinct responsibility, and approximately 9,000 lines of Go across roughly 70 source files.

---

## Package Overview

```
forge.lthn.ai/core/go-scm
├── forge/        Forgejo API client (repos, issues, PRs, labels, webhooks, orgs)
├── gitea/        Gitea API client (repos, issues, meta) for public mirror
├── git/          Multi-repo git operations (status, commit, push, pull)
├── agentci/      Clotho Protocol orchestrator — agent config and security
├── jobrunner/    PR automation pipeline (poll → dispatch → journal)
│   ├── forgejo/  Forgejo signal source (epic issue parsing)
│   └── handlers/ Pipeline action handlers
└── collect/      Data collection (BitcoinTalk, GitHub, market, papers, events)
```

---

## SCM Abstraction Layer

### forge/ — Forgejo API Client

`forge/` wraps the Forgejo SDK (`codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2`) with config-based authentication and contextual error wrapping. It provides thin, typed wrappers for every API surface used by the Lethean platform.

**Client construction:**

```go
// Config-resolved client (preferred)
client, err := forge.NewFromConfig(flagURL, flagToken)

// Direct construction
client, err := forge.New(url, token)
```

**Auth resolution** follows a fixed priority order:

1. `~/.core/config.yaml` keys `forge.url` and `forge.token` (lowest priority)
2. `FORGE_URL` and `FORGE_TOKEN` environment variables
3. Flag overrides passed at call time (highest priority)
4. Default URL `http://localhost:4000` if nothing is configured

**Available operations:**

| File | Operations |
|------|-----------|
| `repos.go` | `CreateRepo`, `ListRepos`, `CreateMirrorRepo`, `CreateOrgRepo` |
| `issues.go` | `GetIssue`, `CreateIssue`, `ListIssues`, `CreateIssueComment`, `AssignIssue`, `CloseIssue`, `EditIssue` |
| `prs.go` | `CreatePullRequest`, `ListPullRequests`, `MergePullRequest`, `SetPRDraft`, `GetCombinedStatus` |
| `labels.go` | `CreateLabel`, `GetLabelByName`, `EnsureLabel`, `AddIssueLabels`, `RemoveIssueLabel` |
| `webhooks.go` | `CreateWebhook`, `ListWebhooks`, `DeleteWebhook` |
| `orgs.go` | `CreateOrg`, `ListOrgs`, `ListOrgRepos` |
| `meta.go` | `GetVersion` |

**SDK limitation:** The Forgejo SDK v2 does not accept `context.Context` on any API method. All SDK calls are synchronous and blocking. Context propagation through forge/ and gitea/ wrappers is therefore nominal — contexts are accepted at the wrapper boundary but cannot be passed to the SDK. This will be resolved when the SDK adds context support.

### gitea/ — Gitea API Client

`gitea/` mirrors the structure of `forge/` but wraps the Gitea SDK (`code.gitea.io/sdk/gitea`) for the public mirror instance at `git.lthn.ai`. The two clients are intentionally structurally identical — same pattern of `client.go`, `config.go`, `repos.go`, `issues.go`, `meta.go` — to reduce cognitive load when working across both.

**Auth resolution** follows the same priority order as forge/, using `GITEA_URL`/`GITEA_TOKEN` environment variables and `gitea.url`/`gitea.token` config keys. The default URL is `https://gitea.snider.dev`.

**Infrastructure split:**

- `forge.lthn.ai` — production Forgejo instance, source of truth, full IP/research data
- `git.lthn.ai` — public Gitea mirror with sensitive data stripped, breach-safe

### git/ — Multi-Repo Git Operations

`git/` provides context-aware git operations across multiple repositories. Unlike the API clients, all operations in this package propagate `context.Context` correctly via `exec.CommandContext`.

**Core types:**

```go
type RepoStatus struct {
    Name, Path, Branch string
    Modified, Untracked, Staged int  // working tree counts
    Ahead, Behind               int  // commits vs upstream
    Error                       error
}

func (s *RepoStatus) IsDirty() bool    { ... }
func (s *RepoStatus) HasUnpushed() bool { ... }
func (s *RepoStatus) HasUnpulled() bool { ... }
```

**Parallel status across repos:**

```go
statuses := git.Status(ctx, git.StatusOptions{
    Paths: []string{"/path/to/repo-a", "/path/to/repo-b"},
    Names: map[string]string{"/path/to/repo-a": "repo-a"},
})
```

Status checks run in parallel via goroutines. Push and pull operations are sequential because SSH passphrase prompts require terminal interaction.

**Service integration:** `git.Service` embeds `framework.ServiceRuntime` and registers query/task handlers on the core framework's message bus. Queries (`QueryStatus`, `QueryDirtyRepos`, `QueryAheadRepos`) return from a cached `lastStatus` field. Tasks (`TaskPush`, `TaskPull`, `TaskPushMultiple`) execute immediately.

---

## AgentCI Dispatch Pipeline

### Overview

The AgentCI pipeline automates the lifecycle of issues assigned to AI agents: detecting unstarted work, dispatching tickets to agent machines, monitoring PR state, and updating the parent epic on merge.

```
Forgejo instance
      │
      │ poll (epic issues, child PRs, combined status)
      ▼
ForgejoSource.Poll()
      │
      │ []PipelineSignal
      ▼
Poller.RunOnce()
      │
      │ Match(signal) → first matching handler
      ├─► DispatchHandler     — NeedsCoding=true, known agent assignee
      ├─► TickParentHandler   — PRState=MERGED
      ├─► EnableAutoMerge     — checks passing, mergeable
      ├─► PublishDraft        — draft PR ready
      ├─► SendFixCommand      — checks failing
      └─► CompletionHandler   — agent completion signal
      │
      │ ActionResult
      ▼
Journal.Append()   — JSONL audit trail
ForgejoSource.Report()  — comment on epic issue
```

### jobrunner/ — Poller and Interfaces

`jobrunner/` defines the interfaces and orchestration loop shared by all pipeline participants.

**Interfaces:**

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

**PipelineSignal** carries the full structural snapshot of a child issue/PR at the moment of polling:

```go
type PipelineSignal struct {
    EpicNumber, ChildNumber, PRNumber int
    RepoOwner, RepoName               string
    PRState                           string  // OPEN, MERGED, CLOSED
    IsDraft                           bool
    Mergeable                         string  // MERGEABLE, CONFLICTING, UNKNOWN
    CheckStatus                       string  // SUCCESS, FAILURE, PENDING
    ThreadsTotal, ThreadsResolved     int
    NeedsCoding                       bool    // true if no PR exists yet
    Assignee                          string  // Forgejo username
    IssueTitle, IssueBody             string  // for dispatch prompt
}
```

**Poller** runs a blocking poll-dispatch loop. On each tick it snapshots sources and handlers (under a read lock), calls each source's `Poll`, matches the first applicable handler per signal, executes it, appends to the journal, and calls `Report` on the source. Dry-run mode logs what would execute without running handlers.

```go
poller := jobrunner.NewPoller(jobrunner.PollerConfig{
    Sources:      []jobrunner.JobSource{forgejoSrc},
    Handlers:     []jobrunner.JobHandler{dispatch, tickParent, autoMerge},
    Journal:      journal,
    PollInterval: 60 * time.Second,
})
poller.Run(ctx)  // blocks until ctx cancelled
```

### jobrunner/forgejo/ — Signal Source

`ForgejoSource` polls a list of repositories for epic issues (labelled `epic`, state `open`). For each epic, it parses the issue body for unchecked task list items (`- [ ] #N`), then for each unchecked child either:

- Builds a `PipelineSignal` with PR state, draft status, check status, and thread counts (if a linked PR exists), or
- Builds a `NeedsCoding=true` signal carrying the child issue title and body (if no PR exists and the issue has an assignee)

Combined commit status is fetched per head SHA via `forge.GetCombinedStatus`.

### jobrunner/handlers/ — Action Handlers

| Handler | Match condition | Action |
|---------|----------------|--------|
| `DispatchHandler` | `NeedsCoding=true`, assignee is a known agent | Build `DispatchTicket` JSON, transfer via SSH, post comment |
| `TickParentHandler` | `PRState=MERGED` | Tick checkbox in epic body, close child issue |
| `EnableAutoMergeHandler` | `CheckStatus=SUCCESS`, `Mergeable=MERGEABLE`, not draft | Enable auto-merge on PR |
| `PublishDraftHandler` | Is draft, threads resolved | Publish draft PR |
| `SendFixCommandHandler` | `CheckStatus=FAILURE` | Post fix command comment to agent |
| `CompletionHandler` | `Type=agent_completion` | Record agent completion result |

### agentci/ — Clotho Protocol

`agentci/` manages agent configuration and the Clotho Protocol for dual-run verification.

**Agent configuration** is loaded from `~/.core/config.yaml` under the `agentci.agents` key:

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
      runner: claude             # claude, codex, or gemini
      verify_model: gemini-1.5-pro
      dual_run: false
      active: true
```

**Spinner** is the Clotho orchestrator. Its `DeterminePlan` method decides between `standard` and `dual` run modes:

1. If the global strategy is not `clotho-verified`, always `standard`.
2. If the agent's `dual_run` flag is set, `dual`.
3. If the repository name is `core` or contains `security`, `dual` (Axiom 1: critical repos always verified).
4. Otherwise, `standard`.

In dual-run mode, `DispatchHandler` populates `DispatchTicket.VerifyModel` and `DispatchTicket.DualRun=true`. The agent runner is responsible for executing both the primary and verifier models and calling `Spinner.Weave` to compare outputs. `Weave` currently performs a byte-equal comparison; semantic diff logic is reserved for a future phase.

**Security functions** in `agentci/security.go`:

- `SanitizePath(input string)` — returns `filepath.Base(input)` after validating against `^[a-zA-Z0-9\-\_\.]+$`. Protects against path traversal by stripping directory components rather than rejecting the input.
- `EscapeShellArg(arg string)` — wraps a string in single quotes with internal single-quote escaping, for safe insertion into SSH remote commands.
- `SecureSSHCommandContext(ctx, host, cmd string)` — constructs an `exec.Cmd` with `StrictHostKeyChecking=yes`, `BatchMode=yes`, and `ConnectTimeout=10`.
- `MaskToken(token string)` — returns a masked version safe for logging.

**Dispatch ticket transfer:**

```
DispatchHandler.Execute()
  ├── SanitizePath(owner), SanitizePath(repo)
  ├── EnsureLabel(in-progress)
  ├── Check issue not already in-progress or completed
  ├── AssignIssue, AddIssueLabels
  ├── DeterminePlan(signal, agentName) → runMode
  ├── Marshal DispatchTicket to JSON
  ├── ticketExists() via SSH (dedup check)
  ├── secureTransfer(ticket JSON, 0644)  ← cat > path via SSH stdin
  ├── secureTransfer(.env with FORGE_TOKEN, 0600)
  └── CreateIssueComment (dispatch confirmation)
```

The Forge token is written as a separate `.env.$ticketID` file with `0600` permissions rather than embedded in the ticket JSON, to avoid the token appearing in queue directory listings.

### Journal

`jobrunner.Journal` writes append-only JSONL files partitioned by date and repository:

```
{baseDir}/{owner}/{repo}/2026-02-20.jsonl
```

Each line is a `JournalEntry` with a signal snapshot (PR state at time of action) and a result snapshot (success, error, duration). Path components are validated against a strict regex and resolved to absolute paths to prevent traversal. Writes are mutex-protected for concurrent safety.

**Replay filtering** (via `journal_replay_test.go` patterns, not yet a public API): entries can be filtered by action name, repo full name, and time range by scanning the JSONL file.

---

## Data Collection

### collect/ — Collection Pipeline

`collect/` provides a pluggable pipeline for gathering data from external sources.

**Collector interface:**

```go
type Collector interface {
    Name() string
    Collect(ctx context.Context, cfg *Config) (*Result, error)
}
```

**Available collectors:**

| File | Source | Rate limit |
|------|--------|-----------|
| `bitcointalk.go` | BitcoinTalk forum (HTTP scraping) | 2 s per request |
| `github.go` | GitHub API via `gh` CLI | 500 ms, pauses at 75% usage |
| `market.go` | CoinGecko market data | 1.5 s per request |
| `papers.go` | IACR and arXiv research papers | 1 s per request |
| `events.go` | Event tracking | — |

**Excavator** orchestrates sequential execution of multiple collectors with state-based resume support:

```go
exc := &collect.Excavator{
    Collectors: []collect.Collector{githubCollector, marketCollector},
    Resume:     true,
}
result, err := exc.Run(ctx, cfg)
```

If `Resume=true`, collectors that already have a non-zero item count in the persisted state file are skipped. Context cancellation is checked between collectors.

**Rate limiter** tracks per-source last-request timestamps. `Wait(ctx, source)` blocks for the configured delay minus elapsed time, then releases. The mutex is released during the wait to avoid holding it across a timer. GitHub rate limiting queries the `gh api rate_limit` endpoint and automatically increases the GitHub delay to 5 s when usage exceeds 75%.

**State** persists collection progress to a JSON file via an `io.Medium` abstraction, enabling incremental runs. Each `StateEntry` stores the last run timestamp, item count, and an opaque cursor for pagination resumption.

**Process pipeline** (`process.go`) handles post-collection transformation. The `Dispatcher` in `events.go` emits typed events (`start`, `progress`, `error`, `complete`) during collection runs.

---

## Dependency Graph

```
collect/  ─────────────────────────────────────────────┐
                                                        │
git/      ──────────────────────────────────────────┐  │
                                                    │  │
gitea/    ────────────────────────────────────┐     │  │
                                              │     │  │
forge/    ────────────────────────────┐       │     │  │
                                      │       │     │  │
agentci/  ──────────────────────────┐ │       │     │  │
                                    │ │       │     │  │
jobrunner/          ────────────────┘ │       │     │  │
jobrunner/forgejo/  ──────────────────┘       │     │  │
jobrunner/handlers/ ──────────────────────────┘     │  │
                                                    │  │
forge.lthn.ai/core/go  (framework, log, config) ───┴──┘
```

External SDK dependencies:
- `codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2` — Forgejo API
- `code.gitea.io/sdk/gitea` — Gitea API
- `github.com/stretchr/testify` — test assertions
- `golang.org/x/net` — HTTP utilities
