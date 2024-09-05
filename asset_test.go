package stac_test

import (
	"encoding/json"
	"testing"

	"github.com/planetlabs/go-stac"
	"github.com/planetlabs/go-stac/extensions/pl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssetMarshal(t *testing.T) {
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
			"image": {
				Title: "An Image",
				Href:  "https://example.com/image.tif",
				Type:  "image/tiff",
				Roles: []string{"data", "reflectance"},
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
			"image": {
				"title": "An Image",
				"href": "https://example.com/image.tif",
				"type": "image/tiff",
				"roles": ["data", "reflectance"]
			}
		}
	}`

	assert.JSONEq(t, expected, string(data))
}

func TestAssetExtendedMarshal(t *testing.T) {
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
			"extended": {
				Href: "https://example.com/image.tif",
				Type: "image/tiff",
				Extensions: []stac.Extension{
					&pl.Asset{
						AssetType:  "ortho_analytic_4b_sr",
						BundleType: "analytic_sr_udm2",
					},
				},
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
			"extended": {
				"href": "https://example.com/image.tif",
				"type": "image/tiff",
				"pl:asset_type": "ortho_analytic_4b_sr",
				"pl:bundle_type": "analytic_sr_udm2"
			}
		},
		"stac_extensions": [
			"https://planetlabs.github.io/stac-extension/v1.0.0-beta.3/schema.json"
		]
	}`

	assert.JSONEq(t, expected, string(data))
}
