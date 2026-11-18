# go-scm

`go-scm` contains source-control, forge, marketplace, plugin, and agent CI helpers built on the `github.com/dappcore/go` core primitives.

## Layout

- `go/` contains the Go module.
- Dependencies are pinned `github.com/dappcore/*` module tags in `go/go.mod` (no `go.work`, no vendored `external/` checkouts).
- `docs/` contains architecture and development notes.
