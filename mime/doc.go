// Package mime provides a strongly typed representation of Internet Media Types
// (MIME types) as defined by IANA.
//
// It suits code that sets or inspects Content-Type and Accept headers, where a
// plain string invites typos that compile fine and fail at runtime.
//
// # Entry points
//
// The [MIME] type is a string alias with a String method, and the package
// constants cover common media types grouped by category — application, audio,
// video, image, font, and text.
//
// The implementation relies on the Go standard library only.
package mime
