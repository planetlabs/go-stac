package pl

import "github.com/planetlabs/go-stac"

const extensionUri = "https://planetlabs.github.io/stac-extension/v1.0.0-beta.1/schema.json"

type Asset struct {
	AssetType  string `json:"pl:asset_type,omitempty"`
	BundleType string `json:"pl:bundle_type,omitempty"`
}

var _ stac.AssetExtension = (*Asset)(nil)

func (*Asset) URI() string {
	return extensionUri
}

func (*Asset) Apply(*stac.Asset) {}

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

var _ stac.ItemExtension = (*Item)(nil)

func (*Item) URI() string {
	return extensionUri
}

func (*Item) Apply(*stac.Item) {}
