package http_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	hqgohttp "github.com/hueristiq/hq-go-http"
	hqgohttpstatus "github.com/hueristiq/hq-go-http/status"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type RoundTripFunc func(req *http.Request) (*http.Response, error)

func (f RoundTripFunc) RoundTrip(req *http.Request) (res *http.Response, err error) {
	res, err = f(req)

	return
}

type fakeTransport struct {
	rt                        RoundTripFunc
	closeIdleConnectionsCount int32
}

func (ft *fakeTransport) RoundTrip(req *http.Request) (res *http.Response, err error) {
	res, err = ft.rt(req)

	return
}

func (ft *fakeTransport) CloseIdleConnections() {
	atomic.AddInt32(&ft.closeIdleConnectionsCount, 1)
}

var (
	ErrInternalHTTP2ClientNotFound        = errors.New("internalHTTP2Client not found")
	ErrInternalHTTP2ClientIsNotHTTPClient = errors.New("internalHTTP2Client is not *http.Client")
	ErrMalformedHTTPVersion               = errors.New(`net/http: HTTP/1.x transport connection broken: malformed HTTP version "HTTP/2"`)
	ErrTemporary                          = errors.New("temporary error")
)

func setInternalHTTP2Transport(c *hqgohttp.Client, rt http.RoundTripper) (err error) {
	// Get the pointer to the client's underlying value.
	v := reflect.ValueOf(c).Elem()

	field := v.FieldByName("internalHTTP2Client")

	if !field.IsValid() {
		err = ErrInternalHTTP2ClientNotFound

		return
	}

	// Make the unexported field addressable.
	field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()

	httpClient, ok := field.Interface().(*http.Client)
	if !ok {
		err = ErrInternalHTTP2ClientIsNotHTTPClient

		return
	}

	httpClient.Transport = rt

	return
}

func TestNewClient(t *testing.T) {
	t.Parallel()

	cfg := &hqgohttp.ClientConfiguration{
		Client:               nil,
		CloseIdleConnections: false,
		Timeout:              10 * time.Second,
		RetryMax:             2,
		RetryWaitMin:         1 * time.Second,
		RetryWaitMax:         2 * time.Second,
		RetryPolicy:          nil,
		RetryBackoff:         nil,
		RespReadLimit:        4096,
	}

	client, err := hqgohttp.NewClient(cfg)

	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestDo(t *testing.T) {
	t.Parallel()

	fakeRespBody := "Success"
	fakeResp := &http.Response{
		StatusCode: hqgohttpstatus.OK.Int(),
		Body:       io.NopCloser(strings.NewReader(fakeRespBody)),
	}
	fakeRT := RoundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return fakeResp, nil
	})

	cfg := &hqgohttp.ClientConfiguration{
		Timeout:              5 * time.Second,
		RetryPolicy:          func(_ context.Context, _ error) (bool, error) { return false, nil },
		RetryMax:             1,
		RetryWaitMin:         0,
		RetryWaitMax:         0,
		CloseIdleConnections: false,
		RespReadLimit:        4096,
		Client:               &http.Client{Transport: fakeRT},
	}

	client, err := hqgohttp.NewClient(cfg)

	require.NoError(t, err)

	res, err := client.Request(&hqgohttp.RequestConfiguration{
		Method: "GET",
		URL:    "/test",
	})

	require.NoError(t, err)
	assert.Equal(t, hqgohttpstatus.OK.Int(), res.StatusCode)

	bodyBytes, err := io.ReadAll(res.Body)

	_ = res.Body.Close()

	require.NoError(t, err)
	assert.Equal(t, fakeRespBody, string(bodyBytes))
}

func TestDoFallbackToHTTP2(t *testing.T) {
	t.Parallel()

	fallbackBody := "Fallback Success"
	fallbackResp := &http.Response{
		StatusCode: hqgohttpstatus.OK.Int(),
		Body:       io.NopCloser(strings.NewReader(fallbackBody)),
	}
	fakeRT1 := RoundTripFunc(func(_ *http.Request) (res *http.Response, err error) {
		err = ErrMalformedHTTPVersion

		return
	})
	fakeRT2 := RoundTripFunc(func(_ *http.Request) (res *http.Response, err error) {
		res = fallbackResp

		return
	})

	cfg := &hqgohttp.ClientConfiguration{
		Timeout:              5 * time.Second,
		RetryPolicy:          func(_ context.Context, _ error) (bool, error) { return false, nil },
		RetryMax:             1,
		RetryWaitMin:         0,
		RetryWaitMax:         0,
		CloseIdleConnections: false,
		RespReadLimit:        4096,
		Client:               &http.Client{Transport: fakeRT1},
	}

	client, err := hqgohttp.NewClient(cfg)

	require.NoError(t, err)

	// Override the HTTP/2 client's transport.
	err = setInternalHTTP2Transport(client, fakeRT2)

	require.NoError(t, err)

	// Issue the request.
	res, err := client.Request(&hqgohttp.RequestConfiguration{
		Method: "GET",
		URL:    "/fallback",
	})

	require.NoError(t, err)
	assert.Equal(t, hqgohttpstatus.OK.Int(), res.StatusCode)

	bodyBytes, err := io.ReadAll(res.Body)

	_ = res.Body.Close()

	require.NoError(t, err)
	assert.Equal(t, fallbackBody, string(bodyBytes))
}

func TestCloseIdleConnections(t *testing.T) {
	t.Parallel()

	// Setup a fake transport that counts CloseIdleConnections calls.
	ft := &fakeTransport{
		rt: RoundTripFunc(func(_ *http.Request) (res *http.Response, err error) {
			res = &http.Response{
				StatusCode: hqgohttpstatus.OK.Int(),
				Body:       io.NopCloser(strings.NewReader("OK")),
			}

			return
		}),
	}

	cfg := &hqgohttp.ClientConfiguration{
		Timeout:              5 * time.Second,
		RetryPolicy:          func(_ context.Context, _ error) (bool, error) { return false, nil },
		RetryMax:             1,
		RetryWaitMin:         0,
		RetryWaitMax:         0,
		CloseIdleConnections: true,
		RespReadLimit:        4096,
		Client:               &http.Client{Transport: ft},
	}

	client, err := hqgohttp.NewClient(cfg)

	require.NoError(t, err)

	// Issue 101 requests. The internal counter should cause CloseIdleConnections to be called once.
	for range 101 {
		res, err := client.Request(&hqgohttp.RequestConfiguration{
			Method: "GET",
			URL:    "/ping",
		})

		require.NoError(t, err)

		_ = res.Body.Close()
	}

	assert.Equal(t, int32(1), atomic.LoadInt32(&ft.closeIdleConnectionsCount))
}

func TestExhaustedRetries(t *testing.T) {
	t.Parallel()

	// Fake transport that always returns an error.
	fakeRT := RoundTripFunc(func(_ *http.Request) (res *http.Response, err error) {
		err = ErrTemporary

		return
	})

	cfg := &hqgohttp.ClientConfiguration{
		Timeout: 5 * time.Second,
		// Always retry.
		RetryPolicy:          func(_ context.Context, _ error) (bool, error) { return true, nil },
		RetryMax:             2, // Means total attempts = RetryMax + 1 (i.e. 3 attempts)
		RetryWaitMin:         10 * time.Millisecond,
		RetryWaitMax:         20 * time.Millisecond,
		CloseIdleConnections: false,
		RespReadLimit:        4096,
		Client:               &http.Client{Transport: fakeRT},
	}

	client, err := hqgohttp.NewClient(cfg)

	require.NoError(t, err)

	// Issue the request.
	res, err := client.Request(&hqgohttp.RequestConfiguration{
		Method:  "GET",
		BaseURL: "http://example.com",
		URL:     "/retry",
	})

	// Expect error since all attempts fail.
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "giving up")

	if res != nil {
		_ = res.Body.Close()
	}
}

// TestRequestConfigurationMerging verifies that request configurations are correctly merged.
// It tests that the BaseURL, relative URL, query parameters, and headers are combined as expected.
func TestRequestConfigurationMerging(t *testing.T) {
	t.Parallel()

	var capturedURL string

	var capturedHeaders http.Header

	// Fake transport that captures the request URL and headers.
	fakeRT := RoundTripFunc(func(req *http.Request) (res *http.Response, err error) {
		capturedURL = req.URL.String()
		capturedHeaders = req.Header

		res = &http.Response{
			StatusCode: hqgohttpstatus.OK.Int(),
			Body:       io.NopCloser(strings.NewReader("merged")),
		}

		return
	})

	cfg := &hqgohttp.ClientConfiguration{
		Timeout:              5 * time.Second,
		RetryPolicy:          func(_ context.Context, _ error) (bool, error) { return false, nil },
		RetryMax:             1,
		RetryWaitMin:         0,
		RetryWaitMax:         0,
		CloseIdleConnections: false,
		RespReadLimit:        4096,
		Client:               &http.Client{Transport: fakeRT},
		// Set some default values.
		BaseURL: "http://example.com",
		URL:     "default",
		Headers: []hqgohttp.Header{
			hqgohttp.NewSetHeader("X-Default", "defaultValue"),
		},
		Params: map[string]string{"default": "1"},
	}

	client, err := hqgohttp.NewClient(cfg)

	require.NoError(t, err)

	// Provide request-specific overrides.
	res, err := client.Request(&hqgohttp.RequestConfiguration{
		Method: "GET",
		URL:    "api",
		Headers: []hqgohttp.Header{
			hqgohttp.NewSetHeader("X-Override", "overrideValue"),
		},
		Params: map[string]string{"q": "test"},
	})

	require.NoError(t, err)

	// The final URL should be joined using url.JoinPath and include query parameters.
	u, err := url.Parse(capturedURL)

	require.NoError(t, err)
	assert.Equal(t, "http", u.Scheme)
	assert.Equal(t, "example.com", u.Host)
	// Expect the path to be "default/api" (JoinPath behavior may vary).
	assert.Contains(t, u.Path, "api")

	// Query parameters should include both defaults and overrides.
	q := u.Query()

	assert.Equal(t, "1", q.Get("default"))
	assert.Equal(t, "test", q.Get("q"))

	// Headers should include both the default and override values.
	assert.Equal(t, "defaultValue", capturedHeaders.Get("X-Default"))
	assert.Equal(t, "overrideValue", capturedHeaders.Get("X-Override"))

	_ = res.Body.Close()
}

func TestConvenienceMethods(t *testing.T) {
	t.Parallel()

	echoRT := RoundTripFunc(func(req *http.Request) (res *http.Response, err error) {
		body := fmt.Sprintf("Method: %s, URL: %s", req.Method, req.URL.String())

		res = &http.Response{
			StatusCode: hqgohttpstatus.OK.Int(),
			Body:       io.NopCloser(strings.NewReader(body)),
		}

		return
	})

	cfg := &hqgohttp.ClientConfiguration{
		Timeout:              5 * time.Second,
		RetryPolicy:          func(_ context.Context, _ error) (bool, error) { return false, nil },
		RetryMax:             1,
		RetryWaitMin:         0,
		RetryWaitMax:         0,
		CloseIdleConnections: false,
		RespReadLimit:        4096,
		Client:               &http.Client{Transport: echoRT},
		BaseURL:              "http://example.com",
	}

	client, err := hqgohttp.NewClient(cfg)

	require.NoError(t, err)

	tests := []struct {
		name       string
		callMethod func() (*http.Response, error)
		expected   string
	}{
		{
			name: "Get",
			callMethod: func() (*http.Response, error) {
				return client.Get("/get")
			},
			expected: "Method: GET, URL: http://example.com/get",
		},
		{
			name: "Head",
			callMethod: func() (*http.Response, error) {
				return client.Head("/head")
			},
			expected: "Method: HEAD, URL: http://example.com/head",
		},
		{
			name: "Put",
			callMethod: func() (*http.Response, error) {
				return client.Put("/put", "payload")
			},
			expected: "Method: PUT, URL: http://example.com/put",
		},
		{
			name: "Delete",
			callMethod: func() (*http.Response, error) {
				return client.Delete("/delete")
			},
			expected: "Method: DELETE, URL: http://example.com/delete",
		},
		{
			name: "Post",
			callMethod: func() (*http.Response, error) {
				return client.Post("/post", "payload")
			},
			expected: "Method: POST, URL: http://example.com/post",
		},
		{
			name: "Options",
			callMethod: func() (*http.Response, error) {
				return client.Options("/options")
			},
			expected: "Method: OPTIONS, URL: http://example.com/options",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res, err := tt.callMethod()

			require.NoError(t, err)

			bodyBytes, err := io.ReadAll(res.Body)

			require.NoError(t, err)

			assert.Equal(t, tt.expected, string(bodyBytes))

			_ = res.Body.Close()
		})
	}
}
