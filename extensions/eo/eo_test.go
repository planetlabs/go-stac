package eo_test

import (
	"encoding/json"
	"github.com/planetlabs/go-stac/extensions/eo"
	"testing"

	"github.com/planetlabs/go-stac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestItemExtendedMarshal(t *testing.T) {
	cloudCover := float64(25)
	snowCover := float64(10)
	centerWavelength := float64(0.85)
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
					&eo.Asset{
						CloudCover: &cloudCover,
						SnowCover:  &snowCover,
						Bands: []*eo.Band{
							{
								Name:             "NIR",
								CenterWavelength: &centerWavelength,
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
                "eo:cloud_cover": 25,
                "eo:snow_cover": 10,
                "eo:bands": [
                    {
                        "name": "NIR",
                        "center_wavelength": 0.85
                    }
                ]
            }
        },
        "stac_extensions": [
            "https://stac-extensions.github.io/eo/v1.1.0/schema.json"
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
                "eo:cloud_cover": 25,
                "eo:snow_cover": 10,
                "eo:bands": [
                    {
                        "name": "NIR",
                        "center_wavelength": 0.85
                    }
                ]
            }
        },
        "stac_extensions": [
            "https://stac-extensions.github.io/eo/v1.1.0/schema.json"
        ]
    }`)

	item := &stac.Item{}
	require.NoError(t, json.Unmarshal(data, item))

	cloudCover := float64(25)
	snowCover := float64(10)
	centerWavelength := float64(0.85)
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
					&eo.Asset{
						CloudCover: &cloudCover,
						SnowCover:  &snowCover,
						Bands: []*eo.Band{
							{
								Name:             "NIR",
								CenterWavelength: &centerWavelength,
							},
						},
					},
				},
			},
		},
	}

	assert.Equal(t, expected, item)
}
