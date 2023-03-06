package stac

import (
	"encoding/json"
	"fmt"

	"github.com/mitchellh/mapstructure"
)

var coreItemProperties = map[string]bool{
	"datetime":      true,
	"created":       true,
	"updated":       true,
	"gsd":           true,
	"constellation": true,
	"instruments":   true,
	"platform":      true,
}

func IsCoreItemProperty(prop string) bool {
	return coreItemProperties[prop]
}

type Item struct {
	Version    string            `json:"stac_version"`
	Id         string            `json:"id"`
	Geometry   any               `json:"geometry"`
	Bbox       []float64         `json:"bbox,omitempty"`
	Properties map[string]any    `json:"properties"`
	Links      []*Link           `json:"links"`
	Assets     map[string]*Asset `json:"assets"`
	Collection string            `json:"collection,omitempty"`
	Extensions []ItemExtension   `json:"-"`
}

var _ json.Marshaler = (*Item)(nil)

type ItemExtension interface {
	Apply(*Item)
	URI() string
}

func PopulateExtensionFromProperties(extension ItemExtension, properties map[string]any) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  extension,
	})
	if err != nil {
		return err
	}

	return decoder.Decode(properties)
}

func (item Item) MarshalJSON() ([]byte, error) {
	itemMap := map[string]any{"type": "Feature"}
	decoder, decoderErr := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &itemMap,
	})
	if decoderErr != nil {
		return nil, decoderErr
	}

	decodeErr := decoder.Decode(item)
	if decodeErr != nil {
		return nil, decodeErr
	}

	propDecoder, propDecoderErr := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &item.Properties,
	})
	if propDecoderErr != nil {
		return nil, propDecoderErr
	}

	extensionUris := []string{}
	lookup := map[string]bool{}

	for _, asset := range item.Assets {
		for _, extension := range asset.Extensions {
			uri := extension.URI()
			if !lookup[uri] {
				extensionUris = append(extensionUris, uri)
				lookup[uri] = true
			}
		}
	}

	for _, extension := range item.Extensions {
		extension.Apply(&item)
		uri := extension.URI()
		if !lookup[uri] {
			extensionUris = append(extensionUris, uri)
			lookup[uri] = true
		}

		if decodeErr := propDecoder.Decode(extension); decodeErr != nil {
			return nil, fmt.Errorf("trouble encoding JSON for %s item properties: %w", uri, decodeErr)
		}
	}

	if len(extensionUris) > 0 {
		itemMap["stac_extensions"] = extensionUris
	}

	return json.Marshal(itemMap)
}

type ItemsList struct {
	Type  string  `json:"type"`
	Items []*Item `json:"features"`
	Links []*Link `json:"links,omitempty"`
}
