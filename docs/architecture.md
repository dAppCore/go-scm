# Architecture

`go-scm` is split into small packages around source-control workflows:

- `git`, `repos`, `forge`, and `gitea` wrap repository and forge operations.
- `manifest`, `marketplace`, and `plugin` manage package metadata and installation.
- `collect` and `jobrunner` support automation and agent CI workflows.

Packages should depend on `dappco.re/go` for filesystem, JSON, string, path, error, and process primitives.
