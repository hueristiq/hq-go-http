package parser

import (
	"errors"
	"fmt"
	"strings"
)

// Link represents a single link parsed from a Link header.
// It contains the target URL, the "rel" attribute (if present),
// and any additional parameters associated with the link.
//
// Fields:
//   - URL (string): The target URI for the link.
//   - Rel (string): The link relation type (e.g., "next", "prev", "author").
//   - Parameters (map[string]string): A map of additional key/value pairs associated with the link,
//     where keys are stored in lower-case for consistent access.
type Link struct {
	URL        string
	Rel        string
	Parameters map[string]string
}

// String returns a string representation of the Link in the format:
//
//	<URL>; key1="value1"; key2="value2"; rel="relation"
//
// The URL is enclosed within angle brackets and followed by a semicolon-separated
// list of parameters. If the "rel" field is non-empty, it is appended to the list.
//
// Returns:
//   - link (string): A formatted string representation of the Link.
func (l Link) String() (link string) {
	params := make([]string, 0, len(l.Parameters)+1)

	for k, v := range l.Parameters {
		params = append(params, fmt.Sprintf("%s=%q", k, v))
	}

	if l.Rel != "" {
		params = append(params, fmt.Sprintf("rel=%q", l.Rel))
	}

	link = fmt.Sprintf("<%s>; %s", l.URL, strings.Join(params, "; "))

	return
}

// HasParameter checks whether the Link contains a parameter with the specified key.
//
// Arguments:
//   - key (string): The parameter name to check for (string).
//
// Returns:
//   - hasParameter (bool): A boolean value that is true if the parameter exists, false otherwise.
func (l Link) HasParameter(key string) (hasParameter bool) {
	_, hasParameter = l.Parameters[key]

	return
}

// Parameter retrieves the value of the parameter identified by the provided key.
// If the parameter does not exist, an empty string is returned.
//
// Arguments:
//   - key (string): The parameter name to retrieve (string).
//
// Returns:
//   - parameter (string): The corresponding value from the Parameters map, or an empty string if not found.
func (l Link) Parameter(key string) (parameter string) {
	parameter = l.Parameters[key]

	return
}

// Links is a slice of Link objects.
// It represents a collection of links, such as those parsed from a Link header.
type Links []Link

// String returns a single string representation of the collection of Links.
// Each Link is formatted using its String method, and the resulting strings are
// joined with a comma and a space. If the Links slice is empty, an empty string is returned.
//
// Returns:
//   - links (string): A single string containing all Link representations joined by ", ".
func (l Links) String() (links string) {
	if len(l) == 0 {
		return
	}

	strs := make([]string, 0, len(l))

	for _, link := range l {
		strs = append(strs, link.String())
	}

	links = strings.Join(strs, ", ")

	return
}

// FilterByRel returns a new Links collection containing only those Link objects
// that have a "rel" attribute matching the provided rel argument.
// The comparison is case-sensitive.
//
// Arguments:
//   - rel (string): The relation type to filter by (string). The comparison is case-sensitive.
//
// Returns:
//   - links (Links): A new Links slice containing only the Link objects with a matching Rel value.
func (l Links) FilterByRel(rel string) (links Links) {
	links = make(Links, 0, len(l))

	for _, link := range l {
		if link.Rel == rel {
			links = append(links, link)
		}
	}

	return
}

var errEmptyParameter = errors.New("empty parameter")

// ParseLinkHeader parses a raw HTTP Link header string into a collection of Links.
// The header string may contain one or more comma-separated link entries.
// Each entry should have the format: <URL>; param1="value1"; param2="value2", etc.
// If the input string is empty, an empty Links slice is returned.
//
// Arguments:
//   - raw (string): The raw HTTP Link header string to be parsed (string).
//
// Returns:
//   - links (Links): A Links slice containing the parsed Link objects. If the raw string is empty,
//     an empty slice is returned.
func ParseLinkHeader(raw string) (links Links) {
	if raw == "" {
		return
	}

	raw = strings.TrimSpace(raw)

	for _, chunk := range strings.Split(raw, ",") {
		chunk = strings.TrimSpace(chunk)
		if chunk == "" {
			continue
		}

		link := Link{
			URL:        "",
			Rel:        "",
			Parameters: make(map[string]string),
		}

		for _, piece := range strings.Split(chunk, ";") {
			piece = strings.TrimSpace(piece)
			if piece == "" {
				continue
			}

			if strings.HasPrefix(piece, "<") && strings.HasSuffix(piece, ">") {
				link.URL = strings.Trim(piece, "<>")

				continue
			}

			key, val, err := parseParameter(piece)
			if err != nil {
				continue
			}

			if key == "" {
				continue
			}

			if strings.EqualFold(key, "rel") {
				link.Rel = val
			} else {
				link.Parameters[strings.ToLower(key)] = val
			}
		}

		if link.URL != "" {
			links = append(links, link)
		}
	}

	return
}

// ParseLinkHeaders processes multiple raw HTTP Link header strings and returns a
// combined collection of Links parsed from all headers.
// This is useful when the link information is spread across several header lines.
//
// Arguments:
//   - headers ([]string): A slice of raw HTTP Link header strings (each string may contain multiple links).
//
// Returns:
//   - links (Links): A combined Links slice containing all parsed Link objects from the provided headers.
func ParseLinkHeaders(headers []string) (links Links) {
	links = make(Links, 0)

	for _, header := range headers {
		links = append(links, ParseLinkHeader(header)...)
	}

	return
}

// parseParameter is an unexported helper function that parses a raw parameter string.
// The expected format of raw is "key=value". It returns the key and value as separate strings.
// If the value is enclosed in double quotes, they are removed.
// If the raw string is empty or improperly formatted, an error is returned.
//
// Arguments:
//   - raw (raw): The raw parameter string to be parsed (e.g., 'rel="next"') (string).
//
// Returns:
//   - key (string): The parsed parameter name (string).
//   - value (string): The parsed parameter value with any surrounding double quotes removed (string).
//   - err (error): An error value which is non-nil if the raw string is empty or improperly formatted.
func parseParameter(raw string) (key, value string, err error) {
	raw = strings.TrimSpace(raw)

	if raw == "" {
		err = errEmptyParameter

		return
	}

	parts := strings.SplitN(raw, "=", 2)
	key = strings.TrimSpace(parts[0])

	if len(parts) == 1 {
		return
	}

	value = strings.Trim(parts[1], "\"")

	return
}
