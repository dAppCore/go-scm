---
title: go-scm
description: SCM integration, AgentCI automation, and data collection for the Lethean ecosystem.
---

# go-scm

`go-scm` provides source control management integration for the Lethean ecosystem. It wraps the Forgejo and Gitea APIs behind ergonomic Go clients, runs an automated PR pipeline for AI agent workflows, collects data from external sources, and manages multi-repo workspaces via a declarative registry.

**Module path:** `forge.lthn.ai/core/go-scm`
**Go version:** 1.26
**Licence:** EUPL-1.2

## Quick Start

### Forgejo API Client

```go
import "forge.lthn.ai/core/go-scm/forge"

// Create a client from config file / env / flags
client, err := forge.NewFromConfig("", "")

// List open issues
issues, err := client.ListIssues("core", "go-scm", forge.ListIssuesOpts{
    State: "open",
})

// List repos in an organisation (paginated iterator)
for repo, err := range client.ListOrgReposIter("core") {
    fmt.Println(repo.Name)
}
```

### Multi-Repo Git Status

```go
import "forge.lthn.ai/core/go-scm/git"

statuses := git.Status(ctx, git.StatusOptions{
    Paths: []string{"/home/dev/core/go-scm", "/home/dev/core/go-ai"},
    Names: map[string]string{"/home/dev/core/go-scm": "go-scm"},
})

for _, s := range statuses {
    if s.IsDirty() {
        fmt.Printf("%s: %d modified, %d untracked\n", s.Name, s.Modified, s.Untracked)
    }
}
```

### AgentCI Poll-Dispatch Loop

```go
import (
    "forge.lthn.ai/core/go-scm/jobrunner"
    "forge.lthn.ai/core/go-scm/jobrunner/forgejo"
    "forge.lthn.ai/core/go-scm/jobrunner/handlers"
)

source := forgejo.New(forgejo.Config{Repos: []string{"core/go-scm"}}, forgeClient)
poller := jobrunner.NewPoller(jobrunner.PollerConfig{
    Sources:      []jobrunner.JobSource{source},
    Handlers:     []jobrunner.JobHandler{
        handlers.NewDispatchHandler(forgeClient, forgeURL, token, spinner),
        handlers.NewTickParentHandler(forgeClient),
        handlers.NewEnableAutoMergeHandler(forgeClient),
    },
    PollInterval: 60 * time.Second,
})
poller.Run(ctx)
```

### Data Collection

```go
import "forge.lthn.ai/core/go-scm/collect"

cfg := collect.NewConfig("/tmp/collected")
excavator := &collect.Excavator{
    Collectors: []collect.Collector{
        &collect.GitHubCollector{Org: "lethean-io"},
        &collect.MarketCollector{CoinID: "lethean", Historical: true},
        &collect.PapersCollector{Source: "all", Query: "cryptography VPN"},
    },
    Resume: true,
}
result, err := excavator.Run(ctx, cfg)
```

## Package Layout

| Package | Import Path | Description |
|---------|-------------|-------------|
| `forge` | `go-scm/forge` | Forgejo API client -- repos, issues, PRs, labels, webhooks, organisations, PR metadata |
| `gitea` | `go-scm/gitea` | Gitea API client -- repos, issues, PRs, mirroring, PR metadata |
| `git` | `go-scm/git` | Multi-repo git operations -- parallel status checks, push, pull; Core DI service |
| `jobrunner` | `go-scm/jobrunner` | AgentCI pipeline engine -- signal types, poller loop, JSONL audit journal |
| `jobrunner/forgejo` | `go-scm/jobrunner/forgejo` | Forgejo job source -- polls epic issues for unchecked children, builds signals |
| `jobrunner/handlers` | `go-scm/jobrunner/handlers` | Pipeline handlers -- dispatch, completion, auto-merge, publish-draft, dismiss-reviews, fix-command, tick-parent |
| `agentci` | `go-scm/agentci` | Clotho Protocol orchestrator -- agent config, SSH security helpers, dual-run verification |
| `collect` | `go-scm/collect` | Data collection framework -- collector interface, rate limiting, state tracking, event dispatch |
| `manifest` | `go-scm/manifest` | Application manifest -- YAML parsing, ed25519 signing and verification |
| `marketplace` | `go-scm/marketplace` | Module marketplace -- catalogue index, search, git-based installer with signature verification |
| `plugin` | `go-scm/plugin` | CLI plugin system -- plugin interface, JSON registry, loader, GitHub-based installer |
| `repos` | `go-scm/repos` | Workspace management -- `repos.yaml` registry, topological sorting, work config, git state, KB config |
| `cmd/forge` | `go-scm/cmd/forge` | CLI commands for the `core forge` subcommand |
| `cmd/gitea` | `go-scm/cmd/gitea` | CLI commands for the `core gitea` subcommand |
| `cmd/collect` | `go-scm/cmd/collect` | CLI commands for data collection |

## Dependencies

### Direct

| Module | Purpose |
|--------|---------|
| `codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2` | Forgejo API SDK |
| `code.gitea.io/sdk/gitea` | Gitea API SDK |
| `forge.lthn.ai/core/cli` | CLI framework (Cobra, TUI) |
| `forge.lthn.ai/core/go-config` | Layered config (`~/.core/config.yaml`) |
| `forge.lthn.ai/core/go-io` | Filesystem abstraction (Medium, Sandbox, Store) |
| `forge.lthn.ai/core/go-log` | Structured logging and contextual error helper |
| `forge.lthn.ai/core/go-i18n` | Internationalisation |
| `github.com/stretchr/testify` | Test assertions |
| `golang.org/x/net` | HTML parsing for collectors |
| `gopkg.in/yaml.v3` | YAML parsing for manifests and registries |

### Indirect

The module transitively pulls in `forge.lthn.ai/core/go` (DI framework) via `go-config`, plus `spf13/viper`, `spf13/cobra`, Charmbracelet TUI libraries, and Go standard library extensions.

## Configuration

Authentication for both Forgejo and Gitea is resolved through a three-tier priority chain:

1. **Config file** -- `~/.core/config.yaml` keys `forge.url`, `forge.token` (or `gitea.*`)
2. **Environment variables** -- `FORGE_URL`, `FORGE_TOKEN` (or `GITEA_URL`, `GITEA_TOKEN`)
3. **CLI flags** -- `--url`, `--token` (highest priority)

Set credentials once:

```bash
core forge config --url https://forge.lthn.ai --token <your-token>
core gitea config --url https://gitea.snider.dev --token <your-token>
```

## Further Reading

- [Architecture](architecture.md) -- internal design, key types, data flow
- [Development Guide](development.md) -- building, testing, contributing
