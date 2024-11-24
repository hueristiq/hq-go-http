package methods

// Method represents HTTP methods as defined by IANA.
// Reference: https://www.iana.org/assignments/http-methods/http-methods.xhtml
type Method string

func (m Method) String() (method string) {
	return string(m)
}

const (
	// The CONNECT method establishes a tunnel to the server identified by the target resource.
	// This method is typically used for HTTPS connections through a proxy.
	// RFC 7231, section 4.3.6 defines the semantics of the CONNECT method.
	Connect Method = "CONNECT" // RFC 7231, 4.3.6

	// The DELETE method requests the server to delete the specified resource.
	// It is a commonly used method in RESTful APIs for resource deletion.
	// The method is defined in RFC 7231, section 4.3.5.
	Delete Method = "DELETE" // RFC 7231, 4.3.5

	// The GET method is used to retrieve data from a specified resource.
	// It is the most commonly used HTTP method and typically requests a representation of the resource.
	// GET requests should not modify server state and are considered safe and idempotent.
	// Defined in RFC 7231, section 4.3.1.
	Get Method = "GET" // RFC 7231, 4.3.1

	// The HEAD method is similar to GET but does not return a message body in the response.
	// It is used to retrieve the headers that a GET request would have obtained, often for checking the existence or meta-information of a resource.
	// Defined in RFC 7231, section 4.3.2.
	Head Method = "HEAD" // RFC 7231, 4.3.2

	// The OPTIONS method is used to describe the communication options for the target resource.
	// It allows a client to determine the capabilities of a server or a resource, such as which HTTP methods are supported.
	// Defined in RFC 7231, section 4.3.7.
	Options Method = "OPTIONS" // RFC 7231, 4.3.7

	// The PATCH method applies partial modifications to a resource.
	// Unlike PUT, which replaces the entire resource, PATCH allows updating only specific fields or data in the resource.
	// It is defined in RFC 5789.
	Patch Method = "PATCH" // RFC 5789

	// The POST method is used to send data to the server, usually resulting in the creation of a new resource or the modification of an existing one.
	// It is often used in web forms and API calls where data is submitted for processing.
	// Defined in RFC 7231, section 4.3.3.
	Post Method = "POST" // RFC 7231, 4.3.3

	// The PUT method replaces all current representations of the target resource with the uploaded content.
	// It is commonly used in RESTful APIs to update a resource completely.
	// Defined in RFC 7231, section 4.3.4.
	Put Method = "PUT" // RFC 7231, 4.3.4

	// The TRACE method performs a message loop-back test along the path to the target resource.
	// TRACE allows the client to see what is being received at the other end of the request chain and is mainly used for diagnostic purposes.
	// Defined in RFC 7231, section 4.3.8.
	Trace Method = "TRACE" // RFC 7231, 4.3.8
)
