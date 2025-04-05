package request_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/hueristiq/hq-go-http/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("ReusableReadCloser", func(t *testing.T) {
		t.Parallel()

		rr, err := request.NewReusableReadCloser("value body")

		require.NoError(t, err)
		require.NotNil(t, rr)

		req, err := request.New("POST", "http://example.com", *rr) //nolint:govet,copylocks

		require.NoError(t, err)
		require.NotNil(t, req)
		require.NotNil(t, req.Body)

		expected := "value body"

		require.Equal(t, int64(len(expected)), req.ContentLength)

		buf := make([]byte, len(expected))

		n, err := req.Body.Read(buf)

		require.NoError(t, err)

		assert.Equal(t, len(expected), n)
		assert.Equal(t, expected, string(buf[:n]))
	})

	t.Run("*ReusableReadCloser", func(t *testing.T) {
		t.Parallel()

		rr, err := request.NewReusableReadCloser("pointer body")

		require.NoError(t, err)
		require.NotNil(t, rr)

		req, err := request.New("POST", "http://example.com", rr)

		require.NoError(t, err)
		require.NotNil(t, req)
		require.NotNil(t, req.Body)

		expected := "pointer body"

		require.Equal(t, int64(len(expected)), req.ContentLength)

		buf := make([]byte, len(expected))

		n, err := req.Body.Read(buf)

		require.NoError(t, err)

		assert.Equal(t, len(expected), n)
		assert.Equal(t, expected, string(buf[:n]))
	})

	t.Run("func() (io.Reader, error)", func(t *testing.T) {
		t.Parallel()

		bodyFunc := func() (io.Reader, error) {
			return strings.NewReader("function body"), nil
		}

		req, err := request.New("POST", "http://example.com", bodyFunc)

		require.NoError(t, err)
		require.NotNil(t, req)
		require.NotNil(t, req.Body)

		expected := "function body"

		require.Equal(t, int64(len(expected)), req.ContentLength)

		buf := make([]byte, len(expected))

		n, err := req.Body.Read(buf)

		require.NoError(t, err)

		assert.Equal(t, len(expected), n)
		assert.Equal(t, expected, string(buf[:n]))
	})

	t.Run("nil body", func(t *testing.T) {
		t.Parallel()

		req, err := request.New("GET", "http://example.com", nil)

		require.NoError(t, err)
		require.NotNil(t, req)
		require.Nil(t, req.Body)
		require.Equal(t, int64(0), req.ContentLength)
	})

	t.Run("string body", func(t *testing.T) {
		t.Parallel()

		bodyStr := "example body"

		req, err := request.New("POST", "http://example.com", bodyStr)

		require.NoError(t, err)
		require.NotNil(t, req)
		require.NotNil(t, req.Body)
		require.Equal(t, int64(len(bodyStr)), req.ContentLength)

		buf := make([]byte, len(bodyStr))

		n, err := req.Body.Read(buf)

		require.NoError(t, err)

		assert.Equal(t, len(bodyStr), n)
		assert.Equal(t, bodyStr, string(buf[:n]))

		buf2 := make([]byte, len(bodyStr))

		n, err = req.Body.Read(buf2)

		require.NoError(t, err)

		assert.Equal(t, len(bodyStr), n)
		assert.Equal(t, bodyStr, string(buf2[:n]))
	})

	t.Run("[]byte body", func(t *testing.T) {
		t.Parallel()

		bodyBytes := []byte("byte body")

		req, err := request.New("POST", "http://example.com", bodyBytes)

		require.NoError(t, err)
		require.NotNil(t, req)
		require.NotNil(t, req.Body)

		expected := "byte body"

		require.Equal(t, int64(len(expected)), req.ContentLength)

		buf := make([]byte, len(expected))

		n, err := req.Body.Read(buf)

		require.NoError(t, err)

		assert.Equal(t, len(expected), n)
		assert.Equal(t, expected, string(buf[:n]))
	})

	t.Run("unsupported body type", func(t *testing.T) {
		t.Parallel()

		req, err := request.New("POST", "http://example.com", 123)

		require.Error(t, err)
		require.Nil(t, req)
	})
}

func TestNewWithContext(t *testing.T) {
	t.Parallel()

	type Key string

	ctx := context.WithValue(t.Context(), Key("key"), "value")

	req, err := request.NewWithContext(ctx, "GET", "http://example.com", nil)

	require.NoError(t, err)
	require.NotNil(t, req)

	val := req.Request.Context().Value(Key("key"))

	require.Equal(t, "value", val)
}
