package http

import "net/http"

// ErrorHandler defines a function type that handles HTTP response errors if retries are expired,
// containing the last status from the http library.
//
// Parameters:
//   - res: The HTTP response returned by the request, which can be nil if the request failed.
//   - err: The error encountered during the request. This can be nil if the request succeeded.
//   - tries: The number of attempts made so far, which can be used for implementing retry logic.
//
// Returns:
//   - req: The processed or retried HTTP response. This may be nil if the function decides not to retry.
//   - herr: An error returned by the error handler, signaling either a failure in retrying or a terminal error condition.
type ErrorHandler func(res *http.Response, err error, tries int) (req *http.Response, herr error)
