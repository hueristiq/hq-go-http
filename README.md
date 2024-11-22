# hq-go-http

![made with go](https://img.shields.io/badge/made%20with-Go-1E90FF.svg) [![go report card](https://goreportcard.com/badge/github.com/hueristiq/xsubfind3r)](https://goreportcard.com/report/github.com/hueristiq/hq-go-http) [![license](https://img.shields.io/badge/license-MIT-gray.svg?color=1E90FF)](https://github.com/hueristiq/hq-go-http/blob/master/LICENSE) ![maintenance](https://img.shields.io/badge/maintained%3F-yes-1E90FF.svg) [![open issues](https://img.shields.io/github/issues-raw/hueristiq/hq-go-http.svg?style=flat&color=1E90FF)](https://github.com/hueristiq/hq-go-http/issues?q=is:issue+is:open) [![closed issues](https://img.shields.io/github/issues-closed-raw/hueristiq/hq-go-http.svg?style=flat&color=1E90FF)](https://github.com/hueristiq/hq-go-http/issues?q=is:issue+is:closed) [![contribution](https://img.shields.io/badge/contributions-welcome-1E90FF.svg)](https://github.com/hueristiq/hq-go-http/blob/master/CONTRIBUTING.md)

`hq-go-http` is a [Go (Golang)](http://golang.org/) HTTP client package for robust web communication. It is built with retry policies, digest authentication support, and automatic fallback to HTTP/2, offering a highly resilient solution for making HTTP requests.

## Resource

* [Features](#features)
* [Installation](#installation)
* [Usage](#usage)
* [Contributing](#contributing)
* [Licensing](#licensing)
* [Credits](#credits)
	* [Contributors](#contributors)
	* [Alternatives](#`)

## Features

* **Retry Logic:** Automatically retries failed requests based on customizable rules. You can define how many times to retry, how long to wait between retries, and how the wait time should increase.
* **Digest Authentication:** Supports digest authentication, a security mechanism used in some web services, to ensure safe and secure communication.
* **HTTP/2 Fallback:** If an HTTP/1.x request fails, the client can automatically retry the request over HTTP/2.
* **Timeout and Error Handling:** Configurable timeouts to avoid hanging requests, and custom error handlers to gracefully manage failures.
* **Idle Connection Management:** Efficient management of idle connections to reduce unnecessary resource consumption.

## Installation

To install the package, run the following command in your terminal:

```bash
go get -v -u github.com/hueristiq/hq-go-http
```

This command will download and install the `hq-go-http` package into your Go workspace, making it available for use in your projects.

## Usage

Here's a simple example demonstrating how to use `hq-go-http`:

```go
package main

import (
	"log"

	http "github.com/hueristiq/hq-go-http"
)

func main() {
	client := http.NewClient(&http.ClientConfiguration{
		RetryMax:     3,              // Max number of retries
		Timeout:      10 * time.Second, // Request timeout
		RetryWaitMin: 1 * time.Second, // Minimum wait between retries
		RetryWaitMax: 5 * time.Second, // Maximum wait between retries
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

We welcome contributions! Feel free to submit [Pull Requests](https://github.com/hueristiq/hq-go-http/pulls) or report [Issues](https://github.com/hueristiq/hq-go-http/issues). For more details, check out the [contribution guidelines](https://github.com/hueristiq/hq-go-http/blob/master/CONTRIBUTING.md).

## Licensing

This package is licensed under the [MIT license](https://opensource.org/license/mit). You are free to use, modify, and distribute it, as long as you follow the terms of the license. You can find the full license text in the repository - [Full MIT license text](https://github.com/hueristiq/hq-go-http/blob/master/LICENSE).

## Credits

### Contributors

A huge thanks to all the contributors who have helped make `hq-go-http` what it is today!

[![contributors](https://contrib.rocks/image?repo=hueristiq/hq-go-http&max=500)](https://github.com/hueristiq/hq-go-http/graphs/contributors)

### Alternatives

If you're interested in more packages like this, check out:

[Hashicorp's go-retryablehttp](https://github.com/hashicorp/go-retryablehttp) ◇ [Project Discovery's retryablehttp-go](https://github.com/projectdiscovery/retryablehttp-go) ◇ [Gojek's heimdall](https://github.com/gojek/heimdall) ◇ [Monaco's request](https://github.com/monaco-io/request) ◇ [Opus Domini's fast-shot](https://github.com/opus-domini/fast-shot)