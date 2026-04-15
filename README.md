[![Go Reference](https://pkg.go.dev/badge/dappco.re/go/core/scm.svg)](https://pkg.go.dev/dappco.re/go/core/scm)
[![License: EUPL-1.2](https://img.shields.io/badge/License-EUPL--1.2-blue.svg)](LICENSE.md)
[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat&logo=go)](go.mod)

# go-scm

SCM integration, AgentCI dispatch automation, and data collection for the Lethean ecosystem. Provides a Forgejo API client and a Gitea client for the public mirror, multi-repo git operations with parallel status checks, the Clotho Protocol orchestrator for dual-run agent verification, the `core dev`/`core pkg` CLI surface, automatic `WorkspacePushed` repo sync handling, a PR automation pipeline (poll → dispatch → journal) driven by epic issue task lists, and pluggable data collectors for BitcoinTalk, GitHub, market data, and research papers.

**Module**: `dappco.re/go/core/scm`
**Licence**: EUPL-1.2
**Language**: Go 1.26

## Quick Start

```go
import (
    "dappco.re/go/core/scm/forge"
    "dappco.re/go/core/scm/git"
    "dappco.re/go/core/scm/jobrunner"
)

// Forgejo client
client, err := forge.NewFromConfig("", "")

// Multi-repo status
statuses := git.Status(ctx, git.StatusOptions{Paths: repoPaths})

// AgentCI dispatch loop
poller := jobrunner.NewPoller(jobrunner.PollerConfig{
    Sources:      []jobrunner.JobSource{forgejoSrc},
    Handlers:     []jobrunner.JobHandler{dispatch, tickParent},
    PollInterval: 60 * time.Second,
})
poller.Run(ctx)
```

## Documentation

- [Architecture](docs/architecture.md) — package overview, AgentCI pipeline, Clotho Protocol, data collection
- [Development Guide](docs/development.md) — building, testing, standards
- [Project History](docs/history.md) — completed phases and known limitations

## Build & Test

```bash
go test ./...
go test -race ./...
go build ./...
```

## Licence

European Union Public Licence 1.2 — see [LICENCE](LICENCE) for details.
