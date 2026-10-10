// Package status provides a strongly typed representation of HTTP status codes
// as defined by the IANA registry.
//
// It suits code that branches on response status codes, where comparing raw
// integers obscures intent and invites magic numbers.
//
// # Entry points
//
// The [Status] type is an int alias: Int returns the numeric code, String the
// standard reason phrase, and the Is* methods — IsInformational, IsSuccess,
// IsRedirection, IsClientError, IsServerError, IsError — classify a code by its
// response class. The package constants cover the standard codes from 100
// Continue through 511 Network Authentication Required.
//
// The implementation relies on the Go standard library only.
package status
