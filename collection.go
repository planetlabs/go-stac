package stac

import (
	"encoding/json"
	"fmt"
	"regexp"

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
	Extensions  []Extension       `json:"-"`
}

var (
	_ json.Marshaler   = (*Collection)(nil)
	_ json.Unmarshaler = (*Collection)(nil)
)

var collectionExtensions = newExtensionRegistry()

func RegisterCollectionExtension(pattern *regexp.Regexp, provider ExtensionProvider) {
	collectionExtensions.register(pattern, provider)
}

func GetCollectionExtension(uri string) Extension {
	return collectionExtensions.get(uri)
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

	extensionUris := []string{}
	lookup := map[string]bool{}

	for _, extension := range collection.Extensions {
		if err := extension.Encode(collectionMap); err != nil {
			return nil, err
		}
		uris, err := GetExtensionUris(collectionMap)
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

	links, linkExtensionUris, err := EncodeLinks(collection.Links)
	if err != nil {
		return nil, err
	}
	collectionMap["links"] = links

	for _, uri := range linkExtensionUris {
		if !lookup[uri] {
			extensionUris = append(extensionUris, uri)
			lookup[uri] = true
		}
	}

	SetExtensionUris(collectionMap, extensionUris)
	return json.Marshal(collectionMap)
}

func (collection *Collection) UnmarshalJSON(data []byte) error {
	collectionMap := map[string]any{}
	if err := json.Unmarshal(data, &collectionMap); err != nil {
		return err
	}

	decoder, decoderErr := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  collection,
	})
	if decoderErr != nil {
		return decoderErr
	}

	if err := decoder.Decode(collectionMap); err != nil {
		return err
	}

	extensionUris, extensionErr := GetExtensionUris(collectionMap)
	if extensionErr != nil {
		return extensionErr
	}
	for _, uri := range extensionUris {
		extension := GetCollectionExtension(uri)
		if extension == nil {
			continue
		}
		if err := extension.Decode(collectionMap); err != nil {
			return fmt.Errorf("decoding error for %s: %w", uri, err)
		}
		collection.Extensions = append(collection.Extensions, extension)
	}

	if err := decodeExtendedLinks(collectionMap, collection.Links, extensionUris); err != nil {
		return err
	}

	return nil
}

type CollectionsList struct {
	Collections []*Collection `json:"collections"`
	Links       []*Link       `json:"links"`
}
