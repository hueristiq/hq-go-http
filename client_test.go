package http_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgohttp "github.com/hueristiq/hq-lib-http-go"
	hqgohttpmethod "github.com/hueristiq/hq-lib-http-go/method"
	hqgohttprequest "github.com/hueristiq/hq-lib-http-go/request"
	hqgohttpstatus "github.com/hueristiq/hq-lib-http-go/status"
	hqgohttpuseragent "github.com/hueristiq/hq-lib-http-go/useragent"
	hqgoretrierbackoff "github.com/hueristiq/hq-lib-retrier-go/backoff"
)

// noWaitBackoff removes the wait between attempts, keeping retry tests fast.
// The retry engine clamps the zero delay to its one-millisecond floor.
func noWaitBackoff(_, _ time.Duration) (backoff hqgoretrierbackoff.Backoff) {
	return func(int) (delay time.Duration) {
		return
	}
}

// readBody reads and closes a response body, failing the test on error.
func readBody(t *testing.T, res *http.Response) (body string) {
	t.Helper()

	data, err := io.ReadAll(res.Body)

	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	body = string(data)

	return
}

func TestNewClient(t *testing.T) {
	t.Parallel()

	client, err := hqgohttp.NewClient()

	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestNewClient_Options(t *testing.T) {
	t.Parallel()

	t.Run("nil HTTP client is ignored", func(t *testing.T) {
		t.Parallel()

		client, err := hqgohttp.NewClient(hqgohttp.WithHTTPClient(nil))

		require.NoError(t, err)
		require.NotNil(t, client)
	})

	t.Run("invalid values are normalized", func(t *testing.T) {
		t.Parallel()

		client, err := hqgohttp.NewClient(
			hqgohttp.WithTimeout(-time.Second),
			hqgohttp.WithMaxRetries(-1),
			hqgohttp.WithRetryWaitMin(-time.Second),
			hqgohttp.WithRetryWaitMax(-time.Second),
			hqgohttp.WithRetryBackoff(nil),
			hqgohttp.WithRetryPolicy(nil),
			hqgohttp.WithRateLimiter(nil),
			hqgohttp.WithOnRequest(nil),
			hqgohttp.WithOnResponse(nil),
			hqgohttp.WithResponseReadLimit(-1),
		)

		require.NoError(t, err)
		require.NotNil(t, client)
	})
}

func TestClient_Do(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("do response"))
	}))

	t.Cleanup(server.Close)

	client, err := hqgohttp.NewClient()

	require.NoError(t, err)

	req, err := hqgohttprequest.New(t.Context(), hqgohttpmethod.GET.String(), server.URL, nil)

	require.NoError(t, err)

	res, err := client.Do(t.Context(), req)

	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, hqgohttpstatus.OK.Int(), res.StatusCode)
	assert.Equal(t, "do response", readBody(t, res))
}

func TestClient_Do_NilRequest(t *testing.T) {
	t.Parallel()

	client, err := hqgohttp.NewClient()

	require.NoError(t, err)

	res, err := client.Do(t.Context(), nil)

	require.Error(t, err)

	if res != nil {
		require.NoError(t, res.Body.Close())
	}

	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "http.Client.Do:")
}

func TestClient_ConvenienceMethods(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//nolint:gosec // Echoing the method and path in a test server is not an XSS vector.
		_, _ = fmt.Fprintf(w, "%s %s", r.Method, r.URL.Path)
	}))

	t.Cleanup(server.Close)

	client, err := hqgohttp.NewClient()

	require.NoError(t, err)

	tests := []struct {
		name     string
		call     func(ctx context.Context, rawURL string) (*http.Response, error)
		expected string
	}{
		{
			name: "Get",
			call: func(ctx context.Context, rawURL string) (*http.Response, error) {
				return client.Get(ctx, rawURL)
			},
			expected: "GET /get",
		},
		{
			name: "Head",
			call: func(ctx context.Context, rawURL string) (*http.Response, error) {
				return client.Head(ctx, rawURL)
			},
			expected: "", // HEAD responses carry no body.
		},
		{
			name: "Post",
			call: func(ctx context.Context, rawURL string) (*http.Response, error) {
				return client.Post(ctx, rawURL, "payload")
			},
			expected: "POST /post",
		},
		{
			name: "Put",
			call: func(ctx context.Context, rawURL string) (*http.Response, error) {
				return client.Put(ctx, rawURL, "payload")
			},
			expected: "PUT /put",
		},
		{
			name: "Patch",
			call: func(ctx context.Context, rawURL string) (*http.Response, error) {
				return client.Patch(ctx, rawURL, "payload")
			},
			expected: "PATCH /patch",
		},
		{
			name: "Delete",
			call: func(ctx context.Context, rawURL string) (*http.Response, error) {
				return client.Delete(ctx, rawURL)
			},
			expected: "DELETE /delete",
		},
		{
			name: "Options",
			call: func(ctx context.Context, rawURL string) (*http.Response, error) {
				return client.Options(ctx, rawURL)
			},
			expected: "OPTIONS /options",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := "/" + strings.ToLower(tt.name)

			res, err := tt.call(t.Context(), server.URL+path)

			require.NoError(t, err)
			require.NotNil(t, res)

			assert.Equal(t, tt.expected, readBody(t, res))
		})
	}
}

func TestClient_RetriesServerErrors(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if attempts.Add(1) < 3 {
			w.WriteHeader(hqgohttpstatus.InternalServerError.Int())

			return
		}

		_, _ = w.Write([]byte("finally"))
	}))

	t.Cleanup(server.Close)

	client, err := hqgohttp.NewClient(
		hqgohttp.WithMaxRetries(3),
		hqgohttp.WithRetryBackoff(noWaitBackoff),
	)

	require.NoError(t, err)

	res, err := client.Get(t.Context(), server.URL)

	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, hqgohttpstatus.OK.Int(), res.StatusCode)
	assert.Equal(t, "finally", readBody(t, res))
	assert.Equal(t, int32(3), attempts.Load(), "Expected two failed attempts followed by a success")
}

func TestClient_Retries429HonoringRetryAfter(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if attempts.Add(1) == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(hqgohttpstatus.TooManyRequests.Int())

			return
		}

		_, _ = w.Write([]byte("after the wait"))
	}))

	t.Cleanup(server.Close)

	client, err := hqgohttp.NewClient(
		hqgohttp.WithMaxRetries(2),
		hqgohttp.WithRetryWaitMax(30*time.Second),
		hqgohttp.WithRetryBackoff(noWaitBackoff),
	)

	require.NoError(t, err)

	start := time.Now()

	res, err := client.Get(t.Context(), server.URL)

	elapsed := time.Since(start)

	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, "after the wait", readBody(t, res))
	assert.Equal(t, int32(2), attempts.Load())
	assert.GreaterOrEqual(t, elapsed, time.Second,
		"Expected the retry to honor the one-second Retry-After header")
}

func TestClient_RetryAfterIsCappedAtWaitMax(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if attempts.Add(1) == 1 {
			w.Header().Set("Retry-After", "10")
			w.WriteHeader(hqgohttpstatus.ServiceUnavailable.Int())

			return
		}

		_, _ = w.Write([]byte("capped"))
	}))

	t.Cleanup(server.Close)

	client, err := hqgohttp.NewClient(
		hqgohttp.WithMaxRetries(1),
		hqgohttp.WithRetryWaitMax(50*time.Millisecond),
		hqgohttp.WithRetryBackoff(noWaitBackoff),
	)

	require.NoError(t, err)

	start := time.Now()

	res, err := client.Get(t.Context(), server.URL)

	elapsed := time.Since(start)

	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, "capped", readBody(t, res))
	assert.Equal(t, int32(2), attempts.Load())
	assert.Less(t, elapsed, 5*time.Second,
		"Expected the honored Retry-After wait to be capped at the maximum retry wait")
}

func TestClient_RetryPolicySeesResponse(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if attempts.Add(1) == 1 {
			w.WriteHeader(hqgohttpstatus.InternalServerError.Int())

			return
		}

		_, _ = w.Write([]byte("recovered"))
	}))

	t.Cleanup(server.Close)

	var statuses []int

	policy := func(_ context.Context, resp *http.Response, _ error) (retry bool, rerr error) {
		if resp != nil {
			statuses = append(statuses, resp.StatusCode)
		}

		if resp != nil && resp.StatusCode == hqgohttpstatus.InternalServerError.Int() {
			retry = true
		}

		return
	}

	client, err := hqgohttp.NewClient(
		hqgohttp.WithRetryPolicy(policy),
		hqgohttp.WithRetryBackoff(noWaitBackoff),
	)

	require.NoError(t, err)

	res, err := client.Get(t.Context(), server.URL)

	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, "recovered", readBody(t, res))
	assert.Equal(t, []int{
		hqgohttpstatus.InternalServerError.Int(),
		hqgohttpstatus.OK.Int(),
	}, statuses, "Expected the policy to observe every attempt's response")
}

func TestClient_RetryPolicyAbort(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(hqgohttpstatus.Teapot.Int())

		_, _ = w.Write([]byte("short and stout"))
	}))

	t.Cleanup(server.Close)

	errTeapot := errors.New("teapot detected")

	policy := func(_ context.Context, resp *http.Response, _ error) (retry bool, rerr error) {
		if resp != nil && resp.StatusCode == hqgohttpstatus.Teapot.Int() {
			rerr = errTeapot
		}

		return
	}

	client, err := hqgohttp.NewClient(hqgohttp.WithRetryPolicy(policy))

	require.NoError(t, err)

	res, err := client.Get(t.Context(), server.URL)

	require.ErrorIs(t, err, errTeapot)
	require.NotNil(t, res, "Expected the attempt's response to be handed to the caller on a policy abort")

	assert.Equal(t, "short and stout", readBody(t, res))
}

func TestClient_ExhaustedRetries(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts.Add(1)

		w.WriteHeader(hqgohttpstatus.ServiceUnavailable.Int())
	}))

	t.Cleanup(server.Close)

	client, err := hqgohttp.NewClient(
		hqgohttp.WithMaxRetries(2),
		hqgohttp.WithRetryBackoff(noWaitBackoff),
	)

	require.NoError(t, err)

	res, err := client.Get(t.Context(), server.URL)

	require.Error(t, err)

	if res != nil {
		require.NoError(t, res.Body.Close())
	}

	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "giving up after 3 attempt(s)")
	assert.Contains(t, err.Error(), "503")
	assert.Equal(t, int32(3), attempts.Load(),
		"Expected the initial attempt plus two retries")
}

func TestClient_RequestOptionOverrides(t *testing.T) {
	t.Parallel()

	// newUnavailableServer starts an httptest server that always answers 503
	// and counts its hits.
	newUnavailableServer := func(t *testing.T) (serverURL string, attempts *atomic.Int32) {
		t.Helper()

		attempts = &atomic.Int32{}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			attempts.Add(1)

			w.WriteHeader(hqgohttpstatus.ServiceUnavailable.Int())
		}))

		t.Cleanup(server.Close)

		serverURL = server.URL

		return
	}

	t.Run("max retries override", func(t *testing.T) {
		t.Parallel()

		serverURL, attempts := newUnavailableServer(t)

		client, err := hqgohttp.NewClient(
			hqgohttp.WithMaxRetries(5),
			hqgohttp.WithRetryBackoff(noWaitBackoff),
		)

		require.NoError(t, err)

		res, err := client.Get(t.Context(), serverURL, hqgohttp.WithRequestMaxRetries(1))

		require.Error(t, err)

		if res != nil {
			require.NoError(t, res.Body.Close())
		}

		assert.Nil(t, res)
		assert.Equal(t, int32(2), attempts.Load(),
			"Expected the per-request override to win over the client configuration")
	})

	t.Run("zero retries disables retrying", func(t *testing.T) {
		t.Parallel()

		serverURL, attempts := newUnavailableServer(t)

		client, err := hqgohttp.NewClient(
			hqgohttp.WithMaxRetries(5),
			hqgohttp.WithRetryBackoff(noWaitBackoff),
		)

		require.NoError(t, err)

		res, err := client.Get(t.Context(), serverURL, hqgohttp.WithRequestMaxRetries(0))

		require.Error(t, err)

		if res != nil {
			require.NoError(t, res.Body.Close())
		}

		assert.Nil(t, res)
		assert.Equal(t, int32(1), attempts.Load())
	})
}

func TestClient_ContextCancellation(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(30 * time.Second):
		}
	}))

	t.Cleanup(server.Close)

	client, err := hqgohttp.NewClient(hqgohttp.WithTimeout(30 * time.Second))

	require.NoError(t, err)

	ctx, cancel := context.WithCancel(t.Context())

	go func() {
		time.Sleep(100 * time.Millisecond)

		cancel()
	}()

	start := time.Now()

	res, err := client.Get(ctx, server.URL)

	elapsed := time.Since(start)

	require.Error(t, err)

	if res != nil {
		require.NoError(t, res.Body.Close())
	}

	assert.Nil(t, res)
	require.ErrorIs(t, err, context.Canceled)
	assert.Less(t, elapsed, 10*time.Second,
		"Expected the cancellation to end the execution promptly")
}

func TestClient_Timeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(30 * time.Second):
		}
	}))

	t.Cleanup(server.Close)

	client, err := hqgohttp.NewClient(
		hqgohttp.WithTimeout(150*time.Millisecond),
		hqgohttp.WithRetryBackoff(noWaitBackoff),
	)

	require.NoError(t, err)

	start := time.Now()

	res, err := client.Get(t.Context(), server.URL)

	elapsed := time.Since(start)

	require.Error(t, err)

	if res != nil {
		require.NoError(t, res.Body.Close())
	}

	assert.Nil(t, res)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, elapsed, 10*time.Second,
		"Expected the configured timeout to bound the execution")
}

func TestClient_Do_ResponseBodyOutlivesExecution(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)

		w.(http.Flusher).Flush()

		// Hold the body back until well after Do has returned, so the caller's
		// read below is in flight across Do's deferred cleanup: if Do cancels
		// the timeout context on return, the read aborts with context canceled.
		time.Sleep(200 * time.Millisecond)

		_, _ = io.WriteString(w, "streamed body")
	}))

	t.Cleanup(server.Close)

	client, err := hqgohttp.NewClient(hqgohttp.WithTimeout(30 * time.Second))

	require.NoError(t, err)

	res, err := client.Get(t.Context(), server.URL)

	require.NoError(t, err)
	require.NotNil(t, res)

	body, err := io.ReadAll(res.Body)

	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	assert.Equal(t, "streamed body", string(body))
}

// fakeRateLimiter is a RateLimiter counting its invocations, optionally
// failing.
//
// Fields:
//   - calls (*atomic.Int32): the number of Wait invocations.
//   - err (error): the error Wait returns; nil means success.
type fakeRateLimiter struct {
	calls *atomic.Int32
	err   error
}

// Compile-time guard ensuring [fakeRateLimiter] satisfies
// [hqgohttp.RateLimiter].
var _ hqgohttp.RateLimiter = (*fakeRateLimiter)(nil)

// Wait counts the invocation and returns the configured error.
func (l *fakeRateLimiter) Wait(_ context.Context) (err error) {
	l.calls.Add(1)

	err = l.err

	return
}

func TestClient_RateLimiter(t *testing.T) {
	t.Parallel()

	t.Run("invoked before every attempt", func(t *testing.T) {
		t.Parallel()

		var attempts atomic.Int32

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			if attempts.Add(1) < 3 {
				w.WriteHeader(hqgohttpstatus.InternalServerError.Int())

				return
			}

			_, _ = w.Write([]byte("limited"))
		}))

		t.Cleanup(server.Close)

		var calls atomic.Int32

		limiter := &fakeRateLimiter{calls: &calls}

		client, err := hqgohttp.NewClient(
			hqgohttp.WithRateLimiter(limiter),
			hqgohttp.WithRetryBackoff(noWaitBackoff),
		)

		require.NoError(t, err)

		res, err := client.Get(t.Context(), server.URL)

		require.NoError(t, err)
		require.NotNil(t, res)

		assert.Equal(t, "limited", readBody(t, res))
		assert.Equal(t, int32(3), calls.Load(),
			"Expected the limiter to be waited on once per attempt")
	})

	t.Run("failure aborts the execution", func(t *testing.T) {
		t.Parallel()

		errLimit := errors.New("limiter unavailable")

		var calls atomic.Int32

		limiter := &fakeRateLimiter{calls: &calls, err: errLimit}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("unreachable"))
		}))

		t.Cleanup(server.Close)

		client, err := hqgohttp.NewClient(hqgohttp.WithRateLimiter(limiter))

		require.NoError(t, err)

		res, err := client.Get(t.Context(), server.URL)

		require.Error(t, err)

		if res != nil {
			require.NoError(t, res.Body.Close())
		}

		assert.Nil(t, res)
		require.ErrorIs(t, err, errLimit)
		assert.Equal(t, int32(1), calls.Load(),
			"Expected the limiter failure to abort before any retry")
	})
}

func TestClient_ObservabilityCallbacks(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if attempts.Add(1) == 1 {
			w.WriteHeader(hqgohttpstatus.InternalServerError.Int())

			return
		}

		_, _ = w.Write([]byte("observed"))
	}))

	t.Cleanup(server.Close)

	var onRequestCalls, onResponseCalls atomic.Int32

	var statuses []int

	client, err := hqgohttp.NewClient(
		hqgohttp.WithRetryBackoff(noWaitBackoff),
		hqgohttp.WithOnRequest(func(req *http.Request) {
			assert.NotNil(t, req)

			onRequestCalls.Add(1)
		}),
		hqgohttp.WithOnResponse(func(resp *http.Response, _ error) {
			onResponseCalls.Add(1)

			if resp != nil {
				statuses = append(statuses, resp.StatusCode)
			}
		}),
	)

	require.NoError(t, err)

	res, err := client.Get(t.Context(), server.URL)

	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, "observed", readBody(t, res))
	assert.Equal(t, int32(2), onRequestCalls.Load(),
		"Expected the request callback once per attempt")
	assert.Equal(t, int32(2), onResponseCalls.Load(),
		"Expected the response callback once per attempt")
	assert.Equal(t, []int{
		hqgohttpstatus.InternalServerError.Int(),
		hqgohttpstatus.OK.Int(),
	}, statuses, "Expected the callback to observe every attempt's outcome")
}

func TestClient_MaxRequestsBeforeIdleClose(t *testing.T) {
	t.Parallel()

	var closes atomic.Int32

	transport := &closeCountingTransport{closes: &closes}

	client, err := hqgohttp.NewClient(
		hqgohttp.WithHTTPClient(&http.Client{Transport: transport}),
		hqgohttp.WithMaxRequestsBeforeIdleClose(2),
	)

	require.NoError(t, err)

	for range 5 {
		res, err := client.Get(t.Context(), "http://example.com/")

		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
	}

	assert.Equal(t, int32(2), closes.Load(),
		"Expected idle connections to be closed on the 2nd and 4th completed requests")
}

// closeCountingTransport is an http.RoundTripper answering 200 OK and
// counting CloseIdleConnections calls.
type closeCountingTransport struct {
	closes *atomic.Int32
}

// RoundTrip answers every request with a 200 OK.
func (t *closeCountingTransport) RoundTrip(_ *http.Request) (res *http.Response, err error) {
	res = &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       io.NopCloser(strings.NewReader("OK")),
		Header:     make(http.Header),
	}

	return
}

// CloseIdleConnections increments the close counter.
func (t *closeCountingTransport) CloseIdleConnections() {
	t.closes.Add(1)
}

func TestClient_UserAgent(t *testing.T) {
	t.Parallel()

	// newCaptureServer starts an httptest server recording the User-Agent of
	// the last request it received, and returns its URL and the record.
	newCaptureServer := func(t *testing.T) (serverURL string, captured *atomic.Value) {
		t.Helper()

		captured = &atomic.Value{}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			captured.Store(r.UserAgent())

			_, _ = w.Write([]byte("ua"))
		}))

		t.Cleanup(server.Close)

		serverURL = server.URL

		return
	}

	t.Run("default", func(t *testing.T) {
		t.Parallel()

		serverURL, captured := newCaptureServer(t)

		client, err := hqgohttp.NewClient()

		require.NoError(t, err)

		res, err := client.Get(t.Context(), serverURL)

		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		assert.Equal(t, hqgohttp.DefaultUserAgent, captured.Load())
	})

	t.Run("fixed", func(t *testing.T) {
		t.Parallel()

		serverURL, captured := newCaptureServer(t)

		client, err := hqgohttp.NewClient(hqgohttp.WithUserAgent("custom-agent/1.0"))

		require.NoError(t, err)

		res, err := client.Get(t.Context(), serverURL)

		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		assert.Equal(t, "custom-agent/1.0", captured.Load())
	})

	t.Run("random rotation draws from the curated list", func(t *testing.T) {
		t.Parallel()

		serverURL, captured := newCaptureServer(t)

		client, err := hqgohttp.NewClient(hqgohttp.WithRandomUserAgent())

		require.NoError(t, err)

		res, err := client.Get(t.Context(), serverURL)

		require.NoError(t, err)
		require.NoError(t, res.Body.Close())

		userAgent, ok := captured.Load().(string)

		require.True(t, ok)
		assert.Contains(t, hqgohttpuseragent.List(), userAgent)
	})

	t.Run("request header wins", func(t *testing.T) {
		t.Parallel()

		serverURL, captured := newCaptureServer(t)

		client, err := hqgohttp.NewClient(hqgohttp.WithUserAgent("custom-agent/1.0"))

		require.NoError(t, err)

		res, err := client.NewRequest().
			URL(serverURL).
			SetHeader("User-Agent", "explicit-agent/2.0").
			Do(t.Context())

		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		assert.Equal(t, "explicit-agent/2.0", captured.Load())
	})
}

func TestClient_ReusableBodyAcrossRetries(t *testing.T) {
	t.Parallel()

	var bodies []string

	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err == nil {
			bodies = append(bodies, string(body))
		}

		if attempts.Add(1) == 1 {
			w.WriteHeader(hqgohttpstatus.InternalServerError.Int())

			return
		}

		_, _ = w.Write([]byte("got it"))
	}))

	t.Cleanup(server.Close)

	client, err := hqgohttp.NewClient(
		hqgohttp.WithMaxRetries(1),
		hqgohttp.WithRetryBackoff(noWaitBackoff),
	)

	require.NoError(t, err)

	res, err := client.Post(t.Context(), server.URL, "the payload")

	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, "got it", readBody(t, res))
	assert.Equal(t, []string{"the payload", "the payload"}, bodies,
		"Expected every attempt to resend the full body")
}

func TestPackageLevelShortcuts(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("package level"))
	}))

	t.Cleanup(server.Close)

	res, err := hqgohttp.Get(t.Context(), server.URL)

	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, "package level", readBody(t, res))
}
