package raster

import (
	"regexp"

	"github.com/planetlabs/go-stac"
)

const (
	extensionUri     = "https://stac-extensions.github.io/raster/v2.0.0/schema.json"
	extensionPattern = `https://stac-extensions.github.io/raster/v2\..*/schema.json`
	prefix           = "raster"
)

func init() {
	stac.RegisterBandExtension(
		regexp.MustCompile(extensionPattern),
		func() stac.Extension {
			return &Band{}
		},
	)

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

type Band struct {
	Sampling          string     `json:"raster:sampling,omitempty"`
	BitsPerSample     *int       `json:"raster:bits_per_sample,omitempty"`
	SpatialResolution *float64   `json:"raster:spatial_resolution,omitempty"`
	Scale             *float64   `json:"raster:scale,omitempty"`
	Offset            *float64   `json:"raster:offset,omitempty"`
	Histogram         *Histogram `json:"raster:histogram,omitempty"`
}

type Item struct {
	Sampling          string     `json:"raster:sampling,omitempty"`
	BitsPerSample     *int       `json:"raster:bits_per_sample,omitempty"`
	SpatialResolution *float64   `json:"raster:spatial_resolution,omitempty"`
	Scale             *float64   `json:"raster:scale,omitempty"`
	Offset            *float64   `json:"raster:offset,omitempty"`
	Histogram         *Histogram `json:"raster:histogram,omitempty"`
}

type Asset struct {
	Sampling          string     `json:"raster:sampling,omitempty"`
	BitsPerSample     *int       `json:"raster:bits_per_sample,omitempty"`
	SpatialResolution *float64   `json:"raster:spatial_resolution,omitempty"`
	Scale             *float64   `json:"raster:scale,omitempty"`
	Offset            *float64   `json:"raster:offset,omitempty"`
	Histogram         *Histogram `json:"raster:histogram,omitempty"`
}

type Histogram struct {
	Count   int       `json:"count"`
	Min     float64   `json:"min"`
	Max     float64   `json:"max"`
	Buckets []float64 `json:"buckets"`
}

var (
	_ stac.Extension = (*Band)(nil)
	_ stac.Extension = (*Item)(nil)
	_ stac.Extension = (*Asset)(nil)
)

func (*Band) URI() string {
	return extensionUri
}

func (e *Band) Encode(bandMap map[string]any) error {
	return stac.EncodeExtendedMap(e, bandMap)
}

func (e *Band) Decode(bandMap map[string]any) error {
	return stac.DecodeExtendedMap(e, bandMap, prefix)
}

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
