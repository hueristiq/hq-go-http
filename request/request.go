package request

import (
	"context"
	"io"
	"net/http"

	"github.com/hueristiq/hq-go-url/parser"
)

// Request is a wrapper around http.Request that enables the request body to be reusable.
// This is useful in scenarios such as retries or logging where the body might need to be read
// more than once. The Request type embeds *http.Request so that it can be used directly with
// standard HTTP libraries and functions.
//
// Fields:
//   - Request (*http.Request): The underlying HTTP request object.
type Request struct {
	*http.Request
}

// New creates a new Request using the specified HTTP method, URL, and body.
// This function is a convenience wrapper that internally delegates to NewFromURL.
//
// Parameters:
//   - method (string): The HTTP method to use (e.g., "GET", "POST").
//   - URL (string): The target URL for the HTTP request.
//   - body (interface{}): An optional parameter representing the request body. The body can be any type
//     supported by getReusableBodyReadCloser (e.g., a ReusableReadCloser, *ReusableReadCloser,
//     or any type accepted by NewReusableReadCloser).
//
// Returns:
//   - req (*Request): A pointer to the newly created Request wrapper containing an http.Request.
//   - err (error): An error value if the request creation fails (for example, due to an unsupported body type).
func New(method, URL string, body interface{}) (req *Request, err error) {
	req, err = NewFromURL(method, URL, body)

	return
}

// NewWithContext creates a new Request with the specified context, HTTP method, URL, and body.
// This function is similar to New but accepts a context.Context, which is useful for cancellation,
// deadlines, or timeouts.
//
// Parameters:
//   - ctx (context.Context): The context to be associated with the HTTP request.
//   - method (string): The HTTP method to use (e.g., "GET", "POST").
//   - URL (string): The target URL for the HTTP request.
//   - body (interface{}): An optional parameter representing the request body. The body can be any
//     type supported by getReusableBodyReadCloser, for example, a ReusableReadCloser,
//     *ReusableReadCloser, or any type accepted by NewReusableReadCloser.
//
// Returns:
//   - req (*Request): A pointer to the newly created Request wrapper containing an http.Request.
//   - err (error): An error value if the request creation fails (for example, due to an unsupported body type).
func NewWithContext(ctx context.Context, method, URL string, body interface{}) (req *Request, err error) {
	req, err = NewFromURLWithContext(ctx, method, URL, body)

	return
}

// NewFromURL creates a new Request using the specified HTTP method, URL, and body.
// It uses a default background context and is a convenience wrapper around NewFromURLWithContext..
//
// Parameters:
//   - method (string): The HTTP method to use (e.g., "GET", "POST").
//   - URL (string): The target URL for the HTTP request.
//   - body (interface{}): An optional parameter representing the request body. The body can be any
//     type supported by getReusableBodyReadCloser, for example, a ReusableReadCloser,
//     *ReusableReadCloser, or any type accepted by NewReusableReadCloser.
//
// Returns:
//   - req (*Request): A pointer to the newly created Request wrapper containing an http.Request.
//   - err (error): An error value if the request creation fails (for example, due to an unsupported body type).
func NewFromURL(method, URL string, body interface{}) (req *Request, err error) {
	req, err = NewFromURLWithContext(context.Background(), method, URL, body)

	return
}

// NewFromURLWithContext creates a new Request using the provided context, HTTP method, URL, and body.
// It performs the following steps:
//  1. Parses the provided URL using a custom parser that applies a default scheme ("http") if needed.
//  2. Constructs an http.Request with a temporary URL (containing only the scheme and host)
//     to avoid overriding any patches applied by the custom parser.
//  3. Replaces the temporary URL in the http.Request with the fully parsed URL.
//  4. Converts the provided body into a reusable ReadCloser using getReusableBodyReadCloser,
//     updating the ContentLength accordingly.
//
// Parameters:
//   - ctx (context.Context): The context to associate with the HTTP request.
//   - method (string): The HTTP method to use (e.g., "GET", "POST").
//   - URL (string): The target URL for the HTTP request.
//   - body (interface{}): An optional parameter representing the request body. The body can be any
//     type supported by getReusableBodyReadCloser, for example, a ReusableReadCloser,
//     *ReusableReadCloser, or any type accepted by NewReusableReadCloser.
//
// Returns:
//   - req (*Request): A pointer to the newly created Request wrapper containing an http.Request.
//   - err (error): An error value if the request creation fails (for example, due to an unsupported body type).
func NewFromURLWithContext(ctx context.Context, method, URL string, body interface{}) (req *Request, err error) {
	parsedURL, err := parser.New(parser.WithDefaultScheme("http")).Parse(URL)
	if err != nil {
		return
	}

	// we provide a url without path to http.NewRequest at start and then replace url instance directly
	// because `http.NewRequest()` internally parses using `url.Parse()` this removes/overrides any
	// patches done by parsed.URL in unsafe mode (ex: https://example.com/%invalid)
	//
	// Note: this does not have any impact on actual path when sending request
	// `http.NewRequestxxx` internally only uses `u.Host` and all other data is stored in `url.URL` instance
	internalHTTPRequest, err := http.NewRequestWithContext(ctx, method, parsedURL.Scheme+"://"+parsedURL.Host, nil) //nolint:gocritic // To be refactored
	if err != nil {
		return
	}

	internalHTTPRequest.URL = parsedURL.URL

	reusableBodyReadCloser, err := getReusableBodyReadCloser(body)
	if err != nil {
		return
	}

	if reusableBodyReadCloser != nil {
		internalHTTPRequest.Body = reusableBodyReadCloser
		internalHTTPRequest.ContentLength = int64(len(reusableBodyReadCloser.data))
	}

	req = &Request{
		Request: internalHTTPRequest,
	}

	return
}

// getReusableBodyReadCloser attempts to convert the provided raw input into a reusable ReadCloser.
// The conversion supports multiple input types, enabling flexibility in how the request body is specified.
//
// Supported types include:
//   - ReusableReadCloser: If raw is a value of this type, its address is taken.
//   - *ReusableReadCloser: If raw is already a pointer, it is used directly.
//   - func() (io.Reader, error): If raw is a function with this signature, the function is invoked to obtain
//     an io.Reader, which is then converted via NewReusableReadCloser.
//   - Other types: For all other types, raw is passed to NewReusableReadCloser, which supports a variety of types.
//
// Parameters:
//   - raw (interface{}): The raw input representing the request body. It may be nil or any type
//     supported by NewReusableReadCloser.
//
// Returns:
//   - reader (*ReusableReadCloser): A pointer to the reusable read-closer if conversion is successful;
//     otherwise, nil.
//   - err (error): An error value if the conversion fails.
func getReusableBodyReadCloser(raw interface{}) (reader *ReusableReadCloser, err error) {
	if raw != nil {
		switch body := raw.(type) {
		case ReusableReadCloser:
			reader = &body
		case *ReusableReadCloser:
			reader = body
		case func() (io.Reader, error):
			var tmp io.Reader

			tmp, err = body()
			if err != nil {
				return
			}

			reader, err = NewReusableReadCloser(tmp)
			if err != nil {
				return
			}
		default:
			reader, err = NewReusableReadCloser(body)
			if err != nil {
				return
			}
		}
	}

	return
}
