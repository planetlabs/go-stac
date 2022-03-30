package validator_test

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"testing"

	"github.com/planetlabs/go-stac/pkg/crawler"
	"github.com/planetlabs/go-stac/pkg/validator"
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
		"v1.0.0/catalog.json",
		"v1.0.0/collection.json",
		"v1.0.0/item.json",
		"v1.0.0/catalog-with-item.json",
		"v1.0.0/catalog-with-multiple-items.json",
		"v1.0.0/item-eo.json",
	}

	v := validator.New(&validator.Options{
		Concurrency: crawler.DefaultOptions.Concurrency,
		Recursion:   crawler.DefaultOptions.Recursion,
	})
	ctx := context.Background()
	for _, c := range cases {
		s.Run(c, func() {
			resourcePath := path.Join("testdata", "cases", c)
			err := v.Validate(ctx, resourcePath)
			s.Assert().NoError(err)
		})
	}
}

func TestSuite(t *testing.T) {
	suite.Run(t, &Suite{})
}
