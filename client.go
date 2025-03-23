package http

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"go.source.hueristiq.com/http/method"
	"go.source.hueristiq.com/http/request"
	"go.source.hueristiq.com/retrier"
	"go.source.hueristiq.com/retrier/backoff"
	"golang.org/x/net/http2"
)

// Client is the primary structure used to perform HTTP requests. It manages
// global and request-specific configurations, as well as separate underlying HTTP
// clients for HTTP/1.x and HTTP/2. It also tracks the number of requests made to
// trigger periodic connection resets, which is useful to prevent resource exhaustion
// in long-running applications.
//
// Fields:
//   - cfg (*ClientConfiguration): Global configuration settings that apply to all requests.
//   - internalHTTPClient (*http.Client): The underlying HTTP client instance used for HTTP/1.x requests.
//   - internalHTTP2Client (*http.Client): A fallback HTTP client instance for HTTP/2 requests.
//   - rc (atomic.Uint32): An atomic counter tracking the number of requests executed.
//     When a preset threshold is reached, idle connections are closed to free resources.
type Client struct {
	cfg                 *ClientConfiguration
	internalHTTPClient  *http.Client
	internalHTTP2Client *http.Client
	rc                  atomic.Uint32
}

// Do executes an HTTP request using the Client. It applies retry logic, error handling,
// and an optional fallback to HTTP/2 if the HTTP/1.x client fails due to a specific error.
// This method supports automatic connection draining (to enable connection reuse) and
// digest authentication if configured. It is the core method that all higher-level HTTP
// convenience methods (e.g. Get, Post) ultimately call.
//
// The retry loop is implemented using retrier.RetryWithData, which repeatedly attempts
// the request based on the retry policy defined in the client's configuration. If hook functions
// (such as onRequest, onResponse, or onError) are provided in a different version, they should be
// called at appropriate stages (not shown in this snippet).
//
// Parameters:
//   - req (*request.Request): A pointer to a request.Request instance that wraps an *http.Request.
//     This wrapper ensures that the request body is reusable between retries.
//
// Returns:
//   - res (*http.Response): A pointer to the http.Response if the request was eventually successful.
//   - err (error): An error if the request ultimately failed after all retry attempts.
func (c *Client) Do(req *request.Request, cfg *RequestConfiguration) (res *http.Response, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.cfg.Timeout)

	defer cancel()

	res, err = retrier.RetryWithData(ctx, func() (res *http.Response, err error) {
		res, err = c.internalHTTPClient.Do(req.Request)

		if err != nil && isErrorHTTP1Broken(err) {
			res, err = c.internalHTTP2Client.Do(req.Request)
		}

		retry, retryPolicyError := cfg.RetryPolicy(req.Context(), err)

		if !retry {
			if retryPolicyError != nil {
				err = retryPolicyError
			}

			c.closeIdleConnections()

			return
		}

		if err == nil && res != nil {
			drainBody(res.Body, cfg.RespReadLimit)
		}

		return
	},
		retrier.WithRetryMax(cfg.RetryMax),
		retrier.WithRetryWaitMin(cfg.RetryWaitMin),
		retrier.WithRetryWaitMax(cfg.RetryWaitMax),
		retrier.WithRetryBackoff(cfg.RetryBackoff),
	)
	if err != nil {
		if res != nil {
			res.Body.Close()

			err = fmt.Errorf("%s %s giving up after %d attempts: response status %d: %w", req.Method, req.URL, cfg.RetryMax, res.StatusCode, err)
		} else {
			err = fmt.Errorf("%s %s giving up after %d attempts: %w", req.Method, req.URL, cfg.RetryMax, err)
		}

		c.closeIdleConnections()
	}

	return
}

// closeIdleConnections is responsible for managing connection reuse by monitoring the number
// of requests made. When the counter reaches a threshold (100 requests), the counter is reset,
// and the client's idle connections are forcefully closed. This helps prevent potential resource leaks
// in applications with a high volume of requests.
//
// Note: The behavior is controlled by the ClientConfiguration.CloseIdleConnections flag.
func (c *Client) closeIdleConnections() {
	if c.cfg.CloseIdleConnections {
		if c.rc.Load() < 100 {
			c.rc.Add(1)
		} else {
			c.rc.Store(0)

			c.internalHTTPClient.CloseIdleConnections()
		}
	}
}

// Request builds and executes an HTTP request based on merged client and request-specific configurations.
// It prepares the request by setting the method, URL, headers, query parameters, and body. It then calls Do
// to execute the prepared request.
//
// Parameters:
//   - configurations (...*RequestConfiguration): One or more pointers to RequestConfiguration that provide
//     request-specific overrides. These are merged with the global client configuration.
//
// Returns:
//   - res (*http.Response): The resulting http.Response from the executed request if successful.
//   - err (error): An error encountered during configuration, request construction, or execution.
func (c *Client) Request(configurations ...*RequestConfiguration) (res *http.Response, err error) {
	var cfg *RequestConfiguration

	cfg, err = c.getRequestConfiguration(configurations...)
	if err != nil {
		return
	}

	var req *request.Request

	req, err = request.New(cfg.Method, cfg.URL, cfg.Body)
	if err != nil {
		return
	}

	for key, value := range cfg.Headers {
		req.Header.Set(key, value)
	}

	res, err = c.Do(req, cfg)

	return
}

// getRequestConfiguration merges one or more RequestConfiguration objects with the client's default configuration.
// It applies overrides for HTTP method, base URL, path URL, query parameters, headers, body, timeout, and retry settings.
// If a BaseURL is specified, it will be combined with the relative URL using url.JoinPath. Additionally, query
// parameters are appended to the URL's query string.
//
// Parameters:
//   - configurations (...*RequestConfiguration): Variadic list of pointers to RequestConfiguration instances.
//
// Returns:
//   - err (error): An error if merging configurations or constructing the final URL fails.
func (c *Client) getRequestConfiguration(configurations ...*RequestConfiguration) (rc *RequestConfiguration, err error) {
	rc = &RequestConfiguration{
		Method:        c.cfg.Method,
		BaseURL:       c.cfg.BaseURL,
		URL:           c.cfg.URL,
		Params:        c.cfg.Params,
		Headers:       c.cfg.Headers,
		Body:          c.cfg.Body,
		RespReadLimit: c.cfg.RespReadLimit,
		RetryPolicy:   c.cfg.RetryPolicy,
		RetryMax:      c.cfg.RetryMax,
		RetryWaitMin:  c.cfg.RetryWaitMin,
		RetryWaitMax:  c.cfg.RetryWaitMax,
		RetryBackoff:  c.cfg.RetryBackoff,
	}

	for _, configuration := range configurations {
		if configuration.Method != "" {
			rc.Method = configuration.Method
		}

		if configuration.BaseURL != "" {
			rc.BaseURL = configuration.BaseURL
		}

		if configuration.URL != "" {
			rc.URL = configuration.URL
		}

		if configuration.Params != nil {
			for k, v := range configuration.Params {
				rc.Params[k] = v
			}
		}

		if configuration.Headers != nil {
			for k, v := range configuration.Headers {
				rc.Headers[k] = v
			}
		}

		if configuration.Body != "" {
			rc.Body = configuration.Body
		}

		if configuration.RespReadLimit > 0 {
			rc.RespReadLimit = configuration.RespReadLimit
		}

		if configuration.RetryPolicy != nil {
			rc.RetryPolicy = configuration.RetryPolicy
		}

		if configuration.RetryMax > 0 {
			rc.RetryMax = configuration.RetryMax
		}

		if configuration.RetryWaitMin > 0 {
			rc.RetryWaitMin = configuration.RetryWaitMin
		}

		if configuration.RetryWaitMax > 0 {
			rc.RetryWaitMax = configuration.RetryWaitMax
		}

		if configuration.RetryBackoff != nil {
			rc.RetryBackoff = configuration.RetryBackoff
		}
	}

	if rc.BaseURL != "" {
		rc.URL, err = url.JoinPath(rc.BaseURL, rc.URL)
		if err != nil {
			return
		}
	}

	if len(rc.Params) > 0 {
		var parsed *url.URL

		parsed, err = url.Parse(rc.URL)
		if err != nil {
			return
		}

		q := parsed.Query()

		for k, v := range rc.Params {
			q.Add(k, v)
		}

		parsed.RawQuery = q.Encode()

		rc.URL = parsed.String()
	}

	return
}

// Get is a convenience method that performs an HTTP GET request.
// It sets the method to GET and calls Request to execute the call.
//
// Parameters:
//   - URL (string): The target URL for the GET request.
//   - configurations (...*RequestConfiguration): Optional overrides for the request configuration.
//
// Returns:
//   - res (*http.Response): The response received from the server.
//   - err (error): An error if the request fails.
func (c *Client) Get(URL string, configurations ...*RequestConfiguration) (res *http.Response, err error) {
	configurations = append(configurations, &RequestConfiguration{
		Method: method.GET.String(),
		URL:    URL,
	})

	res, err = c.Request(configurations...)

	return
}

// Head is a convenience method that performs an HTTP HEAD request.
// It sets the method to HEAD and calls Request to execute the call.
//
// Parameters:
//   - URL (string): The target URL for the HEAD request.
//   - configurations (...*RequestConfiguration): Optional overrides for the request configuration.
//
// Returns:
//   - res (*http.Response): The response received from the server.
//   - err (error): An error if the request fails.
func (c *Client) Head(URL string, configurations ...*RequestConfiguration) (res *http.Response, err error) {
	configurations = append(configurations, &RequestConfiguration{
		Method: method.HEAD.String(),
		URL:    URL,
	})

	res, err = c.Request(configurations...)

	return
}

// Put is a convenience method for performing an HTTP PUT request with a provided body.
// It sets the method to PUT and calls Request to execute the call.
//
// Parameters:
//   - URL (string): The target URL for the PUT request.
//   - body (interface{}): The payload to include in the PUT request.
//   - configurations (...*RequestConfiguration): Optional overrides for the request configuration.
//
// Returns:
//   - res (*http.Response): The response received from the server.
//   - err (error): An error if the request fails.
func (c *Client) Put(URL string, body interface{}, configurations ...*RequestConfiguration) (res *http.Response, err error) {
	configurations = append(configurations, &RequestConfiguration{
		Method: method.PUT.String(),
		URL:    URL,
		Body:   body,
	})

	res, err = c.Request(configurations...)

	return
}

// Delete is a convenience method for performing an HTTP DELETE request.
// It sets the method to DELETE and calls Request to execute the call.
//
// Parameters:
//   - URL (string): The target URL for the DELETE request.
//   - configurations (...*RequestConfiguration): Optional overrides for the request configuration.
//
// Returns:
//   - res (*http.Response): The response received from the server.
//   - err (error): An error if the request fails.
func (c *Client) Delete(URL string, configurations ...*RequestConfiguration) (res *http.Response, err error) {
	configurations = append(configurations, &RequestConfiguration{
		Method: method.DELETE.String(),
		URL:    URL,
	})

	res, err = c.Request(configurations...)

	return
}

// Post is a convenience method for performing an HTTP POST request with a provided body.
// It sets the method to POST and calls Request to execute the call.
//
// Parameters:
//   - URL (string): The target URL for the POST request.
//   - body (interface{}): The payload to include in the POST request.
//   - configurations (...*RequestConfiguration): Optional overrides for the request configuration.
//
// Returns:
//   - res (*http.Response): The response received from the server.
//   - err (error): An error if the request fails.
func (c *Client) Post(URL string, body interface{}, configurations ...*RequestConfiguration) (res *http.Response, err error) {
	configurations = append(configurations, &RequestConfiguration{
		Method: method.POST.String(),
		URL:    URL,
		Body:   body,
	})

	res, err = c.Request(configurations...)

	return
}

// Options is a convenience method for performing an HTTP OPTIONS request.
// It sets the method to OPTIONS and calls Request to execute the call.
//
// Parameters:
//   - URL (string): The target URL for the OPTIONS request.
//   - configurations (...*RequestConfiguration): Optional overrides for the request configuration.
//
// Returns:
//   - res (*http.Response): The response received from the server.
//   - err (error): An error if the request fails.
func (c *Client) Options(URL string, configurations ...*RequestConfiguration) (res *http.Response, err error) {
	configurations = append(configurations, &RequestConfiguration{
		Method: method.OPTIONS.String(),
		URL:    URL,
	})

	res, err = c.Request(configurations...)

	return
}

// ClientConfiguration encapsulates various configuration options for the Client.
// These settings control HTTP client behavior, including timeouts, retry logic,
// backoff strategies, connection management, and default request values.
//
// Fields:
//   - HTTPClient (*http.Client): An optional custom HTTP client that can be provided.
//     If nil, a default pooled or non-pooled client is used depending on CloseIdleConnections.
//   - Timeout (time.Duration): The maximum duration allowed for each HTTP request.
//   - CloseIdleConnections (bool): Determines whether idle connections should be periodically closed.
//     Setting this to true is recommended for high-volume transient request scenarios.
//   - Method (string): Default HTTP method to use (e.g., GET, POST) if not overridden per request.
//   - BaseURL (string): A base URL that is prefixed to all request URLs.
//   - URL (string): Default URL path (which can be combined with BaseURL).
//   - Params (map[string]string): Default query parameters that are appended to every request.
//   - Headers (map[string]string): Default HTTP headers to include in every request.
//   - Body (interface{}): Default request body; can be overridden by individual requests.
//   - RespReadLimit (int64): The maximum number of bytes to drain from a response body to allow connection reuse.
//   - RetryPolicy (RetryPolicy): A function to determine whether a request should be retried.
//   - RetryMax (int): The maximum number of retry attempts before giving up.
//   - RetryWaitMin (time.Duration): The minimum duration to wait between retries.
//   - RetryWaitMax (time.Duration): The maximum duration to wait between retries.
//   - RetryBackoff (backoff.Backoff): The backoff strategy used to calculate wait times between retries.
type ClientConfiguration struct {
	Client               *http.Client
	Timeout              time.Duration
	CloseIdleConnections bool

	Method        string
	BaseURL       string
	URL           string
	Params        map[string]string
	Headers       map[string]string
	Body          interface{}
	RespReadLimit int64
	RetryPolicy   RetryPolicy
	RetryMax      int
	RetryWaitMin  time.Duration
	RetryWaitMax  time.Duration
	RetryBackoff  backoff.Backoff
}

func (c *ClientConfiguration) Configuration() (configuration *ClientConfiguration) {
	configuration = c

	if configuration.Params == nil {
		configuration.Params = make(map[string]string)
	}

	if configuration.Headers == nil {
		configuration.Headers = make(map[string]string)
	}

	if configuration.RetryPolicy == nil {
		configuration.RetryPolicy = DefaultRetryPolicy()
	}

	if configuration.RetryBackoff == nil {
		configuration.RetryBackoff = backoff.Exponential()
	}

	return
}

// RetryPolicy defines a function type that determines if an HTTP request should be retried.
// It is invoked after each request attempt with the request's context and any encountered error.
// If the function returns false, no further retries are attempted. Additionally, a non-nil error
// return value overrides the original error, terminating further retry attempts.
//
// Parameters:
//   - ctx (context.Context): The context for the request, containing cancellation signals and deadlines.
//   - err (error): The error encountered during the HTTP request, or nil if the request succeeded.
//
// Returns:
//   - retry (bool): True if the request should be retried; false if it should not.
//   - errr (error): An error to override the original error, typically when a non-retryable error is encountered.
type RetryPolicy func(ctx context.Context, err error) (retry bool, errr error)

// RequestConfiguration holds settings specific to an individual HTTP request.
// These settings override the global ClientConfiguration on a per-request basis.
// They include the HTTP method, URL (and optionally BaseURL), query parameters, headers,
// body content, timeout, and retry-related options.
//
// Fields:
//   - Method (string): The HTTP method for the request (e.g., GET, POST, PUT, etc.).
//   - BaseURL (string): An optional base URL to be prefixed to the URL, overriding the global BaseURL.
//   - URL (string): The target URL or path for the request.
//   - Params (map[string]string): Query parameters to append to the URL.
//   - Headers (map[string]string): HTTP headers to include with the request.
//   - Body (interface{}): The request body, if applicable.
//   - RespReadLimit (int64): The maximum number of bytes to read from a response body when draining.
//   - RetryPolicy (RetryPolicy): A function to determine retry behavior for this specific request.
//   - RetryMax (int): The maximum number of retry attempts for this request.
//   - RetryWaitMin (time.Duration): The minimum duration to wait between retries.
//   - RetryWaitMax (time.Duration): The maximum duration to wait between retries.
//   - RetryBackoff (backoff.Backoff): The strategy used to calculate backoff delays between retries.
type RequestConfiguration struct {
	Method        string
	BaseURL       string
	URL           string
	Params        map[string]string
	Headers       map[string]string
	Body          interface{}
	RespReadLimit int64
	RetryPolicy   RetryPolicy
	RetryMax      int
	RetryWaitMin  time.Duration
	RetryWaitMax  time.Duration
	RetryBackoff  backoff.Backoff
}

var (
	// redirectsErrorRegex matches error strings that indicate the maximum number of redirects was exceeded.
	// It is used to avoid retrying requests that have failed due to too many redirects.
	redirectsErrorRegex = regexp.MustCompile(`stopped after \d+ redirects\z`)

	// schemeErrorRegex matches error strings indicating an unsupported protocol scheme.
	// This helps in identifying errors that should not be retried.
	schemeErrorRegex = regexp.MustCompile(`unsupported protocol scheme`)
)

// NewClient creates and configures a new Client based on the provided ClientConfiguration.
// It initializes both HTTP/1.x and HTTP/2 clients, configures retry policies and backoff strategies,
// and sets up connection management settings.
//
// Parameters:
//   - cfg (*ClientConfiguration): A pointer to a ClientConfiguration instance containing desired settings.
//
// Returns:
//   - client (*Client): A pointer to the newly created Client.
//   - err (error): An error if client initialization fails.
func NewClient(cfg *ClientConfiguration) (client *Client, err error) {
	client = &Client{
		cfg: cfg.Configuration(),
	}

	internalHTTPClient := DefaultHTTPPooledClient()

	if client.cfg.CloseIdleConnections {
		internalHTTPClient = DefaultHTTPClient()
	}

	if client.cfg.Client != nil {
		internalHTTPClient = client.cfg.Client
	}

	if internalHTTPClient != nil || !client.cfg.CloseIdleConnections {
		if b, ok := internalHTTPClient.Transport.(*http.Transport); ok {
			client.cfg.CloseIdleConnections = b.DisableKeepAlives || b.MaxConnsPerHost < 0
		}
	}

	client.internalHTTPClient = internalHTTPClient

	internalHTTP2Client := DefaultHTTPClient()

	internalHTTP2Transport, ok := internalHTTP2Client.Transport.(*http.Transport)
	if !ok {
		return
	}

	if err = http2.ConfigureTransport(internalHTTP2Transport); err != nil {
		return
	}

	client.internalHTTP2Client = internalHTTP2Client

	if client.cfg.Timeout > 0 {
		client.internalHTTPClient.Timeout = client.cfg.Timeout
		client.internalHTTP2Client.Timeout = client.cfg.Timeout
	}

	return
}

// drainBody reads and discards up to 'limit' bytes from the provided response body.
// Draining the body ensures that the underlying TCP connection can be reused for future requests.
// After reading, the response body is closed to prevent resource leaks.
//
// Parameters:
//   - body (io.ReadCloser): The response body to drain.
//   - limit (int64): The maximum number of bytes to read from the response body.
func drainBody(body io.ReadCloser, limit int64) {
	defer body.Close()

	_, _ = io.Copy(io.Discard, io.LimitReader(body, limit))
}

// isErrorHTTP1Broken checks whether the provided error indicates a specific HTTP/1.x transport issue.
// In particular, it looks for errors that mention a broken HTTP/1.x connection due to a malformed HTTP version.
// Such errors suggest that an HTTP/2 connection might be more appropriate.
//
// Parameters:
//   - err (error): The error to inspect.
//
// Returns:
//   - isErrorHTTP1Broken (bool): True if the error indicates a broken HTTP/1.x connection; otherwise, false.
func isErrorHTTP1Broken(err error) (isErrorHTTP1Broken bool) {
	isErrorHTTP1Broken = err != nil && strings.Contains(err.Error(), "net/http: HTTP/1.x transport connection broken: malformed HTTP version \"HTTP/2\"")

	return
}

// isErrorRecoverable determines whether an error encountered during an HTTP request is recoverable,
// meaning that the request may be retried. It examines the request context and the error details,
// filtering out errors such as context cancellations, excessive redirects, unsupported protocol schemes,
// or TLS certificate verification failures.
//
// Parameters:
//   - ctx (context.Context): The request's context, which may contain cancellation signals or deadlines.
//   - err (error): The error encountered during the HTTP request.
//
// Returns:
//   - recoverable (bool): True if the error is considered recoverable (the request may be retried); otherwise, false.
//   - errr (error): An error value to override the original error in case the context is cancelled or a non-retryable error is detected.
func isErrorRecoverable(ctx context.Context, err error) (recoverable bool, errr error) {
	if ctx.Err() != nil {
		errr = ctx.Err()

		return
	}

	var URLError *url.Error

	if err != nil && errors.As(err, &URLError) {
		if redirectsErrorRegex.MatchString(err.Error()) {
			errr = err

			return
		}

		if schemeErrorRegex.MatchString(err.Error()) {
			errr = err

			return
		}

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
