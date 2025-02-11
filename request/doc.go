// Package request provides a reusable HTTP request wrapper that enables the request body
// to be read multiple times. This is particularly useful in scenarios where the same
// HTTP request (or its body) must be sent repeatedly, such as in retries, logging, or
// proxying situations.
//
// The core functionality of the package is centered around the Request type, which embeds
// the standard *http.Request. By wrapping http.Request, Request seamlessly integrates with
// existing HTTP libraries while adding the ability to “reuse” the request body. This is
// achieved by converting the body into a reusable read-closer (typically an instance of
// *ReusableReadCloser) that can be reset to the beginning once all data has been read.
//
// Usage Example
//
//	To create a new reusable HTTP request with a string as the body:
//
//	    req, err := request.New("POST", "https://example.com/api", "example request body")
//	    if err != nil {
//	        // handle error
//	    }
//	    // req.Request is a standard *http.Request and can be used with http.Client.
//
//	To create a request with a specific context (useful for timeouts or cancellation):
//
//	    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	    defer cancel()
//	    req, err := request.NewWithContext(ctx, "GET", "https://example.com/data", nil)
//	    if err != nil {
//	        // handle error
//	    }
//
// The request package simplifies the creation and management of HTTP requests whose bodies need
// to be read multiple times. By wrapping http.Request and handling the intricacies of buffering,
// resetting, and calculating content lengths, this package allows developers to focus on higher-level
// HTTP logic without worrying about the underlying details of request body management.
package request
