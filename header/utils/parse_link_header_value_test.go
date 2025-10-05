package utils_test

import (
	"testing"

	hqgohttpheaderutils "github.com/hueristiq/hq-go-http/header/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinkString(t *testing.T) {
	t.Parallel()

	link := hqgohttpheaderutils.ParsedLink{
		URL: "http://example.com",
		Rel: "next",
		Parameters: map[string]string{
			"foo": "bar",
			"baz": "qux",
		},
	}

	str := link.String()

	assert.Contains(t, str, "<http://example.com>")
	assert.Contains(t, str, `foo="bar"`)
	assert.Contains(t, str, `baz="qux"`)
	assert.Contains(t, str, `rel="next"`)
	assert.Contains(t, str, "; ")
}

func TestLinkString_EmptyRelAndParams(t *testing.T) {
	t.Parallel()

	link := hqgohttpheaderutils.ParsedLink{
		URL:        "http://example.com",
		Rel:        "",
		Parameters: map[string]string{},
	}

	str := link.String()

	assert.Equal(t, "<http://example.com>; ", str)
}

func TestLinkHasParameter(t *testing.T) {
	t.Parallel()

	link := hqgohttpheaderutils.ParsedLink{
		URL: "http://example.com",
		Parameters: map[string]string{
			"foo": "bar",
		},
	}

	assert.True(t, link.HasParameter("foo"))
	assert.False(t, link.HasParameter("baz"))
}

func TestLinkParameter(t *testing.T) {
	t.Parallel()

	link := hqgohttpheaderutils.ParsedLink{
		URL: "http://example.com",
		Parameters: map[string]string{
			"foo": "bar",
		},
	}

	assert.Equal(t, "bar", link.Parameter("foo"))
	assert.Empty(t, link.Parameter("baz"))
}

func TestLinksString(t *testing.T) {
	t.Parallel()

	links := hqgohttpheaderutils.ParsedLinks{
		{
			URL: "http://example.com",
			Rel: "next",
			Parameters: map[string]string{
				"foo": "bar",
			},
		},
		{
			URL: "http://example.org",
			Rel: "prev",
			Parameters: map[string]string{
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

// TestLinksFilterByRel verifies that FilterByRel returns only the Links with the specified rel attribute.
func TestLinksFilterByRel(t *testing.T) {
	t.Parallel()

	links := hqgohttpheaderutils.ParsedLinks{
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

func TestParseLinkHeaderValue_Empty(t *testing.T) {
	t.Parallel()

	links := hqgohttpheaderutils.ParseLinkHeaderValue("")

	assert.Empty(t, links)
}

func TestParseLinkHeaderValue_Single(t *testing.T) {
	t.Parallel()

	raw := `<http://example.com>; rel="next"; foo="bar"`

	links := hqgohttpheaderutils.ParseLinkHeaderValue(raw)

	require.Len(t, links, 1)

	link := links[0]

	assert.Equal(t, "http://example.com", link.URL)
	assert.Equal(t, "next", link.Rel)
	assert.Equal(t, "bar", link.Parameters["foo"])
}

func TestParseLinkHeaderValue_Multiple(t *testing.T) {
	t.Parallel()

	raw := `<http://example.com>; rel="next"; foo="bar", <http://example.org>; rel="prev"; baz="qux"`

	links := hqgohttpheaderutils.ParseLinkHeaderValue(raw)

	require.Len(t, links, 2)

	link1 := links[0]

	assert.Equal(t, "http://example.com", link1.URL)
	assert.Equal(t, "next", link1.Rel)
	assert.Equal(t, "bar", link1.Parameters["foo"])

	link2 := links[1]

	assert.Equal(t, "http://example.org", link2.URL)
	assert.Equal(t, "prev", link2.Rel)
	assert.Equal(t, "qux", link2.Parameters["baz"])
}

func TestParseLinkHeaderValues(t *testing.T) {
	t.Parallel()

	headers := []string{
		`<http://example.com>; rel="next"; foo="bar"`,
		`<http://example.org>; rel="prev"; baz="qux"`,
	}

	links := hqgohttpheaderutils.ParseLinkHeaderValues(headers)

	require.Len(t, links, 2)

	assert.Equal(t, "http://example.com", links[0].URL)
	assert.Equal(t, "next", links[0].Rel)
	assert.Equal(t, "bar", links[0].Parameters["foo"])
	assert.Equal(t, "http://example.org", links[1].URL)
	assert.Equal(t, "prev", links[1].Rel)
	assert.Equal(t, "qux", links[1].Parameters["baz"])
}
