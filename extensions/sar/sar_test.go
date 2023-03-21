package sar_test

import (
	"encoding/json"
	"testing"

	"github.com/planetlabs/go-stac"
	"github.com/planetlabs/go-stac/extensions/sar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestItemExtendedMarshal(t *testing.T) {
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
			&sar.Item{
				InstrumentMode: "IW",
				FrequencyBand:  "C",
				Polarizations:  []string{"VV", "VH"},
				ProductType:    "GRD",
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
			"sar:instrument_mode": "IW",
			"sar:frequency_band": "C",
			"sar:product_type": "GRD",
			"sar:polarizations": ["VV", "VH"]
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
			"https://stac-extensions.github.io/sar/v1.0.0/schema.json"
		]
	}`

	assert.JSONEq(t, expected, string(data))
}

func TestItemExtendedMarshalOptional(t *testing.T) {
	centerFrequency := 5.405
	resolutionRange := 50.0
	resolutionAzimuth := 50.0
	pixelSpacingRange := 25.0
	pixelSpacingAzimuth := 25.0
	looksRange := 3.0
	looksAzimuth := 1.0
	looksEquivalentNumber := 2.7

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
			&sar.Item{
				InstrumentMode:        "IW",
				FrequencyBand:         "C",
				Polarizations:         []string{"VV", "VH"},
				ProductType:           "GRD",
				CenterFrequency:       &centerFrequency,
				ResolutionRange:       &resolutionRange,
				ResolutionAzimuth:     &resolutionAzimuth,
				PixelSpacingRange:     &pixelSpacingRange,
				PixelSpacingAzimuth:   &pixelSpacingAzimuth,
				LooksRange:            &looksRange,
				LooksAzimuth:          &looksAzimuth,
				LooksEquivalentNumber: &looksEquivalentNumber,
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
			"sar:instrument_mode": "IW",
			"sar:frequency_band": "C",
			"sar:product_type": "GRD",
			"sar:polarizations": ["VV", "VH"],
			"sar:resolution_range": 50,
			"sar:resolution_azimuth": 50,
			"sar:pixel_spacing_range": 25,
			"sar:pixel_spacing_azimuth": 25,
			"sar:looks_range": 3,
			"sar:looks_azimuth": 1,
			"sar:looks_equivalent_number": 2.7,
			"sar:center_frequency": 5.405
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
			"https://stac-extensions.github.io/sar/v1.0.0/schema.json"
		]
	}`

	assert.JSONEq(t, expected, string(data))
}
