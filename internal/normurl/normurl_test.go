package normurl_test

import (
	"encoding/json"
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

func TestJSONRoundTrip(t *testing.T) {
	cases := []string{
		"https://example.com/path/to/file",
		"/path/to/file",
		"file:///path/to/file",
	}

	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			original, newErr := normurl.New(c)
			require.NoError(t, newErr)

			serialized, marshalErr := json.Marshal(original)
			require.NoError(t, marshalErr)

			var deserialized normurl.Locator
			unmarshalErr := json.Unmarshal(serialized, &deserialized)
			require.NoError(t, unmarshalErr)

			assert.Equal(t, original.String(), deserialized.String())
			assert.Equal(t, original.IsFilepath(), deserialized.IsFilepath())
		})
	}
}

func TestUnmarshalJSON(t *testing.T) {
	cases := []struct {
		input          string
		expectedString string
		expectedFile   bool
		expectedErr    error
	}{
		{
			input:          `{"Url": "https://example.com/path/to/file", "File": false}`,
			expectedString: "https://example.com/path/to/file",
			expectedFile:   false,
		},
		{
			input:          `{"Url": "https://example.com/path/to/file"}`,
			expectedString: "https://example.com/path/to/file",
			expectedFile:   false,
		},
		{
			input:          `{"Url": "file:///path/to/file", "File": true}`,
			expectedString: "/path/to/file",
			expectedFile:   true,
		},
		{
			input:       `{"Url": "../path/to/file", "File": true}`,
			expectedErr: errors.New("expected absolute path"),
		},
		{
			input:       `{"Url": "https://example.com/path/to/file", "File": true}`,
			expectedErr: errors.New("file flag mismatch"),
		},
		{
			input:       `{"Url": "/path/to/file", "File": false}`,
			expectedErr: errors.New("file flag mismatch"),
		},
		{
			input:       `{"foo": "bar"}`,
			expectedErr: errors.New("missing url"),
		},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			var l normurl.Locator
			err := json.Unmarshal([]byte(c.input), &l)
			if c.expectedErr != nil {
				assert.EqualError(t, err, c.expectedErr.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, c.expectedString, l.String())
				assert.Equal(t, c.expectedFile, l.IsFilepath())
			}
		})
	}
}

func TestMarshalJSON(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{
			input:    "https://example.com/path/to/file",
			expected: `{"Url": "https://example.com/path/to/file", "File": false}`,
		},
		{
			input:    "/path/to/file",
			expected: `{"Url": "/path/to/file", "File": true}`,
		},
		{
			input:    "file:///path/to/file",
			expected: `{"Url": "/path/to/file", "File": true}`,
		},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			original, newErr := normurl.New(c.input)
			require.NoError(t, newErr)

			serialized, marshalErr := json.Marshal(original)
			require.NoError(t, marshalErr)

			assert.JSONEq(t, c.expected, string(serialized))
		})
	}
}
