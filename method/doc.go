// Package method provides a strongly typed representation of HTTP methods as defined by IANA.
//
// # Overview
//
// The package introduces a custom type, Method, which is an alias for string, to ensure
// compile-time type safety when working with HTTP methods. This design helps prevent common errors,
// such as typos, and guarantees that only valid HTTP methods are used throughout your code.
//
// The package defines a set of constants representing the most common HTTP methods. These constants
// are organized into the following categories:
//
//   - Safe Methods: Methods that are considered safe because they do not modify server state.
//   - GET: Retrieves a resource representation.
//   - HEAD: Similar to GET but does not return a response body.
//   - Idempotent Methods: Methods that can be invoked multiple times without producing different outcomes.
//   - PUT: Updates or creates a resource at the target URL.
//   - DELETE: Removes the specified resource.
//   - Unsafe Methods: Methods that may modify resources on the server and are generally non-idempotent.
//   - POST: Submits data to the server, typically resulting in the creation of a new resource.
//   - PATCH: Applies partial modifications to a resource.
//   - Auxiliary Methods: Special-purpose methods used for options retrieval, debugging, or establishing tunnels.
//   - OPTIONS: Returns the communication options for the target resource.
//   - TRACE: Performs a message loop-back test along the request chain.
//   - CONNECT: Establishes a tunnel to the server, often used for HTTPS over a proxy.
//
// For further details on the HTTP methods defined by IANA, please refer to:
//
//	https://www.iana.org/assignments/http-methods/http-methods.xhtml
//
// # Usage Example
//
// Below is an example demonstrating how to use the method package:
//
//	package main
//
//	import (
//	    "fmt"
//	    "go.source.hueristiq.com/http/method"
//	)
//
//	func main() {
//	    m := method.GET
//	    fmt.Println("HTTP Method:", m.String()) // Output: "GET"
//	}
//
// # Interface
//
// The package also defines an Interface that requires a String() method returning the HTTP method as a string.
// This allows any type that implements this method to be used interchangeably with the Method type, promoting flexibility
// and abstraction in your code.
//
// # Conclusion
//
// The method package enhances code maintainability and robustness by enforcing type safety and organizing HTTP methods
// into logical categories. This results in a clear, consistent API for working with HTTP requests, making it easier to build
// reliable and readable HTTP-based applications.
package method
