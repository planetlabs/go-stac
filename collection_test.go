package stac_test

import (
	"encoding/json"
	"testing"

	"github.com/planetlabs/go-stac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectionMarshal(t *testing.T) {
	collection := &stac.Collection{
		Version:     "1.0.0",
		Id:          "collection-id",
		Description: "Test Collection",
		License:     "various",
		Links: []*stac.Link{
			{Href: "https://example.com/stac/collections/collection-id", Rel: "self", Type: "application/json"},
		},
		Extent: &stac.Extent{
			Spatial: &stac.SpatialExtent{
				Bbox: [][]float64{{-180, -90, 180, 90}},
			},
		},
	}

	data, err := json.Marshal(collection)
	require.Nil(t, err)

	expected := `{
		"type": "Collection",
		"id": "collection-id",
		"description": "Test Collection",
		"extent": {
			"spatial": {
				"bbox": [
					[-180, -90, 180, 90]
				]
			}
		},
		"license": "various",
		"links": [
			{
				"href": "https://example.com/stac/collections/collection-id",
				"rel": "self",
				"type": "application/json"
			}
		],
		"stac_version": "1.0.0"
	}`

	assert.JSONEq(t, expected, string(data))
}
