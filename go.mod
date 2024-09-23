module github.com/planetlabs/go-stac

go 1.21
toolchain go1.22.5

require (
	github.com/dlclark/regexp2 v1.11.4
	github.com/go-logr/logr v1.4.2
	github.com/go-logr/zapr v1.3.0
	github.com/go-viper/mapstructure/v2 v2.1.0
	github.com/google/go-github/v51 v51.0.0
	github.com/hashicorp/go-retryablehttp v0.7.7
	github.com/santhosh-tekuri/jsonschema/v5 v5.3.1
	github.com/schollz/progressbar/v3 v3.16.0
	github.com/stretchr/testify v1.9.0
	github.com/tschaub/retry v1.0.0
	github.com/urfave/cli/v2 v2.27.4
	go.uber.org/zap v1.27.0
	golang.org/x/sync v0.8.0
)

require (
	github.com/ProtonMail/go-crypto v0.0.0-20230217124315-7d5c6f04bbb8 // indirect
	github.com/cloudflare/circl v1.3.7 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.4 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/net v0.23.0 // indirect
	golang.org/x/oauth2 v0.6.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/term v0.24.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/santhosh-tekuri/jsonschema/v5 => github.com/tschaub/jsonschema/v5 v5.1.0-beta.1
