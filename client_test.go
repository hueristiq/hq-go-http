package http_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	hqgohttp "go.source.hueristiq.com/http"
	"go.source.hueristiq.com/http/status"
)

var ErrInternalHTTP2ClientNotFound = errors.New("internalHTTP2Client not found")

var ErrInternalHTTP2ClientType = errors.New("internalHTTP2Client is not *http.Client")

var ErrMalformedHTTPVersion = errors.New(`net/http: HTTP/1.x transport connection broken: malformed HTTP version "HTTP/2"`)

var ErrTemporary = errors.New("temporary error")

type RoundTripFunc func(req *http.Request) (*http.Response, error)

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type fakeTransport struct {
	rt                        RoundTripFunc
	closeIdleConnectionsCount int32
}

func (ft *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return ft.rt(req)
}

func (ft *fakeTransport) CloseIdleConnections() {
	atomic.AddInt32(&ft.closeIdleConnectionsCount, 1)
}

func setInternalHTTP2Transport(c *hqgohttp.Client, rt http.RoundTripper) error {
	// Get the pointer to the client’s underlying value.
	v := reflect.ValueOf(c).Elem()

	field := v.FieldByName("internalHTTP2Client")
	if !field.IsValid() {
		return ErrInternalHTTP2ClientNotFound
	}

	// Make the unexported field addressable using unsafe.
	field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()

	httpClient, ok := field.Interface().(*http.Client)
	if !ok {
		return ErrInternalHTTP2ClientType
	}

	httpClient.Transport = rt

	return nil
}

func TestDoSuccess(t *testing.T) {
	t.Parallel()

	responseBody := "Hello, World!"
	fakeResp := &http.Response{
		StatusCode: status.OK.Int(),
		Body:       io.NopCloser(strings.NewReader(responseBody)),
	}
	fakeRT := RoundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return fakeResp, nil
	})
	cfg := &hqgohttp.ClientConfiguration{
		Timeout:       5 * time.Second,
		RetryPolicy:   func(_ context.Context, _ error) (bool, error) { return false, nil },
		RetryMax:      1,
		RetryWaitMin:  0,
		RetryWaitMax:  0,
		KillIdleConn:  false,
		RespReadLimit: 4096,
		HTTPClient:    &http.Client{Transport: fakeRT},
	}

	client, err := hqgohttp.NewClient(cfg)

	require.NoError(t, err)

	client.WithBaseURL("http://example.com")

	var onRequestCalled, onResponseCalled bool

	client.WithOnRequest(func(_ *http.Request) {
		onRequestCalled = true
	})
	client.WithOnResponse(func(_ *http.Response) {
		onResponseCalled = true
	})

	reqBuilder := client.Request().Method("GET").URL("/test")

	req, err := reqBuilder.Build()

	require.NoError(t, err)

	res, err := client.Do(req)

	require.NoError(t, err)

	assert.Equal(t, status.OK.Int(), res.StatusCode)

	_ = res.Body.Close()

	assert.True(t, onRequestCalled, "Expected onRequest hook to be called")
	assert.True(t, onResponseCalled, "Expected onResponse hook to be called")
}

func TestDoFallbackToHTTP2(t *testing.T) {
	t.Parallel()

	fallbackResp := &http.Response{
		StatusCode: status.OK.Int(),
		Body:       io.NopCloser(strings.NewReader("Fallback Success")),
	}
	// fakeRT1 simulates a broken HTTP/1.x transport.
	fakeRT1 := RoundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return nil, ErrMalformedHTTPVersion
	})
	// fakeRT2 returns a successful fallback response.
	fakeRT2 := RoundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return fallbackResp, nil
	})
	cfg := &hqgohttp.ClientConfiguration{
		Timeout:       5 * time.Second,
		RetryPolicy:   func(_ context.Context, _ error) (bool, error) { return false, nil },
		RetryMax:      1,
		RetryWaitMin:  0,
		RetryWaitMax:  0,
		KillIdleConn:  false,
		RespReadLimit: 4096,
		HTTPClient:    &http.Client{Transport: fakeRT1},
	}

	client, err := hqgohttp.NewClient(cfg)

	require.NoError(t, err)

	client.WithBaseURL("http://example.com")

	// Override the HTTP/2 client's transport via reflection.
	err = setInternalHTTP2Transport(client, fakeRT2)

	require.NoError(t, err)

	reqBuilder := client.Request().Method("GET").URL("/fallback")

	req, err := reqBuilder.Build()

	require.NoError(t, err)

	res, err := client.Do(req)

	require.NoError(t, err)

	assert.Equal(t, status.OK.Int(), res.StatusCode)

	bodyBytes, err := io.ReadAll(res.Body)

	_ = res.Body.Close()

	require.NoError(t, err)

	assert.Equal(t, "Fallback Success", string(bodyBytes))
}

func TestDoExhaustedRetriesOnError(t *testing.T) {
	t.Parallel()

	fakeRT := RoundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return nil, ErrTemporary
	})

	var onErrorCalled bool

	onErrorHook := func(_ *http.Response, err error, _ int) (*http.Response, error) {
		onErrorCalled = true

		// Modify the error message.
		return nil, fmt.Errorf("onError hook: %w", err)
	}

	cfg := &hqgohttp.ClientConfiguration{
		Timeout: 5 * time.Second,
		// Always signal a retry.
		RetryPolicy:   func(_ context.Context, _ error) (bool, error) { return true, nil },
		RetryMax:      2, // total attempts = RetryMax + 1 (i.e. 2 attempts)
		RetryWaitMin:  1 * time.Second,
		RetryWaitMax:  5 * time.Second,
		KillIdleConn:  false,
		RespReadLimit: 4096,
		HTTPClient:    &http.Client{Transport: fakeRT},
	}

	client, err := hqgohttp.NewClient(cfg)

	require.NoError(t, err)

	client.WithBaseURL("http://example.com")
	client.WithOnError(onErrorHook)

	reqBuilder := client.Request().Method("GET").URL("/error")

	req, err := reqBuilder.Build()

	require.NoError(t, err)

	res, err := client.Do(req)
	if res != nil {
		_ = res.Body.Close()
	}

	require.Error(t, err)

	assert.Nil(t, res)
	assert.True(t, onErrorCalled, "Expected onError hook to be called")
	assert.Contains(t, err.Error(), "onError hook")
}

func TestCloseIdleConnections(t *testing.T) {
	t.Parallel()

	fakeRT := RoundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: status.OK.Int(),
			Body:       io.NopCloser(strings.NewReader("OK")),
		}, nil
	})
	ft := &fakeTransport{
		rt: fakeRT,
	}
	cfg := &hqgohttp.ClientConfiguration{
		Timeout:       5 * time.Second,
		RetryPolicy:   func(_ context.Context, _ error) (bool, error) { return false, nil },
		RetryMax:      1,
		RetryWaitMin:  0,
		RetryWaitMax:  0,
		KillIdleConn:  true,
		RespReadLimit: 4096,
		HTTPClient:    &http.Client{Transport: ft},
	}

	client, err := hqgohttp.NewClient(cfg)

	require.NoError(t, err)

	client.WithBaseURL("http://example.com")

	// Call Do 101 times. The internal counter should trigger CloseIdleConnections once.
	for range 101 {
		reqBuilder := client.Request().Method("GET").URL("/ping")

		req, err := reqBuilder.Build()

		require.NoError(t, err)

		res, err := client.Do(req)

		require.NoError(t, err)

		_ = res.Body.Close()
	}

	assert.Equal(t, int32(1), atomic.LoadInt32(&ft.closeIdleConnectionsCount))
}

func TestRequestBuilder_Build(t *testing.T) {
	t.Parallel()

	cfg := &hqgohttp.ClientConfiguration{
		Timeout:       5 * time.Second,
		RetryPolicy:   func(_ context.Context, _ error) (bool, error) { return false, nil },
		RetryMax:      1,
		RetryWaitMin:  0,
		RetryWaitMax:  0,
		KillIdleConn:  false,
		RespReadLimit: 4096,
	}

	client, err := hqgohttp.NewClient(cfg)

	require.NoError(t, err)

	client.WithBaseURL("http://example.com")
	client.WithHeaders(map[string]string{
		"X-Test": "value",
	})

	builder := client.Request().Method("POST").URL("/api")
	builder.AddHeader("Content-Type", "application/json")

	req, err := builder.Build()

	require.NoError(t, err)

	assert.Equal(t, "http://example.com/api", req.URL.String())
	assert.Equal(t, "value", req.Request.Header.Get("X-Test"))
	assert.Equal(t, "application/json", req.Request.Header.Get("Content-Type"))
}

func TestRequestBuilder_Send(t *testing.T) {
	t.Parallel()

	responseBody := "Send success"
	fakeRT := RoundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: status.OK.Int(),
			Body:       io.NopCloser(strings.NewReader(responseBody)),
		}, nil
	})
	cfg := &hqgohttp.ClientConfiguration{
		Timeout:       5 * time.Second,
		RetryPolicy:   func(_ context.Context, _ error) (bool, error) { return false, nil },
		RetryMax:      1,
		RetryWaitMin:  0,
		RetryWaitMax:  0,
		KillIdleConn:  false,
		RespReadLimit: 4096,
		HTTPClient:    &http.Client{Transport: fakeRT},
	}

	client, err := hqgohttp.NewClient(cfg)

	require.NoError(t, err)

	client.WithBaseURL("http://example.com")

	builder := client.Request().Method("GET").URL("/send")

	res, err := builder.Send()

	require.NoError(t, err)

	body, err := io.ReadAll(res.Body)

	_ = res.Body.Close()

	require.NoError(t, err)

	assert.Equal(t, responseBody, string(body))
}
