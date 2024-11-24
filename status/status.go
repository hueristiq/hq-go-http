package status

type Status int

func (s Status) Int() (status int) {
	return int(s)
}

const (
	// Informational responses (100–199): These indicate that the request was received and is being processed, but no final response is available yet.
	Continue           Status = 100 // RFC 7231, 6.2.1 - Request received, continue sending the request.
	SwitchingProtocols Status = 101 // RFC 7231, 6.2.2 - Server is switching protocols based on the client request.
	Processing         Status = 102 // RFC 2518, 10.1 - Server has received and is processing the request, but no response is available yet.
	EarlyHints         Status = 103 // RFC 8297 - Used to send preliminary response headers before the final response.

	// Successful responses (200–299): These indicate that the request was successfully received, understood, and accepted.
	OK                   Status = 200 // RFC 7231, 6.3.1 - The request succeeded, and the server responded with the requested resource.
	Created              Status = 201 // RFC 7231, 6.3.2 - The request was successful, and a new resource was created as a result.
	Accepted             Status = 202 // RFC 7231, 6.3.3 - The request was accepted for processing, but processing is not yet complete.
	NonAuthoritativeInfo Status = 203 // RFC 7231, 6.3.4 - The server successfully processed the request but is returning information from a third-party source.
	NoContent            Status = 204 // RFC 7231, 6.3.5 - The server successfully processed the request, but no content is returned.
	ResetContent         Status = 205 // RFC 7231, 6.3.6 - The server successfully processed the request, but the client should reset its document view.
	PartialContent       Status = 206 // RFC 7233, 4.1 - The server is delivering only part of the resource due to a range header sent by the client.
	MultiStatus          Status = 207 // RFC 4918, 11.1 - The response provides status for multiple independent operations (used in WebDAV).
	AlreadyReported      Status = 208 // RFC 5842, 7.1 - The members of a DAV binding have already been reported in a previous response.
	IMUsed               Status = 226 // RFC 3229, 10.4.1 - The server has fulfilled a request for a resource using delta encoding.

	// Redirection messages (300–399): These indicate that the client must take additional actions to complete the request.
	MultipleChoices   Status = 300 // RFC 7231, 6.4.1 - There are multiple options for the resource that the client may choose from.
	MovedPermanently  Status = 301 // RFC 7231, 6.4.2 - The resource has been permanently moved to a new URL.
	Found             Status = 302 // RFC 7231, 6.4.3 - The resource is temporarily located at a different URL.
	SeeOther          Status = 303 // RFC 7231, 6.4.4 - The client should retrieve the resource using a different URI.
	NotModified       Status = 304 // RFC 7232, 4.1 - The resource has not been modified since the version specified by the request headers.
	UseProxy          Status = 305 // RFC 7231, 6.4.5 - The requested resource must be accessed through a proxy.
	_                 Status = 306 // RFC 7231, 6.4.6 - (Unused) This status code is no longer in use but reserved.
	TemporaryRedirect Status = 307 // RFC 7231, 6.4.7 - The resource is temporarily located at a different URI, but future requests should still use the original URI.
	PermanentRedirect Status = 308 // RFC 7538, 3 - The resource has been permanently moved to a new URI.

	// Client error responses (400–499): These indicate that there was an error in the request sent by the client.
	BadRequest                   Status = 400 // RFC 7231, 6.5.1 - The server could not understand the request due to invalid syntax.
	Unauthorized                 Status = 401 // RFC 7235, 3.1 - Authentication is required to access the resource.
	PaymentRequired              Status = 402 // RFC 7231, 6.5.2 - Reserved for future use, originally intended for payment.
	Forbidden                    Status = 403 // RFC 7231, 6.5.3 - The client does not have permission to access the resource.
	NotFound                     Status = 404 // RFC 7231, 6.5.4 - The server could not find the requested resource.
	MethodNotAllowed             Status = 405 // RFC 7231, 6.5.5 - The request method is not supported for the requested resource.
	NotAcceptable                Status = 406 // RFC 7231, 6.5.6 - The requested resource cannot generate content acceptable according to the Accept headers.
	ProxyAuthRequired            Status = 407 // RFC 7235, 3.2 - The client must authenticate with a proxy.
	RequestTimeout               Status = 408 // RFC 7231, 6.5.7 - The client did not send a complete request within the time allowed by the server.
	Conflict                     Status = 409 // RFC 7231, 6.5.8 - The request could not be completed due to a conflict with the current state of the resource.
	Gone                         Status = 410 // RFC 7231, 6.5.9 - The resource requested is no longer available and will not be available again.
	LengthRequired               Status = 411 // RFC 7231, 6.5.10 - The request did not specify the length of its content, which is required by the server.
	PreconditionFailed           Status = 412 // RFC 7232, 4.2 - One or more preconditions given in the request headers were not met.
	RequestEntityTooLarge        Status = 413 // RFC 7231, 6.5.11 - The request is larger than the server is willing or able to process.
	RequestURITooLong            Status = 414 // RFC 7231, 6.5.12 - The URI requested by the client is longer than the server can handle.
	UnsupportedMediaType         Status = 415 // RFC 7231, 6.5.13 - The media format of the requested data is not supported by the server.
	RequestedRangeNotSatisfiable Status = 416 // RFC 7233, 4.4 - The range specified by the Range header field cannot be fulfilled.
	ExpectationFailed            Status = 417 // RFC 7231, 6.5.14 - The server cannot meet the expectations in the Expect request header.
	Teapot                       Status = 418 // RFC 7168, 2.3.3 - A humorous status code indicating that the server is a teapot and cannot brew coffee.
	MisdirectedRequest           Status = 421 // RFC 7540, 9.1.2 - The request was directed to a server that is not able to produce a response.
	UnprocessableEntity          Status = 422 // RFC 4918, 11.2 - The request is well-formed, but the server is unable to process the contained instructions.
	Locked                       Status = 423 // RFC 4918, 11.3 - The resource being accessed is locked.
	FailedDependency             Status = 424 // RFC 4918, 11.4 - The request failed due to the failure of a previous request.
	UpgradeRequired              Status = 426 // RFC 7231, 6.5.15 - The client should switch to a different protocol.
	PreconditionRequired         Status = 428 // RFC 6585, 3 - The server requires that the request be conditional.
	TooManyRequests              Status = 429 // RFC 6585, 4 - The client has sent too many requests in a given amount of time ("rate limiting").
	RequestHeaderFieldsTooLarge  Status = 431 // RFC 6585, 5 - The server is unwilling to process the request because its header fields are too large.
	UnavailableForLegalReasons   Status = 451 // RFC 7725, 3 - The server is denying access to the resource as a consequence of a legal demand.

	// Server error responses (500–599): These indicate that the server encountered an error or is unable to perform the request.
	InternalServerError           Status = 500 // RFC 7231, 6.6.1 - The server encountered an unexpected condition that prevented it from fulfilling the request.
	NotImplemented                Status = 501 // RFC 7231, 6.6.2 - The server does not support the functionality required to fulfill the request.
	BadGateway                    Status = 502 // RFC 7231, 6.6.3 - The server, while acting as a gateway or proxy, received an invalid response from the upstream server.
	ServiceUnavailable            Status = 503 // RFC 7231, 6.6.4 - The server is currently unable to handle the request due to temporary overloading or maintenance.
	GatewayTimeout                Status = 504 // RFC 7231, 6.6.5 - The server, while acting as a gateway or proxy, did not receive a timely response from the upstream server.
	HTTPVersionNotSupported       Status = 505 // RFC 7231, 6.6.6 - The server does not support the HTTP protocol version used in the request.
	VariantAlsoNegotiates         Status = 506 // RFC 2295, 8.1 - The server has an internal configuration error and cannot complete the request.
	InsufficientStorage           Status = 507 // RFC 4918, 11.5 - The server is unable to store the representation needed to complete the request.
	LoopDetected                  Status = 508 // RFC 5842, 7.2 - The server detected an infinite loop while processing the request.
	NotExtended                   Status = 510 // RFC 2774, 7 - Further extensions to the request are required for the server to fulfill it.
	NetworkAuthenticationRequired Status = 511 // RFC 6585, 6 - The client needs to authenticate to gain network access.
)
