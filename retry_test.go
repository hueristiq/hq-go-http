package http_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgohttp "github.com/hueristiq/hq-go-http"
	hqgohttpstatus "github.com/hueristiq/hq-go-http/status"
)

func TestDefaultRetryPolicy(t *testing.T) {
	t.Parallel()

	errConnection := errors.New("connection refused")

	tests := []struct {
		name      string
		ctx       func(t *testing.T) context.Context
		resp      *http.Response
		err       error
		wantRetry bool
		wantErr   bool
	}{
		{
			name: "connection error is retried",
			err: &url.Error{
				Op:  "Get",
				URL: "http://example.com",
				Err: errConnection,
			},
			wantRetry: true,
		},
		{
			name:      "bare transport error is retried",
			err:       errConnection,
			wantRetry: true,
		},
		{
			name: "canceled context is not retried",
			ctx: func(t *testing.T) context.Context {
				t.Helper()

				ctx, cancel := context.WithCancel(t.Context())

				cancel()

				return ctx
			},
			err:       context.Canceled,
			wantRetry: false,
			wantErr:   true,
		},
		{
			name: "exhausted redirect limit is not retried",
			err: &url.Error{
				Op:  "Get",
				URL: "http://example.com",
				Err: errors.New("stopped after 10 redirects"),
			},
			wantRetry: false,
		},
		{
			name: "unsupported scheme is not retried",
			err: &url.Error{
				Op:  "Get",
				URL: "gopher://example.com",
				Err: errors.New(`unsupported protocol scheme "gopher"`),
			},
			wantRetry: false,
		},
		{
			name: "unknown authority is not retried",
			err: &url.Error{
				Op:  "Get",
				URL: "https://example.com",
				Err: x509.UnknownAuthorityError{},
			},
			wantRetry: false,
		},
		{
			name: "certificate verification failure is not retried",
			err: &url.Error{
				Op:  "Get",
				URL: "https://example.com",
				Err: &tls.CertificateVerificationError{},
			},
			wantRetry: false,
		},
		{
			name:      "429 is retried",
			resp:      &http.Response{StatusCode: hqgohttpstatus.TooManyRequests.Int()},
			wantRetry: true,
		},
		{
			name:      "500 is retried",
			resp:      &http.Response{StatusCode: hqgohttpstatus.InternalServerError.Int()},
			wantRetry: true,
		},
		{
			name:      "503 is retried",
			resp:      &http.Response{StatusCode: hqgohttpstatus.ServiceUnavailable.Int()},
			wantRetry: true,
		},
		{
			name:      "501 is not retried",
			resp:      &http.Response{StatusCode: hqgohttpstatus.NotImplemented.Int()},
			wantRetry: false,
		},
		{
			name:      "zero status is retried",
			resp:      &http.Response{StatusCode: 0},
			wantRetry: true,
		},
		{
			name:      "200 is not retried",
			resp:      &http.Response{StatusCode: hqgohttpstatus.OK.Int()},
			wantRetry: false,
		},
		{
			name:      "404 is not retried",
			resp:      &http.Response{StatusCode: hqgohttpstatus.NotFound.Int()},
			wantRetry: false,
		},
		{
			name:      "nil response and nil error is not retried",
			wantRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			if tt.ctx != nil {
				ctx = tt.ctx(t)
			}

			retry, rerr := hqgohttp.DefaultRetryPolicy(ctx, tt.resp, tt.err)

			assert.Equal(t, tt.wantRetry, retry)

			if tt.wantErr {
				require.Error(t, rerr)
			} else {
				assert.NoError(t, rerr)
			}
		})
	}
}

func TestDefaultRetryPolicy_ReturnsContextError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())

	cancel()

	retry, rerr := hqgohttp.DefaultRetryPolicy(ctx, nil, fmt.Errorf("wrapped: %w", context.Canceled))

	assert.False(t, retry)
	assert.ErrorIs(t, rerr, context.Canceled)
}
