// Package http provides an HTTP client for resilient, configurable HTTP
// communication.
//
// It suits applications that call HTTP services and must absorb transient
// failure — connection errors, rate limiting (429), and server errors (5xx) —
// without hand-rolling retry loops, and tooling that needs control over the
// finer points: request building, header management, raw requests, and
// User-Agent rotation.
//
// # Entry points
//
// [NewClient] creates a [Client]. Requests are executed with [Client.Do],
// with the convenience methods ([Client.Get], [Client.Post], and friends), or
// assembled fluently with [Client.NewRequest] — a [RequestBuilder] chain such
// as:
//
//	res, err := client.NewRequest().
//		Method(method.GET).
//		BaseURL("https://api.example.com").
//		Path("/users").
//		Param("page", "1").
//		SetHeader(header.Accept.String(), mime.JSON.String()).
//		Body(payload).
//		Do(ctx)
//
// The package-level shortcuts ([Get], [Post], and friends) run the same
// requests against a shared default client built lazily on first use; a
// construction failure is reported by every shortcut, never swallowed.
//
// # Configuration
//
// Client behavior is set through functional options passed to [NewClient]:
// [WithBaseURL], [WithTimeout], [WithHTTPClient], [WithHeader],
// [WithUserAgent] and [WithRandomUserAgent], the retry knobs ([WithMaxRetries],
// [WithRetryWaitMin], [WithRetryWaitMax], [WithRetryBackoff],
// [WithRetryPolicy]), [WithRateLimiter], the observability callbacks
// [WithOnRequest] and [WithOnResponse], [WithMaxRequestsBeforeIdleClose], and
// [WithResponseReadLimit]. Unset options fall back to the exported Default*
// constants, and invalid values are normalized to them.
//
// A single request execution overrides the retry knobs, the rate limiter, and
// the drain bound through [RequestOptionFunc] values ([WithRequestMaxRetries]
// and friends) passed to [Client.Do], the convenience methods, or
// [RequestBuilder.WithOptions]. The client configuration stays untouched.
//
// # Retries
//
// The retry engine is [github.com/hueristiq/hq-lib-retrier-go]: attempts are
// bounded by [WithMaxRetries], delayed by the [WithRetryBackoff] strategy
// between [WithRetryWaitMin] and [WithRetryWaitMax]. After each attempt the
// [RetryPolicy] sees the response AND the error, so policies can react to
// status codes; [DefaultRetryPolicy] retries connection and transport errors,
// 429, and 5xx responses (except 501), and never retries a canceled context,
// an exhausted redirect limit, an unsupported scheme, or a TLS certificate
// verification failure. A Retry-After header on a retried response is honored
// as the wait before the next attempt, capped at the maximum retry wait.
//
// # HTTP/1.x and HTTP/2
//
// The primary client negotiates HTTP/2 automatically. When a cleartext
// HTTP/1.x exchange fails because the server answered with HTTP/2 frames —
// net/http reports an unexported, message-only error — the attempt is retried
// once on a fallback client configured for HTTP/2 before the retry policy is
// consulted.
//
// # Context
//
// Every entry point takes the caller's context, and the whole execution —
// every attempt and every wait — is bounded by it, tightened by the
// configured timeout when the context carries no earlier deadline. A
// per-request timeout needs no option: derive it with context.WithTimeout.
//
// # Concurrency
//
// A [Client] is safe for concurrent use: it is immutable after construction,
// and only the counter driving the periodic idle-connection close mutates,
// through atomics. Callbacks and policies are invoked synchronously on the
// goroutine running the attempt and must themselves be safe for concurrent
// use. A [RequestBuilder] is not safe for concurrent use.
//
// # Errors
//
// Errors are wrapped with the producing function ("http.Client.Do: ...") and
// %w, so errors.Is and errors.As keep working; a context that ends an
// execution surfaces as context.Canceled or context.DeadlineExceeded.
//
// # Out of scope
//
// TLS/browser impersonation (surf-style fingerprinting) is deliberately not
// supported: it requires a utls-class dependency, and this library stays
// stdlib-first. The raw subpackage covers byte-exact HTTP/1.x requests.
//
// The implementation relies on the Go standard library, plus
// github.com/hueristiq/hq-lib-url-go (URL parsing with scheme defaulting), and
// github.com/hueristiq/hq-lib-retrier-go (the retry engine).
package http
