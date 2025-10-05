package http

import (
	"time"

	hqgohttpmethod "github.com/hueristiq/hq-go-http/method"
	hqgoretrierbackoff "github.com/hueristiq/hq-go-retrier/backoff"
)

// RequestConfiguration holds settings specific to an individual HTTP request.
//
// This structure allows customization of various aspects of an HTTP request, such as the method,
// URL, query parameters, headers, body, and retry behavior. It is designed to override global
// client settings (e.g., from a ClientConfiguration) on a per-request basis. When used in an
// HTTP client, the fields in this structure are merged with global settings to construct the final
// request before execution.
//
// Fields:
//   - Method (hqgohttpmethod.Method): The HTTP method for the request (e.g., GET, POST, PUT).
//   - BaseURL (string): An optional base URL to prepend to the URL, overriding the global BaseURL.
//   - URL (string): The target URL or path for the request (e.g., "/api/resource" or a full URL).
//   - Params (map[string]string): Query parameters to append to the URL as key-value pairs.
//   - Headers ([]Header): A slice of Header objects specifying HTTP headers to include.
//   - Body (interface{}): The request body, which can be a string, byte slice, or other data type
//     supported by the HTTP client.
//   - RespReadLimit (int64): The maximum number of bytes to read from a response body when draining
//     (e.g., to prevent excessive memory usage).
//   - RetryPolicy (RetryPolicy): A function defining the retry behavior for this request.
//   - RetryMax (int): The maximum number of retry attempts for this request.
//   - RetryWaitMin (time.Duration): The minimum duration to wait between retry attempts.
//   - RetryWaitMax (time.Duration): The maximum duration to wait between retry attempts.
//   - RetryBackoff (hqgoretrierbackoff.Backoff): The strategy used to calculate backoff delays
//     between retries (e.g., exponential, linear).
type RequestConfiguration struct {
	Method        hqgohttpmethod.Method
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
	RetryBackoff  hqgoretrierbackoff.Backoff
}

// Header represents an HTTP header with a key, value, and operation type.
//
// The Header type specifies a header key-value pair and the operation to perform when applying
// it to an HTTP request (e.g., appending or replacing the header). The operation field allows
// fine-grained control over how headers are managed, particularly when multiple headers with the
// same key are present.
//
// Fields:
//   - key (string): The header name (e.g., "Content-Type", "Authorization").
//   - value (string): The header value (e.g., "application/json", "Bearer token").
//   - operation (headerOperation): The operation to perform, either "append" or "replace".
type Header struct {
	key       string
	value     string
	operation headerOperation
}

// headerOperation defines the type of operation to perform when applying a Header.
//
// It determines whether the header should be appended to existing headers with the same key
// (headerOperationAppend) or replace all existing headers with the same key
// (headerOperationReplace). This allows precise control over header behavior in HTTP requests.
type headerOperation string

// Constants defining valid header operations.
//
// These constants specify the possible operations for applying headers in a request:
//   - headerOperationAppend: Adds the header to the existing headers, allowing multiple values
//     for the same key (e.g., multiple "Accept" headers).
//   - headerOperationReplace: Replaces all existing headers with the same key, ensuring only
//     the specified value is used.
const (
	headerOperationAppend  headerOperation = "append"
	headerOperationReplace headerOperation = "replace"
)

// NewAddHeader creates a Header that appends the specified key-value pair to the request headers.
//
// This function constructs a Header with the "append" operation, which adds the header to the
// request without removing existing headers with the same key. This is useful for headers that
// support multiple values, such as "Accept" or "Cookie".
//
// Parameters:
//   - key (string): The header name (e.g., "Accept").
//   - value (string): The header value (e.g., "application/json").
//
// Returns:
//   - h (Header): A Header object configured to append the key-value pair.
func NewAddHeader(key, value string) (h Header) {
	h = Header{
		key:       key,
		value:     value,
		operation: headerOperationAppend,
	}

	return
}

// NewSetHeader creates a Header that replaces the specified key-value pair in the request headers.
//
// This function constructs a Header with the "replace" operation, which replaces all existing
// headers with the same key, ensuring only the specified value is used. This is useful for
// headers that should have a single value, such as "Content-Type" or "Authorization".
//
// Parameters:
//   - key (string): The header name (e.g., "Content-Type").
//   - value (string): The header value (e.g., "application/json").
//
// Returns:
//   - h (Header): A Header object configured to replace the key-value pair.
func NewSetHeader(key, value string) (h Header) {
	h = Header{
		key:       key,
		value:     value,
		operation: headerOperationReplace,
	}

	return
}
