package utils

import (
	"errors"
	"fmt"
	"strings"
)

// ParsedLink represents a single link parsed from an HTTP Link header, as defined by RFC 8288.
//
// It encapsulates the target URL, the relation type ("rel"), and any additional parameters
// associated with the link. The structure is designed to facilitate easy access to link
// components and to support formatting back into a valid Link header string. The Parameters
// map stores keys in lowercase for consistent access, regardless of the case used in the input.
//
// Fields:
//   - URL (string): The target URI for the link (e.g., "https://api.example.com/page/2").
//   - Rel (string): The link relation type (e.g., "next", "prev", "author").
//   - Parameters (map[string]string): A map of additional key/value pairs associated with the link,
//     where keys are stored in lowercase for consistent access (e.g., {"title": "Next Page"}).
type ParsedLink struct {
	URL        string
	Rel        string
	Parameters map[string]string
}

// String returns a string representation of the ParsedLink in the format specified by RFC 8288:
//
//	<URL>; key1="value1"; key2="value2"; rel="relation"
//
// The URL is enclosed in angle brackets (`<` and `>`), followed by semicolon-separated parameters.
// If the Rel field is non-empty, it is appended as a `rel="value"` parameter. The method ensures
// proper quoting of parameter values to comply with HTTP Link header syntax.
//
// Returns:
//   - link (string): A formatted string representation of the ParsedLink.
func (l ParsedLink) String() (link string) {
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

// HasParameter checks whether the ParsedLink contains a parameter with the specified key.
//
// The key is case-sensitive, as the Parameters map stores keys in lowercase (as enforced
// during parsing). This method is useful for checking the presence of specific parameters,
// such as "title" or "type", before attempting to retrieve their values.
//
// Parameters:
//   - key (string): The parameter name to check for.
//
// Returns:
//   - hasParameter (bool): True if the parameter exists, false otherwise.
func (l ParsedLink) HasParameter(key string) (hasParameter bool) {
	_, hasParameter = l.Parameters[key]

	return
}

// Parameter retrieves the value of the parameter identified by the provided key.
//
// If the parameter does not exist, an empty string is returned. The key is case-sensitive,
// as the Parameters map stores keys in lowercase. This method provides a safe way to access
// parameter values without directly interacting with the Parameters map.
//
// Parameters:
//   - key (string): The parameter name to retrieve.
//
// Returns:
//   - parameter (string): The value of the parameter, or an empty string if not found.
func (l ParsedLink) Parameter(key string) (parameter string) {
	parameter = l.Parameters[key]

	return
}

// ParsedLinks is a slice of ParsedLink objects, representing a collection of links.
//
// This type is used to store multiple links parsed from one or more HTTP Link headers,
// such as those used in pagination or hypermedia APIs. It provides methods for formatting
// the collection as a string and filtering links by their relation type.
type ParsedLinks []ParsedLink

// String returns a single string representation of the ParsedLinks collection.
//
// Each ParsedLink is formatted using its String method, and the resulting strings are
// joined with a comma and a space (", ") to conform to the HTTP Link header format.
// If the slice is empty, an empty string is returned.
//
// Returns:
//   - links (string): A comma-separated string of all link representations.
func (l ParsedLinks) String() (links string) {
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

// FilterByRel returns a new ParsedLinks collection containing only the ParsedLink objects
// with a Rel value matching the provided rel argument.
//
// The comparison is case-sensitive, as relation types are typically case-sensitive per RFC 8288.
// This method is useful for extracting links with specific relation types, such as "next" or "prev"
// for pagination, from a larger collection.
//
// Parameters:
//   - rel (string): The relation type to filter by (case-sensitive).
//
// Returns:
//   - links (ParsedLinks): A new ParsedLinks slice containing only links with a matching Rel value.
func (l ParsedLinks) FilterByRel(rel string) (links ParsedLinks) {
	links = make(ParsedLinks, 0, len(l))

	for _, link := range l {
		if link.Rel == rel {
			links = append(links, link)
		}
	}

	return
}

// ParseLinkHeaderValue parses a raw HTTP Link header string into a ParsedLinks collection.
//
// The input string may contain one or more comma-separated link entries, each in the format:
// `<URL>; param1="value1"; param2="value2"`. The function splits the input into individual
// links, extracts the URL (enclosed in angle brackets), and parses parameters, including the
// "rel" attribute. Parameters are stored in a case-insensitive manner (keys are converted to
// lowercase). If the input is empty or contains only invalid entries, an empty ParsedLinks slice
// is returned.
//
// Parameters:
//   - value (string): The raw HTTP Link header string to parse.
//
// Returns:
//   - links (ParsedLinks): A slice of ParsedLink objects representing the parsed links.
func ParseLinkHeaderValue(value string) (links ParsedLinks) {
	if value == "" {
		return
	}

	value = strings.TrimSpace(value)

	for _, chunk := range strings.Split(value, ",") {
		chunk = strings.TrimSpace(chunk)
		if chunk == "" {
			continue
		}

		link := ParsedLink{
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

// ParseLinkHeaderValues processes multiple raw HTTP Link header strings and returns a combined ParsedLinks collection.
//
// This function is useful when Link headers are spread across multiple header lines, as allowed by
// HTTP specifications (e.g., multiple "Link" headers in an HTTP response). It parses each header
// string using ParseLinkHeader and combines the results into a single ParsedLinks slice.
//
// Parameters:
//   - headers ([]string): A slice of raw HTTP Link header strings, each potentially containing multiple links.
//
// Returns:
//   - links (ParsedLinks): A combined ParsedLinks slice containing all parsed ParsedLink objects.
func ParseLinkHeaderValues(headers []string) (links ParsedLinks) {
	links = make(ParsedLinks, 0)

	for _, header := range headers {
		links = append(links, ParseLinkHeaderValue(header)...)
	}

	return
}

// parseParameter parses a raw parameter string in the format "key=value".
//
// It extracts the key and value, removing surrounding double quotes from the value if present.
// The function is unexported, as it is an internal helper used by ParseLinkHeader. It handles
// malformed input by returning an error for empty or improperly formatted parameters.
//
// Parameters:
//   - raw (string): The raw parameter string to parse (e.g., `rel="next"`, `title="Next Page"`).
//
// Returns:
//   - key (string): The parsed parameter name.
//   - value (string): The parsed parameter value, with surrounding quotes removed.
//   - err (error): Non-nil if the raw string is empty or improperly formatted.
func parseParameter(raw string) (key, value string, err error) {
	raw = strings.TrimSpace(raw)

	if raw == "" {
		err = errors.New("empty parameter")

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
