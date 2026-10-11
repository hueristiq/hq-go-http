# hq-lib-http-go

![made with go](https://img.shields.io/badge/made%20with-Go-1E90FF.svg) [![go reference](https://pkg.go.dev/badge/github.com/hueristiq/hq-lib-http-go.svg)](https://pkg.go.dev/github.com/hueristiq/hq-lib-http-go) [![license](https://img.shields.io/badge/license-MIT-gray.svg?color=1E90FF)](https://github.com/hueristiq/hq-lib-http-go/blob/main/LICENSE) ![maintenance](https://img.shields.io/badge/maintained%3F-yes-1E90FF.svg) [![open issues](https://img.shields.io/github/issues-raw/hueristiq/hq-lib-http-go.svg?style=flat&color=1E90FF)](https://github.com/hueristiq/hq-lib-http-go/issues?q=is:issue+is:open) [![closed issues](https://img.shields.io/github/issues-closed-raw/hueristiq/hq-lib-http-go.svg?style=flat&color=1E90FF)](https://github.com/hueristiq/hq-lib-http-go/issues?q=is:issue+is:closed) [![contribution](https://img.shields.io/badge/contributions-welcome-1E90FF.svg)](https://github.com/hueristiq/hq-lib-http-go/blob/main/CONTRIBUTING.md)

`hq-lib-http-go` is a [Go](https://golang.org/) package for resilient, configurable HTTP communication: a client with context-aware retries that see both responses and errors, a fluent request builder, a rate-limiter hook, HTTP/2 fallback, raw HTTP/1.x requests, and User-Agent rotation.

## Resources

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
	- [Basic Client](#basic-client)
	- [Fluent Request Builder](#fluent-request-builder)
	- [Retry Configuration](#retry-configuration)
	- [Raw Requests](#raw-requests)
- [Configuration](#configuration)
- [Contributing](#contributing)
- [Licensing](#licensing)

## Features

- **Context-Aware Retries:** Every entry point honors the caller's context. The retry engine retries connection and transport errors, `429`, and `5xx` responses, and honors `Retry-After` headers (capped at the maximum retry wait). The retry policy sees the response *and* the error of every attempt.
- **Fluent Request Builder:** Assemble requests with a chainable `Method().BaseURL().Path().Param().SetHeader().AddHeader().Body().Do(ctx)` API.
- **HTTP/1.x and HTTP/2:** The primary client negotiates HTTP/2 automatically; when a cleartext HTTP/1.x exchange fails because the server answered with HTTP/2 frames, the attempt is retried once on an HTTP/2 fallback client. The fallback resends the request: reusable bodies rewind, bodies exposing `GetBody` are refetched, and other bodies skip the fallback with the original error.
- **Rate-Limiter Hook:** `WithRateLimiter` accepts any limiter with a `Wait(ctx) error` method — for example [hq-lib-limiter-go](https://github.com/hueristiq/hq-lib-limiter-go) — without imposing a module dependency.
- **Observability Callbacks:** `WithOnRequest` and `WithOnResponse` run around every attempt, for logging, metrics, and tracing.
- **Raw HTTP/1.x Requests:** The `raw` subpackage sends raw request bytes over dialed connections — no normalization, no pooling, one exchange per connection.
- **User-Agent Rotation:** `WithRandomUserAgent()` draws from a curated list of recent browser User-Agents; `WithUserAgent` sets a fixed one.
- **Typed Companion Packages:** `header`, `method`, `mime`, and `status` provide strongly typed constants; `header/utils` canonicalizes header keys and parses RFC 8288 Link headers.
- **Connection Management:** Optional periodic idle-connection closing (`WithMaxRequestsBeforeIdleClose`) bounds resource usage in long-running processes.

TLS/browser impersonation (surf-style fingerprinting) is deliberately out of scope: it requires a utls-class dependency, and this library stays stdlib-first.

## Installation

To install `hq-lib-http-go`, run the following command in your Go project:

```bash
go get -v -u github.com/hueristiq/hq-lib-http-go
```

## Usage

The examples below import the package under the `hqgohttp` alias.

```go
import hqgohttp "github.com/hueristiq/hq-lib-http-go"
```

### Basic Client

`NewClient` takes functional options; the convenience methods take the context first:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	hqgohttp "github.com/hueristiq/hq-lib-http-go"
)

func main() {
	client, err := hqgohttp.NewClient(
		hqgohttp.WithTimeout(10*time.Second),
		hqgohttp.WithMaxRetries(3),
		hqgohttp.WithRetryWaitMin(time.Second),
		hqgohttp.WithRetryWaitMax(5*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to create HTTP client: %v", err)
	}

	response, err := client.Get(context.Background(), "https://example.com")
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	defer response.Body.Close()

	fmt.Println("Response status:", response.Status)
}
```

### Fluent Request Builder

`NewRequest` starts a chainable builder; the client executes the built request with its retry engine, rate limiter, and callbacks:

```go
package main

import (
	"context"
	"fmt"
	"io"
	"log"

	hqgohttp "github.com/hueristiq/hq-lib-http-go"
	"github.com/hueristiq/hq-lib-http-go/header"
	"github.com/hueristiq/hq-lib-http-go/method"
	"github.com/hueristiq/hq-lib-http-go/mime"
)

func main() {
	client, err := hqgohttp.NewClient(
		hqgohttp.WithBaseURL("https://api.example.com"),
	)
	if err != nil {
		log.Fatalf("Failed to create HTTP client: %v", err)
	}

	response, err := client.NewRequest().
		Method(method.POST).
		Path("/users").
		Param("page", "1").
		SetHeader(header.ContentType.String(), mime.JSON.String()).
		AddHeader(header.Accept.String(), mime.JSON.String()).
		Body(`{"name":"ada"}`).
		Do(context.Background())
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Reading response failed: %v", err)
	}

	fmt.Println("Response status:", response.Status)
	fmt.Println("Response body:", string(body))
}
```

### Retry Configuration

The retry policy sees the response and the error of every attempt. The default policy retries connection and transport errors, `429`, and `5xx` (except `501`); a custom policy can react to anything:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	hqgohttp "github.com/hueristiq/hq-lib-http-go"
	hqgoretrierbackoff "github.com/hueristiq/hq-lib-retrier-go/backoff"
)

func main() {
	client, err := hqgohttp.NewClient(
		hqgohttp.WithMaxRetries(4),                        // retries after the initial attempt
		hqgohttp.WithRetryWaitMin(500*time.Millisecond),   // backoff bounds
		hqgohttp.WithRetryWaitMax(10*time.Second),
		hqgohttp.WithRetryBackoff(hqgoretrierbackoff.ExponentialWithFullJitter),
		hqgohttp.WithRetryPolicy(func(ctx context.Context, resp *http.Response, err error) (bool, error) {
			if err == nil && resp != nil && resp.StatusCode == http.StatusForbidden {
				return false, nil // retrying cannot fix a 403
			}

			return hqgohttp.DefaultRetryPolicy(ctx, resp, err)
		}),
	)
	if err != nil {
		log.Fatalf("Failed to create HTTP client: %v", err)
	}

	response, err := client.Get(context.Background(), "https://example.com",
		hqgohttp.WithRequestMaxRetries(1), // per-request override
	)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	defer response.Body.Close()

	fmt.Println("Response status:", response.Status)
}
```

### Raw Requests

The `raw` subpackage sends raw HTTP/1.x request bytes over a freshly dialed connection — one exchange, no pooling:

```go
package main

import (
	"context"
	"fmt"
	"io"
	"log"

	hqgohttpraw "github.com/hueristiq/hq-lib-http-go/raw"
)

func main() {
	request := []byte("GET / HTTP/1.1\r\nHost: example.com\r\nConnection: close\r\n\r\n")

	response, err := hqgohttpraw.Do(context.Background(), "example.com:80", request)
	if err != nil {
		log.Fatalf("Raw request failed: %v", err)
	}

	defer response.Body.Close() // closing the body closes the connection

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Reading response failed: %v", err)
	}

	fmt.Println("Response status:", response.Status)
	fmt.Println("Response body:", string(body))
}
```

## Configuration

### Client options

Behavior is set through functional options passed to `NewClient`. Any option left unset falls back to its default, and invalid values are normalized to the defaults.

| Option | Description | Default |
| --- | --- | --- |
| `WithHTTPClient(client)` | Underlying `*http.Client` executing HTTP/1.x requests. Nil is ignored. | pooled client on `DefaultHTTPPooledTransport` |
| `WithBaseURL(baseURL)` | Base URL prepended to relative request URLs. | none |
| `WithHeader(key, value)` | Default header set on every request; a request-level header with the same key wins. | none |
| `WithUserAgent(userAgent)` | Fixed User-Agent header. Empty falls back to the default. | `DefaultUserAgent` |
| `WithRandomUserAgent()` | Draw a random User-Agent per request from the curated `useragent` list. The last of `WithUserAgent`/`WithRandomUserAgent` wins. | disabled |
| `WithTimeout(d)` | Bound on a whole execution, including retries and waits. `<= 0` falls back to the default. | `DefaultTimeout` (`30s`) |
| `WithMaxRetries(n)` | Retries after the initial attempt (`2` means up to 3 attempts). `0` disables retries; `< 0` falls back to the default. | `DefaultMaxRetries` (`2`) |
| `WithRetryWaitMin(d)` | Minimum delay between attempts. `<= 0` falls back to the default. | `DefaultRetryWaitMin` (`1s`) |
| `WithRetryWaitMax(d)` | Maximum delay between attempts; also caps an honored `Retry-After` wait. `<= 0` falls back to the default. | `DefaultRetryWaitMax` (`30s`) |
| `WithRetryBackoff(fn)` | Constructor for the backoff strategy, called once per execution. Nil selects the default. | `backoff.ExponentialWithDecorrelatedJitter` |
| `WithRetryPolicy(policy)` | `func(ctx, resp, err) (retry bool, rerr error)` deciding per attempt whether to retry. Nil selects the default. | `DefaultRetryPolicy` |
| `WithRateLimiter(l)` | Limiter (`Wait(ctx) error`) waited on before every attempt. Nil disables limiting. | none |
| `WithOnRequest(fn)` | Callback invoked with the outgoing request before every attempt. | none |
| `WithOnResponse(fn)` | Callback invoked with the response/error of every attempt. | none |
| `WithMaxRequestsBeforeIdleClose(n)` | Close idle connections after n completed executions. `<= 0` disables. | disabled |
| `WithResponseReadLimit(n)` | Maximum bytes drained from a retried response's body so the connection can be reused. `<= 0` falls back to the default. | `DefaultResponseReadLimit` (`4096`) |

### Request options

A single execution overrides the execution knobs through `RequestOptionFunc` values passed to `Do`, the convenience methods, or `RequestBuilder.WithOptions`. The client configuration stays untouched; invalid values are ignored, keeping the client configuration.

| Option | Description |
| --- | --- |
| `WithRequestMaxRetries(n)` | Override the number of retries (`0` disables retrying the request). |
| `WithRequestRetryWaitMin(d)` / `WithRequestRetryWaitMax(d)` | Override the backoff bounds. |
| `WithRequestRetryBackoff(fn)` | Override the backoff strategy constructor. |
| `WithRequestRetryPolicy(policy)` | Override the retry policy. |
| `WithRequestRateLimiter(l)` | Override the rate limiter. |
| `WithRequestResponseReadLimit(n)` | Override the drain bound. |

A per-request timeout needs no option: derive it with `context.WithTimeout` on the context handed to the entry point.

## Contributing

Contributions are welcome and encouraged! Feel free to submit [Pull Requests](https://github.com/hueristiq/hq-lib-http-go/pulls) or report [Issues](https://github.com/hueristiq/hq-lib-http-go/issues). For more details, check out the [contribution guidelines](https://github.com/hueristiq/hq-lib-http-go/blob/main/CONTRIBUTING.md).

A big thank you to all the [contributors](https://github.com/hueristiq/hq-lib-http-go/graphs/contributors) for your ongoing support!

![contributors](https://contrib.rocks/image?repo=hueristiq/hq-lib-http-go&max=500)

## Licensing

This package is licensed under the [MIT license](https://opensource.org/license/mit). You are free to use, modify, and distribute it, as long as you follow the terms of the license. You can find the full license text in the repository - [Full MIT license text](https://github.com/hueristiq/hq-lib-http-go/blob/main/LICENSE).
