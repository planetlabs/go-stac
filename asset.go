package stac

import (
	"encoding/json"
	"fmt"

	"github.com/mitchellh/mapstructure"
)

type Asset struct {
	Type        string           `json:"type,omitempty"`
	Href        string           `json:"href"`
	Title       string           `json:"title,omitempty"`
	Description string           `json:"description,omitempty"`
	Created     string           `json:"created,omitempty"`
	Roles       []string         `json:"roles,omitempty"`
	Extensions  []AssetExtension `json:"-"`
}

var _ json.Marshaler = (*Asset)(nil)

type AssetExtension interface {
	Apply(*Asset)
	URI() string
}

func (asset Asset) MarshalJSON() ([]byte, error) {
	assetMap := map[string]any{}
	decoder, decoderErr := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &assetMap,
	})
	if decoderErr != nil {
		return nil, decoderErr
	}

	decodeErr := decoder.Decode(asset)
	if decodeErr != nil {
		return nil, decodeErr
	}

	for _, extension := range asset.Extensions {
		extension.Apply(&asset)
		if decodeErr := decoder.Decode(extension); decodeErr != nil {
			return nil, fmt.Errorf("trouble encoding JSON for %s asset: %w", extension.URI(), decodeErr)
		}
	}

	return json.Marshal(assetMap)
}
