# Verification Pass 2026-03-27

- Repository note: `CODEX.md` was not present under `/workspace`; conventions were taken from `CLAUDE.md`.
- Commands run: `go build ./...`, `go vet ./...`, `go test ./...`
- Command status: all passed

## Banned imports

ZERO FINDINGS for banned imports: `os`, `os/exec`, `encoding/json`, `fmt`, `errors`, `strings`, `path/filepath`.

## Test names not matching `TestFile_Function_{Good,Bad,Ugly}`

597 findings across 67 files:

```text
agentci/config_test.go:23,46,67,85,102,120,127,140,162,177,185,205,223,237,257,270,277,295,302
agentci/security_test.go:12,34,61,79,92,109
cmd/forge/cmd_sync_test.go:11,21,29,39,47
cmd/gitea/cmd_sync_test.go:11,21,29,39,47
collect/bitcointalk_http_test.go:36,76,104,128,152,183,192,210,225
collect/bitcointalk_test.go:16,21,30,42,73,79,87
collect/collect_test.go:10,23,35,57,63
collect/coverage_boost_test.go:18,34,50,62,79,92,109,120,125,130,141,151,176,198,230,264,275,283,291,298,305,313,322,330,335,343,351,358,368,384,400,419,434,452,457,481,497,508,517,533,557,572,592,611,625,639
collect/coverage_phase2_test.go:87,99,115,134,149,167,182,202,216,232,257,281,305,329,353,388,416,444,480,503,530,565,578,594,619,633,646,665,692,718,738,751,764,778,797,807,817,826,835,846,856,866,876,886,896,928,945,962,1022,1059,1098,1110,1132,1157,1181,1193,1230,1246,1261,1300
collect/events_test.go:44,57,72,88,129
collect/excavate_extra_test.go:13,42,68,86,104
collect/excavate_test.go:67,78,99,124,147,164,185
collect/github_test.go:17,22,41,53,65,91
collect/market_extra_test.go:15,59,101,140,175,200,232
collect/market_test.go:19,28,40,95,145,167
collect/papers_http_test.go:112,168,188,208,229,249,281,298,306
collect/papers_test.go:16,21,26,35,44,56,75,83
collect/process_extra_test.go:12,21,29,36,43,51,60,68,76,84,92,101,108,118,134,158,175
collect/process_test.go:16,25,37,58,78,97,114,173,182,191,198
collect/ratelimit_test.go:30,54,63,69,78
collect/state_extra_test.go:11,27,43,56
collect/state_test.go:76,89,98,126,138
forge/client_test.go:16,28,64,92,100,124,135,159,185,211,225,263,326,356,387,409,435,441
forge/config_test.go:19,28,39,50,61,71,77,85,100
forge/issues_test.go:22,44,55,73,94,116,135,154,176,194,211,230,247
forge/labels_test.go:27,45,63,72,81,91,111,127,144
forge/meta_test.go:26,48,66
forge/orgs_test.go:22,40,62
forge/prs_test.go:22,30,38,61,79,96,105,133,142
forge/repos_test.go:22,42,60,81,100,122
forge/webhooks_test.go:26,46
gitea/client_test.go:10,21
gitea/config_test.go:18,27,38,49,59,69,75,83,95
gitea/coverage_boost_test.go:15,26,35,44,90,124,176,220,239,264,294,306
gitea/issues_test.go:22,44,55,73,94,115,137,155,166
gitea/meta_test.go:26,48,66,77,103,115
gitea/repos_test.go:22,42,60,69,79,89,106,127
jobrunner/forgejo/source_test.go:38,109,155,162,166
jobrunner/handlers/dispatch_test.go:38,50,63,76,88,100,120,145,210,232,255,276,302,339
jobrunner/handlers/enable_auto_merge_test.go:29,42,84
jobrunner/handlers/publish_draft_test.go:26,36
jobrunner/handlers/resolve_threads_test.go:26
jobrunner/handlers/send_fix_command_test.go:16,25,37,49
jobrunner/handlers/tick_parent_test.go:26
jobrunner/journal_test.go:116,198,233,249
jobrunner/poller_test.go:121,152,193
jobrunner/types_test.go:28
manifest/compile_test.go:14,38,61,67,74,81,106,123,128,152,163,169
manifest/loader_test.go:12,28,34,51
manifest/manifest_test.go:10,49,67,127,135,146,157,205,210,215,220
manifest/sign_test.go:11,32,44
marketplace/builder_test.go:39,56,71,87,102,121,138,147,154,166,178,189,198,221
marketplace/discovery_test.go:25,59,84,105,122,130,136,147,195,234
marketplace/installer_test.go:78,104,125,142,166,189,212,223,243,270,281,309
marketplace/marketplace_test.go:10,26,38,50,61
pkg/api/provider_handlers_test.go:20,47,66,79,90,105,118,131,142,155,169,183
pkg/api/provider_security_test.go:12,18,23
pkg/api/provider_test.go:92,116,167,181,200,230
plugin/installer_test.go:14,26,36,47,60,81,95,105,120,132,140,148,156,162,168,174,180,186,192,198,204
plugin/loader_test.go:46,71,92,123,132
plugin/manifest_test.go:10,33,50,58,78,89,100
plugin/plugin_test.go:10,25,37
plugin/registry_test.go:28,69,78,129,138,149,160,173
repos/gitstate_test.go:14,50,61,110,132,158,178,189,199,204,210
repos/kbconfig_test.go:13,48,57,76,93
repos/registry_test.go:13,40,72,93,120,127,181,201,209,243,266,286,309,334,350,363,373,382,390,398,403,411,421,427,474,481
repos/workconfig_test.go:14,51,60,79,97,102
```

## Exported functions missing usage-example comments

199 findings across 51 files:

```text
agentci/clotho.go:39,61,71,90
collect/bitcointalk.go:37,46
collect/events.go:72,81,96,105,115,125,135
collect/excavate.go:27,33
collect/github.go:57,65
collect/market.go:33,67
collect/papers.go:41,57
collect/process.go:27,33
collect/ratelimit.go:46,80,87,100,107
collect/state.go:55,81,99,112
forge/client.go:37,40,43,46,55,69
forge/issues.go:21,56,66,76,86,97,130,164,174,185,209
forge/labels.go:15,31,55,65,81,94,105
forge/meta.go:43,100,139
forge/orgs.go:10,34,44
forge/prs.go:18,43,84,108,117
forge/repos.go:12,36,61,85,110,120,130,141
forge/webhooks.go:10,20
git/git.go:32,37,42,283,293
git/service.go:75,113,116,121,132,145,156
gitea/client.go:36,39
gitea/issues.go:20,52,62,72,105,139
gitea/meta.go:43,101,141
gitea/repos.go:12,36,61,85,110,122,145,155
internal/ax/stdio/stdio.go:12,24
jobrunner/forgejo/source.go:36,42,65
jobrunner/handlers/completion.go:33,38,43
jobrunner/handlers/dispatch.go:76,82,91
jobrunner/handlers/enable_auto_merge.go:25,31,40
jobrunner/handlers/publish_draft.go:25,30,37
jobrunner/handlers/resolve_threads.go:30,35,40
jobrunner/handlers/send_fix_command.go:26,32,46
jobrunner/handlers/tick_parent.go:30,35,41
jobrunner/journal.go:98
jobrunner/poller.go:51,58,65,72,79,88,110
jobrunner/types.go:37,42
manifest/manifest.go:46,79,95
marketplace/builder.go:35
marketplace/discovery.go:132,140,145,151,160
marketplace/installer.go:53,115,133,179
marketplace/marketplace.go:39,53,64
pkg/api/provider.go:61,64,67,75,86,108
plugin/installer.go:35,99,133
plugin/loader.go:28,53
plugin/manifest.go:41
plugin/plugin.go:44,47,50,53,56
plugin/registry.go:35,48,54,63,78,105
repos/gitstate.go:101,106,111,120,131,144,162
repos/kbconfig.go:116,121
repos/registry.go:255,265,271,283,325,330
repos/workconfig.go:104
```
