// Package mime provides a strongly typed representation of Internet Media Types as defined by IANA.
//
// # Overview
//
// The mime package introduces the MIME type, which is an alias for string, to represent Internet Media Types
// (or content types) in a strongly typed manner. This design ensures compile-time type safety when working with
// content types in HTTP requests and responses, reducing errors such as typos and improving overall code maintainability.
//
// The package defines a wide range of constants for common MIME types, organized into several logical categories:
//   - Application MIME Types: For non-text and binary data.
//   - Audio MIME Types: For representing audio files.
//   - Video MIME Types: For representing video files.
//   - Image MIME Types: For representing image files.
//   - Font MIME Types: For representing font files.
//   - Text MIME Types: For text-based content.
//   - Document MIME Types: For various document formats.
//   - Other MIME Types: For additional or miscellaneous types.
//
// Reference:
//
//	https://www.iana.org/assignments/media-types/media-types.xhtml
//
// # Usage Example
//
// Below is an example demonstrating how to use the mime package:
//
//	package main
//
//	import (
//	    "fmt"
//	    "go.source.hueristiq.com/http/mime"
//	)
//
//	func main() {
//	    ct := mime.JSON
//	    fmt.Println("Content-Type:", ct.String()) // Output: "application/json"
//	}
//
// # Interface
//
// The package also defines an Interface that requires a String() method returning the MIME type as a string.
// This allows for alternative implementations of MIME-like types that adhere to the same contract, offering additional
// flexibility in your code.
//
// # Conclusion
//
// By providing a strongly typed and well-organized approach to working with Internet Media Types, the mime package
// helps ensure that your applications use valid content types consistently and correctly. This results in clearer, more
// robust HTTP-based code.
package mime
