package http

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

// RetryPolicy decides whether a request is retried after an attempt.
//
// It is invoked after every attempt with the attempt's outcome: the response
// (nil when the transport failed before any response arrived), the error
// (nil when the round trip itself succeeded), and the request's context.
// This mirrors the CheckRetry semantics of hashicorp/go-retryablehttp.
//
// Returning retry as false ends the execution: the response and error are
// returned to the caller as they are. Returning a non-nil rerr also ends the
// execution, but the response and rerr are returned instead — use it to abort
// retries on errors the caller must see unchanged, such as context
// cancellation.
//
// When a response is retried, the client drains and closes its body so the
// connection can be reused. When a policy aborts retries on a response (retry
// false with a non-nil rerr), the response body is handed to the caller
// unclosed — the caller owns it.
//
// A policy is invoked synchronously on the goroutine running the attempt and
// must be safe for concurrent use across requests. A panicking policy
// propagates its panic to the caller of the entry point.
//
// Parameters:
//   - ctx (context.Context): the request's context, carrying cancellation
//     signals and deadlines.
//   - resp (*http.Response): the response received on the attempt, or nil.
//   - err (error): the transport error of the attempt, or nil.
//
// Returns:
//   - retry (bool): true to schedule another attempt, false to stop.
//   - rerr (error): an error replacing the attempt's outcome; non-nil stops
//     retries even when retry is true.
type RetryPolicy func(ctx context.Context, resp *http.Response, err error) (retry bool, rerr error)

var (
	// redirectsErrorRegex matches the error net/http produces when the
	// configured redirect limit is exhausted. The error is not typed
	// specifically, so matching on the message is the only way to detect it;
	// the check is isolated here so the string match lives in exactly one
	// place.
	redirectsErrorRegex = regexp.MustCompile(`stopped after \d+ redirects\z`)

	// schemeErrorRegex matches the error net/http produces for an unsupported
	// protocol scheme. The error is not typed specifically, so matching on the
	// message is the only way to detect it; the check is isolated here so the
	// string match lives in exactly one place.
	schemeErrorRegex = regexp.MustCompile(`unsupported protocol scheme`)
)

// DefaultRetryPolicy is the default [RetryPolicy]: it retries connection and
// transport errors, 429 Too Many Requests, and 5xx responses (except 501 Not
// Implemented), and it never retries a canceled context, an exhausted
// redirect limit, an unsupported protocol scheme, or a TLS certificate
// verification failure.
//
// It is the policy applied when neither [WithRetryPolicy] nor
// [WithRequestRetryPolicy] is supplied.
//
// Parameters:
//   - ctx (context.Context): the request's context, carrying cancellation
//     signals and deadlines.
//   - resp (*http.Response): the response received on the attempt, or nil.
//   - err (error): the transport error of the attempt, or nil.
//
// Returns:
//   - retry (bool): true when another attempt is worthwhile.
//   - rerr (error): the context's error when the context is done; nil
//     otherwise.
func DefaultRetryPolicy(ctx context.Context, resp *http.Response, err error) (retry bool, rerr error) {
	// Do not retry on context.Canceled or context.DeadlineExceeded: the caller
	// is no longer waiting for an answer.
	if ctx.Err() != nil {
		rerr = ctx.Err()

		return
	}

	if err != nil {
		var urlError *url.Error

		if errors.As(err, &urlError) {
			// Do not retry when the error was due to too many redirects.
			if redirectsErrorRegex.MatchString(err.Error()) {
				return
			}

			// Do not retry when the error was due to an unsupported protocol
			// scheme — retrying cannot fix a URL the client cannot speak.
			if schemeErrorRegex.MatchString(err.Error()) {
				return
			}

			// Do not retry TLS certificate verification failures: retrying a
			// certificate the caller does not trust cannot succeed.
			if isCertificateError(err) {
				return
			}
		}

		// The error is likely recoverable — connection refused, reset,
		// timeout, DNS hiccup — so retry.
		retry = true

		return
	}

	if resp != nil {
		// 429 Too Many Requests is recoverable. Servers may add a Retry-After
		// header, which the retry engine honors (capped at the maximum retry
		// wait).
		if resp.StatusCode == http.StatusTooManyRequests {
			retry = true

			return
		}

		// Retry on 500-range responses (except 501 Not Implemented, which
		// retrying cannot fix) to give the server time to recover. The range
		// check also catches invalid codes such as 0.
		if resp.StatusCode == 0 || (resp.StatusCode >= 500 && resp.StatusCode != http.StatusNotImplemented) {
			retry = true

			return
		}
	}

	return
}

// isCertificateError reports whether err wraps a TLS certificate verification
// failure. Unlike the redirect-limit and unsupported-scheme failures, these
// errors are typed, so they are matched with errors.As.
//
// Parameters:
//   - err (error): the error to inspect.
//
// Returns:
//   - ok (bool): true when err is a certificate verification failure.
func isCertificateError(err error) (ok bool) {
	var unknownAuthorityError x509.UnknownAuthorityError

	var hostnameError x509.HostnameError

	var certificateInvalidError x509.CertificateInvalidError

	var certificateVerificationError *tls.CertificateVerificationError

	ok = errors.As(err, &unknownAuthorityError) ||
		errors.As(err, &hostnameError) ||
		errors.As(err, &certificateInvalidError) ||
		errors.As(err, &certificateVerificationError)

	return
}

// retryableError marks an attempt's failure as worth another attempt. The
// retry engine retries only errors carrying this marker — an unwrapped error
// from the attempt operation is final and returned to the caller unchanged —
// and [Client.Do] unwraps the marker when the attempts run out, so the caller
// never sees it.
//
// Fields:
//   - err (error): the attempt's failure: its transport error, or a status
//     error describing the rejected response.
type retryableError struct {
	err error
}

// Error returns the wrapped failure's message.
//
// Returns:
//   - msg (string): the error message.
func (e *retryableError) Error() (msg string) {
	msg = e.err.Error()

	return
}

// Unwrap exposes the wrapped failure for errors.Is and errors.As.
//
// Returns:
//   - err (error): the wrapped failure.
func (e *retryableError) Unwrap() (err error) {
	err = e.err

	return
}

// retryAfterWait parses the Retry-After header of resp, following RFC 9110:
// either a number of seconds or an HTTP-date.
//
// Examples:
//
//	Retry-After: 120
//	Retry-After: Fri, 31 Dec 1999 23:59:59 GMT
//
// Parameters:
//   - resp (*http.Response): the response whose header is parsed.
//
// Returns:
//   - wait (time.Duration): the requested wait; 0 for a date in the past.
//   - ok (bool): true when the header was present and parseable.
func retryAfterWait(resp *http.Response) (wait time.Duration, ok bool) {
	header := resp.Header.Get("Retry-After")
	if header == "" {
		return
	}

	if seconds, err := strconv.ParseInt(header, 10, 64); err == nil {
		// A negative delay makes no sense; leave the header unhonored.
		if seconds < 0 {
			return
		}

		wait = time.Duration(seconds) * time.Second
		ok = true

		return
	}

	retryTime, err := time.Parse(time.RFC1123, header)
	if err != nil {
		return
	}

	if until := time.Until(retryTime); until > 0 {
		wait = until
	}

	ok = true

	return
}
