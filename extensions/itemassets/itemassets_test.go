package itemassets_test

import (
	"encoding/json"
	"testing"

	"github.com/planetlabs/go-stac"
	"github.com/planetlabs/go-stac/extensions/itemassets"
	"github.com/planetlabs/go-stac/extensions/raster"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectionExtendedMarshal(t *testing.T) {
	collection := &stac.Collection{
		Version:     "1.0.0",
		Id:          "collection-id",
		Description: "Test Collection",
		License:     "various",
		Extent: &stac.Extent{
			Spatial: &stac.SpatialExtent{
				Bbox: [][]float64{{-180, -90, 180, 90}},
			},
		},
		Links: []*stac.Link{
			{Href: "https://example.com/stac/collections/collection-id", Rel: "self", Type: "application/json"},
		},
		Extensions: []stac.Extension{
			&itemassets.Collection{
				ItemAssets: map[string]*stac.Asset{
					"vod": {
						Title: "Asset for Vegetation Optical Depth",
						Type:  "image/tiff",
						Roles: []string{"data"},
						Extensions: []stac.Extension{
							&raster.Asset{
								Bands: []*raster.Band{
									{
										NoData:   65535,
										DataType: "uint16",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	data, err := json.Marshal(collection)
	require.NoError(t, err)

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
		"item_assets": {
			"vod": {
				"title": "Asset for Vegetation Optical Depth",
				"type": "image/tiff",
				"roles": ["data"],
				"raster:bands": [
					{
						"nodata": 65535,
						"data_type": "uint16"
					}
				]
			}
		},
		"links": [
			{
				"href": "https://example.com/stac/collections/collection-id",
				"rel": "self",
				"type": "application/json"
			}
		],
		"stac_extensions": [
			"https://stac-extensions.github.io/item-assets/v1.0.0/schema.json",
			"https://stac-extensions.github.io/raster/v1.1.0/schema.json"
		],
		"stac_version": "1.0.0"
	}`

	assert.JSONEq(t, expected, string(data))
}

func TestCollectionExtendedUnmarshal(t *testing.T) {
	data := []byte(`{
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
		"item_assets": {
			"vod": {
				"title": "Asset for Vegetation Optical Depth",
				"type": "image/tiff",
				"roles": ["data"],
				"raster:bands": [
					{
						"nodata": 65535,
						"data_type": "uint16"
					}
				]
			}
		},
		"links": [
			{
				"href": "https://example.com/stac/collections/collection-id",
				"rel": "self",
				"type": "application/json"
			}
		],
		"stac_extensions": [
			"https://stac-extensions.github.io/item-assets/v1.0.0/schema.json",
			"https://stac-extensions.github.io/raster/v1.1.0/schema.json"
		],
		"stac_version": "1.0.0"
	}`)

	collection := &stac.Collection{}
	require.NoError(t, json.Unmarshal(data, collection))

	expected := &stac.Collection{
		Version:     "1.0.0",
		Id:          "collection-id",
		Description: "Test Collection",
		License:     "various",
		Extent: &stac.Extent{
			Spatial: &stac.SpatialExtent{
				Bbox: [][]float64{{-180, -90, 180, 90}},
			},
		},
		Links: []*stac.Link{
			{Href: "https://example.com/stac/collections/collection-id", Rel: "self", Type: "application/json"},
		},
		Extensions: []stac.Extension{
			&itemassets.Collection{
				ItemAssets: map[string]*stac.Asset{
					"vod": {
						Title: "Asset for Vegetation Optical Depth",
						Type:  "image/tiff",
						Roles: []string{"data"},
						Extensions: []stac.Extension{
							&raster.Asset{
								Bands: []*raster.Band{
									{
										NoData:   float64(65535),
										DataType: "uint16",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	assert.Equal(t, expected, collection)
}
