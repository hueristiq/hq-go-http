// Package utils provides utility functions and types for processing HTTP-related data.
//
// The package is designed to be lightweight and reusable in applications that handle HTTP
// headers, such as web servers, clients, or middleware built with packages like net/http.
// It includes functionality to:
//   - Canonicalize HTTP header keys to their standard form (e.g., "content-type" to "Content-Type"),
//     as recommended by RFC 7231 and RFC 9110.
//   - Parse HTTP Link headers into structured ParsedLink and ParsedLinks types, as defined by RFC 8288,
//     to support use cases like pagination, resource preloading, or hypermedia APIs.
//
// Usage Example:
//
//	package main
//
//	import (
//	    "fmt"
//	    hqgohttpheaderutils "github.com/hueristiq/hq-go-http/header/utils"
//	)
//
//	func main() {
//	    // Canonicalize a header key
//	    key := "content-type"
//	    canonical := hqgohttpheaderutils.CanonicalizeHeaderKey(key)
//	    fmt.Println("Canonicalized:", canonical) // Output: Canonicalized: Content-Type
//
//	    // Parse a Link header
//	    header := `<https://api.example.com/page/2>; rel="next"; title="Next Page"`
//	    links := hqgohttpheaderutils.ParseLinkHeader(header)
//	    fmt.Println("Parsed Links:", links.String())
//	    // Output: Parsed Links: <https://api.example.com/page/2>; title="Next Page"; rel="next"
//	}
//
// References:
//   - RFC 7231: HTTP/1.1 Semantics and Content (https://tools.ietf.org/html/rfc7231)
//   - RFC 9110: HTTP Semantics (https://tools.ietf.org/html/rfc9110)
//   - RFC 8288: Web Linking (https://tools.ietf.org/html/rfc8288)
//   - IANA Link Relations: https://www.iana.org/assignments/link-relations/link-relations.xhtml
package utils
