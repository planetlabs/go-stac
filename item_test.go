package stac_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/planetlabs/go-stac"
	"github.com/planetlabs/go-stac/extensions/eo/v1"
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

func TestItemUnmarshal(t *testing.T) {
	data := `{
		"type": "Feature",
		"stac_version": "1.0.0",
		"id": "item-id",
		"geometry": {
			"type": "Point",
			"coordinates": [1, 2]
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

	item := &stac.Item{}
	require.NoError(t, json.Unmarshal([]byte(data), item))

	assert.Equal(t, "item-id", item.Id)

	require.NotNil(t, item.Geometry)

	g, ok := item.Geometry.(map[string]any)
	require.True(t, ok)
	geometryType, ok := g["type"].(string)
	require.True(t, ok)
	assert.Equal(t, "Point", geometryType)
}

type TestGeometry struct {
	data []byte
}

func (g *TestGeometry) UnmarshalJSON(data []byte) error {
	g.data = data
	return nil
}

func TestGeometryUnmarshal(t *testing.T) {
	original := &TestGeometry{}
	stac.GeometryUnmarshaler(original)
	defer stac.GeometryUnmarshaler(nil)

	data := `{
		"type": "Feature",
		"stac_version": "1.0.0",
		"id": "item-id",
		"geometry": {
			"type": "Point",
			"coordinates": [1, 2]
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

	item := &stac.Item{}
	require.NoError(t, json.Unmarshal([]byte(data), item))

	assert.Equal(t, "item-id", item.Id)

	assert.Nil(t, original.data)
	require.NotNil(t, item.Geometry)

	g, ok := item.Geometry.(*TestGeometry)
	require.True(t, ok)
	assert.NotNil(t, g.data)
}

func getExtension(item *stac.Item, uri string) stac.Extension {
	for _, extension := range item.Extensions {
		if extension.URI() == uri {
			return extension
		}
	}
	return nil
}

func TestItemListUnmarshal(t *testing.T) {
	data, err := os.ReadFile("testdata/items.json")
	require.NoError(t, err)

	itemList := &stac.ItemsList{}
	require.NoError(t, json.Unmarshal([]byte(data), itemList))

	require.Len(t, itemList.Items, 2)
	first := itemList.Items[0]

	eoExtension := getExtension(first, "https://stac-extensions.github.io/eo/v1.1.0/schema.json")
	require.NotNil(t, eoExtension)
	eo, ok := eoExtension.(*eo.Item)
	require.True(t, ok)

	require.NotNil(t, eo.CloudCover)
	assert.Equal(t, float64(50), *eo.CloudCover)
}
