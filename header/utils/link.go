package utils

import (
	"errors"
	"fmt"
	"strings"
)

// Link is a single link parsed from an HTTP Link header, as defined by RFC 8288.
//
// It encapsulates the target URL, the relation type ("rel"), and any additional
// parameters associated with the link. The structure is designed to facilitate
// easy access to link components and to support formatting back into a valid
// Link header string.
//
// Fields:
//   - URL (string): The target URI for the link (e.g., "https://api.example.com/page/2").
//   - Rel (string): The link relation type (e.g., "next", "prev", "author").
//   - Params (map[string]string): A map of additional key/value pairs associated with
//     the link, with keys kept exactly as written in the header (e.g.,
//     {"title": "Next Page"}).
type Link struct {
	URL    string
	Rel    string
	Params map[string]string
}

// String returns a string representation of the Link in the format specified by RFC 8288:
//
//	<URL>; key1="value1"; key2="value2"; rel="relation"
//
// The URL is enclosed in angle brackets (`<` and `>`), followed by semicolon-separated
// parameters. If the Rel field is non-empty, it is appended as a `rel="value"`
// parameter. The method ensures proper quoting of parameter values to comply with
// HTTP Link header syntax.
//
// Returns:
//   - link (string): A formatted string representation of the Link.
func (l Link) String() (link string) {
	params := make([]string, 0, len(l.Params)+1)

	for k, v := range l.Params {
		params = append(params, fmt.Sprintf("%s=%q", k, v))
	}

	if l.Rel != "" {
		params = append(params, fmt.Sprintf("rel=%q", l.Rel))
	}

	link = fmt.Sprintf("<%s>; %s", l.URL, strings.Join(params, "; "))

	return
}

// HasParam checks whether the Link contains a parameter with the specified key.
//
// The key is case-sensitive, matching the key exactly as written in the header.
// This method is useful for checking the presence of specific parameters, such
// as "title" or "type", before attempting to retrieve their values.
//
// Parameters:
//   - key (string): The parameter name to check for.
//
// Returns:
//   - hasParam (bool): True if the parameter exists, false otherwise.
func (l Link) HasParam(key string) (hasParam bool) {
	_, hasParam = l.Params[key]

	return
}

// Param retrieves the value of the parameter identified by the provided key.
//
// If the parameter does not exist, an empty string is returned. The key is
// case-sensitive, matching the key exactly as written in the header. This method
// provides a safe way to access parameter values without directly interacting
// with the Params map.
//
// Parameters:
//   - key (string): The parameter name to retrieve.
//
// Returns:
//   - param (string): The value of the parameter, or an empty string if not found.
func (l Link) Param(key string) (param string) {
	param = l.Params[key]

	return
}

// Links is a slice of Link objects, representing a collection of links.
//
// This type is used to store multiple links parsed from one or more HTTP Link
// headers, such as those used in pagination or hypermedia APIs. It provides
// methods for formatting the collection as a string and filtering links by their
// relation type.
type Links []Link

// String returns a single string representation of the Links collection.
//
// Each Link is formatted using its String method, and the resulting strings are
// joined with a comma and a space (", ") to conform to the HTTP Link header
// format. If the slice is empty, an empty string is returned.
//
// Returns:
//   - links (string): A comma-separated string of all link representations.
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

// FilterByRel returns a new Links collection containing only the Link objects
// with a Rel value matching the provided rel argument.
//
// The comparison is case-sensitive, as relation types are typically case-sensitive
// per RFC 8288. This method is useful for extracting links with specific relation
// types, such as "next" or "prev" for pagination, from a larger collection.
//
// Parameters:
//   - rel (string): The relation type to filter by (case-sensitive).
//
// Returns:
//   - links (Links): A new Links slice containing only links with a matching Rel value.
func (l Links) FilterByRel(rel string) (links Links) {
	links = make(Links, 0, len(l))

	for _, link := range l {
		if link.Rel == rel {
			links = append(links, link)
		}
	}

	return
}

// Parse parses a raw HTTP Link header string into a Links collection.
//
// The input string may contain one or more comma-separated link entries, each in
// the format: `<URL>; param1="value1"; param2="value2"`. The function splits the
// input into individual links, extracts the URL (enclosed in angle brackets), and
// parses parameters, including the "rel" attribute (matched case-insensitively).
// Parameter keys are kept exactly as written. If the input is empty or contains
// only invalid entries, an empty Links slice is returned.
//
// Parameters:
//   - raw (string): The raw HTTP Link header string to parse.
//
// Returns:
//   - links (Links): A slice of Link objects representing the parsed links.
func Parse(raw string) (links Links) {
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
			URL:    "",
			Rel:    "",
			Params: make(map[string]string),
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
				link.Params[key] = val
			}
		}

		if link.URL != "" {
			links = append(links, link)
		}
	}

	return
}

// ParseMultiple is like Parse, but accepts a slice of header values rather
// than just one header string.
//
// This function is useful when Link headers are spread across multiple header
// lines, as allowed by HTTP specifications (e.g., multiple "Link" headers in an
// HTTP response). It parses each header string using Parse and combines the
// results into a single Links slice.
//
// Parameters:
//   - headers ([]string): A slice of raw HTTP Link header strings, each
//     potentially containing multiple links.
//
// Returns:
//   - links (Links): A combined Links slice containing all parsed Link objects.
func ParseMultiple(headers []string) (links Links) {
	links = make(Links, 0, len(headers))

	for _, header := range headers {
		links = append(links, Parse(header)...)
	}

	return
}

// parseParameter parses a raw parameter string in the format "key=value".
//
// It extracts the key and value, removing surrounding double quotes from the
// value if present. The function is unexported, as it is an internal helper used
// by Parse. It handles malformed input by returning an error for empty or
// improperly formatted parameters.
//
// Parameters:
//   - raw (string): The raw parameter string to parse (e.g., `rel="next"`,
//     `title="Next Page"`).
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
