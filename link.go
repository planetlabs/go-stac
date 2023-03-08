package stac

import (
	"encoding/json"

	"github.com/mitchellh/mapstructure"
)

type Link struct {
	Href             string         `mapstructure:"href"`
	Rel              string         `mapstructure:"rel"`
	Type             string         `mapstructure:"type,omitempty"`
	Title            string         `mapstructure:"title,omitempty"`
	AdditionalFields map[string]any `mapstructure:",remain"`
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
	// remove if https://github.com/mitchellh/mapstructure/issues/279 is fixed
	delete(m, "AdditionalFields")
	for k, v := range link.AdditionalFields {
		m[k] = v
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
