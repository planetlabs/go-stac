package view_test

import (
	"encoding/json"
	"testing"

	"github.com/planetlabs/go-stac"
	"github.com/planetlabs/go-stac/extensions/view"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestItemExtendedMarshal(t *testing.T) {
	offNadir := 12.3
	sunAzimuth := 145.4
	sunElevation := 17.7

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
		Extensions: []stac.ItemExtension{
			&view.Item{
				OffNadir:     &offNadir,
				SunAzimuth:   &sunAzimuth,
				SunElevation: &sunElevation,
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
			"test": "value",
			"view:off_nadir": 12.3,
			"view:sun_azimuth": 145.4,
			"view:sun_elevation": 17.7
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
		},
		"stac_extensions": [
			"https://stac-extensions.github.io/view/v1.0.0/schema.json"
		]
	}`

	assert.JSONEq(t, expected, string(data))
}
