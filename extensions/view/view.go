package view

import "github.com/planetlabs/go-stac"

const extensionUri = "https://stac-extensions.github.io/view/v1.0.0/schema.json"

type Item struct {
	OffNadir       *float64 `json:"view:off_nadir,omitempty"`
	IncidenceAngle *float64 `json:"view:incidence_angle,omitempty"`
	Azimuth        *float64 `json:"view:azimuth,omitempty"`
	SunAzimuth     *float64 `json:"view:sun_azimuth,omitempty"`
	SunElevation   *float64 `json:"view:sun_elevation,omitempty"`
}

var _ stac.ItemExtension = (*Item)(nil)

func (*Item) URI() string {
	return extensionUri
}

func (*Item) Apply(*stac.Item) {}
