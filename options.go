package http

import (
	"net/http"
	"strings"
	"time"

	hqgoretrierbackoff "github.com/hueristiq/hq-lib-retrier-go/backoff"
)

// clientOptions holds the configurable knobs of a [Client], populated through
// [ClientOptionFunc] values passed to [NewClient].
//
// Fields:
//   - httpClient (*http.Client): the underlying client executing HTTP/1.x
//     requests. Nil means a pooled default client is built.
//   - baseURL (string): the base URL prepended to relative request URLs.
//     Empty means request URLs are used as given.
//   - headers ([]headerField): default headers set on every request, unless
//     the request sets the same key itself.
//   - userAgent (string): the User-Agent header value. Empty selects
//     [DefaultUserAgent]; ignored when randomUserAgent is set.
//   - randomUserAgent (bool): whether each built request draws a random
//     User-Agent from the useragent package instead of using userAgent.
//   - timeout (time.Duration): the bound on a whole request execution,
//     including retries and the waits between them. Non-positive means
//     [DefaultTimeout].
//   - maxRetries (int): the number of retries after the initial attempt.
//     Zero disables retries; negative falls back to [DefaultMaxRetries].
//   - retryWaitMin (time.Duration): the minimum delay between attempts.
//   - retryWaitMax (time.Duration): the maximum delay between attempts, also
//     capping an honored Retry-After wait.
//   - newRetryBackoff (func(minDelay, maxDelay time.Duration) (backoff hqgoretrierbackoff.Backoff)):
//     the constructor for the strategy computing the delay per attempt. Nil selects
//     exponential backoff with decorrelated jitter.
//   - retryPolicy (RetryPolicy): the predicate deciding, per attempt, whether
//     to retry. Nil selects [DefaultRetryPolicy].
//   - rateLimiter (RateLimiter): the limiter waited on before every attempt.
//     Nil means no limiting.
//   - onRequest (OnRequestFunc): the callback invoked before every attempt.
//   - onResponse (OnResponseFunc): the callback invoked after every attempt.
//   - maxRequestsBeforeIdleClose (int): the number of completed executions
//     after which idle connections are closed. Non-positive disables the
//     periodic closing.
//   - responseReadLimit (int64): the maximum number of bytes drained from a
//     retried response's body so the connection can be reused. Non-positive
//     falls back to [DefaultResponseReadLimit].
type clientOptions struct {
	httpClient                 *http.Client
	baseURL                    string
	headers                    []headerField
	userAgent                  string
	randomUserAgent            bool
	timeout                    time.Duration
	maxRetries                 int
	retryWaitMin               time.Duration
	retryWaitMax               time.Duration
	newRetryBackoff            func(minDelay, maxDelay time.Duration) (backoff hqgoretrierbackoff.Backoff)
	retryPolicy                RetryPolicy
	rateLimiter                RateLimiter
	onRequest                  OnRequestFunc
	onResponse                 OnResponseFunc
	maxRequestsBeforeIdleClose int
	responseReadLimit          int64
}

// ClientOptionFunc configures a [Client] through the functional-options idiom.
//
// When the same option is supplied more than once, the last one wins.
//
// Parameters:
//   - opts (*clientOptions): the options bag to mutate.
type ClientOptionFunc func(opts *clientOptions)

// requestOptions holds the per-request overrides of the execution knobs of a
// [Client], populated through [RequestOptionFunc] values passed to
// [Client.Do], the convenience methods, or [RequestBuilder.WithOptions].
//
// The values seed from the client configuration; every option replaces the
// seeded value it touches.
//
// Fields:
//   - maxRetries (int): the number of retries after the initial attempt.
//   - retryWaitMin (time.Duration): the minimum delay between attempts.
//   - retryWaitMax (time.Duration): the maximum delay between attempts, also
//     capping an honored Retry-After wait.
//   - newRetryBackoff (func(minDelay, maxDelay time.Duration) (backoff hqgoretrierbackoff.Backoff)):
//     the constructor for the strategy computing the delay per attempt.
//   - retryPolicy (RetryPolicy): the predicate deciding, per attempt, whether
//     to retry.
//   - rateLimiter (RateLimiter): the limiter waited on before every attempt.
//   - responseReadLimit (int64): the maximum number of bytes drained from a
//     retried response's body.
type requestOptions struct {
	maxRetries        int
	retryWaitMin      time.Duration
	retryWaitMax      time.Duration
	newRetryBackoff   func(minDelay, maxDelay time.Duration) (backoff hqgoretrierbackoff.Backoff)
	retryPolicy       RetryPolicy
	rateLimiter       RateLimiter
	responseReadLimit int64
}

// RequestOptionFunc configures a single request execution through the
// functional-options idiom, overriding the matching client configuration for
// that execution only.
//
// When the same option is supplied more than once, the last one wins.
//
// Parameters:
//   - opts (*requestOptions): the options bag to mutate.
type RequestOptionFunc func(opts *requestOptions)

const (
	// DefaultTimeout is the default bound on a whole request execution,
	// including retries and the waits between them.
	//
	// It applies when [WithTimeout] is not supplied, and when a supplied value
	// less than or equal to 0 is normalized back to the default.
	DefaultTimeout = 30 * time.Second

	// DefaultMaxRetries is the default number of retries after the initial
	// attempt, so a client performs up to 3 attempts per request by default.
	//
	// It applies when [WithMaxRetries] is not supplied, and when a supplied
	// value below 0 is normalized back to the default.
	DefaultMaxRetries = 2

	// DefaultRetryWaitMin is the default minimum delay between retry attempts,
	// serving as the base delay for backoff calculations.
	//
	// It applies when [WithRetryWaitMin] is not supplied, and when a supplied
	// value less than or equal to 0 is normalized back to the default.
	DefaultRetryWaitMin = 1 * time.Second

	// DefaultRetryWaitMax is the default maximum delay between retry attempts.
	// It also caps the wait honored from a Retry-After response header.
	//
	// It applies when [WithRetryWaitMax] is not supplied, and when a supplied
	// value less than or equal to 0 is normalized back to the default.
	DefaultRetryWaitMax = 30 * time.Second

	// DefaultResponseReadLimit is the default maximum number of bytes drained
	// from the body of a response that will be retried, so the underlying
	// connection can be reused.
	//
	// It applies when [WithResponseReadLimit] is not supplied, and when a
	// supplied value less than or equal to 0 is normalized back to the default.
	DefaultResponseReadLimit = 4096

	// DefaultUserAgent is the default User-Agent header sent with requests.
	//
	// It applies when neither [WithUserAgent] nor [WithRandomUserAgent] is
	// supplied and the request itself does not set a User-Agent header.
	DefaultUserAgent = "hq-lib-http-go (https://github.com/hueristiq/hq-lib-http-go)"
)

// WithHTTPClient returns a [ClientOptionFunc] that sets the underlying client
// executing HTTP/1.x requests.
//
// Supplying a custom client hands transport control (proxying, TLS settings,
// connection pooling) to the caller. The client is used as-is; its Timeout
// field is left untouched — use [WithTimeout] or a context deadline instead.
//
// Parameters:
//   - client (*http.Client): the client to execute requests with. Nil is
//     ignored, keeping the pooled default.
//
// Returns:
//   - f (ClientOptionFunc): a functional option that sets the httpClient field.
func WithHTTPClient(client *http.Client) (f ClientOptionFunc) {
	return func(opts *clientOptions) {
		if client != nil {
			opts.httpClient = client
		}
	}
}

// WithBaseURL returns a [ClientOptionFunc] that sets the base URL prepended
// to relative request URLs.
//
// A request whose URL is absolute (carries a scheme) is never prefixed. The
// [RequestBuilder.BaseURL] method overrides this value for a single request.
//
// Parameters:
//   - baseURL (string): the base URL, e.g. "https://api.example.com". Empty
//     clears the base URL.
//
// Returns:
//   - f (ClientOptionFunc): a functional option that sets the baseURL field.
func WithBaseURL(baseURL string) (f ClientOptionFunc) {
	return func(opts *clientOptions) {
		opts.baseURL = baseURL
	}
}

// WithHeader returns a [ClientOptionFunc] that registers a default header set
// on every request, replacing any previously registered default with the same
// key (case-insensitive).
//
// A request-level header with the same key — set through
// [RequestBuilder.SetHeader] or [RequestBuilder.AddHeader] — wins over the
// default.
//
// Parameters:
//   - key (string): the header field name. Empty is ignored.
//   - value (string): the header field value.
//
// Returns:
//   - f (ClientOptionFunc): a functional option that registers the header.
func WithHeader(key, value string) (f ClientOptionFunc) {
	return func(opts *clientOptions) {
		if key == "" {
			return
		}

		for i, h := range opts.headers {
			if strings.EqualFold(h.key, key) {
				opts.headers[i].value = value

				return
			}
		}

		opts.headers = append(opts.headers, headerField{key: key, value: value})
	}
}

// WithUserAgent returns a [ClientOptionFunc] that sets the User-Agent header
// sent with every request that does not set one itself.
//
// Supplying it after [WithRandomUserAgent] disables the rotation; the last of
// the two options wins.
//
// Parameters:
//   - userAgent (string): the User-Agent value. Empty falls back to
//     [DefaultUserAgent].
//
// Returns:
//   - f (ClientOptionFunc): a functional option that sets the userAgent field.
func WithUserAgent(userAgent string) (f ClientOptionFunc) {
	return func(opts *clientOptions) {
		if userAgent == "" {
			userAgent = DefaultUserAgent
		}

		opts.userAgent = userAgent
		opts.randomUserAgent = false
	}
}

// WithRandomUserAgent returns a [ClientOptionFunc] that makes every built
// request draw a random User-Agent from the curated list of the useragent
// package, unless the request sets a User-Agent header itself.
//
// Supplying it after [WithUserAgent] disables the fixed value; the last of
// the two options wins.
//
// Returns:
//   - f (ClientOptionFunc): a functional option that enables the rotation.
func WithRandomUserAgent() (f ClientOptionFunc) {
	return func(opts *clientOptions) {
		opts.randomUserAgent = true
	}
}

// WithTimeout returns a [ClientOptionFunc] that bounds a whole request
// execution, including retries and the waits between them.
//
// The bound derives from the caller's context: an earlier deadline on that
// context still wins. A per-request timeout needs no option — derive it with
// context.WithTimeout on the context handed to the entry point.
//
// Parameters:
//   - timeout (time.Duration): the execution bound. Values less than or equal
//     to 0 fall back to [DefaultTimeout].
//
// Returns:
//   - f (ClientOptionFunc): a functional option that sets the timeout field.
func WithTimeout(timeout time.Duration) (f ClientOptionFunc) {
	return func(opts *clientOptions) {
		opts.timeout = timeout
	}
}

// WithMaxRetries returns a [ClientOptionFunc] that sets the number of retries
// after the initial attempt, so a value of 2 means up to 3 attempts per
// request.
//
// Parameters:
//   - maxRetries (int): the number of retries. Zero disables retries; values
//     below 0 fall back to [DefaultMaxRetries].
//
// Returns:
//   - f (ClientOptionFunc): a functional option that sets the maxRetries field.
func WithMaxRetries(maxRetries int) (f ClientOptionFunc) {
	return func(opts *clientOptions) {
		opts.maxRetries = maxRetries
	}
}

// WithRetryWaitMin returns a [ClientOptionFunc] that sets the minimum delay
// between retry attempts, serving as the base delay for backoff calculations.
//
// Parameters:
//   - retryWaitMin (time.Duration): the minimum delay. Values less than or
//     equal to 0 fall back to [DefaultRetryWaitMin].
//
// Returns:
//   - f (ClientOptionFunc): a functional option that sets the retryWaitMin field.
func WithRetryWaitMin(retryWaitMin time.Duration) (f ClientOptionFunc) {
	return func(opts *clientOptions) {
		opts.retryWaitMin = retryWaitMin
	}
}

// WithRetryWaitMax returns a [ClientOptionFunc] that sets the maximum delay
// between retry attempts, capping the backoff duration. The wait honored from
// a Retry-After response header is capped at this value as well.
//
// Parameters:
//   - retryWaitMax (time.Duration): the maximum delay. Values less than or
//     equal to 0 fall back to [DefaultRetryWaitMax], and a value below the
//     minimum wait is raised to it.
//
// Returns:
//   - f (ClientOptionFunc): a functional option that sets the retryWaitMax field.
func WithRetryWaitMax(retryWaitMax time.Duration) (f ClientOptionFunc) {
	return func(opts *clientOptions) {
		opts.retryWaitMax = retryWaitMax
	}
}

// WithRetryBackoff returns a [ClientOptionFunc] that selects the backoff
// strategy computing the delay for each retry attempt.
//
// It takes a constructor rather than a ready-made strategy: the retry engine
// calls it once per request execution with the normalized wait bounds, so
// stateful strategies such as the decorrelated jitter always run with fresh,
// per-execution state. The constructors in the backoff package of
// github.com/hueristiq/hq-lib-retrier-go match this signature and can be
// passed directly, e.g. WithRetryBackoff(backoff.ExponentialWithFullJitter).
//
// Parameters:
//   - newRetryBackoff (func(minDelay, maxDelay time.Duration) (backoff hqgoretrierbackoff.Backoff)):
//     the backoff strategy constructor. Nil selects the default strategy
//     (exponential backoff with decorrelated jitter).
//
// Returns:
//   - f (ClientOptionFunc): a functional option that sets the newRetryBackoff field.
func WithRetryBackoff(newRetryBackoff func(minDelay, maxDelay time.Duration) (backoff hqgoretrierbackoff.Backoff)) (f ClientOptionFunc) {
	return func(opts *clientOptions) {
		opts.newRetryBackoff = newRetryBackoff
	}
}

// WithRetryPolicy returns a [ClientOptionFunc] that sets the policy deciding,
// after each attempt, whether the request is retried.
//
// Parameters:
//   - retryPolicy (RetryPolicy): the policy. Nil selects [DefaultRetryPolicy],
//     which retries connection and transport errors, 429, and 5xx responses.
//
// Returns:
//   - f (ClientOptionFunc): a functional option that sets the retryPolicy field.
func WithRetryPolicy(retryPolicy RetryPolicy) (f ClientOptionFunc) {
	return func(opts *clientOptions) {
		opts.retryPolicy = retryPolicy
	}
}

// WithRateLimiter returns a [ClientOptionFunc] that registers a rate limiter
// waited on before every attempt of every request.
//
// The limiter is any value with a Wait method honoring a context — for
// example a *Limiter from github.com/hueristiq/hq-lib-limiter-go — so no module
// dependency is imposed. A limiter shared across clients must be safe for
// concurrent use.
//
// Parameters:
//   - rateLimiter (RateLimiter): the limiter. Nil disables limiting.
//
// Returns:
//   - f (ClientOptionFunc): a functional option that sets the rateLimiter field.
func WithRateLimiter(rateLimiter RateLimiter) (f ClientOptionFunc) {
	return func(opts *clientOptions) {
		opts.rateLimiter = rateLimiter
	}
}

// WithOnRequest returns a [ClientOptionFunc] that registers the callback
// invoked with the outgoing request before every attempt of every request.
//
// The callback is invoked synchronously on the goroutine running the attempt:
// a slow callback delays the attempt. It is observational — mutate the
// request only to add orthogonal metadata such as tracing headers. A callback
// must be safe for concurrent use, and a panicking callback propagates its
// panic to the caller of the entry point.
//
// Parameters:
//   - onRequest (OnRequestFunc): the callback. Nil drops the notification.
//
// Returns:
//   - f (ClientOptionFunc): a functional option that sets the onRequest field.
func WithOnRequest(onRequest OnRequestFunc) (f ClientOptionFunc) {
	return func(opts *clientOptions) {
		opts.onRequest = onRequest
	}
}

// WithOnResponse returns a [ClientOptionFunc] that registers the callback
// invoked with the outcome — the response, the error, or both — after every
// attempt of every request, before the retry policy is consulted.
//
// The callback is invoked synchronously on the goroutine running the attempt:
// a slow callback delays the retry decision. It is observational and must not
// close or drain the response body; the retry machinery owns the body until
// the final response is returned. A callback must be safe for concurrent use,
// and a panicking callback propagates its panic to the caller of the entry
// point.
//
// Parameters:
//   - onResponse (OnResponseFunc): the callback. Nil drops the notification.
//
// Returns:
//   - f (ClientOptionFunc): a functional option that sets the onResponse field.
func WithOnResponse(onResponse OnResponseFunc) (f ClientOptionFunc) {
	return func(opts *clientOptions) {
		opts.onResponse = onResponse
	}
}

// WithMaxRequestsBeforeIdleClose returns a [ClientOptionFunc] that sets how
// many request executions complete before the client closes its idle
// connections, freeing file descriptors held by the connection pool.
//
// Long-running processes making many requests to changing hosts can leak
// idle connections; a positive threshold periodically reclaims them.
//
// Parameters:
//   - n (int): the number of completed executions between idle-connection
//     closes. Values less than or equal to 0 disable the periodic closing.
//
// Returns:
//   - f (ClientOptionFunc): a functional option that sets the
//     maxRequestsBeforeIdleClose field.
func WithMaxRequestsBeforeIdleClose(n int) (f ClientOptionFunc) {
	return func(opts *clientOptions) {
		opts.maxRequestsBeforeIdleClose = n
	}
}

// WithResponseReadLimit returns a [ClientOptionFunc] that sets the maximum
// number of bytes drained from the body of a response that will be retried.
//
// Draining the body of a retried response lets the transport reuse the
// underlying connection; the limit bounds the work spent draining large
// bodies.
//
// Parameters:
//   - responseReadLimit (int64): the drain bound in bytes. Values less than
//     or equal to 0 fall back to [DefaultResponseReadLimit].
//
// Returns:
//   - f (ClientOptionFunc): a functional option that sets the responseReadLimit field.
func WithResponseReadLimit(responseReadLimit int64) (f ClientOptionFunc) {
	return func(opts *clientOptions) {
		opts.responseReadLimit = responseReadLimit
	}
}

// WithRequestMaxRetries returns a [RequestOptionFunc] that overrides the
// number of retries after the initial attempt for a single request.
//
// Parameters:
//   - maxRetries (int): the number of retries. Zero disables retries for the
//     request; values below 0 are ignored, keeping the client configuration.
//
// Returns:
//   - f (RequestOptionFunc): a functional option that sets the maxRetries field.
func WithRequestMaxRetries(maxRetries int) (f RequestOptionFunc) {
	return func(opts *requestOptions) {
		if maxRetries >= 0 {
			opts.maxRetries = maxRetries
		}
	}
}

// WithRequestRetryWaitMin returns a [RequestOptionFunc] that overrides the
// minimum delay between retry attempts for a single request.
//
// Parameters:
//   - retryWaitMin (time.Duration): the minimum delay. Values less than or
//     equal to 0 are ignored, keeping the client configuration.
//
// Returns:
//   - f (RequestOptionFunc): a functional option that sets the retryWaitMin field.
func WithRequestRetryWaitMin(retryWaitMin time.Duration) (f RequestOptionFunc) {
	return func(opts *requestOptions) {
		if retryWaitMin > 0 {
			opts.retryWaitMin = retryWaitMin
		}
	}
}

// WithRequestRetryWaitMax returns a [RequestOptionFunc] that overrides the
// maximum delay between retry attempts for a single request. The wait honored
// from a Retry-After response header is capped at this value as well.
//
// Parameters:
//   - retryWaitMax (time.Duration): the maximum delay. Values less than or
//     equal to 0 are ignored, keeping the client configuration.
//
// Returns:
//   - f (RequestOptionFunc): a functional option that sets the retryWaitMax field.
func WithRequestRetryWaitMax(retryWaitMax time.Duration) (f RequestOptionFunc) {
	return func(opts *requestOptions) {
		if retryWaitMax > 0 {
			opts.retryWaitMax = retryWaitMax
		}
	}
}

// WithRequestRetryBackoff returns a [RequestOptionFunc] that overrides the
// backoff strategy computing the delay for each retry attempt of a single
// request.
//
// Parameters:
//   - newRetryBackoff (func(minDelay, maxDelay time.Duration) (backoff hqgoretrierbackoff.Backoff)):
//     the backoff strategy constructor, called once per execution with the
//     normalized wait bounds. Nil is ignored, keeping the client configuration.
//
// Returns:
//   - f (RequestOptionFunc): a functional option that sets the newRetryBackoff field.
func WithRequestRetryBackoff(newRetryBackoff func(minDelay, maxDelay time.Duration) (backoff hqgoretrierbackoff.Backoff)) (f RequestOptionFunc) {
	return func(opts *requestOptions) {
		if newRetryBackoff != nil {
			opts.newRetryBackoff = newRetryBackoff
		}
	}
}

// WithRequestRetryPolicy returns a [RequestOptionFunc] that overrides the
// retry policy for a single request.
//
// Parameters:
//   - retryPolicy (RetryPolicy): the policy. Nil is ignored, keeping the
//     client configuration.
//
// Returns:
//   - f (RequestOptionFunc): a functional option that sets the retryPolicy field.
func WithRequestRetryPolicy(retryPolicy RetryPolicy) (f RequestOptionFunc) {
	return func(opts *requestOptions) {
		if retryPolicy != nil {
			opts.retryPolicy = retryPolicy
		}
	}
}

// WithRequestRateLimiter returns a [RequestOptionFunc] that overrides the
// rate limiter waited on before every attempt of a single request.
//
// Parameters:
//   - rateLimiter (RateLimiter): the limiter. Nil is ignored, keeping the
//     client configuration.
//
// Returns:
//   - f (RequestOptionFunc): a functional option that sets the rateLimiter field.
func WithRequestRateLimiter(rateLimiter RateLimiter) (f RequestOptionFunc) {
	return func(opts *requestOptions) {
		if rateLimiter != nil {
			opts.rateLimiter = rateLimiter
		}
	}
}

// WithRequestResponseReadLimit returns a [RequestOptionFunc] that overrides
// the maximum number of bytes drained from the body of a retried response of
// a single request.
//
// Parameters:
//   - responseReadLimit (int64): the drain bound in bytes. Values less than
//     or equal to 0 are ignored, keeping the client configuration.
//
// Returns:
//   - f (RequestOptionFunc): a functional option that sets the responseReadLimit field.
func WithRequestResponseReadLimit(responseReadLimit int64) (f RequestOptionFunc) {
	return func(opts *requestOptions) {
		if responseReadLimit > 0 {
			opts.responseReadLimit = responseReadLimit
		}
	}
}

// normalize falls invalid values in the client options back to their
// defaults, so a misconfigured client still behaves sanely.
func (opts *clientOptions) normalize() {
	if opts.timeout <= 0 {
		opts.timeout = DefaultTimeout
	}

	if opts.maxRetries < 0 {
		opts.maxRetries = DefaultMaxRetries
	}

	normalizeRetryBounds(&opts.retryWaitMin, &opts.retryWaitMax)

	if opts.newRetryBackoff == nil {
		opts.newRetryBackoff = hqgoretrierbackoff.ExponentialWithDecorrelatedJitter
	}

	if opts.retryPolicy == nil {
		opts.retryPolicy = DefaultRetryPolicy
	}

	if opts.responseReadLimit <= 0 {
		opts.responseReadLimit = DefaultResponseReadLimit
	}
}

// normalize falls invalid values in the request options back to their
// defaults. The request options seed from an already normalized client
// configuration, so this only guards hand-built executions.
func (opts *requestOptions) normalize() {
	if opts.maxRetries < 0 {
		opts.maxRetries = DefaultMaxRetries
	}

	normalizeRetryBounds(&opts.retryWaitMin, &opts.retryWaitMax)

	if opts.newRetryBackoff == nil {
		opts.newRetryBackoff = hqgoretrierbackoff.ExponentialWithDecorrelatedJitter
	}

	if opts.retryPolicy == nil {
		opts.retryPolicy = DefaultRetryPolicy
	}

	if opts.responseReadLimit <= 0 {
		opts.responseReadLimit = DefaultResponseReadLimit
	}
}

// normalizeRetryBounds orders the retry wait bounds: non-positive values fall
// back to their defaults, and a maximum below the minimum is raised to it.
//
// Parameters:
//   - waitMin (*time.Duration): the minimum delay to normalize in place.
//   - waitMax (*time.Duration): the maximum delay to normalize in place.
func normalizeRetryBounds(waitMin, waitMax *time.Duration) {
	if *waitMin <= 0 {
		*waitMin = DefaultRetryWaitMin
	}

	if *waitMax <= 0 {
		*waitMax = DefaultRetryWaitMax
	}

	if *waitMax < *waitMin {
		*waitMax = *waitMin
	}
}

// headerField is a single HTTP header field kept in insertion order.
//
// Fields:
//   - key (string): the header field name.
//   - value (string): the header field value.
type headerField struct {
	key   string
	value string
}
