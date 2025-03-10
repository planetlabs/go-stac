package eo

import (
	"regexp"

	"github.com/planetlabs/go-stac"
)

const (
	extensionUri     = "https://stac-extensions.github.io/eo/v1.1.0/schema.json"
	extensionPattern = `https://stac-extensions.github.io/eo/v1\..*/schema.json`
	prefix           = "eo"
)

func init() {
	stac.RegisterAssetExtension(
		regexp.MustCompile(extensionPattern),
		func() stac.Extension {
			return &Asset{}
		},
	)

	stac.RegisterItemExtension(
		regexp.MustCompile(extensionPattern),
		func() stac.Extension {
			return &Item{}
		},
	)
}

type Item struct {
	CloudCover *float64 `json:"eo:cloud_cover,omitempty"`
	SnowCover  *float64 `json:"eo:snow_cover,omitempty"`
}

var _ stac.Extension = (*Item)(nil)

func (*Item) URI() string {
	return extensionUri
}

func (e *Item) Encode(itemMap map[string]any) error {
	return stac.EncodeExtendedItemProperties(e, itemMap)
}

func (e *Item) Decode(itemMap map[string]any) error {
	if err := stac.DecodeExtendedItemProperties(e, itemMap); err != nil {
		return err
	}
	if e.CloudCover == nil && e.SnowCover == nil {
		return stac.ErrExtensionDoesNotApply
	}
	return nil
}

type Asset struct {
	Bands []*Band `json:"eo:bands,omitempty"`
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
	return stac.EncodeExtendedMap(e, assetMap)
}

func (e *Asset) Decode(assetMap map[string]any) error {
	if err := stac.DecodeExtendedMap(e, assetMap, prefix); err != nil {
		return err
	}
	if len(e.Bands) == 0 {
		return stac.ErrExtensionDoesNotApply
	}
	return nil
}
