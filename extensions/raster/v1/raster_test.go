package raster_test

import (
	"encoding/json"
	"testing"

	"github.com/planetlabs/go-stac"
	"github.com/planetlabs/go-stac/extensions/raster/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestItemExtendedMarshal(t *testing.T) {
	min := float64(20)
	max := float64(100)
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
				Title: "Image",
				Href:  "https://example.com/stac/item-id/image.tif",
				Type:  "image/tif",
				Extensions: []stac.Extension{
					&raster.Asset{
						Bands: []*raster.Band{
							{
								NoData: 10,
								Statistics: &raster.Statistics{
									Minimum: &min,
									Maximum: &max,
								},
								Histogram: &raster.Histogram{
									Count:   3,
									Min:     min,
									Max:     max,
									Buckets: []float64{30, 40, 50},
								},
							},
						},
					},
				},
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
				"title": "Image",
				"href": "https://example.com/stac/item-id/image.tif",
				"type": "image/tif",
				"raster:bands": [
					{
						"nodata": 10,
						"statistics": {
							"minimum": 20,
							"maximum": 100
						},
						"histogram": {
							"min": 20,
							"max": 100,
							"count": 3,
							"buckets": [30, 40, 50]
						}
					}
				]
			}
		},
		"stac_extensions": [
			"https://stac-extensions.github.io/raster/v1.1.0/schema.json"
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
				"title": "Image",
				"href": "https://example.com/stac/item-id/image.tif",
				"type": "image/tif",
				"raster:bands": [
					{
						"nodata": 10,
						"statistics": {
							"minimum": 20,
							"maximum": 100
						},
						"histogram": {
							"min": 20,
							"max": 100,
							"count": 3,
							"buckets": [30, 40, 50]
						}
					}
				]
			}
		},
		"stac_extensions": [
			"https://stac-extensions.github.io/raster/v1.1.0/schema.json"
		]
	}`)

	item := &stac.Item{}
	require.NoError(t, json.Unmarshal(data, item))

	min := float64(20)
	max := float64(100)
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
			"image": {
				Title: "Image",
				Href:  "https://example.com/stac/item-id/image.tif",
				Type:  "image/tif",
				Extensions: []stac.Extension{
					&raster.Asset{
						Bands: []*raster.Band{
							{
								NoData: float64(10),
								Statistics: &raster.Statistics{
									Minimum: &min,
									Maximum: &max,
								},
								Histogram: &raster.Histogram{
									Count:   3,
									Min:     min,
									Max:     max,
									Buckets: []float64{30, 40, 50},
								},
							},
						},
					},
				},
			},
		},
	}

	assert.Equal(t, expected, item)
}
