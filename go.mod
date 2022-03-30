module github.com/planetlabs/go-stac

go 1.18

require (
	github.com/dlclark/regexp2 v1.4.0
	github.com/santhosh-tekuri/jsonschema/v5 v5.0.0
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.1
	github.com/tschaub/workgroup v0.2.0
	github.com/urfave/cli/v2 v2.4.0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/tschaub/limited v0.2.0 // indirect
	golang.org/x/sys v0.0.0-20220330033206-e17cdc41300f // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
)

replace github.com/santhosh-tekuri/jsonschema/v5 v5.0.0 => github.com/tschaub/jsonschema/v5 v5.1.0-beta.1
