module github.com/planetlabs/go-stac

go 1.18

require (
	github.com/dlclark/regexp2 v1.7.0
	github.com/go-logr/logr v1.2.3
	github.com/go-logr/zapr v1.2.3
	github.com/google/go-github/v45 v45.2.0
	github.com/hashicorp/go-retryablehttp v0.7.1
	github.com/santhosh-tekuri/jsonschema/v5 v5.0.0
	github.com/schollz/progressbar/v3 v3.10.1
	github.com/stretchr/testify v1.8.0
	github.com/tschaub/retry v1.0.0
	github.com/urfave/cli/v2 v2.17.1
	go.uber.org/zap v1.23.0
	golang.org/x/sync v0.0.0-20220513210516-0976fa681c29
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/crypto v0.0.0-20220131195533-30dcbda58838 // indirect
	golang.org/x/sys v0.0.0-20220829200755-d48e67d00261 // indirect
	golang.org/x/term v0.0.0-20220722155259-a9ba230a4035 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/santhosh-tekuri/jsonschema/v5 v5.0.0 => github.com/tschaub/jsonschema/v5 v5.1.0-beta.1
