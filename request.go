package http

import (
	"time"

	"github.com/hueristiq/hq-go-http/method"
	"github.com/hueristiq/hq-go-retrier/backoff"
)

// RequestConfiguration holds settings specific to an individual HTTP request.
// These settings override the global ClientConfiguration on a per-request basis.
// The fields in this structure allow you to customize various aspects of an HTTP request,
// such as the method, URL, query parameters, headers, body, and retry behavior.
// This configuration is merged with the global settings to build the final request
// before execution.
//
// Fields:
//   - Method (method.Method): The HTTP method for the request (e.g., GET, POST, PUT, etc.).
//   - BaseURL (string): An optional base URL to be prefixed to the URL, overriding the global BaseURL.
//   - URL (string): The target URL or path for the request.
//   - Params (map[string]string): Query parameters to append to the URL.
//   - Headers ([]Header): HTTP headers to include with the request.
//   - Body (interface{}): The request body, if applicable.
//   - RespReadLimit (int64): The maximum number of bytes to read from a response body when draining.
//   - RetryPolicy (RetryPolicy): A function to determine retry behavior for this specific request.
//   - RetryMax (int): The maximum number of retry attempts for this request.
//   - RetryWaitMin (time.Duration): The minimum duration to wait between retries.
//   - RetryWaitMax (time.Duration): The maximum duration to wait between retries.
//   - RetryBackoff (backoff.Backoff): The strategy used to calculate backoff delays between retries.
type RequestConfiguration struct {
	Method        method.Method
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
