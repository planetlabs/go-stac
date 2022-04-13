package main

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMakeAbsoluteUrl(t *testing.T) {
	cases := []struct {
		dir  string
		link string
		base string
		abs  string
	}{
		{
			dir:  "some/dir",
			link: "./item.json",
			base: "https://example.com/stac",
			abs:  "https://example.com/stac/some/dir/item.json",
		},
		{
			dir:  "some/dir",
			link: "item.json",
			base: "https://example.com/stac",
			abs:  "https://example.com/stac/some/dir/item.json",
		},
		{
			dir:  "some/dir",
			link: "../../catalog.json",
			base: "https://example.com/stac",
			abs:  "https://example.com/stac/catalog.json",
		},
		{
			dir:  "some/dir",
			link: "../../../../../../catalog.json",
			base: "https://example.com/stac",
			abs:  "https://example.com/catalog.json",
		},
		{
			dir:  "../some/dir",
			link: "other/catalog.json",
			base: "https://example.com/stac",
			abs:  "https://example.com/some/dir/other/catalog.json",
		},
		{
			dir:  "some-dir",
			link: "https://example.com/some-dir/item.json",
			base: "https://other.com/stac",
			abs:  "https://example.com/some-dir/item.json",
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			linkUrl, err := url.Parse(c.link)
			require.NoError(t, err)

			baseUrl, err := url.Parse(c.base)
			require.NoError(t, err)

			absUrl := makeAbsolute(linkUrl, c.dir, baseUrl)
			assert.Equal(t, c.abs, absUrl.String())
		})
	}
}
