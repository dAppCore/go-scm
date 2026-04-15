module dappco.re/go/scm

go 1.22

require (
	code.gitea.io/sdk/gitea v0.24.1
	codeberg.org/forgejo/go-sdk v0.0.0
	dappco.re/go/core/config v0.0.0
	github.com/stretchr/testify v1.11.1
	golang.org/x/net v0.53.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
)

replace dappco.re/go/core/config => ./core/config

replace codeberg.org/forgejo/go-sdk => ./third_party/forgejo
