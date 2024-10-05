package http

import (
	"net"
	"net/http"
	"net/url"
	"runtime"
	"time"
)

var DefaultClient *Client

var DefaultSingleClientConfiguration = &ClientConfiguration{
	RetryMax:        5,
	RetryWaitMin:    1 * time.Second,
	RetryWaitMax:    30 * time.Second,
	Timeout:         30 * time.Second,
	RespReadLimit:   4096,
	KillIdleConn:    false,
	NoAdjustTimeout: true,
}

var DefaultSprayingClientConfiguration = &ClientConfiguration{
	RetryMax:        5,
	RetryWaitMin:    1 * time.Second,
	RetryWaitMax:    30 * time.Second,
	Timeout:         30 * time.Second,
	RespReadLimit:   4096,
	KillIdleConn:    true,
	NoAdjustTimeout: true,
}

func init() {
	DefaultClient, _ = NewClient(DefaultSingleClientConfiguration)
}

// DefaultHTTPTransport returns a new http.Transport with similar default values to
// http.DefaultTransport, but with idle connections and keepalives disabled.
// It does this by first creating a transport with pooled connections
// (by calling DefaultHTTPPooledTransport) and then setting DisableKeepAlives
// to true and MaxIdleConnsPerHost to -1.
func DefaultHTTPTransport() (transport *http.Transport) {
	transport = DefaultHTTPPooledTransport()

	transport.DisableKeepAlives = true
	transport.MaxIdleConnsPerHost = -1

	return
}

// DefaultHTTPPooledTransport returns a new http.Transport with similar default
// values to http.DefaultTransport, but with a custom configuration that is
// suitable for transports that will be reused for the same hosts. It sets various
// fields of the http.Transport struct, such as Proxy, DialContext, MaxIdleConns,
// IdleConnTimeout, TLSHandshakeTimeout, ExpectContinueTimeout, ForceAttemptHTTP2, and
// MaxIdleConnsPerHost.
//
// Do not use this for transient transports as it can leak file descriptors over
// time. Only use this for transports that will be re-used for the same host(s).
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

// DefaultHTTPClient returns a new http.Client with similar default values to
// http.Client, but with a non-shared transport, idle connections disabled, and
// keep-alives disabled. It does this by setting the Transport field of the http.Client
// struct to the transport returned by DefaultHTTPTransport.
func DefaultHTTPClient() (client *http.Client) {
	client = &http.Client{
		Transport: DefaultHTTPTransport(),
	}

	return
}

// DefaultPooledClient returns a new http.Client with similar default values to
// http.Client, but with a shared transport. It sets the Transport field of the
// http.Client struct to the transport returned by DefaultHTTPPooledTransport.
//
// Do not use this function for transient clients as it can leak file descriptors
// over time. Only use this for clients that will be re-used for the same host(s).
func DefaultPooledClient() (client *http.Client) {
	client = &http.Client{
		Transport: DefaultHTTPPooledTransport(),
	}

	return
}

func Get(URL string) (res *http.Response, err error) {
	return DefaultClient.Get(URL)
}

func Head(URL string) (res *http.Response, err error) {
	return DefaultClient.Head(URL)
}

func Post(URL, bodyType string, body interface{}) (res *http.Response, err error) {
	return DefaultClient.Post(URL, bodyType, body)
}

func PostForm(URL string, data url.Values) (res *http.Response, err error) {
	return DefaultClient.PostForm(URL, data)
}
