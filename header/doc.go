// Package header provides a strongly typed collection of HTTP header fields as
// defined by IANA.
//
// It suits code that sets, reads, or manipulates HTTP headers — clients,
// servers, middleware built on net/http — where a plain string invites typos
// that compile fine and fail at runtime.
//
// # Entry points
//
// The [Header] type is a string alias with a String method, and the package
// constants cover the standard header fields registered with IANA plus common
// non-standard ones: CORS, caching, security, content negotiation, and
// connection management.
//
// The implementation relies on the Go standard library only.
package header
