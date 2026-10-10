package http

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgohttprequest "github.com/hueristiq/hq-lib-http-go/request"
)

// roundTripFunc adapts a function to an [http.RoundTripper] for scripted
// transports.
type roundTripFunc func(req *http.Request) (res *http.Response, err error)

// RoundTrip executes the scripted function.
//
// Parameters:
//   - req (*http.Request): the request being round-tripped.
//
// Returns:
//   - res (*http.Response): the scripted response.
//   - err (error): the scripted error.
func (f roundTripFunc) RoundTrip(req *http.Request) (res *http.Response, err error) {
	res, err = f(req)

	return
}

// newOKResponse builds a minimal 200 OK response with the given body.
//
// Parameters:
//   - body (string): the response body.
//
// Returns:
//   - res (*http.Response): the response.
func newOKResponse(body string) (res *http.Response) {
	res = &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}

	return
}

// errHTTP1Broken is the failure net/http produces when a cleartext HTTP/1.x
// exchange receives HTTP/2 frames. It is message-only in the standard
// library, so tests reproduce the message.
var errHTTP1Broken = errors.New(`net/http: HTTP/1.x transport connection broken: malformed HTTP version "HTTP/2"`)

//nolint:paralleltest // The test substitutes the package-level newFallbackHTTPClient seam, which is process-wide state.
func TestClient_Do_HTTP1BrokenTransportFallsBack(t *testing.T) {
	var primaryCalls, fallbackCalls atomic.Int32

	original := newFallbackHTTPClient

	t.Cleanup(func() {
		newFallbackHTTPClient = original
	})

	newFallbackHTTPClient = func() (client *http.Client, err error) {
		client = &http.Client{
			Transport: roundTripFunc(func(_ *http.Request) (res *http.Response, err error) {
				fallbackCalls.Add(1)

				res = newOKResponse("fallback body")

				return
			}),
		}

		return
	}

	primary := roundTripFunc(func(_ *http.Request) (res *http.Response, err error) {
		primaryCalls.Add(1)

		err = errHTTP1Broken

		return
	})

	client, err := NewClient(WithHTTPClient(&http.Client{Transport: primary}))

	require.NoError(t, err)

	req, err := hqgohttprequest.New(t.Context(), "GET", "http://example.com/", nil)

	require.NoError(t, err)

	res, err := client.Do(t.Context(), req)

	require.NoError(t, err)
	require.NotNil(t, res)

	body, err := io.ReadAll(res.Body)

	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	assert.Equal(t, "fallback body", string(body))
	assert.Equal(t, int32(1), primaryCalls.Load(), "Expected the primary client to be attempted once")
	assert.Equal(t, int32(1), fallbackCalls.Load(), "Expected the fallback client to be used once")
}

//nolint:paralleltest // The test substitutes the package-level newFallbackHTTPClient seam, which is process-wide state.
func TestClient_Do_OtherTransportErrorsDoNotFallBack(t *testing.T) {
	var fallbackCalls atomic.Int32

	original := newFallbackHTTPClient

	t.Cleanup(func() {
		newFallbackHTTPClient = original
	})

	newFallbackHTTPClient = func() (client *http.Client, err error) {
		client = &http.Client{
			Transport: roundTripFunc(func(_ *http.Request) (res *http.Response, err error) {
				fallbackCalls.Add(1)

				res = newOKResponse("fallback body")

				return
			}),
		}

		return
	}

	errTransport := errors.New("connection refused")

	primary := roundTripFunc(func(_ *http.Request) (res *http.Response, err error) {
		err = errTransport

		return
	})

	client, err := NewClient(
		WithHTTPClient(&http.Client{Transport: primary}),
		WithMaxRetries(0),
	)

	require.NoError(t, err)

	req, err := hqgohttprequest.New(t.Context(), "GET", "http://example.com/", nil)

	require.NoError(t, err)

	res, err := client.Do(t.Context(), req)

	require.ErrorIs(t, err, errTransport)

	if res != nil {
		require.NoError(t, res.Body.Close())
	}

	assert.Nil(t, res)
	assert.Zero(t, fallbackCalls.Load(), "Expected no fallback for an ordinary transport error")
}

func TestRetryAfterWait(t *testing.T) {
	t.Parallel()

	futureDate := time.Now().Add(90 * time.Second).UTC().Format(time.RFC1123)

	tests := []struct {
		name      string
		header    string
		wantWait  time.Duration
		wantOK    bool
		wantRange bool
	}{
		{"absent header", "", 0, false, false},
		{"seconds", "120", 120 * time.Second, true, false},
		{"zero seconds", "0", 0, true, false},
		{"negative seconds", "-5", 0, false, false},
		{"HTTP date in the future", futureDate, 90 * time.Second, true, true},
		{"HTTP date in the past", "Fri, 31 Dec 1999 23:59:59 GMT", 0, true, false},
		{"unparseable", "soon", 0, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res := &http.Response{Header: make(http.Header)}

			if tt.header != "" {
				res.Header.Set("Retry-After", tt.header)
			}

			wait, ok := retryAfterWait(res)

			assert.Equal(t, tt.wantOK, ok)

			if tt.wantRange {
				// The wall clock advanced between formatting the header and
				// parsing it, so the wait lands slightly below the nominal
				// value.
				assert.Greater(t, wait, 80*time.Second)
				assert.LessOrEqual(t, wait, tt.wantWait)
			} else {
				assert.Equal(t, tt.wantWait, wait)
			}
		})
	}
}

func TestClient_CloseIdleConnectionsIfDue(t *testing.T) {
	t.Parallel()

	t.Run("disabled by default", func(t *testing.T) {
		t.Parallel()

		var closes atomic.Int32

		transport := &closeCountingTransport{closes: &closes}

		client, err := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

		require.NoError(t, err)

		for range 10 {
			client.closeIdleConnectionsIfDue()
		}

		assert.Zero(t, closes.Load())
	})

	t.Run("closes at the threshold", func(t *testing.T) {
		t.Parallel()

		var closes atomic.Int32

		transport := &closeCountingTransport{closes: &closes}

		client, err := NewClient(
			WithHTTPClient(&http.Client{Transport: transport}),
			WithMaxRequestsBeforeIdleClose(3),
		)

		require.NoError(t, err)

		for range 7 {
			client.closeIdleConnectionsIfDue()
		}

		assert.Equal(t, int32(2), closes.Load(),
			"Expected a close on the 3rd and 6th completed executions")
	})
}

// closeCountingTransport is an [http.RoundTripper] counting
// CloseIdleConnections calls.
//
// Fields:
//   - closes (*atomic.Int32): the counter incremented per close.
type closeCountingTransport struct {
	closes *atomic.Int32
}

// Compile-time guard ensuring [closeCountingTransport] satisfies
// [http.RoundTripper].
var _ http.RoundTripper = (*closeCountingTransport)(nil)

// RoundTrip answers every request with a 200 OK.
//
// Parameters:
//   - _ (*http.Request): the request being round-tripped.
//
// Returns:
//   - res (*http.Response): a 200 OK response.
//   - err (error): always nil.
func (t *closeCountingTransport) RoundTrip(_ *http.Request) (res *http.Response, err error) {
	res = newOKResponse("OK")

	return
}

// CloseIdleConnections increments the close counter.
func (t *closeCountingTransport) CloseIdleConnections() {
	t.closes.Add(1)
}
