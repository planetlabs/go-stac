package stac

import (
	"encoding/json"

	"github.com/mitchellh/mapstructure"
)

type Collection struct {
	Version     string            `json:"stac_version"`
	Id          string            `json:"id"`
	Title       string            `json:"title,omitempty"`
	Description string            `json:"description"`
	Keywords    []string          `json:"keywords,omitempty"`
	License     string            `json:"license"`
	Providers   []*Provider       `json:"providers,omitempty"`
	Extent      *Extent           `json:"extent"`
	Summaries   map[string]any    `json:"summaries,omitempty"`
	Links       []*Link           `json:"links"`
	Assets      map[string]*Asset `json:"assets,omitempty"`
}

type Provider struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	Url         string   `json:"url,omitempty"`
}

type Extent struct {
	Spatial  *SpatialExtent  `json:"spatial,omitempty"`
	Temporal *TemporalExtent `json:"temporal,omitempty"`
}

type SpatialExtent struct {
	Bbox [][]float64 `json:"bbox"`
}

type TemporalExtent struct {
	Interval [][]any `json:"interval"`
}

var _ json.Marshaler = (*Collection)(nil)

func (collection Collection) MarshalJSON() ([]byte, error) {
	collectionMap := map[string]any{
		"type": "Collection",
	}
	decoder, decoderErr := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &collectionMap,
	})
	if decoderErr != nil {
		return nil, decoderErr
	}

	decodeErr := decoder.Decode(collection)
	if decodeErr != nil {
		return nil, decodeErr
	}

	return json.Marshal(collectionMap)
}

type CollectionsList struct {
	Collections []*Collection `json:"collections"`
	Links       []*Link       `json:"links"`
}
