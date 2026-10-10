package http_test

import (
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgohttp "github.com/hueristiq/hq-go-http"
)

func TestDefaultHTTPPooledTransport(t *testing.T) {
	t.Parallel()

	transport := hqgohttp.DefaultHTTPPooledTransport()

	require.NotNil(t, transport)
	assert.False(t, transport.DisableKeepAlives)
	assert.True(t, transport.ForceAttemptHTTP2)
	assert.NotNil(t, transport.Proxy)
	assert.NotNil(t, transport.DialContext)
}

func TestDefaultHTTPTransport(t *testing.T) {
	t.Parallel()

	transport := hqgohttp.DefaultHTTPTransport()

	require.NotNil(t, transport)
	assert.True(t, transport.DisableKeepAlives)
	assert.Equal(t, -1, transport.MaxIdleConnsPerHost)
}

func TestDefaultHTTPClients(t *testing.T) {
	t.Parallel()

	t.Run("pooled", func(t *testing.T) {
		t.Parallel()

		client := hqgohttp.DefaultHTTPPooledClient()

		require.NotNil(t, client)

		transport, ok := client.Transport.(*http.Transport)

		require.True(t, ok, "Expected the pooled client to run on an *http.Transport")
		assert.False(t, transport.DisableKeepAlives)
	})

	t.Run("non-pooled", func(t *testing.T) {
		t.Parallel()

		client := hqgohttp.DefaultHTTPClient()

		require.NotNil(t, client)

		transport, ok := client.Transport.(*http.Transport)

		require.True(t, ok, "Expected the client to run on an *http.Transport")
		assert.True(t, transport.DisableKeepAlives)
	})
}

func TestClient_CloseIdleConnections(t *testing.T) {
	t.Parallel()

	var closes atomic.Int32

	transport := &closeCountingTransport{closes: &closes}

	client, err := hqgohttp.NewClient(hqgohttp.WithHTTPClient(&http.Client{Transport: transport}))

	require.NoError(t, err)

	client.CloseIdleConnections()

	assert.Equal(t, int32(1), closes.Load(),
		"Expected CloseIdleConnections to reach the primary client's transport")
}
