// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package nameable

import (
	"errors"
	"strings"
)

// Errors returned by ParseSyntax explaining why raw could not be parsed.
var (
	// ErrNoSyntax indicates raw has no '|' at all, i.e. it's an ordinary, plain nameable key.
	ErrNoSyntax = errors.New("nameable: no pipe-delimited syntax present")
	// ErrEmptyLabel indicates the segment before the first '|' is empty.
	ErrEmptyLabel = errors.New("nameable: empty label")
	// ErrNoOptions indicates there's neither a literal option nor the FreeFormToken.
	ErrNoOptions = errors.New("nameable: no options and not free-form")
)

// Special tokens within a Marker's pipe-delimited segments that toggle behavior or carry a tooltip line, rather than
// acting as a literal option.
const (
	AllowEmptyToken = "?"
	FreeFormToken   = "*"
	tooltipPrefix   = "tt("
	tooltipSuffix   = ")"
)

// Marker describes a nameable key of the form "Label|tt(Tooltip line)|option|option|...". Each tt(...) segment (there
// may be more than one) supplies one line of the tooltip; multiple lines are joined with '\n'. AllowEmptyToken and
// FreeFormToken toggle UI behavior instead of being literal choices. Segments are pipe-delimited; a literal '|'
// within a segment is written as '\|', and a literal newline as '\n'. No other character is reserved within the
// label, tooltip, or option text. Recognition of the special segments is purely positional (label is always
// first) or by the tt(...) wrapper.
type Marker struct {
	Raw        string
	Label      string
	Tooltip    string
	Options    []string
	AllowEmpty bool
	FreeForm   bool
}

// splitSegments splits raw on '|', honoring '\|' as an escaped, literal pipe rather than a delimiter, and '\n' as
// an escaped newline.
func splitSegments(raw string) []string {
	segments := make([]string, 0, strings.Count(raw, "|")+1)
	var buf strings.Builder
	for i := 0; i < len(raw); i++ {
		switch {
		case raw[i] == '\\' && i+1 < len(raw) && raw[i+1] == '|':
			buf.WriteByte('|')
			i++
		case raw[i] == '\\' && i+1 < len(raw) && raw[i+1] == 'n':
			buf.WriteByte('\n')
			i++
		case raw[i] == '|':
			segments = append(segments, buf.String())
			buf.Reset()
		default:
			buf.WriteByte(raw[i])
		}
	}
	segments = append(segments, strings.TrimSpace(buf.String()))
	return segments
}

// ParseSyntax attempts to parse raw's pipe-delimited syntax as a Marker. It returns an error (see ErrNoSyntax,
// ErrEmptyLabel, and ErrNoOptions) if raw doesn't use that syntax or if nothing usable survives parsing, in which
// case the returned Marker is the zero value. See ParseMarker for a version that always succeeds, falling back to
// heuristics for old-format markers that don't use this syntax at all.
func ParseSyntax(raw string) (Marker, error) {
	parts := splitSegments(raw)
	if len(parts) < 2 {
		return Marker{}, ErrNoSyntax
	}
	if parts[0] == "" {
		return Marker{}, ErrEmptyLabel
	}
	var marker Marker
	marker.Raw = raw
	marker.Label = parts[0]
	var tooltipLines []string
	for _, part := range parts[1:] {
		switch {
		case part == AllowEmptyToken:
			marker.AllowEmpty = true
		case part == FreeFormToken:
			marker.FreeForm = true
		case strings.HasPrefix(part, tooltipPrefix) && strings.HasSuffix(part, tooltipSuffix):
			tooltipLines = append(tooltipLines, strings.Split(part[len(tooltipPrefix):len(part)-len(tooltipSuffix)], "\n")...)
		default:
			marker.Options = append(marker.Options, part)
		}
	}
	for i, v := range tooltipLines {
		tooltipLines[i] = strings.TrimSpace(v)
	}
	marker.Tooltip = strings.Join(tooltipLines, "\n")
	if len(marker.Options) == 0 && !marker.FreeForm {
		return Marker{}, ErrNoOptions
	}
	return marker, nil
}

// DefaultValue returns the value that should be used for key when no explicit replacement has been recorded for it
// yet. For a key using the pipe-delimited syntax, this is the empty string when AllowEmpty is set (empty is itself
// a valid value there) or its first option otherwise. For any other key, this is the key itself, matching the
// long-standing behavior of showing the raw key text until the user provides a value.
func DefaultValue(key string) string {
	if marker, err := ParseSyntax(key); err == nil {
		if !marker.AllowEmpty && len(marker.Options) > 0 {
			return marker.Options[0]
		}
		return ""
	}
	return key
}
