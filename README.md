# hq-go-http

![made with go](https://img.shields.io/badge/made%20with-Go-1E90FF.svg) [![go report card](https://goreportcard.com/badge/github.com/hueristiq/hq-go-http)](https://goreportcard.com/report/github.com/hueristiq/hq-go-http) [![license](https://img.shields.io/badge/license-MIT-gray.svg?color=1E90FF)](https://github.com/hueristiq/hq-go-http/blob/master/LICENSE) ![maintenance](https://img.shields.io/badge/maintained%3F-yes-1E90FF.svg) [![open issues](https://img.shields.io/github/issues-raw/hueristiq/hq-go-http.svg?style=flat&color=1E90FF)](https://github.com/hueristiq/hq-go-http/issues?q=is:issue+is:open) [![closed issues](https://img.shields.io/github/issues-closed-raw/hueristiq/hq-go-http.svg?style=flat&color=1E90FF)](https://github.com/hueristiq/hq-go-http/issues?q=is:issue+is:closed) [![contribution](https://img.shields.io/badge/contributions-welcome-1E90FF.svg)](https://github.com/hueristiq/hq-go-http/blob/master/CONTRIBUTING.md)

`hq-go-http` is a [Go (Golang)](http://golang.org/) package for robust and flexible HTTP communication. It offers advanced features such as configurable retry policies, automatic fallback between HTTP/1.x and HTTP/2, and fluent request building with connection management.

## Resource

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Contributing](#contributing)
- [Licensing](#licensing)

## Features

- **Configurable Retry Logic:** Customize retry policies and backoff to handle transient network errors gracefully.
- **HTTP/1.x and HTTP/2 Support:** The client maintains both HTTP/1.x and HTTP/2 clients. If the HTTP/1.x client encounters a specific transport error, the package automatically falls back to HTTP/2.
- **Connection Management:** Automatically drain and close idle connections to prevent resource exhaustion.

## Installation

To install `hq-go-http`, run:

```bash
go get -v -u go.source.hueristiq.com/http
```

Make sure your Go environment is set up properly (Go 1.x or later is recommended).

## Usage

Here's a simple example demonstrating how to use `hq-go-http`:

```go
package main

import (
	"log"

	hqgohttp "go.source.hueristiq.com/http"
)

func main() {
	client := hqgohttp.NewClient(&hqgohttp.ClientConfiguration{
		RetryMax:     3,                // Max number of retries
		Timeout:      10 * time.Second, // Request timeout
		RetryWaitMin: 1 * time.Second,  // Minimum wait between retries
		RetryWaitMax: 5 * time.Second,  // Maximum wait between retries
	})

	response, err := client.Get("https://example.com")
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	defer response.Body.Close()
	// Handle response here
}
```

## Contributing

Contributions are welcome and encouraged! Feel free to submit [Pull Requests](https://github.com/hueristiq/hq-go-http/pulls) or report [Issues](https://github.com/hueristiq/hq-go-http/issues). For more details, check out the [contribution guidelines](https://github.com/hueristiq/hq-go-http/blob/master/CONTRIBUTING.md).

A big thank you to all the [contributors](https://github.com/hueristiq/hq-go-http/graphs/contributors) for your ongoing support!

![contributors](https://contrib.rocks/image?repo=hueristiq/hq-go-http&max=500)

## Licensing

This package is licensed under the [MIT license](https://opensource.org/license/mit). You are free to use, modify, and distribute it, as long as you follow the terms of the license. You can find the full license text in the repository - [Full MIT license text](https://github.com/hueristiq/hq-go-http/blob/master/LICENSE).