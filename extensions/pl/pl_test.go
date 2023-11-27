package pl_test

import (
	"encoding/json"
	"testing"

	"github.com/planetlabs/go-stac"
	"github.com/planetlabs/go-stac/extensions/pl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestItemExtendedMarshal(t *testing.T) {
	groundControlRatio := 0.5

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
				Extensions: []stac.Extension{
					&pl.Asset{
						AssetType: "visual",
					},
				},
			},
		},
		Extensions: []stac.Extension{
			&pl.Item{
				ItemType:           "PSScene",
				PixelResolution:    3,
				QualityCategory:    "test",
				StripId:            "123",
				GroundControlRatio: &groundControlRatio,
			},
		},
	}

	data, err := json.Marshal(item)
	require.NoError(t, err)

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
			"pl:item_type": "PSScene",
			"pl:pixel_resolution": 3,
			"pl:quality_category": "test",
			"pl:strip_id": "123",
			"pl:ground_control_ratio": 0.5
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
				"type": "image/png",
				"pl:asset_type": "visual"
			}
		},
		"stac_extensions": [
			"https://planetlabs.github.io/stac-extension/v1.0.0-beta.1/schema.json"
		]
	}`

	assert.JSONEq(t, expected, string(data))
}

func TestItemMarshalGridCell(t *testing.T) {
	gridCell := "1259913"

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
				Extensions: []stac.Extension{
					&pl.Asset{
						AssetType: "visual",
					},
				},
			},
		},
		Extensions: []stac.Extension{
			&pl.Item{
				ItemType: "REOrthoTile",
				GridCell: &gridCell,
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
			"pl:item_type": "REOrthoTile",
			"pl:grid_cell": "1259913"
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
				"type": "image/png",
				"pl:asset_type": "visual"
			}
		},
		"stac_extensions": [
			"https://planetlabs.github.io/stac-extension/v1.0.0-beta.1/schema.json"
		]
	}`

	assert.JSONEq(t, expected, string(data))
}

func TestItemExtendedUnmarshal(t *testing.T) {
	data := []byte(`{
		"type": "Feature",
		"stac_version": "1.0.0",
		"id": "item-id",
		"geometry": {
			"type": "Point",
			"coordinates": [0, 0]
		},
		"properties": {
			"test": "value",
			"pl:item_type": "PSScene",
			"pl:pixel_resolution": 3,
			"pl:quality_category": "test",
			"pl:strip_id": "123",
			"pl:ground_control_ratio": 0.5
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
				"type": "image/png",
				"pl:asset_type": "visual"
			}
		},
		"stac_extensions": [
			"https://planetlabs.github.io/stac-extension/v1.0.0-beta.1/schema.json"
		]
	}`)

	item := &stac.Item{}
	require.NoError(t, json.Unmarshal(data, item))

	groundControlRatio := 0.5

	expected := &stac.Item{
		Version: "1.0.0",
		Id:      "item-id",
		Geometry: map[string]any{
			"type":        "Point",
			"coordinates": []any{float64(0), float64(0)},
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
				Extensions: []stac.Extension{
					&pl.Asset{
						AssetType: "visual",
					},
				},
			},
		},
		Extensions: []stac.Extension{
			&pl.Item{
				ItemType:           "PSScene",
				PixelResolution:    3,
				QualityCategory:    "test",
				StripId:            "123",
				GroundControlRatio: &groundControlRatio,
			},
		},
	}

	assert.Equal(t, expected, item)
}
