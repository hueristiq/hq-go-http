package useragent_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgohttpuseragent "github.com/hueristiq/hq-lib-http-go/useragent"
)

func TestList(t *testing.T) {
	t.Parallel()

	userAgents := hqgohttpuseragent.List()

	require.NotEmpty(t, userAgents, "Expected the curated list to be non-empty")

	for _, userAgent := range userAgents {
		assert.True(t, strings.HasPrefix(userAgent, "Mozilla/5.0"),
			"Expected %q to look like a browser User-Agent", userAgent)
	}
}

func TestList_ReturnsCopy(t *testing.T) {
	t.Parallel()

	userAgents := hqgohttpuseragent.List()

	require.NotEmpty(t, userAgents)

	original := userAgents[0]
	userAgents[0] = "mutated"

	assert.Equal(t, original, hqgohttpuseragent.List()[0],
		"Expected mutating the returned slice to leave the package list untouched")
}

func TestRandom(t *testing.T) {
	t.Parallel()

	allowed := make(map[string]struct{})

	for _, userAgent := range hqgohttpuseragent.List() {
		allowed[userAgent] = struct{}{}
	}

	for range 100 {
		userAgent := hqgohttpuseragent.Random()

		assert.Contains(t, allowed, userAgent,
			"Expected the random User-Agent to come from the curated list")
	}
}
