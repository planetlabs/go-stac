package raster

import (
	"regexp"

	"github.com/planetlabs/go-stac"
)

const (
	extensionUri     = "https://stac-extensions.github.io/raster/v1.1.0/schema.json"
	extensionPattern = `https://stac-extensions.github.io/raster/v1\..*/schema.json`
	prefix           = "raster"
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
	Bands []*Band `json:"raster:bands,omitempty"`
}

type Band struct {
	NoData            any         `json:"nodata,omitempty"`
	Sampling          string      `json:"sampling,omitempty"`
	DataType          string      `json:"data_type,omitempty"`
	BitsPerSample     *int        `json:"bits_per_sample,omitempty"`
	SpatialResolution *float64    `json:"spatial_resolution,omitempty"`
	Statistics        *Statistics `json:"statistics,omitempty"`
	Unit              string      `json:"units,omitempty"`
	Scale             *float64    `json:"scale,omitempty"`
	Offset            *float64    `json:"offset,omitempty"`
	Histogram         *Histogram  `json:"histogram,omitempty"`
}

type Statistics struct {
	Mean         *float64 `json:"mean,omitempty"`
	Minimum      *float64 `json:"minimum,omitempty"`
	Maximum      *float64 `json:"maximum,omitempty"`
	Stdev        *float64 `json:"stdev,omitempty"`
	ValidPercent *float64 `json:"valid_percent,omitempty"`
}

type Histogram struct {
	Count   int       `json:"count"`
	Min     float64   `json:"min"`
	Max     float64   `json:"max"`
	Buckets []float64 `json:"buckets"`
}

var _ stac.Extension = (*Asset)(nil)

func (*Asset) URI() string {
	return extensionUri
}

func (e *Asset) Encode(assetMap map[string]any) error {
	return stac.EncodeExtendedMap(e, assetMap)
}

func (e *Asset) Decode(assetMap map[string]any) error {
	return stac.DecodeExtendedMap(e, assetMap, prefix)
}
