// Package header provides a strongly typed collection of HTTP header fields as defined by IANA.
// The package defines a custom type, Header, which is an alias for string. This approach
// ensures compile-time type safety and helps prevent errors (such as typos) when using HTTP
// header names throughout your application.
//
// # Usage
//
// To use the package, simply import it into your project and refer to the defined constants and types.
// Below is an example that demonstrates how to use the package to reference and print an HTTP header:
//
//	package main
//
//	import (
//	    "fmt"
//	    "go.source.hueristiq.com/http/header"
//	)
//
//	func main() {
//	    // Use the strongly typed constant for the Authorization header.
//	    h := header.Authorization
//	    fmt.Println("HTTP Header:", h.String())
//	}
//
// # Reference
//
//	https://www.iana.org/assignments/http-fields/http-fields.xhtml
package header
