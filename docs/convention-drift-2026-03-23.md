# Convention Drift Check — 2026-03-23

`CODEX.md` is not present under `/workspace`, so this pass uses [CLAUDE.md](/workspace/CLAUDE.md) and [docs/development.md](/workspace/docs/development.md) as the convention baseline.

Scope used for this pass:

- Import ordering was checked on non-test Go files for `stdlib -> core/internal -> third-party`, including blank-line separation between groups.
- UK English findings are limited to repo-authored prose/comments/schema terms, not external API field names, CSS properties, or shield/vendor text beyond repo-owned alt text.
- Missing-test findings come from `go test -coverprofile=/tmp/convention_drift_cover.out ./...` plus `go tool cover -func`.
- SPDX findings are limited to non-test `.go` and `.ts` source files.

## Import Grouping Drift (36)

### Core/internal imports placed after third-party imports (24)

- `cmd/forge/cmd_issues.go:9`: internal/core import follows a third-party SDK import.
- `cmd/forge/cmd_labels.go:8`: internal/core import follows a third-party SDK import.
- `cmd/forge/cmd_migrate.go:8`: internal/core import follows a third-party SDK import.
- `cmd/forge/cmd_prs.go:9`: internal/core import follows a third-party SDK import.
- `cmd/forge/cmd_repos.go:8`: internal/core import follows a third-party SDK import.
- `cmd/forge/cmd_sync.go:12`: internal/core import follows a third-party SDK import.
- `cmd/gitea/cmd_issues.go:9`: internal/core import follows a third-party SDK import.
- `cmd/gitea/cmd_prs.go:9`: internal/core import follows a third-party SDK import.
- `cmd/gitea/cmd_sync.go:12`: internal/core import follows a third-party SDK import.
- `forge/client.go:14`: `dappco.re/go/core/log` follows a third-party SDK import.
- `forge/issues.go:8`: `dappco.re/go/core/log` follows a third-party SDK import.
- `forge/labels.go:8`: `dappco.re/go/core/log` follows a third-party SDK import.
- `forge/meta.go:8`: `dappco.re/go/core/log` follows a third-party SDK import.
- `forge/orgs.go:6`: `dappco.re/go/core/log` follows a third-party SDK import.
- `forge/prs.go:11`: `dappco.re/go/core/log` follows a third-party SDK import.
- `forge/repos.go:8`: `dappco.re/go/core/log` follows a third-party SDK import.
- `forge/webhooks.go:6`: `dappco.re/go/core/log` follows a third-party SDK import.
- `gitea/client.go:14`: `dappco.re/go/core/log` follows a third-party SDK import.
- `gitea/issues.go:8`: `dappco.re/go/core/log` follows a third-party SDK import.
- `gitea/meta.go:8`: `dappco.re/go/core/log` follows a third-party SDK import.
- `gitea/repos.go:8`: `dappco.re/go/core/log` follows a third-party SDK import.
- `jobrunner/forgejo/signals.go:9`: internal jobrunner import follows a third-party SDK import.
- `jobrunner/handlers/resolve_threads.go:10`: `dappco.re/go/core/log` follows a third-party SDK import.
- `jobrunner/handlers/tick_parent.go:11`: `dappco.re/go/core/log` follows a third-party SDK import.

### Missing blank line between import groups (12)

- `collect/bitcointalk.go:13`: no blank line between internal/core and third-party imports.
- `collect/papers.go:14`: no blank line between internal/core and third-party imports.
- `collect/process.go:13`: no blank line between internal/core and third-party imports.
- `manifest/loader.go:9`: no blank line before the third-party YAML import.
- `manifest/manifest.go:5`: no blank line before the third-party YAML import.
- `manifest/sign.go:8`: no blank line before the third-party YAML import.
- `marketplace/discovery.go:11`: no blank line before the third-party YAML import.
- `pkg/api/provider.go:21`: no blank line before the third-party Gin import.
- `repos/gitstate.go:9`: no blank line before the third-party YAML import.
- `repos/kbconfig.go:9`: no blank line before the third-party YAML import.
- `repos/registry.go:13`: no blank line before the third-party YAML import.
- `repos/workconfig.go:9`: no blank line before the third-party YAML import.

## UK English Drift (7)

- `README.md:2`: badge alt text and shield label use `License` rather than `Licence`.
- `CONTRIBUTING.md:34`: section heading uses `License` rather than `Licence`.
- `jobrunner/journal.go:76`: comment uses `normalize` rather than `normalise`.
- `marketplace/marketplace.go:19`: comment uses `catalog` rather than `catalogue`.
- `repos/registry.go:29`: schema field and YAML tag use `License`/`license` rather than `Licence`/`licence`.
- `repos/registry_test.go:48`: fixture uses `license:` rather than `licence:`.
- `docs/architecture.md:508`: `repos.yaml` example uses `license:` rather than `licence:`.

## Missing Tests (50)

### Files or packages with 0% statement coverage, or no tests at all (36)

- `agentci/clotho.go:25`: whole file is at 0% statement coverage.
- `cmd/collect/cmd.go:12`: whole file is at 0% statement coverage.
- `cmd/collect/cmd_bitcointalk.go:16`: whole file is at 0% statement coverage.
- `cmd/collect/cmd_dispatch.go:13`: whole file is at 0% statement coverage.
- `cmd/collect/cmd_excavate.go:19`: whole file is at 0% statement coverage.
- `cmd/collect/cmd_github.go:20`: whole file is at 0% statement coverage.
- `cmd/collect/cmd_market.go:18`: whole file is at 0% statement coverage.
- `cmd/collect/cmd_papers.go:19`: whole file is at 0% statement coverage.
- `cmd/collect/cmd_process.go:12`: whole file is at 0% statement coverage.
- `cmd/forge/cmd_auth.go:17`: whole file is at 0% statement coverage.
- `cmd/forge/cmd_config.go:18`: whole file is at 0% statement coverage.
- `cmd/forge/cmd_forge.go:19`: whole file is at 0% statement coverage.
- `cmd/forge/cmd_issues.go:21`: whole file is at 0% statement coverage.
- `cmd/forge/cmd_labels.go:20`: whole file is at 0% statement coverage.
- `cmd/forge/cmd_migrate.go:21`: whole file is at 0% statement coverage.
- `cmd/forge/cmd_orgs.go:11`: whole file is at 0% statement coverage.
- `cmd/forge/cmd_prs.go:19`: whole file is at 0% statement coverage.
- `cmd/forge/cmd_repos.go:19`: whole file is at 0% statement coverage.
- `cmd/forge/cmd_status.go:11`: whole file is at 0% statement coverage.
- `cmd/forge/cmd_sync.go:25`: whole file is at 0% statement coverage.
- `cmd/forge/helpers.go:11`: whole file is at 0% statement coverage.
- `cmd/gitea/cmd_config.go:18`: whole file is at 0% statement coverage.
- `cmd/gitea/cmd_gitea.go:16`: whole file is at 0% statement coverage.
- `cmd/gitea/cmd_issues.go:21`: whole file is at 0% statement coverage.
- `cmd/gitea/cmd_mirror.go:19`: whole file is at 0% statement coverage.
- `cmd/gitea/cmd_prs.go:19`: whole file is at 0% statement coverage.
- `cmd/gitea/cmd_repos.go:17`: whole file is at 0% statement coverage.
- `cmd/gitea/cmd_sync.go:25`: whole file is at 0% statement coverage.
- `cmd/scm/cmd_compile.go:14`: whole file is at 0% statement coverage.
- `cmd/scm/cmd_export.go:11`: whole file is at 0% statement coverage.
- `cmd/scm/cmd_index.go:11`: whole file is at 0% statement coverage.
- `cmd/scm/cmd_scm.go:14`: whole file is at 0% statement coverage.
- `git/git.go:31`: whole file is at 0% statement coverage.
- `git/service.go:57`: whole file is at 0% statement coverage.
- `jobrunner/handlers/completion.go:23`: whole file is at 0% statement coverage.
- `locales/embed.go:1`: package has no test files.

### 0%-covered functions inside otherwise tested files (14)

- `collect/github.go:123`: `listOrgRepos` has 0% coverage.
- `collect/ratelimit.go:98`: `CheckGitHubRateLimit` has 0% coverage.
- `collect/ratelimit.go:105`: `CheckGitHubRateLimitCtx` has 0% coverage.
- `forge/config.go:73`: `SaveConfig` has 0% coverage.
- `forge/issues.go:129`: `ListPullRequestsIter` has 0% coverage.
- `forge/repos.go:36`: `ListOrgReposIter` has 0% coverage.
- `forge/repos.go:85`: `ListUserReposIter` has 0% coverage.
- `gitea/issues.go:104`: `ListPullRequestsIter` has 0% coverage.
- `gitea/repos.go:36`: `ListOrgReposIter` has 0% coverage.
- `gitea/repos.go:85`: `ListUserReposIter` has 0% coverage.
- `jobrunner/handlers/dispatch.go:272`: `runRemote` has 0% coverage.
- `pkg/api/provider.go:442`: `emitEvent` has 0% coverage.
- `plugin/installer.go:152`: `cloneRepo` has 0% coverage.
- `repos/registry.go:105`: `FindRegistry` has 0% coverage.

## Missing SPDX Headers In Non-Test Go/TS Sources (92)

- `agentci/clotho.go:1`
- `agentci/config.go:1`
- `agentci/security.go:1`
- `cmd/collect/cmd.go:1`
- `cmd/collect/cmd_bitcointalk.go:1`
- `cmd/collect/cmd_dispatch.go:1`
- `cmd/collect/cmd_excavate.go:1`
- `cmd/collect/cmd_github.go:1`
- `cmd/collect/cmd_market.go:1`
- `cmd/collect/cmd_papers.go:1`
- `cmd/collect/cmd_process.go:1`
- `cmd/forge/cmd_auth.go:1`
- `cmd/forge/cmd_config.go:1`
- `cmd/forge/cmd_forge.go:1`
- `cmd/forge/cmd_issues.go:1`
- `cmd/forge/cmd_labels.go:1`
- `cmd/forge/cmd_migrate.go:1`
- `cmd/forge/cmd_orgs.go:1`
- `cmd/forge/cmd_prs.go:1`
- `cmd/forge/cmd_repos.go:1`
- `cmd/forge/cmd_status.go:1`
- `cmd/forge/cmd_sync.go:1`
- `cmd/forge/helpers.go:1`
- `cmd/gitea/cmd_config.go:1`
- `cmd/gitea/cmd_gitea.go:1`
- `cmd/gitea/cmd_issues.go:1`
- `cmd/gitea/cmd_mirror.go:1`
- `cmd/gitea/cmd_prs.go:1`
- `cmd/gitea/cmd_repos.go:1`
- `cmd/gitea/cmd_sync.go:1`
- `cmd/scm/cmd_compile.go:1`
- `cmd/scm/cmd_export.go:1`
- `cmd/scm/cmd_index.go:1`
- `cmd/scm/cmd_scm.go:1`
- `collect/bitcointalk.go:1`
- `collect/collect.go:1`
- `collect/events.go:1`
- `collect/excavate.go:1`
- `collect/github.go:1`
- `collect/market.go:1`
- `collect/papers.go:1`
- `collect/process.go:1`
- `collect/ratelimit.go:1`
- `collect/state.go:1`
- `forge/client.go:1`
- `forge/config.go:1`
- `forge/issues.go:1`
- `forge/labels.go:1`
- `forge/meta.go:1`
- `forge/orgs.go:1`
- `forge/prs.go:1`
- `forge/repos.go:1`
- `forge/webhooks.go:1`
- `git/git.go:1`
- `git/service.go:1`
- `gitea/client.go:1`
- `gitea/config.go:1`
- `gitea/issues.go:1`
- `gitea/meta.go:1`
- `gitea/repos.go:1`
- `jobrunner/forgejo/signals.go:1`
- `jobrunner/forgejo/source.go:1`
- `jobrunner/handlers/completion.go:1`
- `jobrunner/handlers/dispatch.go:1`
- `jobrunner/handlers/enable_auto_merge.go:1`
- `jobrunner/handlers/publish_draft.go:1`
- `jobrunner/handlers/resolve_threads.go:1`
- `jobrunner/handlers/send_fix_command.go:1`
- `jobrunner/handlers/tick_parent.go:1`
- `jobrunner/journal.go:1`
- `jobrunner/poller.go:1`
- `jobrunner/types.go:1`
- `locales/embed.go:1`
- `manifest/compile.go:1`
- `manifest/loader.go:1`
- `manifest/manifest.go:1`
- `manifest/sign.go:1`
- `marketplace/builder.go:1`
- `marketplace/discovery.go:1`
- `marketplace/installer.go:1`
- `marketplace/marketplace.go:1`
- `plugin/config.go:1`
- `plugin/installer.go:1`
- `plugin/loader.go:1`
- `plugin/manifest.go:1`
- `plugin/plugin.go:1`
- `plugin/registry.go:1`
- `repos/gitstate.go:1`
- `repos/kbconfig.go:1`
- `repos/registry.go:1`
- `repos/workconfig.go:1`
- `ui/vite.config.ts:1`
