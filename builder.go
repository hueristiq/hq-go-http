package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	hqgohttpheader "github.com/hueristiq/hq-lib-http-go/header"
	hqgohttpmethod "github.com/hueristiq/hq-lib-http-go/method"
	hqgohttprequest "github.com/hueristiq/hq-lib-http-go/request"
	hqgohttpuseragent "github.com/hueristiq/hq-lib-http-go/useragent"
)

// RequestBuilder is an HTTP request under construction.
//
// The builder methods return the receiver for chaining — an intentional
// fluent-API deviation from the named-return convention; the deviation
// applies to the fluent setters only.
//
// A builder is bound to the [Client] that created it and executes through it:
// the client's retry engine, rate limiter, callbacks, and HTTP/2 fallback all
// apply. Construct through [Client.NewRequest]; the zero value is not usable.
// A builder is not safe for concurrent use.
//
// Fields:
//   - client (*Client): the client executing the built request.
//   - method (hqgohttpmethod.Method): the HTTP method; empty means GET.
//   - baseURL (string): the base URL prepended to a relative URL or path;
//     overrides the client's base URL.
//   - rawURL (string): the request URL — absolute, or relative to the base
//     URL.
//   - path (string): the path joined onto the base URL; used when rawURL is
//     empty.
//   - params (url.Values): query parameters merged into the URL's query.
//   - headersSet ([]headerField): headers replacing all values of their key,
//     in insertion order.
//   - headersAdd ([]headerField): headers appended to the values of their
//     key, in insertion order.
//   - body (any): the request payload, in any form the request package
//     accepts; nil means no body.
//   - requestOptions ([]RequestOptionFunc): per-request overrides of the
//     execution knobs.
type RequestBuilder struct {
	client         *Client
	method         hqgohttpmethod.Method
	baseURL        string
	rawURL         string
	path           string
	params         url.Values
	headersSet     []headerField
	headersAdd     []headerField
	body           any
	requestOptions []RequestOptionFunc
}

// NewRequest creates a [RequestBuilder] executing through the client.
//
// Returns:
//   - builder (*RequestBuilder): a new request builder.
func (c *Client) NewRequest() (builder *RequestBuilder) {
	builder = &RequestBuilder{
		client: c,
		params: make(url.Values),
	}

	return
}

// Method sets the HTTP method. Unset means GET.
//
// Parameters:
//   - method (hqgohttpmethod.Method): the HTTP method, e.g. method.POST.
//
// Returns:
//   - (*RequestBuilder): the receiver, for chaining.
func (b *RequestBuilder) Method(method hqgohttpmethod.Method) *RequestBuilder {
	b.method = method

	return b
}

// BaseURL sets the base URL a relative URL or path is joined onto, overriding
// the client's base URL for this request.
//
// Parameters:
//   - baseURL (string): the base URL, e.g. "https://api.example.com".
//
// Returns:
//   - (*RequestBuilder): the receiver, for chaining.
func (b *RequestBuilder) BaseURL(baseURL string) *RequestBuilder {
	b.baseURL = baseURL

	return b
}

// URL sets the request URL. An absolute URL (carrying "scheme://") is used
// as-is; a relative one is joined onto the base URL.
//
// Parameters:
//   - rawURL (string): the request URL.
//
// Returns:
//   - (*RequestBuilder): the receiver, for chaining.
func (b *RequestBuilder) URL(rawURL string) *RequestBuilder {
	b.rawURL = rawURL

	return b
}

// Path sets the path joined onto the base URL. It is used only when [URL]
// was not set.
//
// Parameters:
//   - path (string): the path, e.g. "/users/1".
//
// Returns:
//   - (*RequestBuilder): the receiver, for chaining.
func (b *RequestBuilder) Path(path string) *RequestBuilder {
	b.path = path

	return b
}

// Param appends a query parameter to the URL's query string. Repeated calls
// with the same key add multiple values.
//
// Parameters:
//   - key (string): the parameter name.
//   - value (string): the parameter value.
//
// Returns:
//   - (*RequestBuilder): the receiver, for chaining.
func (b *RequestBuilder) Param(key, value string) *RequestBuilder {
	b.params.Add(key, value)

	return b
}

// SetHeader sets a header, replacing all previously set values of the same
// key when the request is built.
//
// Parameters:
//   - key (string): the header field name.
//   - value (string): the header field value.
//
// Returns:
//   - (*RequestBuilder): the receiver, for chaining.
func (b *RequestBuilder) SetHeader(key, value string) *RequestBuilder {
	b.headersSet = append(b.headersSet, headerField{key: key, value: value})

	return b
}

// AddHeader appends a header value, keeping any values of the same key
// already present when the request is built.
//
// Parameters:
//   - key (string): the header field name.
//   - value (string): the header field value.
//
// Returns:
//   - (*RequestBuilder): the receiver, for chaining.
func (b *RequestBuilder) AddHeader(key, value string) *RequestBuilder {
	b.headersAdd = append(b.headersAdd, headerField{key: key, value: value})

	return b
}

// Body sets the request payload, in any form the request package accepts: a
// string, a []byte, an io.Reader, an io.ReadSeeker, a *bytes.Buffer, a
// func() (io.Reader, error), or a ReusableReadCloser. The body is buffered,
// so every retry attempt resends it in full.
//
// Parameters:
//   - body (any): the request payload; nil means no body.
//
// Returns:
//   - (*RequestBuilder): the receiver, for chaining.
func (b *RequestBuilder) Body(body any) *RequestBuilder {
	b.body = body

	return b
}

// WithOptions registers per-request overrides of the execution knobs, applied
// on top of the client configuration for this request only.
//
// Parameters:
//   - ofs (...RequestOptionFunc): the overrides.
//
// Returns:
//   - (*RequestBuilder): the receiver, for chaining.
func (b *RequestBuilder) WithOptions(ofs ...RequestOptionFunc) *RequestBuilder {
	b.requestOptions = append(b.requestOptions, ofs...)

	return b
}

// Do builds the request and executes it through the client, honoring the
// caller's context. See [Client.Do] for the execution semantics and outcomes.
//
// Parameters:
//   - ctx (context.Context): the context governing the whole execution.
//
// Returns:
//   - res (*http.Response): the final response; the caller owns and must
//     close its body.
//   - err (error): non-nil when building or executing the request fails.
func (b *RequestBuilder) Do(ctx context.Context) (res *http.Response, err error) {
	req, err := b.build(ctx)
	if err != nil {
		return
	}

	res, err = b.client.Do(ctx, req, b.requestOptions...)

	return
}

// build assembles the request: it resolves method and URL, merges the query
// parameters, applies the client's default headers and User-Agent, then the
// request's own headers.
//
// Parameters:
//   - ctx (context.Context): the context the request is built with.
//
// Returns:
//   - req (*hqgohttprequest.Request): the built request, with a reusable
//     body.
//   - err (error): non-nil when the URL cannot be resolved or the request
//     cannot be constructed.
func (b *RequestBuilder) build(ctx context.Context) (req *hqgohttprequest.Request, err error) {
	rawURL := b.rawURL
	if rawURL == "" {
		rawURL = b.path
	}

	baseURL := b.baseURL
	if baseURL == "" {
		baseURL = b.client.baseURL
	}

	if baseURL != "" && !strings.Contains(rawURL, "://") {
		rawURL, err = url.JoinPath(baseURL, rawURL)
		if err != nil {
			err = fmt.Errorf("http.RequestBuilder.Do: joining base URL: %w", err)

			return
		}
	}

	if rawURL == "" {
		err = errors.New("http.RequestBuilder.Do: no URL configured")

		return
	}

	method := b.method
	if method == "" {
		method = hqgohttpmethod.GET
	}

	req, err = hqgohttprequest.New(ctx, method.String(), rawURL, b.body)
	if err != nil {
		err = fmt.Errorf("http.RequestBuilder.Do: building request: %w", err)

		return
	}

	if req.URL.Host == "" {
		err = fmt.Errorf("http.RequestBuilder.Do: URL %q has no host", req.URL.Redacted())

		return
	}

	if len(b.params) > 0 {
		query := req.URL.Query()

		for key, values := range b.params {
			for _, value := range values {
				query.Add(key, value)
			}
		}

		req.URL.RawQuery = query.Encode()
	}

	for _, h := range b.client.headers {
		if !b.setsHeader(h.key) {
			req.Header.Set(h.key, h.value)
		}
	}

	for _, h := range b.headersSet {
		req.Header.Set(h.key, h.value)
	}

	for _, h := range b.headersAdd {
		req.Header.Add(h.key, h.value)
	}

	if !b.setsHeader(hqgohttpheader.UserAgent.String()) {
		userAgent := b.client.userAgent

		if b.client.randomUserAgent {
			userAgent = hqgohttpuseragent.Random()
		}

		req.Header.Set(hqgohttpheader.UserAgent.String(), userAgent)
	}

	return
}

// setsHeader reports whether the request itself sets a header with the given
// key, so a client default must not overwrite it.
//
// Parameters:
//   - key (string): the header field name, matched case-insensitively.
//
// Returns:
//   - sets (bool): true when the request sets the key.
func (b *RequestBuilder) setsHeader(key string) (sets bool) {
	for _, h := range b.headersSet {
		if strings.EqualFold(h.key, key) {
			sets = true

			return
		}
	}

	for _, h := range b.headersAdd {
		if strings.EqualFold(h.key, key) {
			sets = true

			return
		}
	}

	return
}
