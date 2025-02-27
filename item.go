package stac

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"

	"github.com/go-viper/mapstructure/v2"
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

	extensionUris := []string{}
	lookup := map[string]bool{}

	assetsMap, assetExtensionUris, err := EncodeAssets(item.Assets)
	if err != nil {
		return nil, err
	}
	itemMap["assets"] = assetsMap

	for _, uri := range assetExtensionUris {
		if !lookup[uri] {
			extensionUris = append(extensionUris, uri)
			lookup[uri] = true
		}
	}

	links, linkExtensionUris, err := EncodeLinks(item.Links)
	if err != nil {
		return nil, err
	}
	itemMap["links"] = links

	for _, uri := range linkExtensionUris {
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

var geometryUnmarshaler json.Unmarshaler

// GeometryUnmarshaler allows a custom geometry type that satisfies the json.Unmarshaler interface to be provided.
// If not set, item geometries will be unmarshaled as a map[string]any.
func GeometryUnmarshaler(g json.Unmarshaler) {
	geometryUnmarshaler = g
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
			if errors.Is(err, ErrExtensionDoesNotApply) {
				continue
			}
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

	if err := decodeExtendedAssets(itemMap, item.Assets, extensionUris); err != nil {
		return err
	}

	if err := decodeExtendedLinks(itemMap, item.Links, extensionUris); err != nil {
		return err
	}

	if geometryUnmarshaler != nil {
		geometryMap, ok := itemMap["geometry"]
		if ok {
			geometryData, err := json.Marshal(geometryMap)
			if err != nil {
				return err
			}

			var gv any
			gvt := reflect.TypeOf(geometryUnmarshaler)
			if gvt.Kind() == reflect.Pointer {
				gv = reflect.New(gvt.Elem()).Interface()
			} else {
				gv = reflect.New(gvt).Elem().Interface()
			}
			g, ok := gv.(json.Unmarshaler)
			if !ok {
				return fmt.Errorf("expected %#v to satisfy the json.Unmarshaler interface", gv)
			}
			if err := g.UnmarshalJSON(geometryData); err != nil {
				return err
			}

			item.Geometry = g
		}
	}

	return nil
}

type ItemsList struct {
	Type  string  `json:"type"`
	Items []*Item `json:"features"`
	Links []*Link `json:"links,omitempty"`
}
