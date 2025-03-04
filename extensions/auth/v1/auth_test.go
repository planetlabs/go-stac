package auth_test

import (
	"encoding/json"
	"testing"

	"github.com/planetlabs/go-stac"
	"github.com/planetlabs/go-stac/extensions/auth/v1"
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
			"image": {
				Title: "Image",
				Href:  "https://example.com/stac/item-id/image.tif",
				Type:  "image/tif",
				Extensions: []stac.Extension{
					&auth.Asset{
						Refs: []string{"openid"},
					},
				},
			},
		},
		Extensions: []stac.Extension{
			&auth.Item{
				Schemes: map[string]*auth.Scheme{
					"openid": {
						Type:             "openIdConnect",
						Description:      "Test auth configuration",
						OpenIdConnectUrl: "https://example.com/auth/.well-known/openid-configuration",
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
			"test": "value",
			"auth:schemes": {
				"openid": {
					"type": "openIdConnect",
					"description": "Test auth configuration",
					"openIdConnectUrl": "https://example.com/auth/.well-known/openid-configuration"
				}
			}
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
				"auth:refs": ["openid"]
			}
		},
		"stac_extensions": [
			"https://stac-extensions.github.io/authentication/v1.1.0/schema.json"
		]
	}`

	assert.JSONEq(t, expected, string(data))
}

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
			&auth.Collection{
				Schemes: map[string]*auth.Scheme{
					"openid": {
						Type:             "openIdConnect",
						Description:      "Test auth configuration",
						OpenIdConnectUrl: "https://example.com/auth/.well-known/openid-configuration",
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
		"auth:schemes": {
			"openid": {
				"type": "openIdConnect",
				"description": "Test auth configuration",
				"openIdConnectUrl": "https://example.com/auth/.well-known/openid-configuration"
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
			"https://stac-extensions.github.io/authentication/v1.1.0/schema.json"
		],
		"stac_version": "1.0.0"
	}`

	assert.JSONEq(t, expected, string(data))
}

func TestCollectionLinkExtendedMarshal(t *testing.T) {
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
			{
				Href: "https://example.com/stac/collections/collection-id",
				Rel:  "self",
				Type: "application/json",
				Extensions: []stac.Extension{
					&auth.Link{
						Refs: []string{"openid"},
					},
				},
			},
		},
		Extensions: []stac.Extension{
			&auth.Collection{
				Schemes: map[string]*auth.Scheme{
					"openid": {
						Type:             "openIdConnect",
						Description:      "Test auth configuration",
						OpenIdConnectUrl: "https://example.com/auth/.well-known/openid-configuration",
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
		"auth:schemes": {
			"openid": {
				"type": "openIdConnect",
				"description": "Test auth configuration",
				"openIdConnectUrl": "https://example.com/auth/.well-known/openid-configuration"
			}
		},
		"links": [
			{
				"href": "https://example.com/stac/collections/collection-id",
				"rel": "self",
				"type": "application/json",
				"auth:refs": ["openid"]
			}
		],
		"stac_extensions": [
			"https://stac-extensions.github.io/authentication/v1.1.0/schema.json"
		],
		"stac_version": "1.0.0"
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
			"auth:schemes": {
				"openid": {
					"type": "openIdConnect",
					"description": "Test auth configuration",
					"openIdConnectUrl": "https://example.com/auth/.well-known/openid-configuration"
				}
			}
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
				"auth:refs": ["openid"]
			}
		},
		"stac_extensions": [
			"https://stac-extensions.github.io/authentication/v1.1.0/schema.json"
		]
	}`)

	item := &stac.Item{}
	require.NoError(t, json.Unmarshal(data, item))

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
					&auth.Asset{
						Refs: []string{"openid"},
					},
				},
			},
		},
		Extensions: []stac.Extension{
			&auth.Item{
				Schemes: map[string]*auth.Scheme{
					"openid": {
						Type:             "openIdConnect",
						Description:      "Test auth configuration",
						OpenIdConnectUrl: "https://example.com/auth/.well-known/openid-configuration",
					},
				},
			},
		},
	}

	assert.Equal(t, expected, item)
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
		"auth:schemes": {
			"openid": {
				"type": "openIdConnect",
				"openIdConnectUrl": "https://example.com/auth/.well-known/openid-configuration"
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
			"https://stac-extensions.github.io/authentication/v1.1.0/schema.json"
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
			&auth.Collection{
				Schemes: map[string]*auth.Scheme{
					"openid": {
						Type:             "openIdConnect",
						OpenIdConnectUrl: "https://example.com/auth/.well-known/openid-configuration",
					},
				},
			},
		},
	}

	assert.Equal(t, expected, collection)
}

func TestCollectionLinkExtendedUnmarshal(t *testing.T) {
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
		"auth:schemes": {
			"openid": {
				"type": "openIdConnect",
				"openIdConnectUrl": "https://example.com/auth/.well-known/openid-configuration"
			}
		},
		"links": [
			{
				"href": "https://example.com/stac/collections/collection-id",
				"rel": "self",
				"type": "application/json",
				"auth:refs": ["openid"]
			}
		],
		"stac_extensions": [
			"https://stac-extensions.github.io/authentication/v1.1.0/schema.json"
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
			{
				Href: "https://example.com/stac/collections/collection-id",
				Rel:  "self",
				Type: "application/json",
				Extensions: []stac.Extension{
					&auth.Link{
						Refs: []string{"openid"},
					},
				},
			},
		},
		Extensions: []stac.Extension{
			&auth.Collection{
				Schemes: map[string]*auth.Scheme{
					"openid": {
						Type:             "openIdConnect",
						OpenIdConnectUrl: "https://example.com/auth/.well-known/openid-configuration",
					},
				},
			},
		},
	}

	assert.Equal(t, expected, collection)
}

func TestCatalogExtendedMarshal(t *testing.T) {
	catalog := &stac.Catalog{
		Version:     "1.0.0",
		Id:          "catalog-id",
		Description: "Test Catalog",
		Links: []*stac.Link{
			{Href: "https://example.com/stac/catalog-id", Rel: "self", Type: "application/json"},
		},
		Extensions: []stac.Extension{
			&auth.Catalog{
				Schemes: map[string]*auth.Scheme{
					"oauth": {
						Type:        "oauth2",
						Description: "Test auth configuration",
						Flows: map[string]any{
							"authorizationUrl": "https://example.com/oauth/authorize",
							"tokenUrl":         "https://example.com/oauth/token",
							"scopes":           []any{},
						},
					},
				},
			},
		},
	}

	data, err := json.Marshal(catalog)
	require.NoError(t, err)

	expected := `{
		"type": "Catalog",
		"id": "catalog-id",
		"description": "Test Catalog",
		"auth:schemes": {
			"oauth": {
				"type": "oauth2",
				"description": "Test auth configuration",
				"flows": {
					"authorizationUrl": "https://example.com/oauth/authorize",
					"tokenUrl": "https://example.com/oauth/token",
					"scopes": []
				}
			}
		},
		"links": [
			{
				"href": "https://example.com/stac/catalog-id",
				"rel": "self",
				"type": "application/json"
			}
		],
		"stac_extensions": [
			"https://stac-extensions.github.io/authentication/v1.1.0/schema.json"
		],
		"stac_version": "1.0.0"
	}`

	assert.JSONEq(t, expected, string(data))
}

func TestCatalogLinkExtendedMarshal(t *testing.T) {
	catalog := &stac.Catalog{
		Version:     "1.0.0",
		Id:          "catalog-id",
		Description: "Test Catalog",
		Links: []*stac.Link{
			{
				Href: "https://example.com/stac/catalog-id",
				Rel:  "self",
				Type: "application/json",
				Extensions: []stac.Extension{
					&auth.Link{
						Refs: []string{"oauth"},
					},
				},
			},
		},
		Extensions: []stac.Extension{
			&auth.Catalog{
				Schemes: map[string]*auth.Scheme{
					"oauth": {
						Type:        "oauth2",
						Description: "Test auth configuration",
						Flows: map[string]any{
							"authorizationUrl": "https://example.com/oauth/authorize",
							"tokenUrl":         "https://example.com/oauth/token",
							"scopes":           []any{},
						},
					},
				},
			},
		},
	}

	data, err := json.Marshal(catalog)
	require.NoError(t, err)

	expected := `{
		"type": "Catalog",
		"id": "catalog-id",
		"description": "Test Catalog",
		"auth:schemes": {
			"oauth": {
				"type": "oauth2",
				"description": "Test auth configuration",
				"flows": {
					"authorizationUrl": "https://example.com/oauth/authorize",
					"tokenUrl": "https://example.com/oauth/token",
					"scopes": []
				}
			}
		},
		"links": [
			{
				"href": "https://example.com/stac/catalog-id",
				"rel": "self",
				"type": "application/json",
				"auth:refs": ["oauth"]
			}
		],
		"stac_extensions": [
			"https://stac-extensions.github.io/authentication/v1.1.0/schema.json"
		],
		"stac_version": "1.0.0"
	}`

	assert.JSONEq(t, expected, string(data))
}

func TestCatalogExtendedUnmarshal(t *testing.T) {
	data := []byte(`{
		"type": "Catalog",
		"id": "catalog-id",
		"description": "Test Catalog",
		"auth:schemes": {
			"oauth": {
				"type": "oauth2",
				"flows": {
					"authorizationUrl": "https://example.com/oauth/authorize",
					"tokenUrl": "https://example.com/oauth/token",
					"scopes": []
				}
			}
		},
		"links": [
			{
				"href": "https://example.com/stac/catalog-id",
				"rel": "self",
				"type": "application/json"
			}
		],
		"stac_extensions": [
			"https://stac-extensions.github.io/authentication/v1.1.0/schema.json"
		],
		"stac_version": "1.0.0"
	}`)

	catalog := &stac.Catalog{}
	require.NoError(t, json.Unmarshal(data, catalog))

	expected := &stac.Catalog{
		Version:     "1.0.0",
		Id:          "catalog-id",
		Description: "Test Catalog",
		Links: []*stac.Link{
			{Href: "https://example.com/stac/catalog-id", Rel: "self", Type: "application/json"},
		},
		Extensions: []stac.Extension{
			&auth.Catalog{
				Schemes: map[string]*auth.Scheme{
					"oauth": {
						Type: "oauth2",
						Flows: map[string]any{
							"authorizationUrl": "https://example.com/oauth/authorize",
							"tokenUrl":         "https://example.com/oauth/token",
							"scopes":           []any{},
						},
					},
				},
			},
		},
	}

	assert.Equal(t, expected, catalog)
}

func TestCatalogLinkExtendedUnmarshal(t *testing.T) {
	data := []byte(`{
		"type": "Catalog",
		"id": "catalog-id",
		"description": "Test Catalog",
		"auth:schemes": {
			"oauth": {
				"type": "oauth2",
				"flows": {
					"authorizationUrl": "https://example.com/oauth/authorize",
					"tokenUrl": "https://example.com/oauth/token",
					"scopes": []
				}
			}
		},
		"links": [
			{
				"href": "https://example.com/stac/catalog-id",
				"rel": "self",
				"type": "application/json",
				"auth:refs": ["oauth"]
			}
		],
		"stac_extensions": [
			"https://stac-extensions.github.io/authentication/v1.1.0/schema.json"
		],
		"stac_version": "1.0.0"
	}`)

	catalog := &stac.Catalog{}
	require.NoError(t, json.Unmarshal(data, catalog))

	expected := &stac.Catalog{
		Version:     "1.0.0",
		Id:          "catalog-id",
		Description: "Test Catalog",
		Links: []*stac.Link{
			{
				Href: "https://example.com/stac/catalog-id",
				Rel:  "self",
				Type: "application/json",
				Extensions: []stac.Extension{
					&auth.Link{
						Refs: []string{"oauth"},
					},
				},
			},
		},
		Extensions: []stac.Extension{
			&auth.Catalog{
				Schemes: map[string]*auth.Scheme{
					"oauth": {
						Type: "oauth2",
						Flows: map[string]any{
							"authorizationUrl": "https://example.com/oauth/authorize",
							"tokenUrl":         "https://example.com/oauth/token",
							"scopes":           []any{},
						},
					},
				},
			},
		},
	}

	assert.Equal(t, expected, catalog)
}

func TestItemAssetExtendedMarshal(t *testing.T) {
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
					&auth.Asset{
						Refs: []string{"openid"},
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
				"auth:refs": ["openid"]
			}
		},
		"stac_extensions": [
			"https://stac-extensions.github.io/authentication/v1.1.0/schema.json"
		]
	}`

	assert.JSONEq(t, expected, string(data))
}

func TestItemAssetExtendedUnmarshal(t *testing.T) {
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
					"auth:refs": ["openid"]
				}
			},
			"stac_extensions": [
				"https://stac-extensions.github.io/authentication/v1.1.0/schema.json"
			]
    }`)

	item := &stac.Item{}
	require.NoError(t, json.Unmarshal(data, item))

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
					&auth.Asset{
						Refs: []string{"openid"},
					},
				},
			},
		},
	}

	assert.Equal(t, expected, item)
}

func TestItemLinkExtendedMarshal(t *testing.T) {
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
			{
				Href: "https://example.com/stac/item-id", Rel: "self",
				Extensions: []stac.Extension{
					&auth.Link{
						Refs: []string{"openid"},
					},
				},
			},
		},
		Assets: map[string]*stac.Asset{
			"image": {
				Title: "Image",
				Href:  "https://example.com/stac/item-id/image.tif",
				Type:  "image/tif",
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
				"href": "https://example.com/stac/item-id",
				"auth:refs": ["openid"]
			}
		],
		"assets": {
			"image": {
				"title": "Image",
				"href": "https://example.com/stac/item-id/image.tif",
				"type": "image/tif"
			}
		},
		"stac_extensions": [
			"https://stac-extensions.github.io/authentication/v1.1.0/schema.json"
		]
}`

	assert.JSONEq(t, expected, string(data))
}

func TestItemLinkExtendedUnmarshal(t *testing.T) {
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
					"href": "https://example.com/stac/item-id",
					"auth:refs": ["openid"]
				}
			],
			"assets": {
				"image": {
					"title": "Image",
					"href": "https://example.com/stac/item-id/image.tif",
					"type": "image/tif"
				}
			},
			"stac_extensions": [
				"https://stac-extensions.github.io/authentication/v1.1.0/schema.json"
			]
    }`)

	item := &stac.Item{}
	require.NoError(t, json.Unmarshal(data, item))

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
			{
				Href: "https://example.com/stac/item-id",
				Rel:  "self",
				Extensions: []stac.Extension{
					&auth.Link{
						Refs: []string{"openid"},
					},
				},
			},
		},
		Assets: map[string]*stac.Asset{
			"image": {
				Title: "Image",
				Href:  "https://example.com/stac/item-id/image.tif",
				Type:  "image/tif",
			},
		},
	}

	assert.Equal(t, expected, item)
}
