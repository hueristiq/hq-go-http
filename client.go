package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	hqgohttpmethod "github.com/hueristiq/hq-lib-http-go/method"
	hqgohttprequest "github.com/hueristiq/hq-lib-http-go/request"
	hqgoretrier "github.com/hueristiq/hq-lib-retrier-go"
	hqgoretrierbackoff "github.com/hueristiq/hq-lib-retrier-go/backoff"
)

// RateLimiter paces request attempts. It is the minimal surface the client
// needs from a rate limiter: *Limiter from github.com/hueristiq/hq-lib-limiter-go
// satisfies it, and so does any compatible implementation — no module
// dependency is imposed.
type RateLimiter interface {
	// Wait blocks until the next attempt may proceed or ctx is done.
	Wait(ctx context.Context) (err error)
}

// OnRequestFunc is the callback invoked with the outgoing request before
// every attempt of every request, registered with [WithOnRequest].
//
// Parameters:
//   - req (*http.Request): the request about to be sent.
type OnRequestFunc func(req *http.Request)

// OnResponseFunc is the callback invoked with the outcome of every attempt of
// every request — before the retry policy is consulted — registered with
// [WithOnResponse].
//
// Parameters:
//   - resp (*http.Response): the response received on the attempt, or nil.
//   - err (error): the transport error of the attempt, or nil.
type OnResponseFunc func(resp *http.Response, err error)

// Client is the primary structure used to perform HTTP requests.
//
// It executes requests with retry logic, honoring the caller's context, a
// configurable retry policy that sees both the response and the error of each
// attempt, a rate limiter hook, and observability callbacks. It maintains two
// underlying clients: the primary one executing requests, and an HTTP/2
// fallback used when a cleartext HTTP/1.x exchange fails because the server
// answered with HTTP/2 frames. It can also periodically close idle
// connections to bound resource usage in long-running processes.
//
// A Client is immutable after construction and safe for concurrent use; only
// the request counter driving the periodic idle-connection close mutates,
// through atomic operations. Construct through [NewClient]; the zero value is
// not usable.
//
// Fields:
//   - clientOptions: the embedded configuration — see its Fields list.
//   - fallbackHTTPClient (*http.Client): the client used when the primary
//     client fails with the HTTP/1.x transport broken error.
//   - requestCount (atomic.Uint64): the number of completed executions since
//     the last idle-connection close.
type Client struct {
	clientOptions

	fallbackHTTPClient *http.Client
	requestCount       atomic.Uint64
}

// NewClient creates a [Client] configured by the given options.
//
// Unset options fall back to defaults: a pooled HTTP/1.x client with HTTP/2
// support, no base URL, [DefaultUserAgent] as User-Agent, [DefaultTimeout]
// per execution, [DefaultMaxRetries] retries with waits between
// [DefaultRetryWaitMin] and [DefaultRetryWaitMax] under exponential backoff
// with decorrelated jitter, [DefaultRetryPolicy], no rate limiter, no
// callbacks, no periodic idle-connection close, and [DefaultResponseReadLimit]
// as the drain bound. Invalid values are normalized the same way — see each
// option's documentation.
//
// Parameters:
//   - ofs (...ClientOptionFunc): functional options configuring the client.
//
// Returns:
//   - client (*Client): a new client, ready to execute requests.
//   - err (error): non-nil when the HTTP/2 fallback transport cannot be
//     configured — practically impossible, as the transport is built from
//     scratch here.
func NewClient(ofs ...ClientOptionFunc) (client *Client, err error) {
	opts := &clientOptions{
		timeout:           DefaultTimeout,
		maxRetries:        DefaultMaxRetries,
		retryWaitMin:      DefaultRetryWaitMin,
		retryWaitMax:      DefaultRetryWaitMax,
		newRetryBackoff:   hqgoretrierbackoff.ExponentialWithDecorrelatedJitter,
		retryPolicy:       DefaultRetryPolicy,
		userAgent:         DefaultUserAgent,
		responseReadLimit: DefaultResponseReadLimit,
	}

	for _, of := range ofs {
		of(opts)
	}

	opts.normalize()

	client = &Client{clientOptions: *opts}

	if client.httpClient == nil {
		client.httpClient = &http.Client{
			Transport: DefaultHTTPPooledTransport(),
		}
	}

	fallbackHTTPClient, err := newFallbackHTTPClient()
	if err != nil {
		err = fmt.Errorf("http.NewClient: configuring HTTP/2 fallback transport: %w", err)

		client = nil

		return
	}

	client.fallbackHTTPClient = fallbackHTTPClient

	return
}

// Do executes req, honoring the caller's context, with retries driven by the
// resolved configuration: the client configuration overlaid with ofs.
//
// The execution — every attempt and every wait between them — is bounded by
// ctx, tightened by the configured timeout when the context carries no
// earlier deadline. A nil ctx falls back to the request's own context;
// otherwise req is shallow-copied onto ctx, so the caller's request is never
// mutated.
//
// Each attempt waits on the configured rate limiter, runs the OnRequest
// callback, executes the request, runs the OnResponse callback, and consults
// the retry policy. A request body built through the request package or
// [Client.NewRequest] is reusable, so every attempt resends it in full. When
// the primary client fails with the HTTP/1.x transport broken error, the
// attempt is retried once on the HTTP/2 fallback client before the policy is
// consulted.
//
// Outcomes:
//   - Success: the final response with its body open — the caller owns and
//     must close it.
//   - Attempts exhausted on retryable failures: a nil response and a
//     "giving up" error wrapping the last failure. The last response's body
//     was drained and closed.
//   - Context done: a nil response and the context's error.
//   - Retry policy abort (non-nil rerr): the attempt's response — body open,
//     caller-owned — and rerr.
//
// Parameters:
//   - ctx (context.Context): the context governing the whole execution.
//   - req (*hqgohttprequest.Request): the request to execute.
//   - ofs (...RequestOptionFunc): per-request overrides of the execution
//     knobs; the client configuration stays untouched.
//
// Returns:
//   - res (*http.Response): the final response, or nil on failure (except a
//     policy abort, which returns the attempt's response).
//   - err (error): non-nil when the execution fails.
func (c *Client) Do(ctx context.Context, req *hqgohttprequest.Request, ofs ...RequestOptionFunc) (res *http.Response, err error) {
	if req == nil || req.Request == nil {
		err = errors.New("http.Client.Do: request is nil")

		return
	}

	if ctx == nil {
		//nolint:contextcheck // Falling back to the request's own context is exactly the documented behavior.
		ctx = req.Context()
	}

	opts := c.requestOptions(ofs...)

	if c.timeout > 0 {
		var cancel context.CancelFunc

		ctx, cancel = context.WithTimeout(ctx, c.timeout)

		defer cancel()
	}

	req = &hqgohttprequest.Request{Request: req.WithContext(ctx)}

	// retryAfter carries the wait honored from a Retry-After header from the
	// attempt operation to the backoff strategy. Both run on the retry loop's
	// goroutine, so a plain variable is safe.
	var retryAfter time.Duration

	operation := func() (attemptRes *http.Response, attemptErr error) {
		retryAfter = 0

		if opts.rateLimiter != nil {
			if werr := opts.rateLimiter.Wait(ctx); werr != nil {
				attemptErr = fmt.Errorf("http.Client.Do: waiting on rate limiter: %w", werr)

				return
			}
		}

		if c.onRequest != nil {
			c.onRequest(req.Request)
		}

		attemptRes, attemptErr = c.httpClient.Do(req.Request)

		if attemptErr != nil && isHTTP1BrokenTransportError(attemptErr) {
			attemptRes, attemptErr = c.fallbackHTTPClient.Do(req.Request)
		}

		if c.onResponse != nil {
			c.onResponse(attemptRes, attemptErr)
		}

		if attemptRes == nil && attemptErr == nil {
			attemptErr = errors.New("the transport returned neither a response nor an error")
		}

		retry, rerr := opts.retryPolicy(ctx, attemptRes, attemptErr)
		if rerr != nil {
			attemptErr = rerr

			return
		}

		if !retry {
			return
		}

		if attemptRes != nil {
			if wait, ok := retryAfterWait(attemptRes); ok {
				retryAfter = wait
			}

			drainBody(attemptRes.Body, opts.responseReadLimit)
		}

		if attemptErr == nil {
			attemptErr = fmt.Errorf("unexpected HTTP status %s", attemptRes.Status)
		}

		attemptErr = &retryableError{err: attemptErr}

		return
	}

	res, err = hqgoretrier.RetryWithData(ctx, operation,
		hqgoretrier.WithMaxAttempts(opts.maxRetries+1),
		hqgoretrier.WithWaitMin(opts.retryWaitMin),
		hqgoretrier.WithWaitMax(opts.retryWaitMax),
		hqgoretrier.WithBackoff(func(minDelay, maxDelay time.Duration) (backoff hqgoretrierbackoff.Backoff) {
			base := opts.newRetryBackoff(minDelay, maxDelay)

			return func(attempt int) (delay time.Duration) {
				if retryAfter > 0 {
					delay = min(retryAfter, maxDelay)

					retryAfter = 0

					return
				}

				delay = base(attempt)

				return
			}
		}),
		hqgoretrier.WithRetryOn(isMarkedRetryable),
	)

	c.closeIdleConnectionsIfDue()

	var marker *retryableError

	switch {
	case err == nil:
		// The final response with its body open; the caller owns it.
	case errors.As(err, &marker):
		// Attempts exhausted: the last response's body was already drained.
		res = nil

		err = fmt.Errorf("http.Client.Do: giving up after %d attempt(s) (%s %s): %w",
			opts.maxRetries+1, req.Method, req.URL.Redacted(), marker.err)
	case ctx.Err() != nil:
		// The context ended the execution; the last response, if any, is in an
		// unknown body state, so it is not handed out.
		res = nil
	}

	return
}

// requestOptions resolves the execution configuration for a single request:
// the client configuration overlaid with the given overrides.
//
// Parameters:
//   - ofs (...RequestOptionFunc): the per-request overrides.
//
// Returns:
//   - opts (*requestOptions): the resolved, normalized configuration.
func (c *Client) requestOptions(ofs ...RequestOptionFunc) (opts *requestOptions) {
	opts = &requestOptions{
		maxRetries:        c.maxRetries,
		retryWaitMin:      c.retryWaitMin,
		retryWaitMax:      c.retryWaitMax,
		newRetryBackoff:   c.newRetryBackoff,
		retryPolicy:       c.retryPolicy,
		rateLimiter:       c.rateLimiter,
		responseReadLimit: c.responseReadLimit,
	}

	for _, of := range ofs {
		of(opts)
	}

	opts.normalize()

	return
}

// CloseIdleConnections closes the idle connections of both underlying
// clients, freeing file descriptors held by their connection pools.
// In-flight requests are unaffected.
func (c *Client) CloseIdleConnections() {
	c.httpClient.CloseIdleConnections()
	c.fallbackHTTPClient.CloseIdleConnections()
}

// closeIdleConnectionsIfDue counts a completed execution and, once the
// configured threshold is reached, closes idle connections and restarts the
// count. A non-positive threshold disables the periodic closing.
func (c *Client) closeIdleConnectionsIfDue() {
	if c.maxRequestsBeforeIdleClose <= 0 {
		return
	}

	for {
		count := c.requestCount.Load()

		if count+1 < uint64(c.maxRequestsBeforeIdleClose) {
			if c.requestCount.CompareAndSwap(count, count+1) {
				return
			}

			continue
		}

		if c.requestCount.CompareAndSwap(count, 0) {
			c.CloseIdleConnections()

			return
		}
	}
}

// Get is a convenience method for performing an HTTP GET request.
//
// Parameters:
//   - ctx (context.Context): the context governing the whole execution.
//   - rawURL (string): the target URL — absolute, or relative when a base URL
//     is configured.
//   - ofs (...RequestOptionFunc): per-request overrides of the execution
//     knobs.
//
// Returns:
//   - res (*http.Response): the response received from the server.
//   - err (error): non-nil when the request fails.
func (c *Client) Get(ctx context.Context, rawURL string, ofs ...RequestOptionFunc) (res *http.Response, err error) {
	res, err = c.NewRequest().Method(hqgohttpmethod.GET).URL(rawURL).WithOptions(ofs...).Do(ctx)

	return
}

// Head is a convenience method for performing an HTTP HEAD request.
//
// Parameters:
//   - ctx (context.Context): the context governing the whole execution.
//   - rawURL (string): the target URL — absolute, or relative when a base URL
//     is configured.
//   - ofs (...RequestOptionFunc): per-request overrides of the execution
//     knobs.
//
// Returns:
//   - res (*http.Response): the response received from the server.
//   - err (error): non-nil when the request fails.
func (c *Client) Head(ctx context.Context, rawURL string, ofs ...RequestOptionFunc) (res *http.Response, err error) {
	res, err = c.NewRequest().Method(hqgohttpmethod.HEAD).URL(rawURL).WithOptions(ofs...).Do(ctx)

	return
}

// Post is a convenience method for performing an HTTP POST request with a
// body.
//
// The body accepts the types supported by the request package: a string, a
// []byte, an io.Reader, an io.ReadSeeker, a *bytes.Buffer, a
// func() (io.Reader, error), or a ReusableReadCloser.
//
// Parameters:
//   - ctx (context.Context): the context governing the whole execution.
//   - rawURL (string): the target URL — absolute, or relative when a base URL
//     is configured.
//   - body (any): the request payload; nil means no body.
//   - ofs (...RequestOptionFunc): per-request overrides of the execution
//     knobs.
//
// Returns:
//   - res (*http.Response): the response received from the server.
//   - err (error): non-nil when the request fails.
func (c *Client) Post(ctx context.Context, rawURL string, body any, ofs ...RequestOptionFunc) (res *http.Response, err error) {
	res, err = c.NewRequest().Method(hqgohttpmethod.POST).URL(rawURL).Body(body).WithOptions(ofs...).Do(ctx)

	return
}

// Put is a convenience method for performing an HTTP PUT request with a body.
// See [Client.Post] for the accepted body types.
//
// Parameters:
//   - ctx (context.Context): the context governing the whole execution.
//   - rawURL (string): the target URL — absolute, or relative when a base URL
//     is configured.
//   - body (any): the request payload; nil means no body.
//   - ofs (...RequestOptionFunc): per-request overrides of the execution
//     knobs.
//
// Returns:
//   - res (*http.Response): the response received from the server.
//   - err (error): non-nil when the request fails.
func (c *Client) Put(ctx context.Context, rawURL string, body any, ofs ...RequestOptionFunc) (res *http.Response, err error) {
	res, err = c.NewRequest().Method(hqgohttpmethod.PUT).URL(rawURL).Body(body).WithOptions(ofs...).Do(ctx)

	return
}

// Patch is a convenience method for performing an HTTP PATCH request with a
// body. See [Client.Post] for the accepted body types.
//
// Parameters:
//   - ctx (context.Context): the context governing the whole execution.
//   - rawURL (string): the target URL — absolute, or relative when a base URL
//     is configured.
//   - body (any): the request payload; nil means no body.
//   - ofs (...RequestOptionFunc): per-request overrides of the execution
//     knobs.
//
// Returns:
//   - res (*http.Response): the response received from the server.
//   - err (error): non-nil when the request fails.
func (c *Client) Patch(ctx context.Context, rawURL string, body any, ofs ...RequestOptionFunc) (res *http.Response, err error) {
	res, err = c.NewRequest().Method(hqgohttpmethod.PATCH).URL(rawURL).Body(body).WithOptions(ofs...).Do(ctx)

	return
}

// Delete is a convenience method for performing an HTTP DELETE request.
//
// Parameters:
//   - ctx (context.Context): the context governing the whole execution.
//   - rawURL (string): the target URL — absolute, or relative when a base URL
//     is configured.
//   - ofs (...RequestOptionFunc): per-request overrides of the execution
//     knobs.
//
// Returns:
//   - res (*http.Response): the response received from the server.
//   - err (error): non-nil when the request fails.
func (c *Client) Delete(ctx context.Context, rawURL string, ofs ...RequestOptionFunc) (res *http.Response, err error) {
	res, err = c.NewRequest().Method(hqgohttpmethod.DELETE).URL(rawURL).WithOptions(ofs...).Do(ctx)

	return
}

// Options is a convenience method for performing an HTTP OPTIONS request.
//
// Parameters:
//   - ctx (context.Context): the context governing the whole execution.
//   - rawURL (string): the target URL — absolute, or relative when a base URL
//     is configured.
//   - ofs (...RequestOptionFunc): per-request overrides of the execution
//     knobs.
//
// Returns:
//   - res (*http.Response): the response received from the server.
//   - err (error): non-nil when the request fails.
func (c *Client) Options(ctx context.Context, rawURL string, ofs ...RequestOptionFunc) (res *http.Response, err error) {
	res, err = c.NewRequest().Method(hqgohttpmethod.OPTIONS).URL(rawURL).WithOptions(ofs...).Do(ctx)

	return
}

// isMarkedRetryable reports whether err carries the retryable marker, i.e.
// whether the retry engine should schedule another attempt. Errors without
// the marker are final and returned to the caller unchanged.
//
// Parameters:
//   - err (error): the failure to classify.
//
// Returns:
//   - retry (bool): true when another attempt may follow.
func isMarkedRetryable(err error) (retry bool) {
	var marker *retryableError

	retry = errors.As(err, &marker)

	return
}

// isHTTP1BrokenTransportError reports whether err is the failure net/http
// produces when a cleartext HTTP/1.x exchange receives an HTTP/2 frame: the
// server speaks HTTP/2 without an upgrade. When it fires, the attempt is
// retried on the HTTP/2 fallback client.
//
// net/http builds this error as
// fmt.Errorf("net/http: HTTP/1.x transport connection broken: %w", err)
// wrapping an unexported badStringError, so no errors.As target exists;
// matching the message is the only way to detect it. The check is isolated
// here so the string match lives in exactly one place.
//
// Parameters:
//   - err (error): the error to inspect.
//
// Returns:
//   - broken (bool): true when the error indicates a broken HTTP/1.x
//     transport connection.
func isHTTP1BrokenTransportError(err error) (broken bool) {
	broken = err != nil &&
		strings.Contains(err.Error(), `net/http: HTTP/1.x transport connection broken: malformed HTTP version "HTTP/2"`)

	return
}

// drainBody reads and discards up to limit bytes from body and closes it, so
// the underlying connection can be reused for future requests.
//
// Parameters:
//   - body (io.ReadCloser): the response body to drain.
//   - limit (int64): the maximum number of bytes to read.
func drainBody(body io.ReadCloser, limit int64) {
	defer func() {
		_ = body.Close()
	}()

	_, _ = io.Copy(io.Discard, io.LimitReader(body, limit))
}
