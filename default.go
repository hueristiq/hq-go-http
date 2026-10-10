package http

import (
	"context"
	"net"
	"net/http"
	"runtime"
	"sync"
	"time"

	"golang.org/x/net/http2"
)

// newFallbackHTTPClient builds the fallback client used when a cleartext
// HTTP/1.x exchange fails because the server answered with HTTP/2 frames. It
// is declared as a package-level variable so tests can substitute a client
// with a scripted transport.
var newFallbackHTTPClient = func() (client *http.Client, err error) {
	transport := DefaultHTTPTransport()

	if err = http2.ConfigureTransport(transport); err != nil {
		return
	}

	client = &http.Client{
		Transport: transport,
	}

	return
}

// defaultClient is the shared client behind the package-level shortcuts,
// built lazily on first use. Construction errors — practically impossible,
// since only the HTTP/2 fallback configuration can fail, on a transport built
// from scratch — are not swallowed: every shortcut reports them.
var defaultClient = sync.OnceValues(func() (client *Client, err error) {
	client, err = NewClient()

	return
})

// DefaultHTTPPooledTransport returns a new *http.Transport configured for
// connection pooling, with HTTP/2 support negotiated automatically.
//
// It is optimized for clients making repeated requests to the same hosts,
// allowing efficient reuse of TCP connections.
//
// Warning: the transport is intended for long-lived clients. For transient
// clients, consider [DefaultHTTPTransport] to avoid leaking file descriptors.
//
// Returns:
//   - transport (*http.Transport): a newly configured pooled transport.
func DefaultHTTPPooledTransport() (transport *http.Transport) {
	transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
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

// DefaultHTTPTransport returns a new *http.Transport derived from
// [DefaultHTTPPooledTransport] but with keep-alives and idle connections
// disabled, so no connection is ever reused.
//
// It suits transient clients and host-spraying scenarios, where connection
// reuse across changing hosts is unwanted.
//
// Returns:
//   - transport (*http.Transport): a newly configured non-pooled transport.
func DefaultHTTPTransport() (transport *http.Transport) {
	transport = DefaultHTTPPooledTransport()

	transport.DisableKeepAlives = true
	transport.MaxIdleConnsPerHost = -1

	return
}

// DefaultHTTPPooledClient returns a new *http.Client on a
// [DefaultHTTPPooledTransport].
//
// Returns:
//   - client (*http.Client): a newly created pooled client.
func DefaultHTTPPooledClient() (client *http.Client) {
	client = &http.Client{
		Transport: DefaultHTTPPooledTransport(),
	}

	return
}

// DefaultHTTPClient returns a new *http.Client on a [DefaultHTTPTransport],
// so no connection is ever reused.
//
// Returns:
//   - client (*http.Client): a newly created non-pooled client.
func DefaultHTTPClient() (client *http.Client) {
	client = &http.Client{
		Transport: DefaultHTTPTransport(),
	}

	return
}

// Get performs an HTTP GET request through the shared default client.
//
// Parameters:
//   - ctx (context.Context): the context governing the whole execution.
//   - rawURL (string): the target URL.
//   - ofs (...RequestOptionFunc): per-request overrides of the execution
//     knobs.
//
// Returns:
//   - res (*http.Response): the response received from the server.
//   - err (error): non-nil when the default client cannot be constructed or
//     the request fails.
func Get(ctx context.Context, rawURL string, ofs ...RequestOptionFunc) (res *http.Response, err error) {
	client, err := defaultClient()
	if err != nil {
		return
	}

	res, err = client.Get(ctx, rawURL, ofs...)

	return
}

// Head performs an HTTP HEAD request through the shared default client.
//
// Parameters:
//   - ctx (context.Context): the context governing the whole execution.
//   - rawURL (string): the target URL.
//   - ofs (...RequestOptionFunc): per-request overrides of the execution
//     knobs.
//
// Returns:
//   - res (*http.Response): the response received from the server.
//   - err (error): non-nil when the default client cannot be constructed or
//     the request fails.
func Head(ctx context.Context, rawURL string, ofs ...RequestOptionFunc) (res *http.Response, err error) {
	client, err := defaultClient()
	if err != nil {
		return
	}

	res, err = client.Head(ctx, rawURL, ofs...)

	return
}

// Post performs an HTTP POST request with a body through the shared default
// client. See [Client.Post] for the accepted body types.
//
// Parameters:
//   - ctx (context.Context): the context governing the whole execution.
//   - rawURL (string): the target URL.
//   - body (any): the request payload; nil means no body.
//   - ofs (...RequestOptionFunc): per-request overrides of the execution
//     knobs.
//
// Returns:
//   - res (*http.Response): the response received from the server.
//   - err (error): non-nil when the default client cannot be constructed or
//     the request fails.
func Post(ctx context.Context, rawURL string, body any, ofs ...RequestOptionFunc) (res *http.Response, err error) {
	client, err := defaultClient()
	if err != nil {
		return
	}

	res, err = client.Post(ctx, rawURL, body, ofs...)

	return
}

// Put performs an HTTP PUT request with a body through the shared default
// client. See [Client.Post] for the accepted body types.
//
// Parameters:
//   - ctx (context.Context): the context governing the whole execution.
//   - rawURL (string): the target URL.
//   - body (any): the request payload; nil means no body.
//   - ofs (...RequestOptionFunc): per-request overrides of the execution
//     knobs.
//
// Returns:
//   - res (*http.Response): the response received from the server.
//   - err (error): non-nil when the default client cannot be constructed or
//     the request fails.
func Put(ctx context.Context, rawURL string, body any, ofs ...RequestOptionFunc) (res *http.Response, err error) {
	client, err := defaultClient()
	if err != nil {
		return
	}

	res, err = client.Put(ctx, rawURL, body, ofs...)

	return
}

// Patch performs an HTTP PATCH request with a body through the shared default
// client. See [Client.Post] for the accepted body types.
//
// Parameters:
//   - ctx (context.Context): the context governing the whole execution.
//   - rawURL (string): the target URL.
//   - body (any): the request payload; nil means no body.
//   - ofs (...RequestOptionFunc): per-request overrides of the execution
//     knobs.
//
// Returns:
//   - res (*http.Response): the response received from the server.
//   - err (error): non-nil when the default client cannot be constructed or
//     the request fails.
func Patch(ctx context.Context, rawURL string, body any, ofs ...RequestOptionFunc) (res *http.Response, err error) {
	client, err := defaultClient()
	if err != nil {
		return
	}

	res, err = client.Patch(ctx, rawURL, body, ofs...)

	return
}

// Delete performs an HTTP DELETE request through the shared default client.
//
// Parameters:
//   - ctx (context.Context): the context governing the whole execution.
//   - rawURL (string): the target URL.
//   - ofs (...RequestOptionFunc): per-request overrides of the execution
//     knobs.
//
// Returns:
//   - res (*http.Response): the response received from the server.
//   - err (error): non-nil when the default client cannot be constructed or
//     the request fails.
func Delete(ctx context.Context, rawURL string, ofs ...RequestOptionFunc) (res *http.Response, err error) {
	client, err := defaultClient()
	if err != nil {
		return
	}

	res, err = client.Delete(ctx, rawURL, ofs...)

	return
}

// Options performs an HTTP OPTIONS request through the shared default client.
//
// Parameters:
//   - ctx (context.Context): the context governing the whole execution.
//   - rawURL (string): the target URL.
//   - ofs (...RequestOptionFunc): per-request overrides of the execution
//     knobs.
//
// Returns:
//   - res (*http.Response): the response received from the server.
//   - err (error): non-nil when the default client cannot be constructed or
//     the request fails.
func Options(ctx context.Context, rawURL string, ofs ...RequestOptionFunc) (res *http.Response, err error) {
	client, err := defaultClient()
	if err != nil {
		return
	}

	res, err = client.Options(ctx, rawURL, ofs...)

	return
}
