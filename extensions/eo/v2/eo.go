package eo

import (
	"regexp"

	"github.com/planetlabs/go-stac"
)

const (
	extensionUri     = "https://stac-extensions.github.io/eo/v2.0.0/schema.json"
	extensionPattern = `https://stac-extensions.github.io/eo/v2\..*/schema.json`
	prefix           = "eo"
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

	stac.RegisterBandExtension(
		regexp.MustCompile(extensionPattern),
		func() stac.Extension {
			return &Band{}
		},
	)
}

type Item struct {
	CloudCover        *float64 `json:"eo:cloud_cover,omitempty"`
	SnowCover         *float64 `json:"eo:snow_cover,omitempty"`
	CommonName        string   `json:"eo:common_name,omitempty"`
	CenterWavelength  *float64 `json:"eo:center_wavelength,omitempty"`
	FullWidthHalfMax  *float64 `json:"eo:full_width_half_max,omitempty"`
	SolarIllumination *float64 `json:"eo:solar_illumination,omitempty"`
}

type Asset struct {
	CloudCover        *float64 `json:"eo:cloud_cover,omitempty"`
	SnowCover         *float64 `json:"eo:snow_cover,omitempty"`
	CommonName        string   `json:"eo:common_name,omitempty"`
	CenterWavelength  *float64 `json:"eo:center_wavelength,omitempty"`
	FullWidthHalfMax  *float64 `json:"eo:full_width_half_max,omitempty"`
	SolarIllumination *float64 `json:"eo:solar_illumination,omitempty"`
}

type Band struct {
	CloudCover        *float64 `json:"eo:cloud_cover,omitempty"`
	SnowCover         *float64 `json:"eo:snow_cover,omitempty"`
	CommonName        string   `json:"eo:common_name,omitempty"`
	CenterWavelength  *float64 `json:"eo:center_wavelength,omitempty"`
	FullWidthHalfMax  *float64 `json:"eo:full_width_half_max,omitempty"`
	SolarIllumination *float64 `json:"eo:solar_illumination,omitempty"`
}

var (
	_ stac.Extension = (*Item)(nil)
	_ stac.Extension = (*Asset)(nil)
	_ stac.Extension = (*Band)(nil)
)

func (*Item) URI() string {
	return extensionUri
}

func (e *Item) Encode(itemMap map[string]any) error {
	return stac.EncodeExtendedItemProperties(e, itemMap)
}

func (e *Item) Decode(itemMap map[string]any) error {
	return stac.DecodeExtendedItemProperties(e, itemMap)
}

func (*Asset) URI() string {
	return extensionUri
}

func (e *Asset) Encode(assetMap map[string]any) error {
	return stac.EncodeExtendedMap(e, assetMap)
}

func (e *Asset) Decode(assetMap map[string]any) error {
	return stac.DecodeExtendedMap(e, assetMap, prefix)
}

func (*Band) URI() string {
	return extensionUri
}

func (e *Band) Encode(bandMap map[string]any) error {
	return stac.EncodeExtendedMap(e, bandMap)
}

func (e *Band) Decode(bandMap map[string]any) error {
	return stac.DecodeExtendedMap(e, bandMap, prefix)
}
