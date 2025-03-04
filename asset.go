package stac

import (
	"regexp"

	"github.com/go-viper/mapstructure/v2"
)

type Asset struct {
	Type        string      `json:"type,omitempty"`
	Href        string      `json:"href,omitempty"`
	Title       string      `json:"title,omitempty"`
	Description string      `json:"description,omitempty"`
	Created     string      `json:"created,omitempty"`
	Roles       []string    `json:"roles,omitempty"`
	Bands       []*Band     `json:"bands,omitempty"`
	Extensions  []Extension `json:"-"`
}

var assetExtensions = newExtensionRegistry()

func RegisterAssetExtension(pattern *regexp.Regexp, provider ExtensionProvider) {
	assetExtensions.register(pattern, provider)
}

func GetAssetExtension(uri string) Extension {
	return assetExtensions.get(uri)
}

func EncodeAssets(assets map[string]*Asset) (map[string]any, []string, error) {
	assetsMap := map[string]any{}
	extensionUris := []string{}
	for key, asset := range assets {
		assetMap := map[string]any{}
		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			TagName: "json",
			Result:  &assetMap,
		})
		if err != nil {
			return nil, nil, err
		}
		if err := decoder.Decode(asset); err != nil {
			return nil, nil, err
		}
		for _, extension := range asset.Extensions {
			extensionUris = append(extensionUris, extension.URI())
			if err := extension.Encode(assetMap); err != nil {
				return nil, nil, err
			}
		}

		bandMaps, uris, err := EncodeBands(asset.Bands)
		if err != nil {
			return nil, nil, err
		}
		if len(bandMaps) > 0 {
			extensionUris = append(extensionUris, uris...)
			assetMap["bands"] = bandMaps
		}

		assetsMap[key] = assetMap
	}
	return assetsMap, extensionUris, nil
}
