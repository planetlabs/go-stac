package stac_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/planetlabs/go-stac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLink(t *testing.T) {
	cases := []struct {
		link *stac.Link
		data string
	}{
		{
			link: &stac.Link{
				Rel:  "test",
				Href: "https://example.com/test",
			},
			data: `{
				"rel": "test",
				"href": "https://example.com/test"
			}`,
		},
		{
			link: &stac.Link{
				Rel:   "test",
				Title: "Test Link",
				Type:  "text/plain",
				Href:  "https://example.com/test",
			},
			data: `{
				"rel": "test",
				"title": "Test Link",
				"type": "text/plain",
				"href": "https://example.com/test"
			}`,
		},
		{
			link: &stac.Link{
				Rel:    "test",
				Href:   "https://example.com/test",
				Method: "GET",
				Headers: map[string]any{
					"Content-Type": "test-content-type",
				},
				Body: map[string]any{
					"foo": "bar",
				},
			},
			data: `{
				"rel": "test",
				"href": "https://example.com/test",
				"method": "GET",
				"headers": {
					"Content-Type": "test-content-type"
				},
				"body": {
					"foo": "bar"
				}
			}`,
		},
		{
			link: &stac.Link{
				Rel:    "test",
				Href:   "https://example.com/test",
				Method: "GET",
				AdditionalFields: map[string]any{
					"custom": "custom-link",
				},
			},
			data: `{
				"rel": "test",
				"href": "https://example.com/test",
				"method": "GET",
				"custom": "custom-link"
			}`,
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			data, err := json.Marshal(c.link)
			require.NoError(t, err)
			assert.JSONEq(t, c.data, string(data))

			l := &stac.Link{}
			require.NoError(t, json.Unmarshal([]byte(c.data), l))
			assert.Equal(t, c.link, l)
		})
	}
}
