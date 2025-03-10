package eo_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/planetlabs/go-stac"
	"github.com/planetlabs/go-stac/extensions/eo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEOItemJSON(t *testing.T) {
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
								Extensions: []stac.Extension{
									&eo.Band{
										CommonName:       "red",
										FullWidthHalfMax: &num1,
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
								"eo:common_name": "red",
								"eo:full_width_half_max": 20
							}
						]
					}
				},
				"stac_extensions": [
					"https://stac-extensions.github.io/eo/v2.0.0/schema.json"
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
							},
						},
					},
				},
				Extensions: []stac.Extension{
					&eo.Item{
						CloudCover: &num1,
						SnowCover:  &num2,
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
					"eo:cloud_cover": 20,
					"eo:snow_cover": 100
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
								"nodata": 10
							}
						]
					}
				},
				"stac_extensions": [
					"https://stac-extensions.github.io/eo/v2.0.0/schema.json"
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
							},
						},
						Extensions: []stac.Extension{
							&eo.Asset{
								CenterWavelength:  &num2,
								SolarIllumination: &num1,
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
						"eo:center_wavelength": 100,
						"eo:solar_illumination": 20,
						"bands": [
							{
								"nodata": 10
							}
						]
					}
				},
				"stac_extensions": [
					"https://stac-extensions.github.io/eo/v2.0.0/schema.json"
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
								Extensions: []stac.Extension{
									&eo.Band{
										CommonName: "blue",
									},
								},
							},
						},
						Extensions: []stac.Extension{
							&eo.Asset{
								CloudCover: &num1,
							},
						},
					},
				},
				Extensions: []stac.Extension{
					&eo.Item{
						SnowCover: &num2,
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
					"eo:snow_cover": 100
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
						"eo:cloud_cover": 20,
						"bands": [
							{
								"nodata": 10,
								"eo:common_name": "blue"
							}
						]
					}
				},
				"stac_extensions": [
					"https://stac-extensions.github.io/eo/v2.0.0/schema.json"
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

func TestEOCollectionJSON(t *testing.T) {
	cases := []struct {
		name       string
		collection *stac.Collection
		data       string
		err        string
	}{
		{
			name: "extended bands",
			collection: &stac.Collection{
				Version:     "1.1.0",
				Id:          "collection-id",
				Description: "Test Collection",
				License:     "various",
				Extent: &stac.Extent{
					Spatial: &stac.SpatialExtent{
						Bbox: [][]float64{{-180, -90, 180, 90}},
					},
				},
				ItemAssets: map[string]*stac.Asset{
					"image": {
						Title: "Image",
						Type:  "image/tif",
						Bands: []*stac.Band{
							{
								NoData: 10.0,
								Extensions: []stac.Extension{
									&eo.Band{
										CommonName: "red",
									},
								},
							},
						},
					},
				},
				Links: []*stac.Link{
					{Href: "https://example.com/stac/collection-id", Rel: "self"},
				},
			},
			data: `{
				"type": "Collection",
				"stac_version": "1.1.0",
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
						"rel": "self",
						"href": "https://example.com/stac/collection-id"
					}
				],
				"item_assets": {
					"image": {
						"title": "Image",
						"type": "image/tif",
						"bands": [
							{
								"nodata": 10,
								"eo:common_name": "red"
							}
						]
					}
				},
				"stac_extensions": [
					"https://stac-extensions.github.io/eo/v2.0.0/schema.json"
				]
			}`,
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("%d-%s", i, c.name), func(t *testing.T) {
			if c.collection != nil {
				data, err := json.Marshal(c.collection)
				if c.err != "" {
					assert.ErrorContains(t, err, c.err)
					return
				}
				require.NoError(t, err)
				assert.JSONEq(t, c.data, string(data))
			}

			if c.data != "" {
				collection := &stac.Collection{}
				err := json.Unmarshal([]byte(c.data), collection)
				if c.err != "" {
					assert.ErrorContains(t, err, c.err)
					return
				}
				require.NoError(t, err)
				assert.Equal(t, c.collection, collection)
			}
		})
	}
}
