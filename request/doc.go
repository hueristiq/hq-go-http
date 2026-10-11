// Package request provides an HTTP request wrapper with a reusable body.
//
// It suits scenarios where the same request is sent repeatedly — retries,
// logging, proxying — and reading the body must not consume it.
//
// # Entry points
//
// [New] builds a [Request] wrapping a standard *http.Request, converting the
// body into a [ReusableReadCloser]: an in-memory read-closer that rewinds
// itself at the end of the data, so every read sees the full payload. The
// body accepts a string, a []byte, an io.Reader, an io.ReadSeeker, a
// *bytes.Buffer, a func() (io.Reader, error), or a *ReusableReadCloser, and is
// buffered in memory, with its length becoming the request's ContentLength.
// A nil ctx means context.Background.
//
// The rewind makes [ReusableReadCloser] safe only under length-bounded
// consumers — the *http.Request bodies it backs. Never read one with
// io.ReadAll or io.Copy: the reader rewinds instead of ending, so the read
// never terminates.
//
// URLs are parsed with a parser that applies the default scheme "http" when
// none is given, so "example.com/path" is a valid URL.
//
// The implementation relies on the Go standard library, plus
// github.com/hueristiq/hq-lib-url-go (URL parsing with scheme defaulting).
package request
