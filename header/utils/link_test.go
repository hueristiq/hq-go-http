package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgohttpheaderutils "github.com/hueristiq/hq-lib-http-go/header/utils"
)

func TestLink_String(t *testing.T) {
	t.Parallel()

	link := hqgohttpheaderutils.Link{
		URL: "http://example.com",
		Rel: "next",
		Params: map[string]string{
			"foo": "bar",
			"baz": "qux",
		},
	}

	str := link.String()

	// Parameter keys are emitted in sorted order, followed by rel.
	assert.Equal(t, `<http://example.com>; baz="qux"; foo="bar"; rel="next"`, str)
}

func TestLink_String_EmptyRelAndParams(t *testing.T) {
	t.Parallel()

	link := hqgohttpheaderutils.Link{
		URL:    "http://example.com",
		Rel:    "",
		Params: map[string]string{},
	}

	str := link.String()

	assert.Equal(t, "<http://example.com>", str)
}

func TestLink_HasParam(t *testing.T) {
	t.Parallel()

	link := hqgohttpheaderutils.Link{
		URL: "http://example.com",
		Params: map[string]string{
			"foo": "bar",
		},
	}

	assert.True(t, link.HasParam("foo"))
	assert.False(t, link.HasParam("baz"))
}

func TestLink_Param(t *testing.T) {
	t.Parallel()

	link := hqgohttpheaderutils.Link{
		URL: "http://example.com",
		Params: map[string]string{
			"foo": "bar",
		},
	}

	assert.Equal(t, "bar", link.Param("foo"))
	assert.Empty(t, link.Param("baz"))
}

func TestLinks_String(t *testing.T) {
	t.Parallel()

	links := hqgohttpheaderutils.Links{
		{
			URL: "http://example.com",
			Rel: "next",
			Params: map[string]string{
				"foo": "bar",
			},
		},
		{
			URL: "http://example.org",
			Rel: "prev",
			Params: map[string]string{
				"baz": "qux",
			},
		},
	}

	str := links.String()

	assert.Contains(t, str, "<http://example.com>")
	assert.Contains(t, str, `foo="bar"`)
	assert.Contains(t, str, `rel="next"`)
	assert.Contains(t, str, "<http://example.org>")
	assert.Contains(t, str, `baz="qux"`)
	assert.Contains(t, str, `rel="prev"`)
	assert.Contains(t, str, ", ")
}

func TestLinks_String_Empty(t *testing.T) {
	t.Parallel()

	var links hqgohttpheaderutils.Links

	assert.Empty(t, links.String())
}

func TestLinks_FilterByRel(t *testing.T) {
	t.Parallel()

	links := hqgohttpheaderutils.Links{
		{
			URL: "http://example.com/next",
			Rel: "next",
		},
		{
			URL: "http://example.com/prev",
			Rel: "prev",
		},
		{
			URL: "http://example.com/also-next",
			Rel: "next",
		},
	}

	filtered := links.FilterByRel("next")

	assert.Len(t, filtered, 2)

	for _, link := range filtered {
		assert.Equal(t, "next", link.Rel)
	}
}

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		raw      string
		expected hqgohttpheaderutils.Links
	}{
		{
			name:     "empty",
			raw:      "",
			expected: nil,
		},
		{
			name: "single link",
			raw:  `<http://example.com>; rel="next"; foo="bar"`,
			expected: hqgohttpheaderutils.Links{
				{
					URL:    "http://example.com",
					Rel:    "next",
					Params: map[string]string{"foo": "bar"},
				},
			},
		},
		{
			name: "multiple links",
			raw:  `<http://example.com>; rel="next"; foo="bar", <http://example.org>; rel="prev"; baz="qux"`,
			expected: hqgohttpheaderutils.Links{
				{
					URL:    "http://example.com",
					Rel:    "next",
					Params: map[string]string{"foo": "bar"},
				},
				{
					URL:    "http://example.org",
					Rel:    "prev",
					Params: map[string]string{"baz": "qux"},
				},
			},
		},
		{
			name: "comma inside URL",
			raw:  `<http://example.com/a,b?x=1,y=2>; rel="next", <http://example.org/z>; rel="prev"`,
			expected: hqgohttpheaderutils.Links{
				{
					URL:    "http://example.com/a,b?x=1,y=2",
					Rel:    "next",
					Params: map[string]string{},
				},
				{
					URL:    "http://example.org/z",
					Rel:    "prev",
					Params: map[string]string{},
				},
			},
		},
		{
			name: "unquoted values",
			raw:  `<http://example.com>; rel=next; type=text/html`,
			expected: hqgohttpheaderutils.Links{
				{
					URL:    "http://example.com",
					Rel:    "next",
					Params: map[string]string{"type": "text/html"},
				},
			},
		},
		{
			name: "rel key is matched case-insensitively",
			raw:  `<http://example.com>; REL="next"`,
			expected: hqgohttpheaderutils.Links{
				{
					URL:    "http://example.com",
					Rel:    "next",
					Params: map[string]string{},
				},
			},
		},
		{
			name: "parameter keys are kept as written",
			raw:  `<http://example.com>; rel="next"; Title="Next Page"`,
			expected: hqgohttpheaderutils.Links{
				{
					URL:    "http://example.com",
					Rel:    "next",
					Params: map[string]string{"Title": "Next Page"},
				},
			},
		},
		{
			name: "parameter without value",
			raw:  `<http://example.com>; rel="next"; hreflang`,
			expected: hqgohttpheaderutils.Links{
				{
					URL:    "http://example.com",
					Rel:    "next",
					Params: map[string]string{"hreflang": ""},
				},
			},
		},
		{
			name: "whitespace around separators",
			raw:  "  <http://example.com>  ;  rel=\"next\"  ,  <http://example.org>  ;  rel=\"prev\"  ",
			expected: hqgohttpheaderutils.Links{
				{
					URL:    "http://example.com",
					Rel:    "next",
					Params: map[string]string{},
				},
				{
					URL:    "http://example.org",
					Rel:    "prev",
					Params: map[string]string{},
				},
			},
		},
		{
			name: "entries without a URL are skipped",
			raw:  `rel="next", <http://example.com>; rel="next"`,
			expected: hqgohttpheaderutils.Links{
				{
					URL:    "http://example.com",
					Rel:    "next",
					Params: map[string]string{},
				},
			},
		},
		{
			name:     "only invalid entries",
			raw:      `rel="next"`,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			links := hqgohttpheaderutils.Parse(tt.raw)

			assert.Equal(t, tt.expected, links)
		})
	}
}

func TestParseMultiple(t *testing.T) {
	t.Parallel()

	headers := []string{
		`<http://example.com>; rel="next"; foo="bar"`,
		`<http://example.org>; rel="prev"; baz="qux"`,
	}

	links := hqgohttpheaderutils.ParseMultiple(headers)

	require.Len(t, links, 2)

	assert.Equal(t, "http://example.com", links[0].URL)
	assert.Equal(t, "next", links[0].Rel)
	assert.Equal(t, "bar", links[0].Params["foo"])
	assert.Equal(t, "http://example.org", links[1].URL)
	assert.Equal(t, "prev", links[1].Rel)
	assert.Equal(t, "qux", links[1].Params["baz"])
}
