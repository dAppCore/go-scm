# go-scm

`go-scm` contains source-control, forge, marketplace, plugin, and agent CI helpers built on the `dappco.re/go` core primitives.

## Layout

- `go/` contains the Go module.
- Dependencies are pinned `dappco.re/*` module tags in `go/go.mod` (no `go.work`, no vendored `external/` checkouts).
- `docs/` contains architecture and development notes.
