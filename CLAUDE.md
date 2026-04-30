# CLAUDE.md

This file provides guidance for agents working in this repository.

## Project Overview

This repository contains the Go module for `dappco.re/go/scm`.

## Repo Layout

Go module files and all Go packages now live under `go/`.

```text
go-scm/
├── go/
│   ├── go.mod
│   ├── go.sum
│   ├── scm.go
│   ├── scm_test.go
│   ├── agentci/
│   ├── cmd/
│   ├── collect/
│   ├── core/
│   ├── forge/
│   ├── git/
│   ├── gitea/
│   ├── internal/
│   ├── jobrunner/
│   ├── manifest/
│   ├── marketplace/
│   ├── pkg/
│   ├── plugin/
│   ├── repos/
│   ├── tests/
│   ├── third_party/
│   ├── README.md (symlink when present at repo root)
│   ├── CLAUDE.md (symlink when present at repo root)
│   └── AGENTS.md (symlink when present at repo root)
├── .woodpecker.yml
└── sonar-project.properties
```

All Go-related checks, tooling, and package commands should be run from
`go/` (for example `cd go && go test ./...`, `cd go && go mod tidy`).

For repository docs layout, keep cross-language files at the repo root and
gate them into `go/` via symlinks when they exist.
