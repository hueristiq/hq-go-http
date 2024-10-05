package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	dac "github.com/Mzack9999/go-http-digest-auth-client"
	"github.com/hueristiq/hq-go-http/methods"
	retrier "github.com/hueristiq/hq-go-retrier"
	"github.com/hueristiq/hq-go-retrier/backoff"
	"golang.org/x/net/http2"
)

// Client defines an HTTP client with retry policies, support for digest authentication, and optional HTTP/2 fallback.
// It is configured with both HTTP/1.x and HTTP/2 clients, as well as error handling and retry logic.
type Client struct {
	HTTPClient  *http.Client // HTTP/1.x client used for standard requests.
	HTTP2Client *http.Client // HTTP/2 client fallback if HTTP/1.x fails with specific errors.

	OnError ErrorHandler // Custom error handler function, invoked after retries are exhausted.

	CheckRetry CheckRetry      // Function to check if the request should be retried based on response or error.
	Backoff    backoff.Backoff // Strategy for delaying retries (e.g., exponential backoff).

	requestCounter atomic.Uint32 // Counts the number of requests made, used to determine when to close idle connections.

	cfg *ClientConfiguration // Configuration object containing client-specific settings like retry limits, timeouts, and backoff strategy.
}

// Do executes an HTTP request with the client, applying retry policies, error handling, and optional HTTP/2 fallback.
// It supports digest authentication and keeps track of request metrics.
//
// Parameters:
//   - req: The HTTP request to be executed.
//
// Returns:
//   - res: The HTTP response from the request, or nil if the request failed.
//   - err: Error encountered during the request or after exhausting retries.
func (c *Client) Do(req *Request) (res *http.Response, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.cfg.Timeout)

	defer cancel()

	// Determine the maximum number of retries from the configuration or request context.
	retryMax := c.cfg.RetryMax

	if ctxRetryMax := req.Context().Value(RetryMax); ctxRetryMax != nil {
		if maxRetriesParsed, ok := ctxRetryMax.(int); ok {
			retryMax = maxRetriesParsed
		}
	}

	res, err = retrier.RetryWithData(ctx, func() (res *http.Response, err error) {
		// Handle digest authentication if required by the request.
		if req.Auth != nil && req.Auth.Type == DigestAuth {
			digestTransport := dac.NewTransport(req.Auth.Username, req.Auth.Password)

			digestTransport.HTTPClient = c.HTTPClient

			res, err = digestTransport.RoundTrip(req.Request)
		} else {
			res, err = c.HTTPClient.Do(req.Request)
		}

		// Check if the request should be retried based on the response or error.
		retry, checkErr := c.CheckRetry(req.Context(), err)

		// Fallback to HTTP/2 if HTTP/1.x transport encounters specific errors.
		if err != nil && strings.Contains(err.Error(), "net/http: HTTP/1.x transport connection broken: malformed HTTP version \"HTTP/2\"") {
			res, err = c.HTTP2Client.Do(req.Request)

			retry, checkErr = c.CheckRetry(req.Context(), err)
		}

		if err != nil {
			req.Metrics.Failures++
		}

		if !retry {
			if checkErr != nil {
				err = checkErr
			}

			c.closeIdleConnections()

			return
		}

		req.Metrics.Retries++

		if err == nil && res != nil {
			c.drainBody(req, res)
		}

		return
	},
		retrier.WithMaxRetries(retryMax),
		retrier.WithMaxDelay(c.cfg.RetryWaitMax),
		retrier.WithMinDelay(c.cfg.RetryWaitMin),
	)

	if c.OnError != nil {
		c.closeIdleConnections()

		res, err = c.OnError(res, err, c.cfg.RetryMax+1)

		return
	}

	if err != nil {
		if res != nil {
			res.Body.Close()
		}

		c.closeIdleConnections()

		err = fmt.Errorf("%s %s giving up after %d attempts: %w", req.Method, req.URL, c.cfg.RetryMax+1, err)
	}

	return
}

// Get sends an HTTP GET request to the specified URL.
// It creates a new request and delegates the actual work to the Do method.
//
// Parameters:
//   - URL: The URL to send the GET request to.
//
// Returns:
//   - res: The HTTP response from the request, or nil if the request failed.
//   - err: Error encountered during the request or after exhausting retries.
func (c *Client) Get(URL string) (res *http.Response, err error) {
	req, err := NewRequest(methods.Get, URL, nil)
	if err != nil {
		return nil, err
	}

	res, err = c.Do(req)

	return
}

// Head sends an HTTP HEAD request to the specified URL.
// Similar to the Get method, but retrieves only the headers.
//
// Parameters:
//   - URL: The URL to send the HEAD request to.
//
// Returns:
//   - res: The HTTP response from the request, or nil if the request failed.
//   - err: Error encountered during the request or after exhausting retries.
func (c *Client) Head(URL string) (res *http.Response, err error) {
	req, err := NewRequest(methods.Head, URL, nil)
	if err != nil {
		return nil, err
	}

	res, err = c.Do(req)

	return
}

// Post sends an HTTP POST request with a specified body to the provided URL.
// It sets the appropriate Content-Type header and sends the request.
//
// Parameters:
//   - URL: The URL to send the POST request to.
//   - bodyType: The MIME type of the body content (e.g., "application/json").
//   - body: The data to send in the POST request.
//
// Returns:
//   - res: The HTTP response from the request, or nil if the request failed.
//   - err: Error encountered during the request or after exhausting retries.
func (c *Client) Post(URL, bodyType string, body interface{}) (res *http.Response, err error) {
	req, err := NewRequest(methods.Post, URL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", bodyType)

	res, err = c.Do(req)

	return
}

// PostForm sends an HTTP POST request with form data to the provided URL.
// The form data is encoded in application/x-www-form-urlencoded format.
//
// Parameters:
//   - URL: The URL to send the POST request to.
//   - data: The form data to be encoded and sent in the request body.
//
// Returns:
//   - res: The HTTP response from the request, or nil if the request failed.
//   - err: Error encountered during the request or after exhausting retries.
func (c *Client) PostForm(URL string, data url.Values) (res *http.Response, err error) {
	res, err = c.Post(URL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))

	return
}

// setKillIdleConnections checks the HTTP client's configuration to determine if idle connections should be killed.
// This is done based on settings like DisableKeepAlives or MaxConnsPerHost.
//
// Parameters: None.
//
// Returns: None.
func (c *Client) setKillIdleConnections() {
	if c.HTTPClient != nil || !c.cfg.KillIdleConn {
		if b, ok := c.HTTPClient.Transport.(*http.Transport); ok {
			c.cfg.KillIdleConn = b.DisableKeepAlives || b.MaxConnsPerHost < 0
		}
	}
}

// closeIdleConnections closes idle connections in the HTTP client if the request count reaches a certain threshold.
//
// Parameters: None.
//
// Returns: None.
func (c *Client) closeIdleConnections() {
	if c.cfg.KillIdleConn {
		if c.requestCounter.Load() < 100 {
			c.requestCounter.Add(1)
		} else {
			c.requestCounter.Store(0)
			c.HTTPClient.CloseIdleConnections()
		}
	}
}

// drainBody drains and discards the response body to prevent connection reuse issues.
// It also closes the response body.
//
// Parameters:
//   - req: The request whose body is being drained.
//   - resp: The response whose body is being drained.
//
// Returns: None.
func (c *Client) drainBody(req *Request, resp *http.Response) {
	_, err := io.Copy(io.Discard, io.LimitReader(resp.Body, c.cfg.RespReadLimit))
	if err != nil {
		req.Metrics.DrainErrors++
	}

	resp.Body.Close()
}

// ClientConfiguration defines the configuration for an HTTP client.
// This includes settings for retry logic, timeouts, backoff strategies, and connection handling.
type ClientConfiguration struct {
	HTTPClient      *http.Client    // Optional HTTP client for use by the main client.
	CheckRetry      CheckRetry      // Function to determine retry logic for failed requests.
	RetryMax        int             // Maximum number of retry attempts for requests.
	RetryWaitMin    time.Duration   // Minimum wait time between retries.
	RetryWaitMax    time.Duration   // Maximum wait time between retries.
	Backoff         backoff.Backoff // Backoff strategy for retrying requests.
	KillIdleConn    bool            // Whether to close idle connections after each request.
	RespReadLimit   int64           // Limit for reading response bodies during draining.
	Timeout         time.Duration   // Global timeout for the HTTP client.
	NoAdjustTimeout bool            // Flag to prevent automatic adjustment of per-request timeouts.
}

// NewClient creates a new HTTP client based on the provided configuration.
// It sets up the HTTP/1.x and HTTP/2 clients, retry logic, and backoff strategy.
//
// Parameters:
//   - cfg: The configuration for the client.
//
// Returns:
//   - client: A pointer to the newly created Client.
//   - err: Any error encountered during client creation.
func NewClient(cfg *ClientConfiguration) (client *Client, err error) {
	client = &Client{}

	client.HTTPClient = DefaultPooledClient()

	if cfg.KillIdleConn {
		client.HTTPClient = DefaultHTTPClient()
	}

	if cfg.HTTPClient != nil {
		client.HTTPClient = cfg.HTTPClient
	}

	client.HTTP2Client = DefaultHTTPClient()

	HTTP2ClientTransport, ok := client.HTTP2Client.Transport.(*http.Transport)
	if !ok {
		return
	}

	if err = http2.ConfigureTransport(HTTP2ClientTransport); err != nil {
		return
	}

	client.CheckRetry = DefaultRetryPolicy()

	if cfg.CheckRetry != nil {
		client.CheckRetry = cfg.CheckRetry
	}

	client.Backoff = backoff.Exponential()

	if cfg.Backoff != nil {
		client.Backoff = cfg.Backoff
	}

	// add timeout to clients
	if cfg.Timeout > 0 {
		client.HTTPClient.Timeout = cfg.Timeout
		client.HTTP2Client.Timeout = cfg.Timeout
	}

	// if necessary adjusts per-request timeout proportionally to general timeout (30%)
	if cfg.Timeout > time.Second*15 && cfg.RetryMax > 1 && !cfg.NoAdjustTimeout {
		client.HTTPClient.Timeout = time.Duration(cfg.Timeout.Seconds()*0.3) * time.Second
	}

	client.cfg = cfg

	client.setKillIdleConnections()

	return
}
