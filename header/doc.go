// Package header provides a strongly typed collection of HTTP header fields as defined by IANA.
//
// The package defines a custom type, Header, which is an alias for string. This approach
// ensures compile-time type safety and helps prevent errors (such as typos) when using HTTP
// header names throughout your application. By using a custom type and predefined constants,
// the package promotes consistency, readability, and maintainability in code that interacts
// with HTTP headers, such as in HTTP clients, servers, or middleware built with packages like
// net/http.
//
// The constants defined in this package correspond to standard HTTP header fields registered
// with IANA, as well as common non-standard headers (e.g., X-Frame-Options). These constants
// cover a wide range of use cases, including CORS, caching, security, content negotiation,
// and connection management. Developers should use these constants instead of raw strings to
// leverage type safety and avoid errors.
//
// Usage Example:
//
//	package main
//
//	import (
//	    "fmt"
//	    "net/http"
//	    hqgohttpheader "github.com/hueristiq/hq-go-http/header"
//	)
//
//	func main() {
//	    // Set an HTTP header using the typed constant
//	    req, _ := http.NewRequest("GET", "https://example.com", nil)
//	    req.Header.Set(hqgohttpheader.ContentType.String(), "application/json")
//	    fmt.Println("HTTP Header:", hqgohttpheader.Authorization.String()) // Output: "Authorization"
//	}
//
// Reference:
//   - IANA HTTP Fields Registry: https://www.iana.org/assignments/http-fields/http-fields.xhtml
//   - Relevant RFCs: RFC 7231 (HTTP/1.1), RFC 7540 (HTTP/2), RFC 9110 (HTTP Semantics)
package header
