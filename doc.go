// Package http provides an advanced HTTP client with a rich set of features including configurable
// retry policies, fallback to HTTP/2, digest authentication support, and robust connection management.
// It encapsulates both HTTP/1.x and HTTP/2 clients to enable resilient and highly configurable HTTP
// communication.
//
// Key Features:
//
//   - **Configurable Retry Logic:**
//     Supports customizable retry policies and backoff strategies to handle transient network errors.
//     The client uses a retry loop (powered by the retrier package) that can be tailored using parameters
//     such as maximum retries, minimum/maximum wait times, and backoff functions.
//
//   - **HTTP/1.x and HTTP/2 Support:**
//     The client maintains separate underlying HTTP/1.x and HTTP/2 clients. If the HTTP/1.x client
//     encounters a specific transport error (e.g., a malformed HTTP version error), the client automatically
//     falls back to the HTTP/2 client.
//
//   - **Fluent Request Building:**
//     A fluent API is provided via the RequestBuilder, allowing easy construction of HTTP requests.
//     Default headers and base URLs can be configured, and request bodies are supported via a reusable
//     request wrapper.
//
//   - **Custom Hooks:**
//     Developers can attach custom hook functions that are executed:
//
//   - **OnRequest:** Before each HTTP request (for logging or modification).
//
//   - **OnResponse:** After receiving an HTTP response (for logging or inspection).
//
//   - **OnError:** When all retry attempts are exhausted (to perform custom error handling).
//
//   - **Connection Management:**
//     The client can automatically drain and close idle connections after a certain number of requests,
//     which helps prevent resource exhaustion in long-running applications. The behavior is configurable
//     via the ClientConfiguration.
//
//   - **Default Configurations:**
//     Predefined configurations (e.g., DefaultSingleClientConfiguration and DefaultSprayingClientConfiguration)
//     are available to simplify setup for common use cases such as single-use clients or host-spraying scenarios.
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
//	    // Create a new client using the default single-use configuration.
//	    client, err := http.NewClient(http.DefaultSingleClientConfiguration)
//	    if err != nil {
//	        log.Fatalf("Failed to create client: %v", err)
//	    }
//
//	    // Build and send an HTTP GET request.
//	    resp, err := client.Request().
//	        Method("GET").
//	        URL("https://api.example.com/data").
//	        Send()
//	    if err != nil {
//	        log.Fatalf("Request failed: %v", err)
//	    }
//	    defer resp.Body.Close()
//
//	    fmt.Println("Response status:", resp.Status)
//	}
//
// For additional details, see the following types and functions:
//   - **Client:** The core type that manages HTTP/1.x and HTTP/2 clients, retry logic, and hooks.
//   - **ClientConfiguration:** Defines various configuration options including timeouts, retry policies,
//     and connection management settings.
//   - **RequestBuilder:** Provides a fluent API to build and send HTTP requests.
//   - **RetryPolicy:** A function type used to determine whether a request should be retried based on
//     the encountered error.
//
// Also, refer to the subpackage "go.source.hueristiq.com/http/request" for the Request wrapper that
// allows for reusable HTTP request bodies.
//
// This package is part of the hueristiq HTTP client library which is designed for resilience,
// flexibility, and ease-of-use when interacting with HTTP services.
package http
