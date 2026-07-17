# go-scm Agent Notes

This repository follows the tag-based core/go consumer layout:

- Go module source lives under `go/`.
- Dependencies are pinned `dappco.re/*` module tags in `go/go.mod` — no `go.work`, no vendored `external/` checkouts.

Use `dappco.re/go` primitives instead of banned standard-library imports in consumer code. Public symbols require matching Good, Bad, and Ugly tests plus examples in the source file's sibling test files.
