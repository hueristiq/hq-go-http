package header_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.source.hueristiq.com/http/header"
)

func TestHeaderString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		header   header.Header
		expected string
	}{
		// Authentication headers.
		{"Authorization", header.Authorization, "Authorization"},
		{"ProxyAuthenticate", header.ProxyAuthenticate, "Proxy-Authenticate"},
		{"ProxyAuthorization", header.ProxyAuthorization, "Proxy-Authorization"},
		{"WWWAuthenticate", header.WWWAuthenticate, "WWW-Authenticate"},

		// Caching headers.
		{"Age", header.Age, "Age"},
		{"CacheControl", header.CacheControl, "Cache-Control"},
		{"ClearSiteData", header.ClearSiteData, "Clear-Site-Data"},
		{"Expires", header.Expires, "Expires"},
		{"Pragma", header.Pragma, "Pragma"},
		{"Warning", header.Warning, "Warning"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual := tc.header.String()

			assert.Equal(t, tc.expected, actual, "Expected header %s to have string value %s", tc.name, tc.expected)
		})
	}
}

func TestCustomHeader(t *testing.T) {
	t.Parallel()

	custom := header.Header("X-Custom-Header")

	assert.Equal(t, "X-Custom-Header", custom.String(), "Custom header should return its underlying string representation")
}
