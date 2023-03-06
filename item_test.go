package stac_test

import (
	"encoding/json"
	"testing"

	"github.com/planetlabs/go-stac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestItemMarshal(t *testing.T) {
	item := &stac.Item{
		Version: "1.0.0",
		Id:      "item-id",
		Geometry: map[string]any{
			"type":        "Point",
			"coordinates": []float64{0, 0},
		},
		Properties: map[string]any{
			"test": "value",
		},
		Links: []*stac.Link{
			{Href: "https://example.com/stac/item-id", Rel: "self"},
		},
		Assets: map[string]*stac.Asset{
			"thumbnail": {
				Title: "Thumbnail",
				Href:  "https://example.com/stac/item-id/thumb.png",
				Type:  "image/png",
			},
		},
	}

	data, err := json.Marshal(item)
	require.Nil(t, err)

	expected := `{
		"type": "Feature",
		"stac_version": "1.0.0",
		"id": "item-id",
		"geometry": {
			"type": "Point",
			"coordinates": [0, 0]
		},
		"properties": {
			"test": "value"
		},
		"links": [
			{
				"rel": "self",
				"href": "https://example.com/stac/item-id"
			}
		],
		"assets": {
			"thumbnail": {
				"title": "Thumbnail",
				"href": "https://example.com/stac/item-id/thumb.png",
				"type": "image/png"
			}
		}
	}`

	assert.JSONEq(t, expected, string(data))
}
