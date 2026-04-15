module dappco.re/go/scm

go 1.22

require (
	dappco.re/go/core/config v0.0.0
	github.com/stretchr/testify v1.11.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
)

replace dappco.re/go/core/config => ./core/config
