package status_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.source.hueristiq.com/http/status"
)

func TestStatusInt(t *testing.T) {
	t.Parallel()

	s := status.OK

	assert.Equal(t, 200, s.Int(), "OK should have code 200")
}

func TestStatusString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		status   status.Status
		expected string
	}{
		{status.OK, "OK"},
		{status.NotFound, "Not Found"},
		{status.Teapot, "I'm a teapot"},
		{status.Status(999), "Unknown Status (999)"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, tc.status.String())
		})
	}
}

func TestStatusCategories(t *testing.T) {
	t.Parallel()

	// Informational (1xx)
	assert.True(t, status.Continue.IsInformational(), "Continue (100) should be informational")
	assert.False(t, status.OK.IsInformational(), "OK (200) should not be informational")

	// Success (2xx)
	assert.True(t, status.OK.IsSuccess(), "OK (200) should indicate success")
	assert.False(t, status.BadRequest.IsSuccess(), "BadRequest (400) should not indicate success")

	// Redirection (3xx)
	assert.True(t, status.TemporaryRedirect.IsRedirection(), "TemporaryRedirect (307) should indicate redirection")
	assert.False(t, status.OK.IsRedirection(), "OK (200) should not indicate redirection")

	// Client Error (4xx)
	assert.True(t, status.BadRequest.IsClientError(), "BadRequest (400) should indicate a client error")
	assert.False(t, status.OK.IsClientError(), "OK (200) should not indicate a client error")

	// Server Error (5xx)
	assert.True(t, status.InternalServerError.IsServerError(), "InternalServerError (500) should indicate a server error")
	assert.False(t, status.OK.IsServerError(), "OK (200) should not indicate a server error")

	// General Error (either client or server error)
	assert.True(t, status.BadRequest.IsError(), "BadRequest (400) should be considered an error")
	assert.True(t, status.InternalServerError.IsError(), "InternalServerError (500) should be considered an error")
	assert.False(t, status.OK.IsError(), "OK (200) should not be considered an error")
}
