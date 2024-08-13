package stac

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"sync"

	"github.com/go-viper/mapstructure/v2"
)

type Extension interface {
	URI() string
	Encode(map[string]any) error
	Decode(map[string]any) error
}

type ExtensionProvider func() Extension

type extensionRegistry struct {
	mutex      *sync.RWMutex
	extensions map[*regexp.Regexp]ExtensionProvider
}

func newExtensionRegistry() *extensionRegistry {
	return &extensionRegistry{
		mutex:      &sync.RWMutex{},
		extensions: map[*regexp.Regexp]ExtensionProvider{},
	}
}

func (r *extensionRegistry) register(pattern *regexp.Regexp, provider ExtensionProvider) {
	r.mutex.Lock()
	r.extensions[pattern] = provider
	r.mutex.Unlock()
}

func (r *extensionRegistry) get(uri string) Extension {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	for pattern, provider := range r.extensions {
		if pattern.Match([]byte(uri)) {
			return provider()
		}
	}
	return nil
}

const uriKey = "stac_extensions"

func GetExtensionUris(data map[string]any) ([]string, error) {
	value, ok := data[uriKey]
	if !ok {
		return nil, nil
	}

	if uris, ok := value.([]string); ok {
		return uris, nil
	}

	values, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("unexpected type for %s: %t", uriKey, value)
	}
	uris := make([]string, len(values))
	for i, v := range values {
		uri, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("expected strings for %s, got %t", uriKey, v)
		}
		uris[i] = uri
	}
	return uris, nil
}

func SetExtensionUris(data map[string]any, uris []string) {
	if len(uris) == 0 {
		return
	}
	slices.Sort(uris)
	data[uriKey] = uris
}

const propertiesKey = "properties"

func EncodeExtendedItemProperties(itemExtension Extension, itemMap map[string]any) error {
	properties := itemMap[propertiesKey]
	if properties == nil {
		properties = map[string]any{}
	}
	encoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &properties,
	})
	if err != nil {
		return err
	}
	if err := encoder.Decode(itemExtension); err != nil {
		return err
	}
	itemMap[propertiesKey] = properties
	return nil
}

func DecodeExtendedItemProperties(itemExtension Extension, itemMap map[string]any) error {
	propertiesValue, ok := itemMap[propertiesKey]
	if !ok {
		return nil
	}
	properties, ok := propertiesValue.(map[string]any)
	if !ok {
		return fmt.Errorf("unexpected properties type: %t", propertiesValue)
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  itemExtension,
	})
	if err != nil {
		return err
	}
	if err := decoder.Decode(properties); err != nil {
		return err
	}

	extendedProperties := map[string]any{}
	encoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &extendedProperties,
	})
	if err != nil {
		return err
	}
	if err := encoder.Decode(itemExtension); err != nil {
		return err
	}
	for key := range extendedProperties {
		delete(properties, key)
	}
	itemMap[propertiesKey] = properties
	return nil
}

func EncodeExtendedMap(extension Extension, data map[string]any) error {
	encoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &data,
	})
	if err != nil {
		return err
	}
	return encoder.Decode(extension)
}

func DecodeExtendedMap(extension Extension, data map[string]any) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  extension,
	})
	if err != nil {
		return err
	}
	return decoder.Decode(data)
}

func decodeExtendedLinks(data map[string]any, links []*Link, extensionUris []string) error {
	linksValue, ok := data["links"]
	if !ok {
		return nil
	}
	linksData, ok := linksValue.([]any)
	if !ok {
		return fmt.Errorf("unexpected type for links: %t", linksValue)
	}

	for i, link := range links {
		for _, uri := range extensionUris {
			extension := GetLinkExtension(uri)
			if extension == nil {
				continue
			}
			linkMap, ok := linksData[i].(map[string]any)
			if !ok {
				return fmt.Errorf("unexpected type for %q link: %t", i, linksData[i])
			}
			if err := extension.Decode(linkMap); err != nil {
				if errors.Is(err, ErrExtensionDoesNotApply) {
					continue
				}
				return fmt.Errorf("decoding error for %s: %w", uri, err)
			}
			link.Extensions = append(link.Extensions, extension)
		}
	}

	return nil
}

func decodeExtendedAssets(data map[string]any, assets map[string]*Asset, extensionUris []string) error {
	assetsValue, ok := data["assets"]
	if !ok {
		return nil
	}
	assetsMap, ok := assetsValue.(map[string]any)
	if !ok {
		return fmt.Errorf("unexpected type for assets: %t", assetsValue)
	}

	for key, asset := range assets {
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
				if errors.Is(err, ErrExtensionDoesNotApply) {
					continue
				}
				return fmt.Errorf("decoding error for %s: %w", uri, err)
			}
			asset.Extensions = append(asset.Extensions, extension)
		}
	}

	return nil
}

// ErrExtensionDoesNotApply is returned when decoding JSON and an extension referenced in stac_extensions does not apply to a value.
var ErrExtensionDoesNotApply = errors.New("extension does not apply")
