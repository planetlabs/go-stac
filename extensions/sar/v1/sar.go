package sar

import (
	"regexp"

	"github.com/planetlabs/go-stac"
)

const (
	extensionUri     = "https://stac-extensions.github.io/sar/v1.0.0/schema.json"
	extensionPattern = `https://stac-extensions.github.io/sar/v1\..*/schema.json`
)

func init() {
	stac.RegisterItemExtension(
		regexp.MustCompile(extensionPattern),
		func() stac.Extension {
			return &Item{}
		},
	)
}

type Item struct {
	InstrumentMode        string   `json:"sar:instrument_mode"`
	FrequencyBand         string   `json:"sar:frequency_band"`
	CenterFrequency       *float64 `json:"sar:center_frequency,omitempty"`
	Polarizations         []string `json:"sar:polarizations"`
	ProductType           string   `json:"sar:product_type"`
	ResolutionRange       *float64 `json:"sar:resolution_range,omitempty"`
	ResolutionAzimuth     *float64 `json:"sar:resolution_azimuth,omitempty"`
	PixelSpacingRange     *float64 `json:"sar:pixel_spacing_range,omitempty"`
	PixelSpacingAzimuth   *float64 `json:"sar:pixel_spacing_azimuth,omitempty"`
	LooksRange            *float64 `json:"sar:looks_range,omitempty"`
	LooksAzimuth          *float64 `json:"sar:looks_azimuth,omitempty"`
	LooksEquivalentNumber *float64 `json:"sar:looks_equivalent_number,omitempty"`
	ObservationDirection  *string  `json:"sar:observation_direction,omitempty"`
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
