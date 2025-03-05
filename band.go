package stac

import (
	"regexp"

	"github.com/go-viper/mapstructure/v2"
)

var bandExtensions = newExtensionRegistry()

func RegisterBandExtension(pattern *regexp.Regexp, provider ExtensionProvider) {
	bandExtensions.register(pattern, provider)
}

func GetBandExtension(uri string) Extension {
	return bandExtensions.get(uri)
}

type Band struct {
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
	NoData      any         `json:"nodata,omitempty"`
	DataType    string      `json:"data_type,omitempty"`
	Statistics  *Statistics `json:"statistics,omitempty"`
	Unit        string      `json:"unit,omitempty"`
	Extensions  []Extension `json:"-"`
}

type Statistics struct {
	Mean         *float64 `json:"mean,omitempty"`
	Minimum      *float64 `json:"minimum,omitempty"`
	Maximum      *float64 `json:"maximum,omitempty"`
	Stdev        *float64 `json:"stdev,omitempty"`
	ValidPercent *float64 `json:"valid_percent,omitempty"`
	Count        *int     `json:"count,omitempty"`
}

func encodeBand(band *Band) (map[string]any, []string, error) {
	extensionUris := []string{}
	bandMap := map[string]any{}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &bandMap,
	})
	if err != nil {
		return nil, nil, err
	}
	if err := decoder.Decode(band); err != nil {
		return nil, nil, err
	}
	for _, extension := range band.Extensions {
		extensionUris = append(extensionUris, extension.URI())
		if err := extension.Encode(bandMap); err != nil {
			return nil, nil, err
		}
	}
	return bandMap, extensionUris, nil
}

func EncodeBands(bands []*Band) ([]map[string]any, []string, error) {
	extensionUris := []string{}
	bandMaps := make([]map[string]any, len(bands))
	for i, band := range bands {
		bandMap, uris, err := encodeBand(band)
		if err != nil {
			return nil, nil, err
		}
		bandMaps[i] = bandMap
		extensionUris = append(extensionUris, uris...)
	}
	return bandMaps, extensionUris, nil
}
