package crawler_test

import (
	"sync/atomic"
	"testing"

	"github.com/planetlabs/go-stac/crawler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssets(t *testing.T) {
	count := uint64(0)

	visitor := func(location string, resource crawler.Resource) error {
		atomic.AddUint64(&count, 1)

		assert.Equal(t, crawler.Item, resource.Type())

		assets := resource.Assets()
		require.Len(t, assets, 3)

		visual, ok := assets["visual"]
		require.True(t, ok)
		assert.Equal(t, "https://storage.googleapis.com/open-cogs/stac-examples/20201211_223832_CS2.tif", visual.Href())
		assert.Equal(t, "image/tiff; application=geotiff; profile=cloud-optimized", visual.Type())
		assert.Equal(t, "3-Band Visual", visual.Title())
		assert.Equal(t, []string{"visual"}, visual.Roles())

		thumbnail, ok := assets["thumbnail"]
		require.True(t, ok)
		assert.Equal(t, "https://storage.googleapis.com/open-cogs/stac-examples/20201211_223832_CS2.jpg", thumbnail.Href())
		assert.Equal(t, "image/jpeg", thumbnail.Type())
		assert.Equal(t, "Thumbnail", thumbnail.Title())
		assert.Equal(t, []string{"thumbnail"}, thumbnail.Roles())

		minimal, ok := assets["minimal"]
		require.True(t, ok)
		assert.Equal(t, "https://example.com/minimal", minimal.Href())
		assert.Equal(t, "", minimal.Type())
		assert.Equal(t, "", minimal.Title())
		assert.Equal(t, []string{}, minimal.Roles())

		return nil
	}

	err := crawler.Crawl("testdata/v1.0.0/item-in-collection.json", visitor)
	assert.NoError(t, err)

	assert.Equal(t, uint64(1), count)
}
