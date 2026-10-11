// Package utils provides utility functions and types for processing HTTP-related data.
//
// It suits applications that handle HTTP headers — clients, servers, middleware —
// where a header key must be rendered in its canonical form or an RFC 8288 Link
// header must be taken apart, for pagination, resource preloading, or hypermedia
// APIs.
//
// # Entry points
//
// [CanonicalizeHeaderKey] converts a header key to its canonical form (e.g.,
// "content-type" to "Content-Type"), as recommended by RFC 9110.
//
// [Parse] parses a raw Link header into a [Links] collection of [Link] values —
// target URL, relation type, and parameters — and [ParseMultiple] merges several
// Link header lines into one collection. The parsing semantics follow
// github.com/tomnomnom/linkheader: parameter keys are kept exactly as written,
// and the "rel" attribute is matched case-insensitively. Commas inside the
// angle-bracketed URL are preserved. Note the shared limitation: quoted
// parameter values containing a comma or a semicolon are not split correctly.
//
// The implementation relies on the Go standard library only.
package utils
