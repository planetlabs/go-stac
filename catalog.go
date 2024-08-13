package stac

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/go-viper/mapstructure/v2"
)

type Catalog struct {
	Version     string      `json:"stac_version"`
	Id          string      `json:"id"`
	Title       string      `json:"title,omitempty"`
	Description string      `json:"description"`
	Links       []*Link     `json:"links"`
	ConformsTo  []string    `json:"conformsTo,omitempty"`
	Extensions  []Extension `json:"-"`
}

var (
	_ json.Marshaler   = (*Catalog)(nil)
	_ json.Unmarshaler = (*Catalog)(nil)
)

var catalogExtensions = newExtensionRegistry()

func RegisterCatalogExtension(pattern *regexp.Regexp, provider ExtensionProvider) {
	catalogExtensions.register(pattern, provider)
}

func GetCatalogExtension(uri string) Extension {
	return catalogExtensions.get(uri)
}

func (catalog Catalog) MarshalJSON() ([]byte, error) {
	catalogMap := map[string]any{
		"type": "Catalog",
	}
	decoder, decoderErr := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &catalogMap,
	})
	if decoderErr != nil {
		return nil, decoderErr
	}

	decodeErr := decoder.Decode(catalog)
	if decodeErr != nil {
		return nil, decodeErr
	}

	extensionUris := []string{}
	lookup := map[string]bool{}

	for _, extension := range catalog.Extensions {
		if err := extension.Encode(catalogMap); err != nil {
			return nil, err
		}
		uris, err := GetExtensionUris(catalogMap)
		if err != nil {
			return nil, err
		}
		uris = append(uris, extension.URI())
		for _, uri := range uris {
			if !lookup[uri] {
				extensionUris = append(extensionUris, uri)
				lookup[uri] = true
			}
		}
	}

	links, linkExtensionUris, err := EncodeLinks(catalog.Links)
	if err != nil {
		return nil, err
	}
	catalogMap["links"] = links

	for _, uri := range linkExtensionUris {
		if !lookup[uri] {
			extensionUris = append(extensionUris, uri)
			lookup[uri] = true
		}
	}

	SetExtensionUris(catalogMap, extensionUris)
	return json.Marshal(catalogMap)
}

func (catalog *Catalog) UnmarshalJSON(data []byte) error {
	catalogMap := map[string]any{}
	if err := json.Unmarshal(data, &catalogMap); err != nil {
		return err
	}

	decoder, decoderErr := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  catalog,
	})
	if decoderErr != nil {
		return decoderErr
	}

	if err := decoder.Decode(catalogMap); err != nil {
		return err
	}

	extensionUris, err := GetExtensionUris(catalogMap)
	if err != nil {
		return err
	}

	for _, uri := range extensionUris {
		extension := GetCatalogExtension(uri)
		if extension == nil {
			continue
		}
		if err := extension.Decode(catalogMap); err != nil {
			return fmt.Errorf("decoding error for %s: %w", uri, err)
		}
		catalog.Extensions = append(catalog.Extensions, extension)
	}

	if err := decodeExtendedLinks(catalogMap, catalog.Links, extensionUris); err != nil {
		return err
	}

	return nil
}
