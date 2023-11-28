package itemassets

import (
	"fmt"
	"regexp"

	"github.com/mitchellh/mapstructure"
	"github.com/planetlabs/go-stac"
)

const (
	extensionUri     = "https://stac-extensions.github.io/item-assets/v1.0.0/schema.json"
	extensionPattern = `https://stac-extensions.github.io/item-assets/v1\..*/schema.json`
)

func init() {
	stac.RegisterCollectionExtension(
		regexp.MustCompile(extensionPattern),
		func() stac.Extension {
			return &Collection{}
		},
	)
}

type Collection struct {
	ItemAssets map[string]*stac.Asset `json:"item_assets,omitempty"`
}

var _ stac.Extension = (*Collection)(nil)

func (*Collection) URI() string {
	return extensionUri
}

func (e *Collection) Encode(collectionMap map[string]any) error {
	assetsMap, extensionUris, err := stac.EncodeAssets(e.ItemAssets)
	if err != nil {
		return err
	}

	collectionMap["item_assets"] = assetsMap
	stac.SetExtensionUris(collectionMap, extensionUris)
	return nil
}

func (e *Collection) Decode(collectionMap map[string]any) error {
	assetsValue, ok := collectionMap["item_assets"]
	if !ok {
		return nil
	}

	assetsMap, ok := assetsValue.(map[string]any)
	if !ok {
		return fmt.Errorf("expected item_assets to be an map[string]any, got %t", assetsValue)
	}

	extensionUris, extensionErr := stac.GetExtensionUris(collectionMap)
	if extensionErr != nil {
		return extensionErr
	}

	itemAssets := map[string]*stac.Asset{}
	for key, assetValue := range assetsMap {
		assetMap, ok := assetValue.(map[string]any)
		if !ok {
			return fmt.Errorf("expected asset to be a map[string]any, got %t", assetValue)
		}
		asset := &stac.Asset{}
		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			TagName: "json",
			Result:  asset,
		})
		if err != nil {
			return err
		}
		if err := decoder.Decode(assetMap); err != nil {
			return err
		}

		for _, uri := range extensionUris {
			if assetExtension := stac.GetAssetExtension(uri); assetExtension != nil {
				if err := assetExtension.Decode(assetMap); err != nil {
					return err
				}
				asset.Extensions = append(asset.Extensions, assetExtension)
			}
		}

		itemAssets[key] = asset
	}

	e.ItemAssets = itemAssets
	return nil
}
