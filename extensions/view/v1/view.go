package view

import (
	"regexp"

	"github.com/planetlabs/go-stac"
)

const (
	extensionUri     = "https://stac-extensions.github.io/view/v1.0.0/schema.json"
	extensionPattern = `https://stac-extensions.github.io/view/v1\..*/schema.json`
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
	OffNadir       *float64 `json:"view:off_nadir,omitempty"`
	IncidenceAngle *float64 `json:"view:incidence_angle,omitempty"`
	Azimuth        *float64 `json:"view:azimuth,omitempty"`
	SunAzimuth     *float64 `json:"view:sun_azimuth,omitempty"`
	SunElevation   *float64 `json:"view:sun_elevation,omitempty"`
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
