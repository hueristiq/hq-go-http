package http_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgohttp "github.com/hueristiq/hq-go-http"
	hqgohttpheader "github.com/hueristiq/hq-go-http/header"
	hqgohttpmethod "github.com/hueristiq/hq-go-http/method"
	hqgohttpmime "github.com/hueristiq/hq-go-http/mime"
	hqgohttpstatus "github.com/hueristiq/hq-go-http/status"
)

// requestMirror captures what a request assembled by the builder looks like
// on the wire.
//
// Fields:
//   - method (string): the HTTP method.
//   - url (string): the full request URL.
//   - headers (http.Header): the request headers.
//   - body (string): the request body.
type requestMirror struct {
	method  string
	url     string
	headers http.Header
	body    string
}

// newMirrorServer starts an httptest server echoing the request's method and
// URL in the response body, recording the full request into mirror.
func newMirrorServer(t *testing.T, mirror *requestMirror) (serverURL string) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			body = nil
		}

		mirror.method = r.Method
		mirror.url = r.URL.String()
		mirror.headers = r.Header.Clone()
		mirror.body = string(body)

		//nolint:gosec // Echoing the method and URL in a test server is not an XSS vector.
		_, _ = w.Write([]byte(r.Method + " " + r.URL.String()))
	}))

	t.Cleanup(server.Close)

	serverURL = server.URL

	return
}

func TestRequestBuilder_EndToEnd(t *testing.T) {
	t.Parallel()

	var mirror requestMirror

	serverURL := newMirrorServer(t, &mirror)

	client, err := hqgohttp.NewClient()

	require.NoError(t, err)

	res, err := client.NewRequest().
		Method(hqgohttpmethod.POST).
		BaseURL(serverURL).
		Path("/users").
		Param("page", "1").
		Param("filter", "active").
		SetHeader(hqgohttpheader.ContentType.String(), hqgohttpmime.JSON.String()).
		AddHeader(hqgohttpheader.Accept.String(), hqgohttpmime.JSON.String()).
		Body(`{"name":"ada"}`).
		Do(t.Context())

	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, hqgohttpstatus.OK.Int(), res.StatusCode)
	assert.Equal(t, "POST /users?filter=active&page=1", readBody(t, res))

	assert.Equal(t, "POST", mirror.method)
	assert.Equal(t, hqgohttpmime.JSON.String(), mirror.headers.Get(hqgohttpheader.ContentType.String()))
	assert.JSONEq(t, `{"name":"ada"}`, mirror.body)

	parsed, err := url.Parse(serverURL + mirror.url)

	require.NoError(t, err)

	query := parsed.Query()

	assert.Equal(t, "1", query.Get("page"))
	assert.Equal(t, "active", query.Get("filter"))
}

func TestRequestBuilder_URLResolution(t *testing.T) {
	t.Parallel()

	t.Run("client base URL with builder path", func(t *testing.T) {
		t.Parallel()

		var mirror requestMirror

		serverURL := newMirrorServer(t, &mirror)

		client, err := hqgohttp.NewClient(hqgohttp.WithBaseURL(serverURL))

		require.NoError(t, err)

		res, err := client.NewRequest().
			Path("/api/v1/items").
			Do(t.Context())

		require.NoError(t, err)

		assert.Equal(t, "GET /api/v1/items", readBody(t, res))
	})

	t.Run("builder base URL overrides client base URL", func(t *testing.T) {
		t.Parallel()

		var mirror requestMirror

		serverURL := newMirrorServer(t, &mirror)

		client, err := hqgohttp.NewClient(hqgohttp.WithBaseURL("http://unused.example.com"))

		require.NoError(t, err)

		res, err := client.NewRequest().
			BaseURL(serverURL).
			Path("/here").
			Do(t.Context())

		require.NoError(t, err)

		assert.Equal(t, "GET /here", readBody(t, res))
	})

	t.Run("absolute URL is never prefixed", func(t *testing.T) {
		t.Parallel()

		var mirror requestMirror

		serverURL := newMirrorServer(t, &mirror)

		client, err := hqgohttp.NewClient(hqgohttp.WithBaseURL("http://unused.example.com"))

		require.NoError(t, err)

		res, err := client.NewRequest().
			URL(serverURL + "/absolute").
			Do(t.Context())

		require.NoError(t, err)

		assert.Equal(t, "GET /absolute", readBody(t, res))
	})

	t.Run("scheme defaulting", func(t *testing.T) {
		t.Parallel()

		var mirror requestMirror

		serverURL := newMirrorServer(t, &mirror)

		client, err := hqgohttp.NewClient()

		require.NoError(t, err)

		host := serverURL[len("http://"):]

		res, err := client.NewRequest().
			URL(host + "/schemeless").
			Do(t.Context())

		require.NoError(t, err)

		assert.Equal(t, "GET /schemeless", readBody(t, res))
	})
}

func TestRequestBuilder_Headers(t *testing.T) {
	t.Parallel()

	t.Run("set replaces, add appends", func(t *testing.T) {
		t.Parallel()

		var mirror requestMirror

		serverURL := newMirrorServer(t, &mirror)

		client, err := hqgohttp.NewClient()

		require.NoError(t, err)

		res, err := client.NewRequest().
			URL(serverURL).
			SetHeader("X-Set", "first").
			SetHeader("X-Set", "second").
			AddHeader("X-Add", "one").
			AddHeader("X-Add", "two").
			Do(t.Context())

		require.NoError(t, err)
		require.NoError(t, res.Body.Close())

		assert.Equal(t, []string{"second"}, mirror.headers.Values("X-Set"))
		assert.Equal(t, []string{"one", "two"}, mirror.headers.Values("X-Add"))
	})

	t.Run("request header overrides client default", func(t *testing.T) {
		t.Parallel()

		var mirror requestMirror

		serverURL := newMirrorServer(t, &mirror)

		client, err := hqgohttp.NewClient(hqgohttp.WithHeader("X-Default", "client-value"))

		require.NoError(t, err)

		res, err := client.NewRequest().
			URL(serverURL).
			SetHeader("X-Default", "request-value").
			Do(t.Context())

		require.NoError(t, err)
		require.NoError(t, res.Body.Close())

		assert.Equal(t, []string{"request-value"}, mirror.headers.Values("X-Default"))
	})

	t.Run("client default header is applied", func(t *testing.T) {
		t.Parallel()

		var mirror requestMirror

		serverURL := newMirrorServer(t, &mirror)

		client, err := hqgohttp.NewClient(hqgohttp.WithHeader("X-Default", "client-value"))

		require.NoError(t, err)

		res, err := client.NewRequest().
			URL(serverURL).
			Do(t.Context())

		require.NoError(t, err)
		require.NoError(t, res.Body.Close())

		assert.Equal(t, []string{"client-value"}, mirror.headers.Values("X-Default"))
	})
}

func TestRequestBuilder_DefaultMethod(t *testing.T) {
	t.Parallel()

	var mirror requestMirror

	serverURL := newMirrorServer(t, &mirror)

	client, err := hqgohttp.NewClient()

	require.NoError(t, err)

	res, err := client.NewRequest().
		URL(serverURL).
		Do(t.Context())

	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	assert.Equal(t, hqgohttpmethod.GET.String(), mirror.method)
}

func TestRequestBuilder_Invalid(t *testing.T) {
	t.Parallel()

	client, err := hqgohttp.NewClient()

	require.NoError(t, err)

	t.Run("no URL configured", func(t *testing.T) {
		t.Parallel()

		res, err := client.NewRequest().
			Method(hqgohttpmethod.GET).
			Do(t.Context())

		require.Error(t, err)

		if res != nil {
			require.NoError(t, res.Body.Close())
		}

		assert.Nil(t, res)
		assert.Contains(t, err.Error(), "no URL configured")
	})

	t.Run("relative URL without a base URL", func(t *testing.T) {
		t.Parallel()

		res, err := client.NewRequest().
			Path("/relative").
			Do(t.Context())

		require.Error(t, err)

		if res != nil {
			require.NoError(t, res.Body.Close())
		}

		assert.Nil(t, res)
	})
}

func TestRequestBuilder_WithOptions(t *testing.T) {
	t.Parallel()

	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++

		w.WriteHeader(hqgohttpstatus.ServiceUnavailable.Int())
	}))

	t.Cleanup(server.Close)

	client, err := hqgohttp.NewClient(
		hqgohttp.WithMaxRetries(5),
		hqgohttp.WithRetryBackoff(noWaitBackoff),
	)

	require.NoError(t, err)

	res, err := client.NewRequest().
		URL(server.URL).
		WithOptions(hqgohttp.WithRequestMaxRetries(1)).
		Do(t.Context())

	require.Error(t, err)

	if res != nil {
		require.NoError(t, res.Body.Close())
	}

	assert.Nil(t, res)
	assert.Equal(t, int32(2), attempts,
		"Expected the builder-registered request options to override the client configuration")
}
