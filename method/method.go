package method

// Method represents HTTP methods as defined by IANA.
//
// The Method type provides a strongly typed representation of HTTP methods, ensuring correctness
// and maintainability when constructing and handling HTTP requests.
//
// Reference: https://www.iana.org/assignments/http-methods/http-methods.xhtml
type Method string

// String returns the string representation of the Method type.
//
// It converts the Method (a string alias) to its underlying string value.
// This is useful when you need to output the method in HTTP requests, logs, or debug messages.
//
// Returns:
//   - method (string): The HTTP method as a string value.
//
// Example:
//
//	m := GET
//	fmt.Println(m.String()) // Output: "GET"
func (m Method) String() (method string) {
	return string(m)
}

// Interface defines a common interface for HTTP methods.
//
// Any type that implements a String() method returning the HTTP method as a string is considered
// to satisfy this interface. This is useful when you want to abstract the representation of HTTP methods
// or allow different implementations that adhere to the same contract.
type Interface interface {
	String() (method string)
}

// Safe Methods
// These methods are considered safe because they are intended solely for retrieval
// of data and should not cause side effects on the server.
const (
	// GET retrieves a resource representation.
	// It should not have side effects on the server.
	// For further details, see RFC 7231, section 4.3.1.
	GET Method = "GET"

	// HEAD is similar to GET but does not return a response body.
	// It is used for obtaining meta-information without transferring the resource itself.
	// For further details, see RFC 7231, section 4.3.2.
	HEAD Method = "HEAD"
)

// Idempotent Methods
// These methods can be invoked multiple times with the same outcome.
// Using idempotent methods ensures that repeating the request does not lead to additional changes
// beyond the first request.
const (
	// PUT updates or creates a resource at the target URL.
	// It fully replaces the existing resource with the provided data.
	// For further details, see RFC 7231, section 4.3.4.
	PUT Method = "PUT"

	// DELETE removes the specified resource.
	// Repeated calls to DELETE should produce the same outcome as a single call.
	// For further details, see RFC 7231, section 4.3.5.
	DELETE Method = "DELETE"
)

// Unsafe Methods
// These methods may modify resources on the server and are generally not idempotent.
// They should be used with care since multiple identical requests might result in duplicate actions.
const (
	// POST submits data to the server, typically resulting in the creation of a new resource.
	// This method is not idempotent and should be used when side effects are intended.
	// For further details, see RFC 7231, section 4.3.3.
	POST Method = "POST"

	// PATCH applies partial modifications to a resource.
	// Unlike PUT, it does not require the complete replacement of the resource.
	// For further details, see RFC 5789.
	PATCH Method = "PATCH"
)

// Auxiliary Methods
// These methods serve special use cases that do not fall into the standard safe,
// idempotent, or unsafe categories.
const (
	// OPTIONS returns the communication options available for a target resource.
	// This method is often used to discover the capabilities or requirements of the server.
	// For further details, see RFC 7231, section 4.3.7.
	OPTIONS Method = "OPTIONS"

	// TRACE performs a message loop-back test along the path to the target resource.
	// It helps diagnose or debug intermediary behaviors (e.g., proxies altering the request).
	// For further details, see RFC 7231, section 4.3.8.
	TRACE Method = "TRACE"

	// CONNECT establishes a tunnel to the server, usually to facilitate HTTPS communication over a proxy.
	// It is used to create a direct connection for encrypted communication.
	// For further details, see RFC 7231, section 4.3.6.
	CONNECT Method = "CONNECT"
)

// This compile-time assertion ensures that the Method type correctly implements the Interface interface.
// If it does not, the assignment will cause a compile-time error.
var _ Interface = (*Method)(nil)
