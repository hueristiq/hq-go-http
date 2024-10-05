package http

import (
	"context"
	"crypto/x509"
	"errors"
	"net/url"
	"regexp"
)

// CheckRetry defines a function type that determines whether a request should be retried.
// It is called following each request with the response and error values returned by
// the http.Client. If CheckRetry returns false, the Client stops retrying
// and returns the response to the caller. If CheckRetry returns an error,
// that error value is returned in lieu of the error from the request. The
// Client will close any response body when retrying, but if the retry is
// aborted it is up to the CheckRetry callback to properly close any
// response body before returning.
//
// Parameters:
//   - ctx: The request's context, which may contain deadlines or cancellation signals.
//   - err: The error encountered during the request. Can be nil if the request succeeded.
//
// Returns:
//   - retry: A boolean indicating whether the request should be retried.
//   - perr: An error if there was an issue while checking for retry logic.
type CheckRetry func(ctx context.Context, err error) (retry bool, perr error)

var (
	// redirectsErrorRegex is a regular expression to match the error returned by net/http when the
	// configured number of redirects is exhausted. This error isn't typed
	// specifically so we resort to matching on the error string.
	redirectsErrorRegex = regexp.MustCompile(`stopped after \d+ redirects\z`)

	// schemeErrorRegex is a regular expression to match the error returned by net/http when the
	// scheme specified in the URL is unsupported or invalid URL. This error isn't typed
	// specifically so we resort to matching on the error string.
	schemeErrorRegex = regexp.MustCompile(`unsupported protocol scheme`)
)

// DefaultRetryPolicy returns a function that applies a default retry policy based
// on the recoverability of the error encountered or the response status.
//
// Parameters: None.
//
// Returns:
//   - A CheckRetry function that determines if the request should be retried.
func DefaultRetryPolicy() func(ctx context.Context, err error) (retry bool, perr error) {
	return IsErrorRecoverable
}

// HostSprayRetryPolicy returns a retry policy function similar to the default one.
// This can be used in scenarios where host-spraying or distributed requests are being made.
//
// Parameters: None.
//
// Returns:
//   - A CheckRetry function that determines if the request should be retried based on recoverable errors.
func HostSprayRetryPolicy() func(ctx context.Context, err error) (retry bool, perr error) {
	return IsErrorRecoverable
}

// IsErrorRecoverable checks if an error or HTTP response can be considered recoverable,
// meaning the request could be retried.
//
// Parameters:
//   - ctx: The request's context, which may contain deadlines or cancellation signals.
//   - res: The HTTP response returned by the request. Can be nil if the request failed.
//   - target: The error encountered during the request.
//
// Returns:
//   - recoverable: A boolean indicating whether the error is recoverable and the request can be retried.
//   - err: An error if the context encountered an issue (e.g., context.Canceled or context.DeadlineExceeded).
func IsErrorRecoverable(ctx context.Context, target error) (recoverable bool, err error) {
	// Do not retry if the context has been canceled or the deadline has been exceeded
	if ctx.Err() != nil {
		err = ctx.Err()

		return
	}

	var URLError *url.Error

	if target != nil && errors.As(target, &URLError) {
		// Do not retry if the error was caused by exceeding the maximum number of redirects
		if redirectsErrorRegex.MatchString(target.Error()) {
			return
		}

		// Do not retry if the error was caused by an unsupported protocol scheme
		if schemeErrorRegex.MatchString(target.Error()) {
			return
		}

		// Do not retry if the error was caused by a TLS certificate verification failure
		var UnknownAuthorityError x509.UnknownAuthorityError

		if errors.As(target, &UnknownAuthorityError) {
			return
		}
	}

	if target != nil {
		recoverable = true

		return
	}

	// If no error is present, return false as the request did not fail
	return
}
