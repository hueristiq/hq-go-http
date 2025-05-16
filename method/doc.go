// Package method provides a strongly typed representation of HTTP methods as defined by IANA.
// The package introduces a custom type, Method, which is an alias for string, to ensure
// compile-time type safety when working with HTTP methods. This design helps prevent common errors,
// such as typos, and guarantees that only valid HTTP methods are used throughout your code.
//
// # Usage Example
//
//	package main
//
//	import (
//	    "fmt"
//	    hqgohttpmethod "github.com/hueristiq/hq-go-http/method"
//	)
//
//	func main() {
//	    m := hqgohttpmethod.GET
//	    fmt.Println("HTTP Method:", m.String()) // Output: "GET"
//	}
//
// Reference:
//
//	https://www.iana.org/assignments/http-methods/http-methods.xhtml
package method
