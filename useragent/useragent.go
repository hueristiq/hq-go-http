package useragent

import (
	"math/rand/v2"
)

// defaultUserAgents is the curated list of User-Agent values [Random] draws
// from: recent releases of the major browsers across the common desktop and
// mobile platforms. The list favors widespread versions so rotated requests
// blend in with ordinary traffic.
var defaultUserAgents = []string{
	// Chrome on Windows, macOS, and Linux.
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36",
	// Firefox on Windows, macOS, and Linux.
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:143.0) Gecko/20100101 Firefox/143.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14.7; rv:143.0) Gecko/20100101 Firefox/143.0",
	"Mozilla/5.0 (X11; Linux x86_64; rv:143.0) Gecko/20100101 Firefox/143.0",
	// Safari on macOS and iOS.
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_7_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.4 Safari/605.1.15",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 18_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.6 Mobile/15E148 Safari/604.1",
	// Edge on Windows.
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36 Edg/141.0.0.0",
	// Chrome on Android.
	"Mozilla/5.0 (Linux; Android 15; Pixel 9) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Mobile Safari/537.36",
}

// List returns the curated list of User-Agent values [Random] draws from.
//
// The returned slice is a copy: mutating it never affects the package.
//
// Returns:
//   - userAgents ([]string): the curated User-Agent values.
func List() (userAgents []string) {
	userAgents = make([]string, len(defaultUserAgents))

	copy(userAgents, defaultUserAgents)

	return
}

// Random returns a random User-Agent value from the curated list.
//
// The draw uses a pseudorandom source, not a cryptographic one: it suits
// request rotation, not security decisions.
//
// Returns:
//   - userAgent (string): a random User-Agent value.
func Random() (userAgent string) {
	//nolint:gosec // The draw only rotates request cosmetics; predictability is not a security property here.
	userAgent = defaultUserAgents[rand.IntN(len(defaultUserAgents))]

	return
}
