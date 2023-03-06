package stac_test

import (
	"encoding/json"
	"testing"

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
