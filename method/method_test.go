package method_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.source.hueristiq.com/http/method"
)

func TestMethodString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		m        method.Method
		expected string
	}{
		// Safe Methods
		{"GET", method.GET, "GET"},
		{"HEAD", method.HEAD, "HEAD"},

		// Idempotent Methods
		{"PUT", method.PUT, "PUT"},
		{"DELETE", method.DELETE, "DELETE"},

		// Unsafe Methods
		{"POST", method.POST, "POST"},
		{"PATCH", method.PATCH, "PATCH"},

		// Auxiliary Methods
		{"OPTIONS", method.OPTIONS, "OPTIONS"},
		{"TRACE", method.TRACE, "TRACE"},
		{"CONNECT", method.CONNECT, "CONNECT"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual := tc.m.String()

			assert.Equal(t, tc.expected, actual, "Expected method %s to return %s", tc.name, tc.expected)
		})
	}
}

func TestCustomMethod(t *testing.T) {
	t.Parallel()

	custom := method.Method("CUSTOM")

	assert.Equal(t, "CUSTOM", custom.String(), "Custom method should return its underlying string representation")
}
