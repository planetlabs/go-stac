package raster_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/planetlabs/go-stac"
	"github.com/planetlabs/go-stac/extensions/raster/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRasterJSON(t *testing.T) {
	num1 := 20.0
	num2 := 100.0

	cases := []struct {
		name string
		item *stac.Item
		data string
		err  string
	}{
		{
			name: "extended bands",
			item: &stac.Item{
				Version: "1.1.0",
				Id:      "item-id",
				Geometry: map[string]any{
					"type":        "Point",
					"coordinates": []any{1.1, 2.2},
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
						Bands: []*stac.Band{
							{
								NoData: 10.0,
								Statistics: &stac.Statistics{
									Minimum: &num1,
									Maximum: &num2,
								},
								Extensions: []stac.Extension{
									&raster.Band{
										Histogram: &raster.Histogram{
											Count:   3,
											Min:     num1,
											Max:     num2,
											Buckets: []float64{30, 40, 50},
										},
									},
								},
							},
						},
					},
				},
			},
			data: `{
				"type": "Feature",
				"stac_version": "1.1.0",
				"id": "item-id",
				"geometry": {
					"type": "Point",
					"coordinates": [1.1, 2.2]
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
						"bands": [
							{
								"nodata": 10,
								"statistics": {
									"minimum": 20,
									"maximum": 100
								},
								"raster:histogram": {
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
					"https://stac-extensions.github.io/raster/v2.0.0/schema.json"
				]
			}`,
		},
		{
			name: "extended item",
			item: &stac.Item{
				Version: "1.1.0",
				Id:      "item-id",
				Geometry: map[string]any{
					"type":        "Point",
					"coordinates": []any{1.1, 2.2},
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
						Bands: []*stac.Band{
							{
								NoData: 10.0,
								Statistics: &stac.Statistics{
									Minimum: &num1,
									Maximum: &num2,
								},
							},
						},
					},
				},
				Extensions: []stac.Extension{
					&raster.Item{
						Scale:  &num1,
						Offset: &num2,
					},
				},
			},
			data: `{
				"type": "Feature",
				"stac_version": "1.1.0",
				"id": "item-id",
				"geometry": {
					"type": "Point",
					"coordinates": [1.1, 2.2]
				},
				"properties": {
					"test": "value",
					"raster:scale": 20,
					"raster:offset": 100
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
						"bands": [
							{
								"nodata": 10,
								"statistics": {
									"minimum": 20,
									"maximum": 100
								}
							}
						]
					}
				},
				"stac_extensions": [
					"https://stac-extensions.github.io/raster/v2.0.0/schema.json"
				]
			}`,
		},
		{
			name: "extended asset",
			item: &stac.Item{
				Version: "1.1.0",
				Id:      "item-id",
				Geometry: map[string]any{
					"type":        "Point",
					"coordinates": []any{1.1, 2.2},
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
						Bands: []*stac.Band{
							{
								NoData: 10.0,
								Statistics: &stac.Statistics{
									Minimum: &num1,
									Maximum: &num2,
								},
							},
						},
						Extensions: []stac.Extension{
							&raster.Asset{
								Sampling:          "area",
								SpatialResolution: &num1,
							},
						},
					},
				},
			},
			data: `{
				"type": "Feature",
				"stac_version": "1.1.0",
				"id": "item-id",
				"geometry": {
					"type": "Point",
					"coordinates": [1.1, 2.2]
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
						"raster:sampling": "area",
						"raster:spatial_resolution": 20,
						"bands": [
							{
								"nodata": 10,
								"statistics": {
									"minimum": 20,
									"maximum": 100
								}
							}
						]
					}
				},
				"stac_extensions": [
					"https://stac-extensions.github.io/raster/v2.0.0/schema.json"
				]
			}`,
		},
		{
			name: "extended item, asset, and bands",
			item: &stac.Item{
				Version: "1.1.0",
				Id:      "item-id",
				Geometry: map[string]any{
					"type":        "Point",
					"coordinates": []any{1.1, 2.2},
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
						Bands: []*stac.Band{
							{
								NoData: 10.0,
								Statistics: &stac.Statistics{
									Minimum: &num1,
									Maximum: &num2,
								},
								Extensions: []stac.Extension{
									&raster.Band{
										Histogram: &raster.Histogram{
											Count:   3,
											Min:     num1,
											Max:     num2,
											Buckets: []float64{30, 40, 50},
										},
									},
								},
							},
						},
						Extensions: []stac.Extension{
							&raster.Asset{
								Sampling:          "area",
								SpatialResolution: &num1,
							},
						},
					},
				},
				Extensions: []stac.Extension{
					&raster.Item{
						Scale:  &num1,
						Offset: &num2,
					},
				},
			},
			data: `{
				"type": "Feature",
				"stac_version": "1.1.0",
				"id": "item-id",
				"geometry": {
					"type": "Point",
					"coordinates": [1.1, 2.2]
				},
				"properties": {
					"test": "value",
					"raster:scale": 20,
					"raster:offset": 100
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
						"raster:sampling": "area",
						"raster:spatial_resolution": 20,
						"bands": [
							{
								"nodata": 10,
								"statistics": {
									"minimum": 20,
									"maximum": 100
								},
								"raster:histogram": {
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
					"https://stac-extensions.github.io/raster/v2.0.0/schema.json"
				]
			}`,
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("%d-%s", i, c.name), func(t *testing.T) {
			if c.item != nil {
				data, err := json.Marshal(c.item)
				if c.err != "" {
					assert.ErrorContains(t, err, c.err)
					return
				}
				require.NoError(t, err)
				assert.JSONEq(t, c.data, string(data))
			}

			if c.data != "" {
				item := &stac.Item{}
				err := json.Unmarshal([]byte(c.data), item)
				if c.err != "" {
					assert.ErrorContains(t, err, c.err)
					return
				}
				require.NoError(t, err)
				assert.Equal(t, c.item, item)
			}
		})
	}
}
