# Development

Run module commands from `go/` with workspace mode disabled for release-compatible checks:

```sh
GOWORK=off GOPROXY=direct GOSUMDB=off go build ./...
GOWORK=off GOPROXY=direct GOSUMDB=off go vet ./...
GOWORK=off GOPROXY=direct GOSUMDB=off go test -count=1 -short ./...
```

Run the v0.9.0 audit from the repository root:

```sh
bash /Users/snider/Code/core/go/tests/cli/v090-upgrade/audit.sh .
```
