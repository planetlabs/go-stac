package stac

import (
	"encoding/json"
	"regexp"

	"github.com/go-viper/mapstructure/v2"
)

type Link struct {
	Href             string         `mapstructure:"href"`
	Rel              string         `mapstructure:"rel"`
	Type             string         `mapstructure:"type,omitempty"`
	Title            string         `mapstructure:"title,omitempty"`
	Method           string         `mapstructure:"method,omitempty"`
	Headers          map[string]any `mapstructure:"headers,omitempty"`
	Body             any            `mapstructure:"body,omitempty"`
	Extensions       []Extension    `mapstructure:"-"`
	AdditionalFields map[string]any `mapstructure:",remain"`
}

var linkExtensions = newExtensionRegistry()

func RegisterLinkExtension(pattern *regexp.Regexp, provider ExtensionProvider) {
	linkExtensions.register(pattern, provider)
}

func GetLinkExtension(uri string) Extension {
	return linkExtensions.get(uri)
}

func EncodeLinks(links []*Link) ([]map[string]any, []string, error) {
	extensionUris := []string{}
	linksData := make([]map[string]any, len(links))
	for i, link := range links {
		linkMap := map[string]any{}
		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Result: &linkMap,
		})
		if err != nil {
			return nil, nil, err
		}
		if err := decoder.Decode(link); err != nil {
			return nil, nil, err
		}

		for _, extension := range link.Extensions {
			extensionUris = append(extensionUris, extension.URI())
			if err := extension.Encode(linkMap); err != nil {
				return nil, nil, err
			}
		}
		linksData[i] = linkMap
	}
	return linksData, extensionUris, nil
}

var (
	_ json.Marshaler   = (*Link)(nil)
	_ json.Unmarshaler = (*Link)(nil)
)

func (link *Link) MarshalJSON() ([]byte, error) {
	m := map[string]any{}
	if err := mapstructure.Decode(link, &m); err != nil {
		return nil, err
	}
	return json.Marshal(m)
}

func (link *Link) UnmarshalJSON(data []byte) error {
	m := map[string]any{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	return mapstructure.Decode(m, link)
}
