package normurl_test

import (
	"errors"
	"fmt"
	"net/url"
	"testing"

	"github.com/planetlabs/go-stac/internal/normurl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	cases := []struct {
		input      string
		expected   string
		isFilepath bool
		err        error
	}{
		{
			input:      "https://example.com",
			expected:   "https://example.com",
			isFilepath: false,
		},
		{
			input:      "http://example.com/foo/bar",
			expected:   "http://example.com/foo/bar",
			isFilepath: false,
		},
		{
			input:      "/foo/bar",
			expected:   "/foo/bar",
			isFilepath: true,
		},
		{
			input:      "file:///foo/bar",
			expected:   "/foo/bar",
			isFilepath: true,
		},
		{
			input: "foo/bar",
			err:   errors.New("expected absolute path"),
		},
		{
			input: "bogus://foo/bar",
			err:   errors.New("unsupported scheme bogus"),
		},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			l, err := normurl.New(c.input)
			if c.err != nil {
				assert.Nil(t, l)
				require.Error(t, err)
				assert.EqualError(t, err, c.err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, c.expected, l.String())
				assert.Equal(t, c.isFilepath, l.IsFilepath())
			}
		})
	}
}

func TestSetQueryParam(t *testing.T) {
	cases := []struct {
		input    string
		key      string
		value    string
		expected string
	}{
		{
			input:    "https://example.com",
			key:      "foo",
			value:    "bar",
			expected: "https://example.com?foo=bar",
		},
		{
			input:    "https://example.com?foo=bar",
			key:      "baz",
			value:    "qux",
			expected: "https://example.com?foo=bar&baz=qux",
		},
		{
			input:    "https://example.com?foo=bar&baz=qux",
			key:      "baz",
			value:    "bam",
			expected: "https://example.com?foo=bar&baz=bam",
		},
		{
			input:    "https://example.com?foo=bar&baz=qux",
			key:      "baz",
			value:    "",
			expected: "https://example.com?foo=bar",
		},
		{
			input:    "https://example.com?foo=bar",
			key:      "baz",
			value:    "",
			expected: "https://example.com?foo=bar",
		},
		{
			input:    "/path/to/file",
			key:      "foo",
			value:    "bar",
			expected: "/path/to/file",
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			l, err := normurl.New(c.input)
			require.NoError(t, err)

			l.SetQueryParam(c.key, c.value)

			expUrl, err := url.Parse(c.expected)
			require.NoError(t, err)
			expQuery := expUrl.Query()

			gotUrl, err := url.Parse(l.String())
			require.NoError(t, err)
			gotQuery := gotUrl.Query()

			assert.Equal(t, expQuery, gotQuery)
		})
	}
}

func TestResolve(t *testing.T) {
	cases := []struct {
		base     string
		input    string
		expected string
		err      error
	}{
		{
			base:     "https://example.com",
			input:    "foo/bar",
			expected: "https://example.com/foo/bar",
		},
		{
			base:     "https://example.com",
			input:    "/foo/bar",
			expected: "https://example.com/foo/bar",
		},
		{
			base:     "https://example.com",
			input:    "../foo/bar",
			expected: "https://example.com/foo/bar",
		},
		{
			base:     "/foo/bar",
			input:    "bam",
			expected: "/foo/bam",
		},
		{
			base:     "/foo/bar",
			input:    "./bam",
			expected: "/foo/bam",
		},
		{
			base:     "/foo/bar",
			input:    "../../bam",
			expected: "/bam",
		},
		{
			base:     "/foo/bar/",
			input:    "bam",
			expected: "/foo/bar/bam",
		},
		{
			base:     "/foo/bar/",
			input:    "./bam",
			expected: "/foo/bar/bam",
		},
		{
			base:     "/foo/bar/",
			input:    "../../bam",
			expected: "/bam",
		},
		{
			base:     "/foo/bar/",
			input:    "https://example.com/bam",
			expected: "https://example.com/bam",
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			l, err := normurl.New(c.base)
			require.NoError(t, err)

			resolved, err := l.Resolve(c.input)
			if c.err != nil {
				assert.Nil(t, resolved)
				require.Error(t, err)
				assert.EqualError(t, err, c.err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, c.expected, resolved.String())
			}
		})
	}
}
