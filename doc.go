// Package http provides an advanced HTTP client library that offers a rich set of features
// for resilient and highly configurable HTTP communication. It integrates support for both
// HTTP/1.x and HTTP/2, configurable retry logic with customizable backoff strategies, digest
// authentication support, and robust connection management.
//
// Usage Example:
//
//	package main
//
//	import (
//	    "fmt"
//	    "log"
//	    "time"
//
//	    "go.source.hueristiq.com/http"
//	)
//
//	func main() {
//	    // Initialize a new HTTP client using the default single-use configuration.
//	    client, err := http.NewClient(http.DefaultSingleClientConfiguration)
//	    if err != nil {
//	        log.Fatalf("Failed to create HTTP client: %v", err)
//	    }
//
//	    // Construct and send an HTTP GET request using the fluent RequestBuilder.
//	    // The RequestBuilder API allows for setting the method, URL, headers, and more.
//	    resp, err := client.Get("https://api.example.com/data")
//	    if err != nil {
//	        log.Fatalf("HTTP request failed: %v", err)
//	    }
//
//	    defer resp.Body.Close()
//
//	    fmt.Println("Response status:", resp.Status)
//	}
package http
