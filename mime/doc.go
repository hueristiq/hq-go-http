// Package mime provides a strongly typed representation of Internet Media Types as defined by IANA.
// The mime package introduces the MIME type, which is an alias for string, to represent Internet Media Types
// (or content types) in a strongly typed manner. This design ensures compile-time type safety when working with
// content types in HTTP requests and responses, reducing errors such as typos and improving overall code maintainability.
//
// # Usage Example
//
//	package main
//
//	import (
//	    "fmt"
//	    hqgohttpmime "github.com/hueristiq/hq-go-http/mime"
//	)
//
//	func main() {
//	    ct := hqgohttpmime.JSON
//	    fmt.Println("Content-Type:", ct.String()) // Output: "application/json"
//	}
//
// Reference:
//
//	https://www.iana.org/assignments/media-types/media-types.xhtml
package mime
