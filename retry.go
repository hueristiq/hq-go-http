package http

import (
	"context"
	"crypto/x509"
	"errors"
	"net/url"
	"regexp"
)

// RetryPolicy defines a function type that determines if an HTTP request should be retried.
// It is invoked after each request attempt, passing the request's context and any encountered error.
// The function returns a boolean indicating whether the request should be retried,
// and a secondary error value that, if non-nil, overrides the original error and terminates further retry attempts.
//
// Parameters:
//   - ctx (context.Context): The request's context, carrying cancellation signals and deadlines.
//   - err (error): The error encountered during the HTTP request, or nil if the request succeeded.
//
// Returns:
//   - retry (bool): True if the request should be retried; false otherwise.
//   - errr (error): An error to override the original error, typically when a non-retryable condition is met.
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
// meaning that the request may be retried. It examines both the request context and the error details,
// filtering out conditions such as context cancellation, excessive redirects, unsupported protocol schemes,
// or TLS certificate verification failures (e.g., unknown authority).
//
// The function first checks the context for cancellation or deadline expiration. If the context
// has an error, it immediately returns that error, as the request should not be retried.
// It then inspects the error, particularly if it is of type *url.Error, and checks for specific error
// conditions using regular expressions and type assertions. If any non-retryable condition is detected,
// the error is returned and further retries are prevented.
//
// Parameters:
//   - ctx (context.Context): The request's context containing cancellation signals or deadlines.
//   - err (error): The error encountered during the HTTP request.
//
// Returns:
//   - recoverable (bool): True if the error is considered recoverable and the request may be retried;
//     false if the error is non-retryable.
//   - errr (error): An error to override the original error when a non-retryable condition is detected,
//     such as a cancelled context or a specific transport error.
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
