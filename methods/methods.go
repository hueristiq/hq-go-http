package methods

// This file lists the HTTP methods as constants.
//
// HTTP methods, sometimes referred to as HTTP verbs,
// are actions that the client can request from the server when making an HTTP request.
//
// Each method is followed by a comment that indicates the corresponding
// section of the RFC (Request for Comments) document that defines the method.

const (
	// The CONNECT method establishes a tunnel to the server identified by the target resource. It is defined in section 4.3.6 of RFC 7231.
	Connect = "CONNECT" // RFC 7231, 4.3.6
	// The DELETE method is used to delete the specified resource. It is defined in section 4.3.5 of RFC 7231.
	Delete = "DELETE" // RFC 7231, 4.3.5
	// The GET method is used to request data from a specified resource. It is one of the most common HTTP methods and is defined in section 4.3.1 of RFC 7231.
	Get = "GET" // RFC 7231, 4.3.1
	// The HEAD method is almost identical to GET, but without the response body. It is used to retrieve the meta-information written in HTTP headers, without transferring the entire content. It is defined in section 4.3.2 of RFC 7231.
	Head = "HEAD" // RFC 7231, 4.3.2
	// The OPTIONS method is used to describe the communication options for the target resource. It is defined in section 4.3.7 of RFC 7231.
	Options = "OPTIONS" // RFC 7231, 4.3.7
	//  The PATCH method is used to apply partial modifications to a resource. It is defined in RFC 5789.
	Patch = "PATCH" // RFC 5789
	// The POST method is used to send data to a server to create/update a resource. It is defined in section 4.3.3 of RFC 7231.
	Post = "POST" // RFC 7231, 4.3.3
	// The PUT method is used to update a current resource with new data. It is defined in section 4.3.4 of RFC 7231.
	Put = "PUT" // RFC 7231, 4.3.4
	// The TRACE method performs a message loop-back test along the path to the target resource. It is defined in section 4.3.8 of RFC 7231.
	Trace = "TRACE" // RFC 7231, 4.3.8
)
