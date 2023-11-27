package stac

import (
	"fmt"
	"regexp"
	"slices"
	"sync"

	"github.com/mitchellh/mapstructure"
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

func EncodeExtendedAsset(assetExtension Extension, assetMap map[string]any) error {
	encoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &assetMap,
	})
	if err != nil {
		return err
	}
	return encoder.Decode(assetExtension)
}

func DecodeExtendedAsset(assetExtension Extension, assetMap map[string]any) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  assetExtension,
	})
	if err != nil {
		return err
	}
	return decoder.Decode(assetMap)
}
