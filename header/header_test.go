package header_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	hqgohttpheader "github.com/hueristiq/hq-lib-http-go/header"
)

func TestHeader_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		header   hqgohttpheader.Header
		expected string
	}{
		{"Accept", hqgohttpheader.Accept, "Accept"},
		{"Authorization", hqgohttpheader.Authorization, "Authorization"},
		{"ContentType", hqgohttpheader.ContentType, "Content-Type"},
		{"ETag", hqgohttpheader.ETag, "ETag"},
		{"RetryAfter", hqgohttpheader.RetryAfter, "Retry-After"},
		{"StrictTransportSecurity", hqgohttpheader.StrictTransportSecurity, "Strict-Transport-Security"},
		{"UserAgent", hqgohttpheader.UserAgent, "User-Agent"},
		{"WWWAuthenticate", hqgohttpheader.WWWAuthenticate, "WWW-Authenticate"},
		{"XContentTypeOptions", hqgohttpheader.XContentTypeOptions, "X-Content-Type-Options"},
		{"XForwardedFor", hqgohttpheader.XForwardedFor, "X-Forwarded-For"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.header.String())
		})
	}
}

func TestHeader_ConstantsAreDistinct(t *testing.T) {
	t.Parallel()

	headers := []hqgohttpheader.Header{
		hqgohttpheader.Accept,
		hqgohttpheader.AcceptEncoding,
		hqgohttpheader.Authorization,
		hqgohttpheader.CacheControl,
		hqgohttpheader.ContentType,
		hqgohttpheader.Cookie,
		hqgohttpheader.Date,
		hqgohttpheader.ETag,
		hqgohttpheader.Host,
		hqgohttpheader.IfModifiedSince,
		hqgohttpheader.LastModified,
		hqgohttpheader.Link,
		hqgohttpheader.Location,
		hqgohttpheader.Referer,
		hqgohttpheader.RetryAfter,
		hqgohttpheader.Server,
		hqgohttpheader.SetCookie,
		hqgohttpheader.TransferEncoding,
		hqgohttpheader.UserAgent,
		hqgohttpheader.Vary,
		hqgohttpheader.WWWAuthenticate,
		hqgohttpheader.XForwardedFor,
	}

	seen := make(map[string]int, len(headers))

	for _, header := range headers {
		assert.NotEmpty(t, header.String())

		seen[header.String()]++
	}

	for value, count := range seen {
		assert.Equal(t, 1, count, "Expected %q to be defined exactly once", value)
	}
}
