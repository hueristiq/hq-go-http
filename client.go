package http

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"go.source.hueristiq.com/http/request"
	"go.source.hueristiq.com/retrier"
	"go.source.hueristiq.com/retrier/backoff"
	"golang.org/x/net/http2"
)

// Client defines an HTTP client with advanced features such as configurable retry policies,
// digest authentication support, and an optional fallback to HTTP/2. It encapsulates separate
// HTTP/1.x and HTTP/2 clients, along with logic for error handling and connection management.
//
// The Client structure is designed to handle transient network errors gracefully by retrying
// requests according to a user-specified policy and backoff strategy.
//
// Fields:
//   - internalHTTPClient (*http.Client): internalHTTPClient is the primary HTTP/1.x client used for executing requests.
//   - internalHTTP2Client (*http.Client): internalHTTP2Client is the fallback HTTP/2 client used when HTTP/1.x fails.
//   - onRequest (OnRequest): onRequest is an optional hook function that is executed before each HTTP request.
//     It can be used for logging, metrics collection, or modifying the request.
//   - onResponse (OnResponse): onResponse is an optional hook function that is executed after receiving an HTTP response.
//     It can be used for logging response details or performing additional validations.
//   - onError (OnError): onError is an optional hook function that is executed when all retry attempts are exhausted.
//     It allows custom error handling or modification of the final error and response.
//   - requestCounter (atomic.Uint32): requestCounter tracks the number of requests made.
//     It is used to determine when to close idle connections to prevent resource exhaustion.
//   - baseURL (string): baseURL holds the base URL to be prepended to request paths.
//   - headers (map[string]string): headers holds default headers that are applied to every request created using this client.
//   - cfg (*ClientConfiguration): cfg contains the client configuration, including timeout settings, retry policies,
//     backoff strategies, and connection management options.
type Client struct {
	internalHTTPClient  *http.Client
	internalHTTP2Client *http.Client
	onRequest           OnRequest
	onResponse          OnResponse
	onError             OnError
	requestCounter      atomic.Uint32
	baseURL             string
	headers             map[string]string
	cfg                 *ClientConfiguration
}

// Do executes an HTTP request using the Client. It applies retry logic, error handling,
// and an optional fallback to HTTP/2 if the HTTP/1.x client fails with a specific error.
// It also supports digest authentication and connection management by draining or closing idle connections.
//
// The retry loop is implemented using retrier.RetryWithData, which repeatedly attempts the request
// according to the retry policy specified in the client configuration. If hook functions (onRequest,
// onResponse, onError) are provided, they are called at appropriate stages.
//
// Parameters:
//   - req (*request.Request): A pointer to a request.Request containing the HTTP request to be executed.
//     This type wraps a standard *http.Request and ensures that the request body is reusable.
//
// Returns:
//   - res (*http.Response): A pointer to the http.Response from the executed request, if successful.
//   - err (error): An error encountered during the request, or after all retry attempts have been exhausted.
func (c *Client) Do(req *request.Request) (res *http.Response, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.cfg.Timeout)

	defer cancel()

	res, err = retrier.RetryWithData(ctx, func() (res *http.Response, err error) {
		if c.onRequest != nil {
			c.onRequest(req.Request)
		}

		res, err = c.internalHTTPClient.Do(req.Request)

		if err != nil && isErrorHTTP1Broken(err) {
			res, err = c.internalHTTP2Client.Do(req.Request)
		}

		retry, checkErr := c.cfg.RetryPolicy(req.Context(), err)

		// Call this here to maintain the behavior of logging all requests,
		// even if CheckRetry signals to stop.
		if err == nil && res != nil {
			if c.onResponse != nil {
				c.onResponse(res)
			}
		}

		if !retry {
			if checkErr != nil {
				err = checkErr
			}

			c.closeIdleConnections()

			return
		}

		if err == nil && res != nil {
			c.drainBody(res.Body)
		}

		return
	},
		retrier.WithRetryMax(c.cfg.RetryMax),
		retrier.WithRetryWaitMin(c.cfg.RetryWaitMin),
		retrier.WithRetryWaitMax(c.cfg.RetryWaitMax),
		retrier.WithRetryBackoff(c.cfg.RetryBackoff),
	)

	if c.onError != nil {
		c.closeIdleConnections()

		res, err = c.onError(res, err, c.cfg.RetryMax)

		return
	}

	if err != nil {
		if res != nil {
			res.Body.Close()

			err = fmt.Errorf("%s %s giving up after %d attempts: response status %d: %w", req.Method, req.URL, c.cfg.RetryMax, res.StatusCode, err)
		} else {
			err = fmt.Errorf("%s %s giving up after %d attempts: %w", req.Method, req.URL, c.cfg.RetryMax, err)
		}

		c.closeIdleConnections()
	}

	return
}

// setKillIdleConnections examines the internal HTTP client's transport configuration and
// determines if idle connections should be forcefully closed. The decision is based on the
// transport's DisableKeepAlives setting and MaxConnsPerHost value. If these conditions are met,
// the client's configuration is updated accordingly.
func (c *Client) setKillIdleConnections() {
	if c.internalHTTPClient != nil || !c.cfg.KillIdleConn {
		if b, ok := c.internalHTTPClient.Transport.(*http.Transport); ok {
			c.cfg.KillIdleConn = b.DisableKeepAlives || b.MaxConnsPerHost < 0
		}
	}
}

// closeIdleConnections manages connection reuse by tracking the number of requests made.
// If the internal request counter exceeds a threshold (100 requests), the counter is reset and
// the HTTP client's idle connections are forcefully closed. This helps prevent resource exhaustion
// in long-running applications.
func (c *Client) closeIdleConnections() {
	if c.cfg.KillIdleConn {
		if c.requestCounter.Load() < 100 {
			c.requestCounter.Add(1)
		} else {
			c.requestCounter.Store(0)
			c.internalHTTPClient.CloseIdleConnections()
		}
	}
}

// drainBody reads and discards up to a configured number of bytes from the provided response body.
// This ensures that the underlying connection can be safely reused. The method also closes the body,
// preventing resource leaks.
//
// Arguments:
//   - body (io.ReadCloser): The response body that needs to be drained.
func (c *Client) drainBody(body io.ReadCloser) {
	defer body.Close()

	_, _ = io.Copy(io.Discard, io.LimitReader(body, c.cfg.RespReadLimit))
}

// WithBaseURL sets the base URL for the Client. This base URL will be prepended to paths when
// building new requests.
//
// Arguments:
//   - baseURL (string): The base URL to be used for all requests.
func (c *Client) WithBaseURL(baseURL string) {
	c.baseURL = baseURL
}

// WithHeaders sets default headers for the Client. These headers are applied to every request
// built using this client.
//
// Arguments:
//   - headers (map[string]string): A map containing header keys and their corresponding values.
func (c *Client) WithHeaders(headers map[string]string) {
	c.headers = headers
}

// WithOnRequest sets a hook function that is called before each HTTP request is executed.
// This hook receives the http.Request object, allowing for logging or modification.
//
// Arguments:
//   - onRequest (OnRequest): The hook function to be called prior to request execution.
func (c *Client) WithOnRequest(onRequest OnRequest) {
	c.onRequest = onRequest
}

// WithOnResponse sets a hook function that is called after receiving an HTTP response.
// This hook receives the http.Response object, allowing for logging or further processing.
//
// Arguments:
//   - onResponse (OnResponse): The hook function to be called after a response is received.
func (c *Client) WithOnResponse(onResponse OnResponse) {
	c.onResponse = onResponse
}

// WithOnError sets a hook function that is called when all retry attempts are exhausted.
// The hook receives the final http.Response (if any), the encountered error, and the total number of attempts.
//
// Arguments:
//   - onError (OnError): The hook function for custom error handling after retries fail.
func (c *Client) WithOnError(onError OnError) {
	c.onError = onError
}

// Request returns a new RequestBuilder instance configured with the Client's base URL and default headers.
// The RequestBuilder provides a fluent API for constructing and sending HTTP requests.
//
// Returns:
//   - builder (*RequestBuilder): A pointer to a newly created RequestBuilder instance.
func (c *Client) Request() (builder *RequestBuilder) {
	builder = &RequestBuilder{
		client: c,
		_URL:   c.baseURL,
		header: make(http.Header),
	}

	for k, v := range c.headers {
		builder.header.Set(k, v)
	}

	return
}

// OnRequest is a hook function type that is called before an HTTP request is executed.
// The provided http.Request can be inspected or modified, and is typically used for logging.
type OnRequest func(req *http.Request)

// OnResponse is a hook function type that is called after an HTTP response is received.
// The provided http.Response can be inspected, logged, or modified. Note that reading or closing
// the response body in this hook may affect the response returned by Do().
type OnResponse func(res *http.Response)

// OnError is a hook function type that is called when all retry attempts are exhausted.
// It receives the final *http.Response (if any), the encountered error, and the total number of attempts.
// The function should close the response body if necessary and can modify the error or response before returning.
//
// Arguments:
//   - res (*http.Response): The final response received (may be nil).
//   - err (error): The error encountered after exhausting retries.
//   - tries (int): The total number of attempts made.
//
// Returns:
//   - ress (*http.Response): A potentially modified response to be returned.
//   - errr (error): A potentially modified error to be returned.
type OnError func(res *http.Response, err error, tries int) (ress *http.Response, errr error)

// RequestBuilder provides a fluent API for constructing and sending HTTP requests using the Client.
// It encapsulates request details such as method, URL, body, and headers.
//
// Fields:
//   - client (*Client): client is the Client that will execute the request.
//   - method (string): method holds the HTTP method (e.g., GET, POST) for the request.
//   - _URL (string): _URL holds the full URL for the request, starting with the base URL.
//   - body (interface{}): body is the request body, which can be of any type supported by the request package.
//   - header (http.Header):  header contains the HTTP headers to be sent with the request.
type RequestBuilder struct {
	client *Client

	method string
	_URL   string
	body   interface{}
	header http.Header
}

// Method sets the HTTP method for the request and returns the updated RequestBuilder.
//
// Arguments:
//   - m (string): The HTTP method to use (e.g., "GET", "POST").
//
// Returns:
//   - builder (*RequestBuilder): The updated RequestBuilder with the method set.
func (b *RequestBuilder) Method(m string) (builder *RequestBuilder) {
	b.method = m

	return b
}

// URL appends the provided URL segment to the builder's current URL and returns the updated RequestBuilder.
//
// Arguments:
//   - u (string): A URL segment or path to append.
//
// Returns:
//   - builder (*RequestBuilder): The updated RequestBuilder with the new URL.
func (b *RequestBuilder) URL(u string) (builder *RequestBuilder) {
	b._URL += u

	return b
}

// Body sets the request body for the RequestBuilder and returns the updated builder.
//
// Arguments:
//   - body (interface{}): The request body, which can be of any type supported by the request package.
//
// Returns:
//   - builder (*RequestBuilder): The updated RequestBuilder with the body set.
func (b *RequestBuilder) Body(body interface{}) (builder *RequestBuilder) {
	b.body = body

	return b
}

// AddHeader adds the specified header key and value to the request and returns the updated RequestBuilder.
//
// Arguments:
//   - key (string): The header key.
//   - value (string): The header value.
//
// Returns:
//   - builder (*RequestBuilder): The updated RequestBuilder with the header added.
func (b *RequestBuilder) AddHeader(key, value string) (builder *RequestBuilder) {
	b.header.Add(key, value)

	return b
}

// SetHeader sets the specified header key to the given value, replacing any existing values,
// and returns the updated RequestBuilder.
//
// Arguments:
//   - key (string): The header key.
//   - value (string): The header value.
//
// Returns:
//   - builder (*RequestBuilder): The updated RequestBuilder with the header set.
func (b *RequestBuilder) SetHeader(key, value string) (builder *RequestBuilder) {
	b.header.Set(key, value)

	return b
}

// Build constructs a request.Request using the builder's method, URL, and body, and applies the headers.
// It returns the constructed request or an error if the construction fails.
//
// Returns:
//   - req (*request.Request): A pointer to the constructed request.Request.
//   - err (error): An error if building the request fails.
func (b *RequestBuilder) Build() (req *request.Request, err error) {
	req, err = request.New(b.method, b._URL, b.body)
	if err != nil {
		return
	}

	req.Request.Header = b.header

	return
}

// Send builds and sends the HTTP request using the associated Client's Do method.
// It returns the resulting http.Response or an error if the request fails.
//
// Returns:
//   - res (*http.Response): A pointer to the http.Response received.
//   - err (error): An error if sending the request or receiving the response fails.
func (b *RequestBuilder) Send() (res *http.Response, err error) {
	var req *request.Request

	req, err = b.Build()
	if err != nil {
		return
	}

	res, err = b.client.Do(req)

	return
}

// ClientConfiguration encapsulates various configuration options for the Client.
// These settings control timeouts, retry behavior, backoff strategies, connection management,
// and limits on reading response bodies.
//
// Fields:
//   - HTTPClient: Optionally provides a custom HTTP client to use for requests.
//   - Timeout: The maximum duration allowed for an HTTP request.
//   - RetryPolicy: A function that determines whether a request should be retried based on errors.
//   - RetryMax: The maximum number of retry attempts.
//   - RetryWaitMin: The minimum wait duration between retry attempts.
//   - RetryWaitMax: The maximum wait duration between retry attempts.
//   - RetryBackoff: A backoff strategy used to compute delays between retries.
//   - KillIdleConn: If true, idle connections will be periodically closed to free resources.
//   - RespReadLimit: The maximum number of bytes to read from a response body when draining.
type ClientConfiguration struct {
	HTTPClient *http.Client

	Timeout time.Duration

	RetryPolicy  RetryPolicy
	RetryMax     int
	RetryWaitMin time.Duration
	RetryWaitMax time.Duration
	RetryBackoff backoff.Backoff

	KillIdleConn  bool
	RespReadLimit int64
}

// RetryPolicy defines a function type that determines whether an HTTP request should be retried.
// It is invoked after each request attempt with the request's context and the encountered error.
// If the function returns false, no further retries are attempted. If an error is returned,
// that error overrides the original error.
//
// Arguments:
//   - ctx (context.Context): The request's context, which may contain cancellation signals or deadlines.
//   - err (error): The error encountered during the request (can be nil if the request succeeded).
//
// Returns:
//   - retry (bool): True if the request should be retried; false otherwise.
//   - errr (error): An optional error that, if non-nil, overrides the original error.
type RetryPolicy func(ctx context.Context, err error) (retry bool, errr error)

var (
	// redirectsErrorRegex matches error strings indicating that the maximum number of redirects was exceeded.
	// This is used to prevent retrying requests that have already failed due to excessive redirects.
	redirectsErrorRegex = regexp.MustCompile(`stopped after \d+ redirects\z`)

	// schemeErrorRegex matches error strings indicating an unsupported protocol scheme.
	// This helps detect errors that should not be retried.
	schemeErrorRegex = regexp.MustCompile(`unsupported protocol scheme`)
)

// DefaultSingleClientConfiguration defines a default configuration for a single-use HTTP client.
// It is intended for standard scenarios where connection pooling is acceptable.
var DefaultSingleClientConfiguration = &ClientConfiguration{
	Timeout:       30 * time.Second,
	RetryMax:      5,
	RetryWaitMin:  1 * time.Second,
	RetryWaitMax:  30 * time.Second,
	KillIdleConn:  false,
	RespReadLimit: 4096,
}

// DefaultSprayingClientConfiguration defines a default configuration for scenarios such as host spraying,
// where killing idle connections is desirable to reduce resource usage.
var DefaultSprayingClientConfiguration = &ClientConfiguration{
	Timeout:       30 * time.Second,
	RetryMax:      5,
	RetryWaitMin:  1 * time.Second,
	RetryWaitMax:  30 * time.Second,
	KillIdleConn:  true,
	RespReadLimit: 4096,
}

// DefaultClient is a package-level default Client instance, initialized during package startup.
var DefaultClient *Client

// init is executed when the package is initialized. It creates a default Client instance using
// DefaultSingleClientConfiguration. Any errors during client creation are ignored.
func init() {
	DefaultClient, _ = NewClient(DefaultSingleClientConfiguration)
}

// NewClient creates and configures a new Client based on the provided ClientConfiguration.
// It sets up both HTTP/1.x and HTTP/2 clients, configures retry policies and backoff strategies,
// and applies connection management settings.
//
// Arguments:
//   - cfg (*ClientConfiguration): A pointer to a ClientConfiguration containing desired settings.
//
// Returns:
//   - client (*Client): A pointer to the newly created Client.
//   - err (error): An error if client initialization fails.
func NewClient(cfg *ClientConfiguration) (client *Client, err error) {
	client = &Client{}

	client.internalHTTPClient = DefaultPooledClient()

	if cfg.KillIdleConn {
		client.internalHTTPClient = DefaultHTTPClient()
	}

	if cfg.HTTPClient != nil {
		client.internalHTTPClient = cfg.HTTPClient
	}

	client.internalHTTP2Client = DefaultHTTPClient()

	internalHTTP2ClientTransport, ok := client.internalHTTP2Client.Transport.(*http.Transport)
	if !ok {
		return
	}

	if err = http2.ConfigureTransport(internalHTTP2ClientTransport); err != nil {
		return
	}

	if cfg.Timeout > 0 {
		client.internalHTTPClient.Timeout = cfg.Timeout
		client.internalHTTP2Client.Timeout = cfg.Timeout
	}

	if cfg.RetryPolicy == nil {
		cfg.RetryPolicy = DefaultRetryPolicy()
	}

	if cfg.RetryBackoff == nil {
		cfg.RetryBackoff = backoff.Exponential()
	}

	client.cfg = cfg

	client.setKillIdleConnections()

	return
}

// DefaultHTTPTransport returns a new http.Transport configured to disable idle connections
// and keep-alives. It is derived from DefaultHTTPPooledTransport but modified to avoid
// connection reuse, making it suitable for transient requests.
//
// Returns:
//   - transport (*http.Transport): A pointer to a newly configured http.Transport.
func DefaultHTTPTransport() (transport *http.Transport) {
	transport = DefaultHTTPPooledTransport()

	transport.DisableKeepAlives = true
	transport.MaxIdleConnsPerHost = -1

	return
}

// DefaultHTTPPooledTransport returns a new http.Transport configured for connection pooling.
// It sets various parameters such as timeouts, keep-alives, and idle connection limits to
// optimize connection reuse for repeated requests to the same host.
//
// Warning: This transport is intended for long-lived clients; using it for transient clients
// may lead to file descriptor leaks.
//
// Returns:
//   - transport (*http.Transport): A pointer to a newly configured pooled http.Transport.
func DefaultHTTPPooledTransport() (transport *http.Transport) {
	transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
	}

	return
}

// DefaultHTTPClient returns a new http.Client configured with a non-shared transport.
// Idle connections and keep-alives are disabled, making it suitable for transient requests
// where connection reuse is not desired.
//
// Returns:
//   - client (*http.Client): A pointer to a newly created http.Client.
func DefaultHTTPClient() (client *http.Client) {
	client = &http.Client{
		Transport: DefaultHTTPTransport(),
	}

	return
}

// DefaultPooledClient returns a new http.Client configured with a shared transport that supports
// connection pooling. It is optimized for clients that make repeated requests to the same host,
// allowing efficient reuse of TCP connections.
//
// Warning: This client should not be used for short-lived operations, as it may leak file descriptors.
//
// Returns:
//   - client (*http.Client): A pointer to a newly created pooled http.Client.
func DefaultPooledClient() (client *http.Client) {
	client = &http.Client{
		Transport: DefaultHTTPPooledTransport(),
	}

	return
}

// DefaultRetryPolicy returns a default RetryPolicy function that determines if a request should
// be retried based on whether the encountered error is recoverable. It delegates the decision
// to the isErrorRecoverable function.
//
// Returns:
//   - A RetryPolicy function.
func DefaultRetryPolicy() func(ctx context.Context, err error) (retry bool, errr error) {
	return isErrorRecoverable
}

// HostSprayRetryPolicy returns a RetryPolicy function tailored for scenarios where distributed
// (or host-spraying) requests are made. It currently delegates to the same isErrorRecoverable
// function as the default policy.
//
// Returns:
//   - A RetryPolicy function.
func HostSprayRetryPolicy() func(ctx context.Context, err error) (retry bool, errr error) {
	return isErrorRecoverable
}

// Request is a helper function that returns a new RequestBuilder instance from the DefaultClient.
// This provides a convenient way to start building a request using the package-level default client.
//
// Returns:
//   - builder (*RequestBuilder): A pointer to a new RequestBuilder.
func Request() (builder *RequestBuilder) {
	builder = DefaultClient.Request()

	return
}

// isErrorHTTP1Broken checks whether the provided error indicates a specific HTTP/1.x transport issue.
// In particular, it looks for errors that mention a broken HTTP/1.x connection due to a malformed HTTP version,
// which suggests that an HTTP/2 connection might be more appropriate.
//
// Arguments:
//   - err (error): The error to inspect.
//
// Returns:
//   - isErrorHTTP1Broken (bool): True if the error indicates a broken HTTP/1.x transport connection; false otherwise.
func isErrorHTTP1Broken(err error) (isErrorHTTP1Broken bool) {
	isErrorHTTP1Broken = err != nil && strings.Contains(err.Error(), "net/http: HTTP/1.x transport connection broken: malformed HTTP version \"HTTP/2\"")

	return
}

// isErrorRecoverable determines whether an error encountered during an HTTP request is recoverable,
// meaning that the request may be retried. It examines the request context and the error details,
// filtering out errors such as context cancellations, excessive redirects, unsupported protocol schemes,
// or TLS certificate verification failures.
//
// Arguments:
//   - ctx (context.Context): The request's context, which may contain cancellation signals or deadlines.
//   - err (error): The error encountered during the HTTP request.
//
// Returns:
//   - recoverable (bool): True if the error is considered recoverable and the request may be retried; false otherwise.
//   - errr (error): An error value if the context signals cancellation or the deadline is exceeded, or if a non-retryable error is detected.
func isErrorRecoverable(ctx context.Context, err error) (recoverable bool, errr error) {
	// Do not retry if the context has been canceled or the deadline has been exceeded
	if ctx.Err() != nil {
		errr = ctx.Err()

		return
	}

	var URLError *url.Error

	if err != nil && errors.As(err, &URLError) {
		// Do not retry if the error was caused by exceeding the maximum number of redirects
		if redirectsErrorRegex.MatchString(err.Error()) {
			errr = err

			return
		}

		// Do not retry if the error was caused by an unsupported protocol scheme
		if schemeErrorRegex.MatchString(err.Error()) {
			errr = err

			return
		}

		// Do not retry if the error was caused by a TLS certificate verification failure
		var UnknownAuthorityError x509.UnknownAuthorityError

		if errors.As(err, &UnknownAuthorityError) {
			errr = err

			return
		}
	}

	if err != nil {
		recoverable = true

		return
	}

	return
}
