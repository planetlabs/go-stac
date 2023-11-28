package pl

import (
	"regexp"

	"github.com/planetlabs/go-stac"
)

const (
	extensionUri     = "https://planetlabs.github.io/stac-extension/v1.0.0-beta.1/schema.json"
	extensionPattern = `https://planetlabs.github.io/stac-extension/v1\..*/schema.json`
)

func init() {
	stac.RegisterItemExtension(
		regexp.MustCompile(extensionPattern),
		func() stac.Extension {
			return &Item{}
		},
	)
	stac.RegisterAssetExtension(
		regexp.MustCompile(extensionPattern),
		func() stac.Extension {
			return &Asset{}
		},
	)
}

type Asset struct {
	AssetType  string `json:"pl:asset_type,omitempty"`
	BundleType string `json:"pl:bundle_type,omitempty"`
}

var _ stac.Extension = (*Asset)(nil)

func (*Asset) URI() string {
	return extensionUri
}

func (e *Asset) Encode(assetMap map[string]any) error {
	return stac.EncodeExtendedAsset(e, assetMap)
}

func (e *Asset) Decode(assetMap map[string]any) error {
	return stac.DecodeExtendedAsset(e, assetMap)
}

type Item struct {
	ItemType           string   `json:"pl:item_type,omitempty"`
	PixelResolution    float64  `json:"pl:pixel_resolution,omitempty"`
	PublishingStage    string   `json:"pl:publishing_stage,omitempty"`
	QualityCategory    string   `json:"pl:quality_category,omitempty"`
	StripId            string   `json:"pl:strip_id,omitempty"`
	BlackFill          *float64 `json:"pl:black_fill,omitempty"`
	ClearPercent       *float64 `json:"pl:clear_percent,omitempty"`
	GridCell           *string  `json:"pl:grid_cell,omitempty"`
	GroundControl      *bool    `json:"pl:ground_control,omitempty"`
	GroundControlRatio *float64 `json:"pl:ground_control_ratio,omitempty"`
}

var _ stac.Extension = (*Item)(nil)

func (*Item) URI() string {
	return extensionUri
}

func (e *Item) Encode(itemMap map[string]any) error {
	return stac.EncodeExtendedItemProperties(e, itemMap)
}

func (e *Item) Decode(itemMap map[string]any) error {
	return stac.DecodeExtendedItemProperties(e, itemMap)
}
