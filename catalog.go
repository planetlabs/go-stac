package stac

import (
	"encoding/json"

	"github.com/mitchellh/mapstructure"
)

type Catalog struct {
	Version     string   `json:"stac_version"`
	Id          string   `json:"id"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description"`
	Links       []*Link  `json:"links"`
	ConformsTo  []string `json:"conformsTo,omitempty"`
}

var _ json.Marshaler = (*Catalog)(nil)

func (catalog Catalog) MarshalJSON() ([]byte, error) {
	collectionMap := map[string]any{
		"type": "Catalog",
	}
	decoder, decoderErr := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &collectionMap,
	})
	if decoderErr != nil {
		return nil, decoderErr
	}

	decodeErr := decoder.Decode(catalog)
	if decodeErr != nil {
		return nil, decodeErr
	}

	return json.Marshal(collectionMap)
}
