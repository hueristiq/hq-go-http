package http

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httputil"
)

// Request wraps the standard http.Request struct and adds fields for tracking
// request metrics and custom authentication details.
//
// NOTE: Request is not threadsafe. A request cannot be used by multiple goroutines
// concurrently.
type Request struct {
	// Embedded standard http.Request. This makes a *Request act exactly
	// like an *http.Request so that all meta methods are supported.
	*http.Request

	Metrics Metrics // Tracks various metrics related to request handling
}

// WithContext creates a new Request with the provided context. This allows you
// to pass metadata, deadlines, or cancellation signals throughout the request lifecycle.
//
// Parameters:
//   - ctx: The new context to associate with the request.
//
// Returns:
//   - req: A new Request with the updated context.
func (r *Request) WithContext(ctx context.Context) (req *Request) {
	req = r

	req.Request = req.Request.WithContext(ctx)

	return
}

// BodyBytes reads the request body and returns it as a byte slice.
//
// Parameters: None.
//
// Returns:
//   - body: The body content as a byte slice, or an empty slice if the body is nil.
//   - err: An error if the body reading fails.
func (r *Request) BodyBytes() (body []byte, err error) {
	if r.Request.Body == nil {
		return
	}

	buf := new(bytes.Buffer)

	_, err = buf.ReadFrom(r.Body)
	if err != nil {
		return
	}

	body = buf.Bytes()

	return
}

// Clone creates a deep copy of the Request, resetting its Metrics and duplicating
// the Auth data, if available. This is useful for generating a new request instance
// while retaining most properties.
//
// Parameters:
//   - ctx: The context to associate with the cloned request.
//
// Returns:
//   - req: A new Request with the same data but reset Metrics and context.
func (r *Request) Clone(ctx context.Context) (req *Request) {
	req = &Request{
		Request: r.Request.Clone(ctx),
		Metrics: Metrics{},
	}

	return
}

// Dump serializes the Request into a byte slice. Optionally includes the body
// in the dump if it is present and non-empty.
//
// Parameters: None.
//
// Returns:
//   - dump: A byte slice representation of the request, including headers and optionally the body.
//   - err: An error if dumping the request fails.
func (r *Request) Dump() (dump []byte, err error) {
	resplen := int64(0)
	dumpbody := true

	clone := r.Clone(context.TODO())
	if clone.Body != nil {
		resplen, _ = getReaderLength(clone.Body)
	}

	if resplen == 0 {
		dumpbody = false

		clone.ContentLength = 0
		clone.Body = nil

		delete(clone.Header, "Content-length")
	} else {
		clone.ContentLength = resplen
	}

	dump, err = httputil.DumpRequestOut(clone.Request, dumpbody)
	if err != nil {
		return
	}

	return
}

// Metrics represents statistics related to request handling. These metrics are
// useful for tracking performance or issues encountered during the request lifecycle.
type Metrics struct {
	Failures    int // Failures is the number of failed requests
	Retries     int // Retries is the number of retries for the request
	DrainErrors int // DrainErrors is number of errors occurred in draining response body
}

// NewRequest creates a new Request without context using the specified HTTP method, URL, and body.
//
// Parameters:
//   - method: The HTTP method to use (e.g., GET, POST).
//   - url: The URL to send the request to.
//   - body: The request body, which can be nil.
//
// Returns:
//   - req: A new Request object.
//   - err: An error if request creation fails.
func NewRequest(method, url string, body interface{}) (req *Request, err error) {
	req, err = NewRequestFromURL(url, method, body)

	return
}

// NewRequestWithContext creates a new Request with the specified context.
//
// Parameters:
//   - ctx: The context to associate with the request.
//   - method: The HTTP method to use (e.g., GET, POST).
//   - url: The URL to send the request to.
//   - body: The request body, which can be nil.
//
// Returns:
//   - req: A new Request object with the provided context.
//   - err: An error if request creation fails.
func NewRequestWithContext(ctx context.Context, method, url string, body interface{}) (req *Request, err error) {
	req, err = NewRequestFromURLWithContext(ctx, url, method, body)

	return
}

// NewRequestFromURL creates a new Request without a context.
//
// Parameters:
//   - url: The URL to send the request to.
//   - method: The HTTP method to use.
//   - body: The request body, which can be nil.
//
// Returns:
//   - req: A new Request object with the default context (context.Background()).
//   - err: An error if request creation fails.
func NewRequestFromURL(url, method string, body interface{}) (req *Request, err error) {
	req, err = NewRequestFromURLWithContext(context.Background(), url, method, body)

	return
}

// NewRequestFromURLWithContext creates a new Request with the specified context and body.
// It also calculates the content length if a body is provided and sets the appropriate headers.
//
// Parameters:
//   - ctx: The context to associate with the request.
//   - url: The URL to send the request to.
//   - method: The HTTP method to use.
//   - body: The request body, which can be nil.
//
// Returns:
//   - req: A new Request object with the provided context.
//   - err: An error if request creation fails.
func NewRequestFromURLWithContext(ctx context.Context, url, method string, body interface{}) (req *Request, err error) {
	reqBodyReader, reqContentLength, err := getReusableBodyandContentLength(body)
	if err != nil {
		return
	}

	// we provide a url without path to http.NewRequest at start and then replace url instance directly
	// because `http.NewRequest()` internally parses using `url.Parse()` this removes/overrides any
	// patches done by urlutil.URL in unsafe mode (ex: https://scanme.sh/%invalid)
	// Note: this does not have any impact on actual path when sending request
	// `http.NewRequestxxx` internally only uses `u.Host` and all other data is stored in `url.URL` instance
	httpReq, err := http.NewRequestWithContext(ctx, method, url, nil) //nolint:gocritic // To be refactored
	if err != nil {
		return
	}

	// content-length and body should be assigned only
	// if request has body
	if reqBodyReader != nil {
		httpReq.ContentLength = reqContentLength
		httpReq.Body = reqBodyReader
	}

	req = &Request{
		Request: httpReq,
		Metrics: Metrics{},
	}

	return
}

// getReaderLength reads the entire content of an io.Reader and returns its length. The data is discarded.
//
// Parameters:
//   - reader: The io.Reader containing the data.
//
// Returns:
//   - length: The number of bytes read from the reader.
//   - err: An error if reading fails.
func getReaderLength(reader io.Reader) (length int64, err error) {
	length, err = io.Copy(io.Discard, reader)

	return
}
