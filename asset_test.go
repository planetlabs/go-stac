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
	asset := &stac.Asset{
		Title: "An Image",
		Href:  "https://example.com/image.tif",
		Type:  "image/tiff",
		Roles: []string{"data", "reflectance"},
	}

	data, err := json.Marshal(asset)
	require.Nil(t, err)

	expected := `{
		"title": "An Image",
		"href": "https://example.com/image.tif",
		"type": "image/tiff",
		"roles": ["data", "reflectance"]
	}`

	assert.JSONEq(t, expected, string(data))
}

func TestAssetExtendedMarshal(t *testing.T) {
	asset := &stac.Asset{
		Href: "https://example.com/image.tif",
		Type: "image/tiff",
		Extensions: []stac.AssetExtension{
			&pl.Asset{
				AssetType:  "ortho_analytic_4b_sr",
				BundleType: "analytic_sr_udm2",
			},
		},
	}

	data, err := json.Marshal(asset)
	require.Nil(t, err)

	expected := `{
		"href": "https://example.com/image.tif",
		"type": "image/tiff",
		"pl:asset_type": "ortho_analytic_4b_sr",
		"pl:bundle_type": "analytic_sr_udm2"
	}`

	assert.JSONEq(t, expected, string(data))
}
