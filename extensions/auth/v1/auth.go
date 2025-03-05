package auth

import (
	"fmt"
	"regexp"

	"github.com/go-viper/mapstructure/v2"
	"github.com/planetlabs/go-stac"
)

const (
	extensionUri     = "https://stac-extensions.github.io/authentication/v1.1.0/schema.json"
	extensionPattern = `https://stac-extensions.github.io/authentication/v1\..*/schema.json`
	prefix           = "auth"
	schemesKey       = "auth:schemes"
	refsKey          = "auth:refs"
)

func init() {
	r := regexp.MustCompile(extensionPattern)

	stac.RegisterItemExtension(r, func() stac.Extension { return &Item{} })
	stac.RegisterCollectionExtension(r, func() stac.Extension { return &Collection{} })
	stac.RegisterCatalogExtension(r, func() stac.Extension { return &Catalog{} })
	stac.RegisterAssetExtension(r, func() stac.Extension { return &Asset{} })
	stac.RegisterLinkExtension(r, func() stac.Extension { return &Link{} })
}

type Item struct {
	Schemes map[string]*Scheme `json:"auth:schemes,omitempty"`
}

var _ stac.Extension = (*Item)(nil)

func (*Item) URI() string {
	return extensionUri
}

func (e *Item) Encode(itemMap map[string]any) error {
	return stac.EncodeExtendedItemProperties(e, itemMap)
}

func (e *Item) Decode(itemMap map[string]any) error {
	if err := stac.DecodeExtendedItemProperties(e, itemMap); err != nil {
		return err
	}
	if e.Schemes == nil {
		return stac.ErrExtensionDoesNotApply
	}
	return nil
}

type Collection struct {
	Schemes map[string]*Scheme
}

var _ stac.Extension = (*Collection)(nil)

func (*Collection) URI() string {
	return extensionUri
}

func (e *Collection) Encode(collectionMap map[string]any) error {
	collectionMap[schemesKey] = e.Schemes
	return nil
}

func (e *Collection) Decode(collectionMap map[string]any) error {
	schemes, err := decodeSchemes(collectionMap)
	if err != nil {
		return err
	}
	e.Schemes = schemes
	return nil
}

type Catalog struct {
	Schemes map[string]*Scheme
}

var _ stac.Extension = (*Catalog)(nil)

func (*Catalog) URI() string {
	return extensionUri
}

func (e *Catalog) Encode(catalogMap map[string]any) error {
	catalogMap[schemesKey] = e.Schemes
	return nil
}

func (e *Catalog) Decode(catalogMap map[string]any) error {
	schemes, err := decodeSchemes(catalogMap)
	if err != nil {
		return err
	}
	e.Schemes = schemes
	return nil
}

func decodeSchemes(data map[string]any) (map[string]*Scheme, error) {
	schemesValue, ok := data[schemesKey]
	if !ok {
		return nil, nil
	}

	schemesMap, ok := schemesValue.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected %s to be an map[string]any, got %T", schemesKey, schemesValue)
	}

	schemes := map[string]*Scheme{}
	for key, schemeValue := range schemesMap {
		schemeMap, ok := schemeValue.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("expected scheme to be a map[string]any, got %T", schemeValue)
		}
		scheme := &Scheme{}
		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			TagName: "json",
			Result:  scheme,
		})
		if err != nil {
			return nil, err
		}
		if err := decoder.Decode(schemeMap); err != nil {
			return nil, err
		}

		schemes[key] = scheme
	}
	return schemes, nil
}

type Scheme struct {
	Type             string         `json:"type"`
	Description      string         `json:"description,omitempty"`
	Name             string         `json:"name,omitempty"`
	In               string         `json:"in,omitempty"`
	Scheme           string         `json:"scheme,omitempty"`
	Flows            map[string]any `json:"flows,omitempty"`
	OpenIdConnectUrl string         `json:"openIdConnectUrl,omitempty"`
}

type Asset struct {
	Refs []string `json:"auth:refs,omitempty"`
}

var _ stac.Extension = (*Asset)(nil)

func (*Asset) URI() string {
	return extensionUri
}

func (e *Asset) Encode(assetMap map[string]any) error {
	return stac.EncodeExtendedMap(e, assetMap)
}

func (e *Asset) Decode(assetMap map[string]any) error {
	if _, ok := assetMap[refsKey]; !ok {
		return stac.ErrExtensionDoesNotApply
	}
	return stac.DecodeExtendedMap(e, assetMap, prefix)
}

type Link struct {
	Refs []string `json:"auth:refs,omitempty"`
}

var _ stac.Extension = (*Link)(nil)

func (*Link) URI() string {
	return extensionUri
}

func (e *Link) Encode(linkMap map[string]any) error {
	return stac.EncodeExtendedMap(e, linkMap)
}

func (e *Link) Decode(linkMap map[string]any) error {
	if _, ok := linkMap[refsKey]; !ok {
		return stac.ErrExtensionDoesNotApply
	}
	return stac.DecodeExtendedMap(e, linkMap, prefix)
}
