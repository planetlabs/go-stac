package eo

import (
	"regexp"

	"github.com/planetlabs/go-stac"
)

const (
	extensionUri     = "https://stac-extensions.github.io/eo/v1.1.0/schema.json"
	extensionPattern = `https://stac-extensions.github.io/eo/v1\..*/schema.json`
)

func init() {
	stac.RegisterAssetExtension(
		regexp.MustCompile(extensionPattern),
		func() stac.Extension {
			return &Asset{}
		},
	)
}

type Asset struct {
	CloudCover *float64 `json:"eo:cloud_cover,omitempty"`
	SnowCover  *float64 `json:"eo:snow_cover,omitempty"`
	Bands      []*Band  `json:"eo:bands,omitempty"`
}

type Band struct {
	Name              string   `json:"name,omitempty"`
	CommonName        string   `json:"common_name,omitempty"`
	Description       string   `json:"description,omitempty"`
	CenterWavelength  *float64 `json:"center_wavelength,omitempty"`
	FullWidthHalfMax  *float64 `json:"full_width_half_max,omitempty"`
	SolarIllumination *float64 `json:"solar_illumination,omitempty"`
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
