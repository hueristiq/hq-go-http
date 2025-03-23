package http

import (
	"context"
	"crypto/x509"
	"errors"
	"net/url"
	"regexp"
)

// RetryPolicy defines a function type that determines if an HTTP request should be retried.
// It is invoked after each request attempt with the request's context and any encountered error.
// If the function returns false, no further retries are attempted. Additionally, a non-nil error
// return value overrides the original error, terminating further retry attempts.
//
// Parameters:
//   - ctx (context.Context): The context for the request, containing cancellation signals and deadlines.
//   - err (error): The error encountered during the HTTP request, or nil if the request succeeded.
//
// Returns:
//   - retry (bool): True if the request should be retried; false if it should not.
//   - errr (error): An error to override the original error, typically when a non-retryable error is encountered.
type RetryPolicy func(ctx context.Context, err error) (retry bool, errr error)

var (
	// redirectsErrorRegex matches error strings that indicate the maximum number of redirects was exceeded.
	// It is used to avoid retrying requests that have failed due to too many redirects.
	redirectsErrorRegex = regexp.MustCompile(`stopped after \d+ redirects\z`)

	// schemeErrorRegex matches error strings indicating an unsupported protocol scheme.
	// This helps in identifying errors that should not be retried.
	schemeErrorRegex = regexp.MustCompile(`unsupported protocol scheme`)
)

// isErrorRecoverable determines whether an error encountered during an HTTP request is recoverable,
// meaning that the request may be retried. It examines the request context and the error details,
// filtering out errors such as context cancellations, excessive redirects, unsupported protocol schemes,
// or TLS certificate verification failures.
//
// Parameters:
//   - ctx (context.Context): The request's context, which may contain cancellation signals or deadlines.
//   - err (error): The error encountered during the HTTP request.
//
// Returns:
//   - recoverable (bool): True if the error is considered recoverable (the request may be retried); otherwise, false.
//   - errr (error): An error value to override the original error in case the context is cancelled or a non-retryable error is detected.
func isErrorRecoverable(ctx context.Context, err error) (recoverable bool, errr error) {
	if ctx.Err() != nil {
		errr = ctx.Err()

		return
	}

	var URLError *url.Error

	if err != nil && errors.As(err, &URLError) {
		if redirectsErrorRegex.MatchString(err.Error()) {
			errr = err

			return
		}

		if schemeErrorRegex.MatchString(err.Error()) {
			errr = err

			return
		}

		var UnknownAuthorityError x509.UnknownAuthorityError

		if errors.As(err, &UnknownAuthorityError) {
			errr = err

			return
		}
	}

	if err != nil {
		recoverable = true

		return
	}

	return
}
