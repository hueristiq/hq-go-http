// Package status provides a strongly typed representation of HTTP status codes as defined by the IANA registry.
//
// # Overview
//
// The status package defines the Status type, an alias for int, to represent HTTP status codes in a structured
// and type-safe manner. In addition to exposing the raw numerical value of a status code via the Int() method,
// the package offers a String() method that returns a human-readable description for each code. Utility methods are
// also provided to help categorize responses:
//   - IsInformational() determines if a status code is in the 1xx range.
//   - IsSuccess() determines if a status code is in the 2xx range.
//   - IsRedirection() determines if a status code is in the 3xx range.
//   - IsClientError() determines if a status code is in the 4xx range.
//   - IsServerError() determines if a status code is in the 5xx range.
//   - IsError() returns true if the status code represents either a client or server error.
//
// Reference:
//
//	https://www.iana.org/assignments/http-status-codes/http-status-codes.xhtml
//
// # Usage Example
//
// The following example demonstrates how to use the status package:
//
//	package main
//
//	import (
//	    "fmt"
//	    "go.source.hueristiq.com/http/status"
//	)
//
//	func main() {
//	    // Create a status using one of the predefined constants.
//	    s := status.OK
//	    fmt.Println("Status code:", s.Int())    // Output: 200
//	    fmt.Println("Status text:", s.String())   // Output: "OK"
//
//	    // Check the category of the status code.
//	    if s.IsSuccess() {
//	        fmt.Println("The request was successful.")
//	    }
//	}
//
// # Interface
//
// The package defines an Interface that standardizes the behavior of types representing HTTP status codes. The interface
// includes the following methods:
//   - Int() int
//   - String() string
//   - IsInformational() bool
//   - IsSuccess() bool
//   - IsRedirection() bool
//   - IsClientError() bool
//   - IsServerError() bool
//   - IsError() bool
//
// # Conclusion
//
// By providing a type-safe, well-structured representation of HTTP status codes, the status package enhances code clarity,
// maintainability, and correctness when handling HTTP responses in your applications.
package status
