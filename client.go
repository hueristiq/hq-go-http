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

	"github.com/hueristiq/hq-go-http/method"
	"github.com/hueristiq/hq-go-http/request"
	hqgoretrier "github.com/hueristiq/hq-go-retrier"
	"github.com/hueristiq/hq-go-retrier/backoff"
	"golang.org/x/net/http2"
)

// Client is the primary structure used to perform HTTP requests.
// It manages both global and request-specific configurations, and maintains
// separate underlying HTTP clients for HTTP/1.x and HTTP/2.
// Additionally, it tracks the number of requests made to trigger periodic
// connection resets, preventing resource exhaustion in long-running applications.
//
// Fields:
//   - cfg (*ClientConfiguration): Global configuration settings that apply to all requests.
//   - internalHTTPClient (*http.Client): The underlying HTTP client instance used for HTTP/1.x requests.
//   - internalHTTP2Client (*http.Client): A fallback HTTP client instance used for HTTP/2 requests.
//   - rc (atomic.Uint32): An atomic counter tracking the number of requests executed.
//     When a preset threshold is reached, idle connections are closed to free resources.
type Client struct {
	cfg                 *ClientConfiguration
	internalHTTPClient  *http.Client
	internalHTTP2Client *http.Client
	rc                  atomic.Uint32
}

// Do executes an HTTP request using the Client.
// It applies retry logic, error handling, and, when necessary, falls back to the HTTP/2 client
// if the HTTP/1.x client fails due to a specific transport error.
// It supports automatic connection draining (to enable connection reuse) and integrates
// with the retry policy defined in the request configuration.
//
// Parameters:
//   - req (*request.Request): A pointer to a request.Request instance that wraps an *http.Request.
//     This wrapper allows the request body to be reusable between retries.
//   - cfg (*RequestConfiguration): Request-specific configuration overrides including retry settings,
//     response read limits, and additional parameters.
//
// Returns:
//   - res (*http.Response): The HTTP response received upon success.
//   - err (error): An error if the request ultimately fails after all retry attempts.
func (c *Client) Do(req *request.Request, cfg *RequestConfiguration) (res *http.Response, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.cfg.Timeout)

	defer cancel()

	res, err = hqgoretrier.RetryWithData(ctx, func() (res *http.Response, err error) {
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
		hqgoretrier.WithRetryMax(cfg.RetryMax),
		hqgoretrier.WithRetryWaitMin(cfg.RetryWaitMin),
		hqgoretrier.WithRetryWaitMax(cfg.RetryWaitMax),
		hqgoretrier.WithRetryBackoff(cfg.RetryBackoff),
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

// closeIdleConnections manages connection reuse by monitoring the number of requests made.
// When the request counter reaches a threshold (100 requests), the counter is reset and the
// client's idle connections are forcefully closed. This mechanism helps prevent potential
// resource leaks in high-volume scenarios.
//
// Note: The behavior is enabled only if the ClientConfiguration.CloseIdleConnections flag is true.
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
// It constructs the request by setting the HTTP method, URL (including base URL and query parameters),
// headers, and body. After preparing the request, it delegates execution to the Do method.
//
// Parameters:
//   - configurations (...*RequestConfiguration): One or more pointers to RequestConfiguration instances
//     that provide request-specific overrides. These configurations are merged with the global client settings.
//
// Returns:
//   - res (*http.Response): The HTTP response from the executed request if successful.
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

	for _, header := range cfg.Headers {
		switch header.mode {
		case HeaderModeAdd:
			req.Header.Add(header.key, header.value)
		case HeaderModeSet:
			req.Header.Set(header.key, header.value)
		}
	}

	res, err = c.Do(req, cfg)

	return
}

// getRequestConfiguration merges one or more RequestConfiguration objects with the client's default configuration.
// It applies overrides for HTTP method, base URL, URL path, query parameters, headers, body, timeout,
// and retry settings. If a BaseURL is provided, it is combined with the relative URL. Additionally,
// query parameters are appended to the URL's query string.
//
// Parameters:
//   - configurations (...*RequestConfiguration): Variadic list of pointers to RequestConfiguration instances.
//
// Returns:
//   - cfg (*RequestConfiguration): The merged RequestConfiguration.
//   - err (error): An error if merging configurations or constructing the final URL fails.
func (c *Client) getRequestConfiguration(configurations ...*RequestConfiguration) (cfg *RequestConfiguration, err error) {
	cfg = &RequestConfiguration{
		Method:        c.cfg.Method,
		BaseURL:       c.cfg.BaseURL,
		URL:           c.cfg.URL,
		Params:        make(map[string]string),
		Headers:       []Header{},
		Body:          c.cfg.Body,
		RespReadLimit: c.cfg.RespReadLimit,
		RetryPolicy:   c.cfg.RetryPolicy,
		RetryMax:      c.cfg.RetryMax,
		RetryWaitMin:  c.cfg.RetryWaitMin,
		RetryWaitMax:  c.cfg.RetryWaitMax,
		RetryBackoff:  c.cfg.RetryBackoff,
	}

	if c.cfg.Params != nil {
		for k, v := range c.cfg.Params {
			cfg.Params[k] = v
		}
	}

	if c.cfg.Headers != nil {
		cfg.Headers = append(cfg.Headers, c.cfg.Headers...)
	}

	for _, configuration := range configurations {
		if configuration.Method != "" {
			cfg.Method = configuration.Method
		}

		if configuration.BaseURL != "" {
			cfg.BaseURL = configuration.BaseURL
		}

		if configuration.URL != "" {
			cfg.URL = configuration.URL
		}

		if configuration.Params != nil {
			for k, v := range configuration.Params {
				cfg.Params[k] = v
			}
		}

		if configuration.Headers != nil {
			cfg.Headers = append(cfg.Headers, configuration.Headers...)
		}

		if configuration.Body != "" {
			cfg.Body = configuration.Body
		}

		if configuration.RespReadLimit > 0 {
			cfg.RespReadLimit = configuration.RespReadLimit
		}

		if configuration.RetryPolicy != nil {
			cfg.RetryPolicy = configuration.RetryPolicy
		}

		if configuration.RetryMax > 0 {
			cfg.RetryMax = configuration.RetryMax
		}

		if configuration.RetryWaitMin > 0 {
			cfg.RetryWaitMin = configuration.RetryWaitMin
		}

		if configuration.RetryWaitMax > 0 {
			cfg.RetryWaitMax = configuration.RetryWaitMax
		}

		if configuration.RetryBackoff != nil {
			cfg.RetryBackoff = configuration.RetryBackoff
		}
	}

	if cfg.BaseURL != "" {
		cfg.URL, err = url.JoinPath(cfg.BaseURL, cfg.URL)
		if err != nil {
			return
		}
	}

	if len(cfg.Params) > 0 {
		var parsed *url.URL

		parsed, err = url.Parse(cfg.URL)
		if err != nil {
			return
		}

		q := parsed.Query()

		for k, v := range cfg.Params {
			q.Add(k, v)
		}

		parsed.RawQuery = q.Encode()

		cfg.URL = parsed.String()
	}

	return
}

// Get is a convenience method for performing an HTTP GET request.
// It sets the HTTP method to GET and delegates request execution to the Request method.
//
// Parameters:
//   - URL (string): The target URL for the GET request.
//   - configurations (...*RequestConfiguration): Optional request configuration overrides.
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

// Head is a convenience method for performing an HTTP HEAD request.
// It sets the HTTP method to HEAD and delegates request execution to the Request method.
//
// Parameters:
//   - URL (string): The target URL for the HEAD request.
//   - configurations (...*RequestConfiguration): Optional request configuration overrides.
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
// It sets the HTTP method to PUT and delegates request execution to the Request method.
//
// Parameters:
//   - URL (string): The target URL for the PUT request.
//   - body (interface{}): The payload to include in the PUT request.
//   - configurations (...*RequestConfiguration): Optional request configuration overrides.
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
// It sets the HTTP method to DELETE and delegates request execution to the Request method.
//
// Parameters:
//   - URL (string): The target URL for the DELETE request.
//   - configurations (...*RequestConfiguration): Optional request configuration overrides.
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
// It sets the HTTP method to POST and delegates request execution to the Request method.
//
// Parameters:
//   - URL (string): The target URL for the POST request.
//   - body (interface{}): The payload to include in the POST request.
//   - configurations (...*RequestConfiguration): Optional request configuration overrides.
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
// It sets the HTTP method to OPTIONS and delegates request execution to the Request method.
//
// Parameters:
//   - URL (string): The target URL for the OPTIONS request.
//   - configurations (...*RequestConfiguration): Optional request configuration overrides.
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
// It controls HTTP client behavior including timeouts, retry logic, backoff strategies,
// connection management, and default request values.
//
// Fields:
//   - Client (*http.Client): An optional custom HTTP client to be used. If nil, a default client is used.
//   - Timeout (time.Duration): The maximum duration allowed for each HTTP request.
//   - CloseIdleConnections (bool): Determines whether idle connections should be periodically closed.
//   - Method (string): The default HTTP method to use (e.g., GET, POST) if not overridden.
//   - BaseURL (string): A base URL that is prefixed to all request URLs.
//   - URL (string): The default URL path that can be combined with BaseURL.
//   - Params (map[string]string): Default query parameters appended to every request.
//   - Headers ([]Header): Default HTTP headers included in every request.
//   - Body (interface{}): The default request body; can be overridden per request.
//   - RespReadLimit (int64): The maximum number of bytes to drain from a response body to allow connection reuse.
//   - RetryPolicy (RetryPolicy): A function that determines whether a request should be retried.
//   - RetryMax (int): The maximum number of retry attempts before giving up.
//   - RetryWaitMin (time.Duration): The minimum wait time between retries.
//   - RetryWaitMax (time.Duration): The maximum wait time between retries.
//   - RetryBackoff (backoff.Backoff): The backoff strategy used to calculate wait times between retries.
type ClientConfiguration struct {
	Client               *http.Client
	Timeout              time.Duration
	CloseIdleConnections bool

	Method        string
	BaseURL       string
	URL           string
	Params        map[string]string
	Headers       []Header
	Body          interface{}
	RespReadLimit int64
	RetryPolicy   RetryPolicy
	RetryMax      int
	RetryWaitMin  time.Duration
	RetryWaitMax  time.Duration
	RetryBackoff  backoff.Backoff
}

// Configuration ensures that all configuration fields are properly initialized.
// It creates empty maps for Params and Headers if they are nil, and sets default
// retry policy and backoff strategy if they are not provided.
//
// Returns:
//   - configuration (*ClientConfiguration): A pointer to the initialized configuration.
func (c *ClientConfiguration) Configuration() (configuration *ClientConfiguration) {
	configuration = c

	if configuration.Params == nil {
		configuration.Params = make(map[string]string)
	}

	if configuration.Headers == nil {
		configuration.Headers = []Header{}
	}

	if configuration.RetryPolicy == nil {
		configuration.RetryPolicy = DefaultRetryPolicy()
	}

	if configuration.RetryBackoff == nil {
		configuration.RetryBackoff = backoff.Exponential()
	}

	return
}

// NewClient creates and configures a new Client based on the provided ClientConfiguration.
// It initializes both HTTP/1.x and HTTP/2 clients, applies retry policies, backoff strategies,
// and sets up connection management settings.
//
// Parameters:
//   - cfg (*ClientConfiguration): A pointer to a ClientConfiguration containing desired settings.
//
// Returns:
//   - client (*Client): A pointer to the newly created Client instance.
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

// isErrorHTTP1Broken checks whether the provided error indicates a specific HTTP/1.x transport issue.
// In particular, it detects errors that mention a broken HTTP/1.x connection due to a malformed HTTP version.
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
