package stac_test

import (
	"encoding/json"
	"regexp"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/planetlabs/go-stac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCatalogMarshal(t *testing.T) {
	catalog := &stac.Catalog{
		Version:     "1.0.0",
		Id:          "catalog-id",
		Description: "Test Catalog",
		Links: []*stac.Link{
			{Href: "https://example.com/stac/catalog", Rel: "self", Type: "application/json"},
		},
	}

	data, err := json.Marshal(catalog)
	require.Nil(t, err)

	expected := `{
		"type": "Catalog",
		"id": "catalog-id",
		"description": "Test Catalog",
		"links": [
			{
				"href": "https://example.com/stac/catalog",
				"rel": "self",
				"type": "application/json"
			}
		],
		"stac_version": "1.0.0"
	}`

	assert.JSONEq(t, expected, string(data))
}

const (
	extensionAlias   = "test-catalog-extension"
	extensionUri     = "https://example.com/test-catalog-extension/v1.0.0/schema.json"
	extensionPattern = `https://example.com/test-catalog-extension/v1\..*/schema.json`
)

type CatalogExtension struct {
	RequiredNum  float64 `json:"required_num"`
	OptionalBool *bool   `json:"optional_bool,omitempty"`
}

var _ stac.Extension = (*CatalogExtension)(nil)

func (*CatalogExtension) URI() string {
	return extensionUri
}

func (e *CatalogExtension) Encode(catalogMap map[string]any) error {
	extendedProps := map[string]any{}
	encoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &extendedProps,
	})
	if err != nil {
		return err
	}
	if err := encoder.Decode(e); err != nil {
		return err
	}
	catalogMap[extensionAlias] = extendedProps
	return nil
}

func (e *CatalogExtension) Decode(catalogMap map[string]any) error {
	extendedProps, present := catalogMap[extensionAlias]
	if !present {
		return nil
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  e,
	})
	if err != nil {
		return err
	}
	return decoder.Decode(extendedProps)
}

func TestExtendedCatalogMarshal(t *testing.T) {
	stac.RegisterCatalogExtension(
		regexp.MustCompile(extensionPattern),
		func() stac.Extension {
			return &CatalogExtension{}
		},
	)

	catalog := &stac.Catalog{
		Description: "Test catalog with extension",
		Id:          "catalog-id",
		Extensions: []stac.Extension{
			&CatalogExtension{
				RequiredNum: 42,
			},
		},
		Links:   []*stac.Link{},
		Version: "1.2.3",
	}

	data, err := json.Marshal(catalog)
	require.NoError(t, err)

	expected := `{
		"type": "Catalog",
		"description": "Test catalog with extension",
		"id": "catalog-id",
		"test-catalog-extension": {
			"required_num": 42
		},
		"links": [],
		"stac_extensions": [
			"https://example.com/test-catalog-extension/v1.0.0/schema.json"
		],
		"stac_version": "1.2.3"
	}`

	assert.JSONEq(t, expected, string(data))
}

func TestExtendedCatalogUnmarshal(t *testing.T) {
	stac.RegisterCatalogExtension(
		regexp.MustCompile(extensionPattern),
		func() stac.Extension {
			return &CatalogExtension{}
		},
	)

	data := []byte(`{
		"type": "Catalog",
		"description": "Test catalog with extension",
		"id": "catalog-id",
		"test-catalog-extension": {
			"required_num": 100,
			"optional_bool": true
		},
		"links": [],
		"stac_extensions": [
			"https://example.com/test-catalog-extension/v1.0.0/schema.json"
		],
		"stac_version": "1.2.3"
	}`)

	catalog := &stac.Catalog{}
	require.NoError(t, json.Unmarshal(data, catalog))

	b := true
	expected := &stac.Catalog{
		Description: "Test catalog with extension",
		Id:          "catalog-id",
		Extensions: []stac.Extension{
			&CatalogExtension{
				RequiredNum:  100,
				OptionalBool: &b,
			},
		},
		Links:   []*stac.Link{},
		Version: "1.2.3",
	}

	assert.Equal(t, expected, catalog)
}
