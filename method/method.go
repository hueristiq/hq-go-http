package method

// Method represents an HTTP method as defined by IANA.
//
// The Method type is a string alias that encapsulates common HTTP methods such as GET, POST, PUT, etc.
// This strong typing improves code clarity and type safety by ensuring only valid HTTP method names are used.
type Method string

// String returns the underlying string representation of the Method.
//
// This method converts the Method (a string alias) to its plain string value. This conversion is
// particularly useful when the method needs to be included in HTTP requests, log messages, or debug output.
//
// Returns:
//   - method (string): The HTTP method as a string.
func (m Method) String() (method string) {
	method = string(m)

	return
}

// Interface defines the contract for any type representing an HTTP method.
//
// Any type that implements the String() method returning a string is considered to satisfy this interface.
// This abstraction is useful when you need to work with various types that represent HTTP methods while
// ensuring they can all be converted to a string for processing.
type Interface interface {
	String() (method string)
}

// Predefined HTTP method constants.
//
// These constants represent the most common HTTP methods and are declared as type Method to ensure
// type safety and to prevent common errors such as misspelling method names. Although some of these
// methods (like POST, PUT, DELETE, and PATCH) may cause changes on the server, they are included
// here to cover a full range of common HTTP operations.
const (
	CONNECT Method = "CONNECT"
	DELETE  Method = "DELETE"
	GET     Method = "GET"
	HEAD    Method = "HEAD"
	OPTIONS Method = "OPTIONS"
	PATCH   Method = "PATCH"
	POST    Method = "POST"
	PUT     Method = "PUT"
	TRACE   Method = "TRACE"
)

// Compile-time interface check to ensure Method implements Interface.
//
// This assignment forces a compile-time error if Method does not implement
// the required String() method of the Interface, ensuring consistency.
var _ Interface = (*Method)(nil)
