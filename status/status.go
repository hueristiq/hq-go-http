package status

import (
	"fmt"
)

// Status represents HTTP status codes as defined by the IANA registry.
//
// The Status type provides a structured and type-safe representation of HTTP status codes,
// ensuring correctness when working with HTTP responses. It allows you to retrieve both the
// numerical and human-readable representations of a status code, as well as to determine
// the category of the response (informational, success, redirection, or error).
//
// Reference: https://www.iana.org/assignments/http-status-codes/http-status-codes.xhtml
type Status int

// Int returns the integer (numeric) representation of the Status type.
//
// This method extracts the underlying integer value from a Status instance,
// which corresponds to the actual HTTP status code.
//
// Returns:
//   - status (int): The numerical HTTP status code.
//
// Example:
//
//	s := OK
//	fmt.Println(s.Int()) // Output: 200
func (s Status) Int() (status int) {
	return int(s)
}

// String returns the human-readable string representation of the Status type.
//
// It maps the Status value to its descriptive text according to the HTTP standard.
// If an unknown status code is provided, it returns a formatted string indicating the code is unknown.
//
// Returns:
//   - status (string): The human-readable description of the HTTP status code.
//
// Example:
//
//	s := NotFound
//	fmt.Println(s.String()) // Output: "Not Found"
func (s Status) String() (status string) {
	switch s {
	case Continue:
		return "Continue"
	case SwitchingProtocols:
		return "Switching Protocols"
	case Processing:
		return "Processing"
	case EarlyHints:
		return "Early Hints"
	case OK:
		return "OK"
	case Created:
		return "Created"
	case Accepted:
		return "Accepted"
	case NonAuthoritativeInfo:
		return "Non-Authoritative Information"
	case NoContent:
		return "No Content"
	case ResetContent:
		return "Reset Content"
	case PartialContent:
		return "Partial Content"
	case MultiStatus:
		return "Multi-Status"
	case AlreadyReported:
		return "Already Reported"
	case IMUsed:
		return "IM Used"
	case MultipleChoices:
		return "Multiple Choices"
	case MovedPermanently:
		return "Moved Permanently"
	case Found:
		return "Found"
	case SeeOther:
		return "See Other"
	case NotModified:
		return "Not Modified"
	case UseProxy:
		return "Use Proxy"
	case TemporaryRedirect:
		return "Temporary Redirect"
	case PermanentRedirect:
		return "Permanent Redirect"
	case BadRequest:
		return "Bad Request"
	case Unauthorized:
		return "Unauthorized"
	case PaymentRequired:
		return "Payment Required"
	case Forbidden:
		return "Forbidden"
	case NotFound:
		return "Not Found"
	case MethodNotAllowed:
		return "Method Not Allowed"
	case NotAcceptable:
		return "Not Acceptable"
	case ProxyAuthRequired:
		return "Proxy Authentication Required"
	case RequestTimeout:
		return "Request Timeout"
	case Conflict:
		return "Conflict"
	case Gone:
		return "Gone"
	case LengthRequired:
		return "Length Required"
	case PreconditionFailed:
		return "Precondition Failed"
	case RequestEntityTooLarge:
		return "Request Entity Too Large"
	case RequestURITooLong:
		return "Request URI Too Long"
	case UnsupportedMediaType:
		return "Unsupported Media Type"
	case RequestedRangeNotSatisfiable:
		return "Requested Range Not Satisfiable"
	case ExpectationFailed:
		return "Expectation Failed"
	case Teapot:
		return "I'm a teapot"
	case MisdirectedRequest:
		return "Misdirected Request"
	case UnprocessableEntity:
		return "Unprocessable Entity"
	case Locked:
		return "Locked"
	case FailedDependency:
		return "Failed Dependency"
	case UpgradeRequired:
		return "Upgrade Required"
	case PreconditionRequired:
		return "Precondition Required"
	case TooManyRequests:
		return "Too Many Requests"
	case RequestHeaderFieldsTooLarge:
		return "Request Header Fields Too Large"
	case UnavailableForLegalReasons:
		return "Unavailable For Legal Reasons"
	case InternalServerError:
		return "Internal Server Error"
	case NotImplemented:
		return "Not Implemented"
	case BadGateway:
		return "Bad Gateway"
	case ServiceUnavailable:
		return "Service Unavailable"
	case GatewayTimeout:
		return "Gateway Timeout"
	case HTTPVersionNotSupported:
		return "HTTP Version Not Supported"
	case VariantAlsoNegotiates:
		return "Variant Also Negotiates"
	case InsufficientStorage:
		return "Insufficient Storage"
	case LoopDetected:
		return "Loop Detected"
	case NotExtended:
		return "Not Extended"
	case NetworkAuthenticationRequired:
		return "Network Authentication Required"
	default:
		return fmt.Sprintf("Unknown Status (%d)", s)
	}
}

// IsInformational checks if the status code falls within the 1xx range (100–199).
//
// Returns:
//   - isInformational (bool): True if the status code is informational, false otherwise.
func (s Status) IsInformational() (isInformational bool) {
	return s >= 100 && s < 200
}

// IsSuccess checks if the status code falls within the 2xx range (200–299).
//
// Returns:
//   - isSuccess (bool): True if the status code indicates success, false otherwise.
func (s Status) IsSuccess() (isSuccess bool) {
	return s >= 200 && s < 300
}

// IsRedirection checks if the status code falls within the 3xx range (300–399).
//
// Returns:
//   - isRedirection (bool): True if the status code indicates redirection, false otherwi
func (s Status) IsRedirection() (isRedirection bool) {
	return s >= 300 && s < 400
}

// IsError checks if the status code represents an error.
// It returns true if the status code falls within either the 4xx (client error) or 5xx (server error) range.
//
// Returns:
//   - isError (bool): True if the status code is an error, false otherwise.
func (s Status) IsError() (isError bool) {
	return s.IsClientError() || s.IsServerError()
}

// IsClientError checks if the status code falls within the 4xx range (400–499).
//
// Returns:
//   - isClientError (bool): True if the status code indicates a client error, false otherwise.
func (s Status) IsClientError() (isClientError bool) {
	return s >= 400 && s < 500
}

// IsServerError checks if the status code falls within the 5xx range (500–599).
//
// Returns:
//   - isServerError (bool): True if the status code indicates a server error, false otherwise.
func (s Status) IsServerError() (isServerError bool) {
	return s >= 500 && s < 600
}

// Interface defines a common interface for types representing HTTP status codes.
//
// Any type that implements the methods below (to retrieve the integer and string representation,
// as well as categorizing the status code) can be used interchangeably with Status.
// This promotes flexibility when extending or wrapping HTTP status functionality.
type Interface interface {
	Int() (status int)
	IsInformational() (isInformational bool)
	IsSuccess() (isSuccess bool)
	IsRedirection() (isRedirection bool)
	IsError() (isError bool)
	IsClientError() (isClientError bool)
	IsServerError() (isServerError bool)
}

// Informational responses (100–199):
// These indicate that the request has been received and is being processed,
// but no final response is available yet.
const (
	Continue           Status = 100 // RFC 7231, 6.2.1 - Request received; continue sending the request.
	SwitchingProtocols Status = 101 // RFC 7231, 6.2.2 - Server is switching protocols as requested by the client.
	Processing         Status = 102 // RFC 2518, 10.1 - Server is processing the request, but no response is yet available.
	EarlyHints         Status = 103 // RFC 8297 - Preliminary response headers before the final response.
)

// Successful responses (200–299):
// These indicate that the request was successfully received, understood, and accepted.
const (
	OK                   Status = 200 // RFC 7231, 6.3.1 - The request succeeded and the server returned the requested resource.
	Created              Status = 201 // RFC 7231, 6.3.2 - A new resource was created as a result of the request.
	Accepted             Status = 202 // RFC 7231, 6.3.3 - The request was accepted for processing, but processing is not complete.
	NonAuthoritativeInfo Status = 203 // RFC 7231, 6.3.4 - The returned meta-information is from a third-party source.
	NoContent            Status = 204 // RFC 7231, 6.3.5 - The request succeeded, but there is no content to return.
	ResetContent         Status = 205 // RFC 7231, 6.3.6 - The client should reset its document view.
	PartialContent       Status = 206 // RFC 7233, 4.1 - The server is delivering only part of the resource (due to a range header).
	MultiStatus          Status = 207 // RFC 4918, 11.1 - Multiple status codes for different parts of the response (WebDAV).
	AlreadyReported      Status = 208 // RFC 5842, 7.1 - The members of a DAV binding have already been reported.
	IMUsed               Status = 226 // RFC 3229, 10.4.1 - The server fulfilled the request using delta encoding.
)

// Redirection messages (300–399):
// These indicate that the client must take additional actions to complete the request.
const (
	MultipleChoices   Status = 300 // RFC 7231, 6.4.1 - Multiple options for the resource are available.
	MovedPermanently  Status = 301 // RFC 7231, 6.4.2 - The resource has permanently moved to a new URL.
	Found             Status = 302 // RFC 7231, 6.4.3 - The resource is temporarily located at a different URL.
	SeeOther          Status = 303 // RFC 7231, 6.4.4 - The resource can be found at a different URI.
	NotModified       Status = 304 // RFC 7232, 4.1 - The resource has not been modified since the last request.
	UseProxy          Status = 305 // RFC 7231, 6.4.5 - The requested resource must be accessed through a proxy.
	_                 Status = 306 // RFC 7231, 6.4.6 - This code is no longer used but reserved.
	TemporaryRedirect Status = 307 // RFC 7231, 6.4.7 - The resource is temporarily at a different URI.
	PermanentRedirect Status = 308 // RFC 7538, 3 - The resource has permanently moved to a new URI.
)

// Client error responses (400–499):
// These indicate that there was an error in the request sent by the client.
const (
	BadRequest                   Status = 400 // RFC 7231, 6.5.1 - The server could not understand the request due to invalid syntax.
	Unauthorized                 Status = 401 // RFC 7235, 3.1 - Authentication is required to access the resource.
	PaymentRequired              Status = 402 // RFC 7231, 6.5.2 - Reserved for future use.
	Forbidden                    Status = 403 // RFC 7231, 6.5.3 - The client does not have permission to access the resource.
	NotFound                     Status = 404 // RFC 7231, 6.5.4 - The server could not find the requested resource.
	MethodNotAllowed             Status = 405 // RFC 7231, 6.5.5 - The request method is not supported for the resource.
	NotAcceptable                Status = 406 // RFC 7231, 6.5.6 - The requested resource cannot generate content acceptable per the Accept headers.
	ProxyAuthRequired            Status = 407 // RFC 7235, 3.2 - The client must authenticate with a proxy.
	RequestTimeout               Status = 408 // RFC 7231, 6.5.7 - The client did not send a complete request within the allowed time.
	Conflict                     Status = 409 // RFC 7231, 6.5.8 - The request conflicts with the current state of the resource.
	Gone                         Status = 410 // RFC 7231, 6.5.9 - The resource is no longer available.
	LengthRequired               Status = 411 // RFC 7231, 6.5.10 - The request did not specify the length of its content.
	PreconditionFailed           Status = 412 // RFC 7232, 4.2 - One or more preconditions in the request header were not met.
	RequestEntityTooLarge        Status = 413 // RFC 7231, 6.5.11 - The request is too large for the server to process.
	RequestURITooLong            Status = 414 // RFC 7231, 6.5.12 - The provided URI is too long for the server to handle.
	UnsupportedMediaType         Status = 415 // RFC 7231, 6.5.13 - The media type of the request is unsupported by the server.
	RequestedRangeNotSatisfiable Status = 416 // RFC 7233, 4.4 - The requested range cannot be satisfied.
	ExpectationFailed            Status = 417 // RFC 7231, 6.5.14 - The server cannot meet the expectations in the Expect header.
	Teapot                       Status = 418 // RFC 7168, 2.3.3 - A humorous status code indicating the server is a teapot.
	MisdirectedRequest           Status = 421 // RFC 7540, 9.1.2 - The request was misdirected to a server that cannot produce a response.
	UnprocessableEntity          Status = 422 // RFC 4918, 11.2 - The request is well-formed but cannot be processed.
	Locked                       Status = 423 // RFC 4918, 11.3 - The resource is locked.
	FailedDependency             Status = 424 // RFC 4918, 11.4 - A previous request failed, causing this request to fail.
	UpgradeRequired              Status = 426 // RFC 7231, 6.5.15 - The client should switch to a different protocol.
	PreconditionRequired         Status = 428 // RFC 6585, 3 - The server requires that the request be conditional.
	TooManyRequests              Status = 429 // RFC 6585, 4 - The client has sent too many requests in a given time.
	RequestHeaderFieldsTooLarge  Status = 431 // RFC 6585, 5 - The server is unwilling to process the request due to large header fields.
	UnavailableForLegalReasons   Status = 451 // RFC 7725, 3 - The server is denying access due to legal reasons.
)

// Server error responses (500–599):
// These indicate that the server encountered an error or is unable to complete the request.
const (
	InternalServerError           Status = 500 // RFC 7231, 6.6.1 - The server encountered an unexpected condition.
	NotImplemented                Status = 501 // RFC 7231, 6.6.2 - The server does not support the functionality required.
	BadGateway                    Status = 502 // RFC 7231, 6.6.3 - The server received an invalid response from an upstream server.
	ServiceUnavailable            Status = 503 // RFC 7231, 6.6.4 - The server is currently unable to handle the request.
	GatewayTimeout                Status = 504 // RFC 7231, 6.6.5 - The server did not receive a timely response from an upstream server.
	HTTPVersionNotSupported       Status = 505 // RFC 7231, 6.6.6 - The server does not support the HTTP protocol version used.
	VariantAlsoNegotiates         Status = 506 // RFC 2295, 8.1 - The server encountered an internal configuration error.
	InsufficientStorage           Status = 507 // RFC 4918, 11.5 - The server is unable to store the representation needed.
	LoopDetected                  Status = 508 // RFC 5842, 7.2 - The server detected an infinite loop while processing the request.
	NotExtended                   Status = 510 // RFC 2774, 7 - Additional extensions to the request are required.
	NetworkAuthenticationRequired Status = 511 // RFC 6585, 6 - The client must authenticate to gain network access.
)

// This compile-time assertion ensures that the Status type correctly implements the Interface interface.
// If it does not, the assignment will cause a compile-time error..
var _ Interface = (*Status)(nil)
