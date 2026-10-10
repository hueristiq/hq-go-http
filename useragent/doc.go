// Package useragent provides User-Agent values for HTTP requests.
//
// It suits clients that should not announce themselves with a library default
// — crawlers, scrapers, security tooling — where blending in with ordinary
// browser traffic matters more than identifying the tool.
//
// # Entry points
//
// [List] returns the curated list of recent browser User-Agent values across
// the common desktop and mobile platforms. [Random] draws one value from that
// list at random, for per-request rotation.
//
// The draw is pseudorandom, not cryptographic: it suits rotation, not
// security decisions.
//
// The implementation relies on the Go standard library only.
package useragent
