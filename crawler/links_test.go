package crawler_test

import (
	"fmt"
	"testing"

	"github.com/planetlabs/go-stac/crawler"
	"github.com/stretchr/testify/assert"
)

func TestLinksRel(t *testing.T) {
	cases := []struct {
		links         crawler.Links
		rel           string
		matchers      []crawler.LinkMatcher
		expectedIndex int
	}{
		{
			links: crawler.Links{
				{"rel": "self", "type": "application/json"},
			},
			rel:           "self",
			matchers:      []crawler.LinkMatcher{crawler.LinkTypeApplicationJSON},
			expectedIndex: 0,
		},
		{
			links: crawler.Links{
				{"rel": "other", "type": "text/html"},
				{"rel": "self", "type": "application/json"},
			},
			rel:           "self",
			matchers:      []crawler.LinkMatcher{crawler.LinkTypeApplicationJSON},
			expectedIndex: 1,
		},
		{
			links: crawler.Links{
				{"rel": "self", "type": "text/html"},
				{"rel": "self", "type": "application/json"},
			},
			rel:           "self",
			matchers:      []crawler.LinkMatcher{crawler.LinkTypeApplicationJSON},
			expectedIndex: 1,
		},
		{
			links: crawler.Links{
				{"rel": "self", "type": "text/html"},
				{"rel": "item", "type": "application/ld+json"},
				{"rel": "item", "type": "application/json"},
				{"rel": "item", "type": "application/geo+json"},
			},
			rel:           "item",
			matchers:      []crawler.LinkMatcher{crawler.LinkTypeGeoJSON, crawler.LinkTypeApplicationJSON, crawler.LinkTypeAnyJSON},
			expectedIndex: 3,
		},
		{
			links: crawler.Links{
				{"rel": "self", "type": "text/html"},
				{"rel": "item", "type": "application/ld+json"},
				{"rel": "item", "type": "application/json"},
				{"rel": "item"},
			},
			rel:           "item",
			matchers:      []crawler.LinkMatcher{crawler.LinkTypeGeoJSON, crawler.LinkTypeApplicationJSON, crawler.LinkTypeAnyJSON},
			expectedIndex: 2,
		},
		{
			links: crawler.Links{
				{"rel": "self", "type": "text/html"},
				{"rel": "item", "type": "text/html"},
				{"rel": "item"},
			},
			rel:           "item",
			matchers:      []crawler.LinkMatcher{crawler.LinkTypeGeoJSON, crawler.LinkTypeApplicationJSON, crawler.LinkTypeAnyJSON, crawler.LinkTypeNone},
			expectedIndex: 2,
		},
		{
			links: crawler.Links{
				{"rel": "self", "type": "text/html"},
				{"rel": "item", "type": "application/geo+json"},
				{"rel": "item"},
			},
			rel:           "item",
			matchers:      []crawler.LinkMatcher{crawler.LinkTypeApplicationJSON, crawler.LinkTypeAnyJSON, crawler.LinkTypeNone},
			expectedIndex: 1,
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			link := c.links.Rel(c.rel, c.matchers...)
			if c.expectedIndex == -1 {
				assert.Nil(t, link)
			} else {
				assert.Equal(t, c.links[c.expectedIndex], link)
			}
		})
	}
}
