package validator_test

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"testing"

	"github.com/planetlabs/go-stac/crawler"
	"github.com/planetlabs/go-stac/validator"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/stretchr/testify/suite"
)

func loadSchema(schemaURL string) (io.ReadCloser, error) {
	u, err := url.Parse(schemaURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schema url: %w", err)
	}
	schemaPath := path.Join("testdata", "schema", u.Host, u.Path)
	file, openErr := os.Open(schemaPath)
	if openErr != nil {
		return nil, fmt.Errorf("failed to open schema file: %w", openErr)
	}
	return file, nil
}

type Suite struct {
	suite.Suite
	originalHttpLoader  func(string) (io.ReadCloser, error)
	originalHttpsLoader func(string) (io.ReadCloser, error)
}

func (s *Suite) SetupSuite() {
	s.originalHttpLoader = jsonschema.Loaders["http"]
	s.originalHttpsLoader = jsonschema.Loaders["https"]
	jsonschema.Loaders["http"] = loadSchema
	jsonschema.Loaders["https"] = loadSchema
}

func (s *Suite) TearDownSuite() {
	jsonschema.Loaders["http"] = s.originalHttpLoader
	jsonschema.Loaders["https"] = s.originalHttpsLoader
}

func (s *Suite) TestValidCases() {
	cases := []string{
		"v1.0.0-beta.2/LC08_L1TP_097073_20130319_20200913_02_T1.json",
		"v1.0.0/catalog.json",
		"v1.0.0/collection.json",
		"v1.0.0/item.json",
		"v1.0.0/catalog-with-item.json",
		"v1.0.0/catalog-with-multiple-items.json",
		"v1.0.0/item-eo.json",
	}

	v := validator.New()
	ctx := context.Background()
	for _, c := range cases {
		s.Run(c, func() {
			resourcePath := path.Join("testdata", "cases", c)
			err := v.Validate(ctx, resourcePath)
			s.Assert().NoError(err)
		})
	}
}

func (s *Suite) TestSchemaMap() {
	v := validator.New(&validator.Options{
		Concurrency: crawler.DefaultOptions.Concurrency,
		Recursion:   crawler.DefaultOptions.Recursion,
		SchemaMap: map[string]string{
			"https://stac-extensions.github.io/custom/v1.0.0/schema.json": "https://example.com//extensions/custom.json",
		},
	})
	ctx := context.Background()
	resourcePath := path.Join("testdata", "cases", "v1.0.0", "item-custom.json")
	err := v.Validate(ctx, resourcePath)
	s.Assert().NoError(err)
}

func TestSuite(t *testing.T) {
	suite.Run(t, &Suite{})
}

func ExampleValidator_Validate_children() {
	v := validator.New(&validator.Options{
		Recursion: crawler.Children,
	})

	err := v.Validate(context.Background(), "testdata/cases/v1.0.0/catalog-with-item-missing-id.json")
	fmt.Printf("%#v\n", err)
	// Output:
	// invalid item: testdata/cases/v1.0.0/item-missing-id.json
	// [I#] [S#] doesn't validate with https://schemas.stacspec.org/v1.0.0/item-spec/json-schema/item.json#
	//   [I#] [S#/allOf/0] allOf failed
	//     [I#] [S#/allOf/0/$ref] doesn't validate with '/definitions/core'
	//       [I#] [S#/definitions/core/allOf/2] allOf failed
	//         [I#] [S#/definitions/core/allOf/2/required] missing properties: 'id'
}

func ExampleValidator_Validate_single() {
	v := validator.New(&validator.Options{
		Recursion: crawler.None,
	})

	err := v.Validate(context.Background(), "testdata/cases/v1.0.0/item-missing-id.json")
	fmt.Printf("%#v\n", err)
	// Output:
	// invalid item: testdata/cases/v1.0.0/item-missing-id.json
	// [I#] [S#] doesn't validate with https://schemas.stacspec.org/v1.0.0/item-spec/json-schema/item.json#
	//   [I#] [S#/allOf/0] allOf failed
	//     [I#] [S#/allOf/0/$ref] doesn't validate with '/definitions/core'
	//       [I#] [S#/definitions/core/allOf/2] allOf failed
	//         [I#] [S#/definitions/core/allOf/2/required] missing properties: 'id'
}
