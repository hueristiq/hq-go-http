package http

import (
	"context"
	"net"
	"net/http"
	"runtime"
	"time"
)

var (
	// DefaultClient is a package-level default Client instance that is automatically
	// initialized during package startup in the init() function. It uses the
	// DefaultSingleClientConfiguration, which is suitable for standard connection pooling.
	DefaultClient *Client

	// DefaultSingleClientConfiguration defines a default configuration for a standard HTTP client.
	// It is intended for scenarios where connection pooling is acceptable.
	DefaultSingleClientConfiguration = &ClientConfiguration{
		Timeout:              30 * time.Second,
		CloseIdleConnections: false,
		RetryMax:             3,
		RetryWaitMin:         1 * time.Second,
		RetryWaitMax:         30 * time.Second,
		RespReadLimit:        4096,
	}

	// DefaultSprayingClientConfiguration defines a default configuration for scenarios such as host spraying,
	// where killing idle connections is desirable to reduce resource usage.
	DefaultSprayingClientConfiguration = &ClientConfiguration{
		Timeout:              30 * time.Second,
		CloseIdleConnections: true,
		RetryMax:             3,
		RetryWaitMin:         1 * time.Second,
		RetryWaitMax:         30 * time.Second,
		RespReadLimit:        4096,
	}
)

// init is executed during package initialization. It creates a default Client instance
// using the DefaultSingleClientConfiguration. Any errors during client creation are ignored.
func init() {
	DefaultClient, _ = NewClient(DefaultSingleClientConfiguration)
}

// DefaultHTTPPooledClient returns a new *http.Client configured with a shared transport that supports
// connection pooling. It is optimized for clients that make repeated requests to the same host,
// allowing efficient reuse of TCP connections.
//
// Warning: This client should be used for long-lived operations. For short-lived operations,
// consider using DefaultHTTPClient to avoid leaking file descriptors.
//
// Returns:
//   - client (*http.Client): A pointer to a newly created pooled http.Client.
func DefaultHTTPPooledClient() (client *http.Client) {
	client = &http.Client{
		Transport: DefaultHTTPPooledTransport(),
	}

	return
}

// DefaultHTTPPooledTransport returns a new *http.Transport configured for connection pooling.
// It sets various parameters such as timeouts, keep-alives, and idle connection limits to
// optimize connection reuse for repeated requests.
//
// Warning: This transport is intended for long-lived clients. Using it for transient clients
// may lead to file descriptor leaks.
//
// Returns:
//   - transport (*http.Transport): A pointer to a newly configured pooled http.Transport.
func DefaultHTTPPooledTransport() (transport *http.Transport) {
	transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
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

// DefaultHTTPTransport returns a new *http.Transport configured to disable idle connections
// and keep-alives. It is derived from DefaultHTTPPooledTransport but modified to avoid
// connection reuse, making it suitable for transient requests.
//
// Returns:
//   - transport (*http.Transport): A pointer to a newly configured http.Transport with idle connections disabled.
func DefaultHTTPTransport() (transport *http.Transport) {
	transport = DefaultHTTPPooledTransport()

	transport.DisableKeepAlives = true
	transport.MaxIdleConnsPerHost = -1

	return
}

// DefaultHTTPClient returns a new *http.Client configured with a non-shared transport.
// Idle connections and keep-alives are disabled, making it suitable for transient requests
// where connection reuse is not desired.
//
// Returns:
//   - client (*http.Client): A pointer to a newly created http.Client with a non-pooled transport.
func DefaultHTTPClient() (client *http.Client) {
	client = &http.Client{
		Transport: DefaultHTTPTransport(),
	}

	return
}

// DefaultRetryPolicy returns a default RetryPolicy function that determines if a request should be retried.
// It bases its decision on whether the encountered error is recoverable by delegating to isErrorRecoverable.
//
// Returns:
//   - A RetryPolicy function that accepts a context and error, and returns a boolean indicating retry and an error.
func DefaultRetryPolicy() func(ctx context.Context, err error) (retry bool, errr error) {
	return isErrorRecoverable
}

// HostSprayRetryPolicy returns a RetryPolicy function tailored for scenarios where multiple hosts are being
// targeted concurrently (host spraying). Currently, it delegates to the same isErrorRecoverable function as the default.
//
// Returns:
//   - A RetryPolicy function suitable for host spraying scenarios.
func HostSprayRetryPolicy() func(ctx context.Context, err error) (retry bool, errr error) {
	return isErrorRecoverable
}

// Get performs an HTTP GET request using the DefaultClient.
// It is a shortcut for DefaultClient.Get.
//
// Parameters:
//   - URL (string): The target URL for the GET request.
//   - configurations (...*RequestConfiguration): Optional overrides for the request configuration.
//
// Returns:
//   - res (*http.Response): The response received from the server.
//   - err (error): An error if the request fails.
func Get(URL string, configurations ...*RequestConfiguration) (res *http.Response, err error) {
	return DefaultClient.Get(URL, configurations...)
}

// Head performs an HTTP HEAD request using the DefaultClient.
// It is a shortcut for DefaultClient.Head.
//
// Parameters:
//   - URL (string): The target URL for the HEAD request.
//   - configurations (...*RequestConfiguration): Optional overrides for the request configuration.
//
// Returns:
//   - res (*http.Response): The response received from the server.
//   - err (error): An error if the request fails.
func Head(URL string, configurations ...*RequestConfiguration) (res *http.Response, err error) {
	return DefaultClient.Head(URL, configurations...)
}

// Put performs an HTTP PUT request using the DefaultClient.
// It is a shortcut for DefaultClient.Put.
//
// Parameters:
//   - URL (string): The target URL for the PUT request.
//   - body (interface{}): The payload to include in the PUT request.
//   - configurations (...*RequestConfiguration): Optional overrides for the request configuration.
//
// Returns:
//   - res (*http.Response): The response received from the server.
//   - err (error): An error if the request fails.
func Put(URL string, body interface{}, configurations ...*RequestConfiguration) (res *http.Response, err error) {
	return DefaultClient.Put(URL, body, configurations...)
}

// Delete performs an HTTP DELETE request using the DefaultClient.
// It is a shortcut for DefaultClient.Delete.
//
// Parameters:
//   - URL (string): The target URL for the DELETE request.
//   - configurations (...*RequestConfiguration): Optional overrides for the request configuration.
//
// Returns:
//   - res (*http.Response): The response received from the server.
//   - err (error): An error if the request fails.
func Delete(URL string, configurations ...*RequestConfiguration) (res *http.Response, err error) {
	return DefaultClient.Delete(URL, configurations...)
}

// Post performs an HTTP POST request using the DefaultClient.
// It is a shortcut for DefaultClient.Post.
//
// Parameters:
//   - URL (string): The target URL for the POST request.
//   - body (interface{}): The payload to include in the POST request.
//   - configurations (...*RequestConfiguration): Optional overrides for the request configuration.
//
// Returns:
//   - res (*http.Response): The response received from the server.
//   - err (error): An error if the request fails.
func Post(URL string, body interface{}, configurations ...*RequestConfiguration) (res *http.Response, err error) {
	return DefaultClient.Post(URL, body, configurations...)
}

// Options performs an HTTP OPTIONS request using the DefaultClient.
// It is a shortcut for DefaultClient.Options.
//
// Parameters:
//   - URL (string): The target URL for the OPTIONS request.
//   - configurations (...*RequestConfiguration): Optional overrides for the request configuration.
//
// Returns:
//   - res (*http.Response): The response received from the server.
//   - err (error): An error if the request fails.
func Options(URL string, configurations ...*RequestConfiguration) (res *http.Response, err error) {
	return DefaultClient.Options(URL, configurations...)
}
