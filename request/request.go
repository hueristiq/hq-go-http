package request

import (
	"context"
	"io"
	"net/http"

	hqgourlparser "github.com/hueristiq/hq-go-url/parser"
)

// schemeDefaultingParser is the parser [New] builds URLs with. It is
// constructed once at package load: parser construction materializes the TLD
// set, which is far too expensive to repeat per request. Parse only reads the
// parser's configuration, so the shared instance is safe for concurrent use.
var schemeDefaultingParser = hqgourlparser.New(hqgourlparser.WithDefaultScheme("http"))

// Request is an HTTP request whose body is reusable.
//
// It wraps *http.Request so it integrates directly with the standard library
// and any HTTP tooling, while the body — converted to a
// [*ReusableReadCloser] — rewinds itself at the end of the data, so every
// read sees the full payload. This suits scenarios where the same request is
// sent repeatedly, such as retries or logging.
//
// Construct through [New]; the zero value is not usable.
//
// Fields:
//   - Request (*http.Request): the underlying HTTP request.
type Request struct {
	*http.Request
}

// New creates a [Request] for the given method, URL, and body, bound to ctx.
//
// The URL is parsed with a parser that applies the default scheme "http" when
// none is given, so "example.com/path" is a valid URL. The body accepts a
// string, a []byte, an io.Reader, an io.ReadSeeker, a *bytes.Buffer, a
// func() (io.Reader, error), a ReusableReadCloser, or a *ReusableReadCloser;
// it is buffered in memory and its length becomes the request's
// ContentLength. A nil body means no body.
//
// Parameters:
//   - ctx (context.Context): the context associated with the request.
//   - method (string): the HTTP method (e.g., "GET", "POST"); empty means
//     GET.
//   - rawURL (string): the target URL for the request.
//   - body (any): an optional request body.
//
// Returns:
//   - req (*Request): the newly created request wrapper.
//   - err (error): non-nil when the URL does not parse or the body type is
//     unsupported.
func New(ctx context.Context, method, rawURL string, body any) (req *Request, err error) {
	parsedURL, err := schemeDefaultingParser.Parse(rawURL)
	if err != nil {
		return
	}

	// The request is first built against a URL carrying only the scheme and
	// host, and the parsed URL is then assigned directly: NewRequestWithContext
	// parses its URL argument with url.Parse, which would strip patches the
	// parser applied in unsafe mode (e.g. https://example.com/%invalid).
	//
	// This has no impact on the wire: the transport reads only the host from
	// the request — everything else travels in the assigned url.URL instance.
	internalHTTPRequest, err := http.NewRequestWithContext(ctx, method, parsedURL.Scheme+"://"+parsedURL.Host, nil) //nolint:gocritic // nil keeps Request.Body nil, matching net/http's no-body convention; http.NoBody would leak a non-nil body.
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

// getReusableBodyReadCloser converts the provided raw input into a reusable
// ReadCloser.
//
// Supported types include:
//   - ReusableReadCloser: if raw is a value of this type, its address is taken.
//   - *ReusableReadCloser: if raw is already a pointer, it is used directly.
//   - func() (io.Reader, error): the function is invoked to obtain an
//     io.Reader, which is then converted via NewReusableReadCloser.
//   - Other types: raw is passed to NewReusableReadCloser, which supports a
//     variety of types.
//
// Parameters:
//   - raw (any): the raw input representing the request body; nil yields no
//     body.
//
// Returns:
//   - reader (*ReusableReadCloser): the reusable read-closer, or nil when raw
//     is nil.
//   - err (error): non-nil when the conversion fails.
func getReusableBodyReadCloser(raw any) (reader *ReusableReadCloser, err error) {
	if raw == nil {
		return
	}

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

	return
}
