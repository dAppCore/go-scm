# go-scm Agent Notes

This repository follows the core/go v0.9.0 consumer layout:

- Go module source lives under `go/`.
- The repository root owns `go.work`.
- Local dependency checkouts live under `external/` and must not be edited directly.

Use `dappco.re/go` primitives instead of banned standard-library imports in consumer code. Public symbols require matching Good, Bad, and Ugly tests plus examples in the source file's sibling test files.
