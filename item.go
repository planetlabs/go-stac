package stac

import (
	"encoding/json"
	"fmt"
	"regexp"

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
	Extensions []Extension       `json:"-"`
}

var (
	_ json.Marshaler   = (*Item)(nil)
	_ json.Unmarshaler = (*Item)(nil)
)

var itemExtensions = newExtensionRegistry()

func RegisterItemExtension(pattern *regexp.Regexp, provider ExtensionProvider) {
	itemExtensions.register(pattern, provider)
}

func GetItemExtension(uri string) Extension {
	return itemExtensions.get(uri)
}

func PopulateExtensionFromProperties(extension Extension, properties map[string]any) error {
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

	assetsMap, assetExtensionUris, err := EncodeAssets(item.Assets)
	if err != nil {
		return nil, err
	}
	itemMap["assets"] = assetsMap

	extensionUris := []string{}
	lookup := map[string]bool{}

	for _, uri := range assetExtensionUris {
		if !lookup[uri] {
			extensionUris = append(extensionUris, uri)
			lookup[uri] = true
		}
	}

	for _, extension := range item.Extensions {
		if err := extension.Encode(itemMap); err != nil {
			return nil, err
		}
		uri := extension.URI()
		if !lookup[uri] {
			extensionUris = append(extensionUris, uri)
			lookup[uri] = true
		}
	}

	SetExtensionUris(itemMap, extensionUris)

	return json.Marshal(itemMap)
}

func (item *Item) UnmarshalJSON(data []byte) error {
	itemMap := map[string]any{}
	if err := json.Unmarshal(data, &itemMap); err != nil {
		return err
	}

	extensionUris, extensionErr := GetExtensionUris(itemMap)
	if extensionErr != nil {
		return extensionErr
	}
	for _, uri := range extensionUris {
		extension := GetItemExtension(uri)
		if extension == nil {
			continue
		}
		if err := extension.Decode(itemMap); err != nil {
			return fmt.Errorf("decoding error for %s: %w", uri, err)
		}
		item.Extensions = append(item.Extensions, extension)
	}

	decoder, decoderErr := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  item,
	})
	if decoderErr != nil {
		return decoderErr
	}

	if err := decoder.Decode(itemMap); err != nil {
		return err
	}

	assetsValue, ok := itemMap["assets"]
	if !ok {
		return nil
	}
	assetsMap, ok := assetsValue.(map[string]any)
	if !ok {
		return fmt.Errorf("unexpected type for assets: %t", assetsValue)
	}

	for key, asset := range item.Assets {
		for _, uri := range extensionUris {
			extension := GetAssetExtension(uri)
			if extension == nil {
				continue
			}
			assetMap, ok := assetsMap[key].(map[string]any)
			if !ok {
				return fmt.Errorf("unexpected type for %q asset: %t", key, assetsMap[key])
			}
			if err := extension.Decode(assetMap); err != nil {
				return fmt.Errorf("decoding error for %s: %w", uri, err)
			}
			asset.Extensions = append(asset.Extensions, extension)
		}
	}

	return nil
}

type ItemsList struct {
	Type  string  `json:"type"`
	Items []*Item `json:"features"`
	Links []*Link `json:"links,omitempty"`
}
