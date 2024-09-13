package validator_test

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"

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

func (s *Suite) TestValidateBytes() {
	cases := []string{
		"v1.0.0-beta.2/LC08_L1TP_097073_20130319_20200913_02_T1.json",
		"v1.0.0/catalog.json",
		"v1.0.0/collection.json",
		"v1.0.0/item.json",
		"v1.0.0/catalog-with-item.json",
		"v1.0.0/catalog-with-multiple-items.json",
		"v1.0.0/item-eo.json",
	}

	ctx := context.Background()
	for _, c := range cases {
		s.Run(c, func() {
			location := path.Join("testdata", "cases", c)
			data, err := os.ReadFile(location)
			s.Require().NoError(err)
			s.NoError(validator.ValidateBytes(ctx, data, location))
		})
	}
}

func (s *Suite) TestValidateBytesInvalidItem() {
	location := "testdata/cases/v1.0.0/item-missing-id.json"

	data, readErr := os.ReadFile(location)
	s.Require().NoError(readErr)
	ctx := context.Background()

	err := validator.ValidateBytes(ctx, data, location)
	s.Require().Error(err)
	s.Assert().True(strings.HasSuffix(fmt.Sprintf("%#v", err), "missing properties: 'id'"))
}

func (s *Suite) TestSchemaMap() {
	v := validator.New(&validator.Options{
		SchemaMap: map[string]string{
			"https://stac-extensions.github.io/custom/v1.0.0/schema.json": "https://example.com//extensions/custom.json",
		},
	})
	ctx := context.Background()
	resourcePath := path.Join("testdata", "cases", "v1.0.0", "item-custom.json")
	err := v.Validate(ctx, resourcePath)
	s.Assert().NoError(err)
}

func (s *Suite) TestCatalogWithInvalidItem() {
	v := validator.New()

	err := v.Validate(context.Background(), "testdata/cases/v1.0.0/catalog-with-item-missing-id.json")
	s.Require().Error(err)
	s.Assert().True(strings.HasSuffix(fmt.Sprintf("%#v", err), "missing properties: 'id'"))
}

func (s *Suite) TestInvalidItem() {
	v := validator.New()

	err := v.Validate(context.Background(), "testdata/cases/v1.0.0/item-missing-id.json")
	s.Require().Error(err)
	s.Assert().True(strings.HasSuffix(fmt.Sprintf("%#v", err), "missing properties: 'id'"))
}

func TestSuite(t *testing.T) {
	suite.Run(t, &Suite{})
}
