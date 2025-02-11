// Package header provides a strongly typed collection of HTTP header fields as defined by IANA.
//
// # Overview
//
// The package defines a custom type, Header, which is an alias for string. This approach
// ensures compile-time type safety and helps prevent errors (such as typos) when using HTTP
// header names throughout your application.
//
// In addition to the Header type, the package exposes a wide range of HTTP header constants,
// grouped into logical categories. These categories include:
//
//   - Authentication: Contains headers related to authentication and authorization.
//   - Caching: Defines headers that control caching behavior between clients and servers.
//   - Client Hints: Provides headers that offer hints about the client's device characteristics or network conditions.
//   - Conditionals: Contains headers used for conditional HTTP requests, such as ETag handling.
//   - Connection Management: Defines headers that manage aspects of the network connection.
//   - Content Negotiation: Specifies headers for negotiating content type, language, encoding, etc.
//   - Controls: General headers that manage cookies, expectations, and related directives.
//   - CORS (Cross-Origin Resource Sharing): Contains headers that manage cross-origin requests.
//   - Do Not Track: Expresses the user's tracking preferences.
//   - Downloads: Relates to the handling of downloadable content.
//   - Message Body Information: Describes properties of the HTTP message body, including its encoding, length, and content type.
//   - Proxies: Contains headers that provide details when requests pass through proxy servers.
//   - Redirects: Specifies redirection information.
//   - Request Context: Provides contextual information about the request (e.g., host, user agent, referrer).
//   - Response Context: Offers details about the response, such as supported methods and server information.
//   - Range Requests: Enables partial content retrieval through byte-range specifications.
//   - Security: Enforces various security policies and mechanisms.
//   - Server-Sent Events (SSE): Contains headers used in Server-Sent Events for real-time communications.
//   - Transfer Coding: Specifies headers that govern transfer encoding mechanisms.
//   - WebSockets: Provides headers necessary for initiating and managing WebSocket connections.
//   - Other: Miscellaneous headers that do not fit into the above categories.
//
// The package also defines an Interface that includes a String() method. Any type that implements
// this method can be used interchangeably with Header values, allowing for custom header implementations
// that adhere to the same contract.
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
// For a complete list of standard HTTP header fields, please refer to the IANA registry:
//
//	https://www.iana.org/assignments/http-fields/http-fields.xhtml
//
// # Conclusion
//
// The header package is designed to improve code maintainability and robustness by providing
// a type-safe and well-organized way to work with HTTP headers. By grouping headers into meaningful
// categories and offering a clear interface, the package simplifies handling HTTP communications
// in your applications.
package header
