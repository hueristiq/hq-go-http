// Package status provides a strongly typed representation of HTTP status codes as defined by the IANA registry.
// The status package defines the Status type, an alias for int, to represent HTTP status codes in a structured
// and type-safe manner. In addition to exposing the raw numerical value of a status code via the Int() method,
// the package offers a String() method that returns a human-readable description for each code.
//
// # Usage Example
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
// Reference:
//
//	https://www.iana.org/assignments/http-status-codes/http-status-codes.xhtml
package status
