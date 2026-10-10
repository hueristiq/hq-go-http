package raw_test

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgohttpraw "github.com/hueristiq/hq-lib-http-go/raw"
)

// serveRawTCP accepts a single connection, captures the raw request bytes up
// to the end of the headers, and replies with the canned response.
func serveRawTCP(t *testing.T, response string) (address string, received <-chan string) {
	t.Helper()

	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")

	require.NoError(t, err)

	t.Cleanup(func() {
		_ = listener.Close()
	})

	captured := make(chan string, 1)

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}

		defer func() { _ = conn.Close() }()

		var request strings.Builder

		reader := bufio.NewReader(conn)

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}

			request.WriteString(line)

			if line == "\r\n" {
				break
			}
		}

		captured <- request.String()

		_, _ = io.WriteString(conn, response)
	}()

	return listener.Addr().String(), captured
}

func TestDo_RoundTrip(t *testing.T) {
	t.Parallel()

	address, received := serveRawTCP(t, "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 5\r\n\r\nhello")

	request := []byte("GET /path?q=1 HTTP/1.1\r\nHost: " + address + "\r\nX-Custom: raw\r\n\r\n")

	res, err := hqgohttpraw.Do(t.Context(), address, request)

	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "text/plain", res.Header.Get("Content-Type"))

	body, err := io.ReadAll(res.Body)

	require.NoError(t, err)
	assert.Equal(t, "hello", string(body))

	require.NoError(t, res.Body.Close())

	assert.Equal(t, string(request), <-received,
		"Expected the server to receive the request bytes exactly as given")
}

func TestDo_HeadResponseHasNoBody(t *testing.T) {
	t.Parallel()

	address, _ := serveRawTCP(t, "HTTP/1.1 200 OK\r\nContent-Length: 5\r\n\r\n")

	request := []byte("HEAD / HTTP/1.1\r\nHost: " + address + "\r\n\r\n")

	res, err := hqgohttpraw.Do(t.Context(), address, request)

	require.NoError(t, err)
	require.NotNil(t, res)

	body, err := io.ReadAll(res.Body)

	require.NoError(t, err)
	assert.Empty(t, body)

	require.NoError(t, res.Body.Close())
}

func TestDo_TLS(t *testing.T) {
	t.Parallel()

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("tls hello"))
	}))

	t.Cleanup(server.Close)

	certPool := x509.NewCertPool()
	certPool.AddCert(server.Certificate())

	tlsConfig := &tls.Config{
		RootCAs:    certPool,
		MinVersion: tls.VersionTLS12,
	}

	serverURL, err := url.Parse(server.URL)

	require.NoError(t, err)

	request := []byte("GET / HTTP/1.1\r\nHost: " + serverURL.Host + "\r\n\r\n")

	res, err := hqgohttpraw.Do(t.Context(), serverURL.Host, request, hqgohttpraw.WithTLSConfig(tlsConfig))

	require.NoError(t, err)
	require.NotNil(t, res)

	body, err := io.ReadAll(res.Body)

	require.NoError(t, err)
	assert.Equal(t, "tls hello", string(body))

	require.NoError(t, res.Body.Close())
}

func TestDo_Validation(t *testing.T) {
	t.Parallel()

	t.Run("empty address", func(t *testing.T) {
		t.Parallel()

		res, err := hqgohttpraw.Do(t.Context(), "", []byte("GET / HTTP/1.1\r\n\r\n"))

		require.Error(t, err)

		if res != nil {
			require.NoError(t, res.Body.Close())
		}

		assert.Nil(t, res)
		assert.Contains(t, err.Error(), "raw.Do:")
	})

	t.Run("empty request", func(t *testing.T) {
		t.Parallel()

		res, err := hqgohttpraw.Do(t.Context(), "127.0.0.1:80", nil)

		require.Error(t, err)

		if res != nil {
			require.NoError(t, res.Body.Close())
		}

		assert.Nil(t, res)
		assert.Contains(t, err.Error(), "raw.Do:")
	})
}

func TestDo_DialFailure(t *testing.T) {
	t.Parallel()

	// Port 1 on the loopback interface is reserved and refuses connections.
	res, err := hqgohttpraw.Do(t.Context(), "127.0.0.1:1", []byte("GET / HTTP/1.1\r\n\r\n"),
		hqgohttpraw.WithTimeout(time.Second))

	require.Error(t, err)

	if res != nil {
		require.NoError(t, res.Body.Close())
	}

	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "raw.Do:")
}

func TestDo_ContextDeadline(t *testing.T) {
	t.Parallel()

	// The server accepts the connection but never answers.
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")

	require.NoError(t, err)

	t.Cleanup(func() {
		_ = listener.Close()
	})

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}

		defer func() { _ = conn.Close() }()

		_, _ = io.Copy(io.Discard, conn)
	}()

	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)

	defer cancel()

	start := time.Now()

	res, err := hqgohttpraw.Do(ctx, listener.Addr().String(), []byte("GET / HTTP/1.1\r\nHost: x\r\n\r\n"))

	require.Error(t, err)

	if res != nil {
		require.NoError(t, res.Body.Close())
	}

	assert.Nil(t, res)
	assert.Less(t, time.Since(start), 5*time.Second,
		"Expected the context deadline to bound the exchange")
}
