// Package header provides a strongly typed collection of HTTP header fields as defined by IANA.
// The package defines a custom type, Header, which is an alias for string. This approach
// ensures compile-time type safety and helps prevent errors (such as typos) when using HTTP
// header names throughout your application.
//
// # Usage Example:
//
//	package main
//
//	import (
//	    "fmt"
//	    "go.source.hueristiq.com/http/header"
//	)
//
//	func main() {
//	    h := header.Authorization
//	    fmt.Println("HTTP Header:", h.String()) // Output: "Authorization"
//	}
//
// # Reference
//
//	https://www.iana.org/assignments/http-fields/http-fields.xhtml
package header
