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
	"regexp"
	"strings"
)

// Special tokens within a Marker's pipe-delimited segments that toggle behavior or carry a tooltip line, rather than
// acting as a literal option.
const (
	EscapeChar         = '\\'
	MarkerDelimiter    = '@'
	SegmentDelimiter   = '|'
	AllowEmptyToken    = "?"
	FreeFormToken      = "*"
	TooltipPrefixToken = "tt("
	TooltipSuffixToken = ")"
)

var (
	// legacyExampleListPattern matches an old-format marker shaped like "Label: item, item[, item...], etc." -- a
	// short leading label, then two or more comma-separated example items, ending in a trailing "etc."/"etc" that
	// signals the list isn't exhaustive. The item list itself is captured without the trailing ", etc." part.
	legacyExampleListPattern = regexp.MustCompile(`^([^:\n]{1,40}): ((?:[^,\n]+, )+[^,\n]+), ?etc\.?$`)

	// legacyLabeledPattern matches an old-format marker shaped like "Label: anything" -- just a short leading label
	// before the first colon, with no other requirement placed on what follows.
	legacyLabeledPattern = regexp.MustCompile(`^([^:\n]{1,40}): (.+)$`)
)

// Marker is a parsed nameable key of the form "Label|tt(Tooltip line)|option|option|...".
//
// Segments are pipe-delimited.
// The first segment MUST be the label and MUST NOT be empty.
// Each tt(...) segment (there may be more than one) supplies one line of the tooltip.
// Multiple tooltip segments are joined with '\n' into a final single tooltip string.
// AllowEmptyToken and FreeFormToken toggle UI behavior instead of being literal choices.
// A literal `|` or `\` can be escaped by prefixing with `\`. No other character is reserved within the
// label, tooltip, or option text.
type Marker struct {
	Raw        string
	Label      string
	Tooltip    string
	Options    []string
	AllowEmpty bool
	FreeForm   bool
	Legacy     bool
}

// DefaultValue return the default value for this marker when no replacement is provided
func (m *Marker) DefaultValue() string {
	if !m.AllowEmpty && len(m.Options) > 0 {
		return m.Options[0]
	}
	if m.Legacy {
		return m.Raw
	}
	return ""
}

// ParseMarker parses a single nameable marker string into a Marker struct
func ParseMarker(raw string) (Marker, bool) {
	segments := splitSegments(unescapeChars(raw, MarkerDelimiter))

	label := segments[0]
	if label == "" {
		return Marker{Raw: raw}, false
	}

	if len(segments) == 1 {
		if m := legacyExampleListPattern.FindStringSubmatch(label); m != nil {
			tooltip := strings.TrimPrefix(label, m[1]+": ")
			var options []string
			for item := range strings.SplitSeq(m[2], ", ") {
				if item = strings.TrimSpace(item); item != "" {
					options = append(options, item)
				}
			}
			if len(options) >= 2 {
				return Marker{Raw: raw, Label: m[1], Tooltip: tooltip, Options: options, AllowEmpty: true, FreeForm: true, Legacy: true}, true
			}
		}
		if m := legacyLabeledPattern.FindStringSubmatch(label); m != nil {
			return Marker{Raw: raw, Label: m[1], Tooltip: m[2], AllowEmpty: true, FreeForm: true, Legacy: true}, true
		}

		return Marker{Raw: raw, Label: label, AllowEmpty: true, FreeForm: true, Legacy: true}, true
	}

	var tooltips []string
	var options []string
	var allowEmpty bool
	var freeForm bool

	for i := 1; i < len(segments); i++ {
		if segments[i] == "" {
			continue
		}
		if strings.HasPrefix(segments[i], TooltipPrefixToken) && strings.HasSuffix(segments[i], TooltipSuffixToken) {
			tooltips = append(tooltips, segments[i][len(TooltipPrefixToken):len(segments[i])-len(TooltipSuffixToken)])
			continue
		}
		if segments[i] == AllowEmptyToken {
			allowEmpty = true
			continue
		}
		if segments[i] == FreeFormToken {
			freeForm = true
			continue
		}
		options = append(options, segments[i])
	}

	// A marker with no literal options is treated as free-form, even if the FreeFormToken wasn't given -- there's
	// nothing else it could mean.
	if len(options) == 0 {
		freeForm = true
	}

	return Marker{Raw: raw, Label: label, Tooltip: strings.Join(tooltips, "\n"), Options: options, AllowEmpty: allowEmpty, FreeForm: freeForm}, true
}

// ExtractMarkers extracts and parses markers from input strings
func ExtractMarkers(in ...string) map[string]Marker {
	markers := map[string]Marker{}

	for _, src := range in {
		if src == "" {
			continue
		}
		start := -1
		escape := false
		for i := 0; i < len(src); i++ {
			if escape {
				escape = false
				// If the previous character was an unescaped eascape character, skip an escape or marker delimiter
				if src[i] == EscapeChar || src[i] == MarkerDelimiter {
					continue
				}
			} else if src[i] == EscapeChar {
				escape = true
				continue
			}

			// A newline breaks any in-progress marker stride
			if src[i] == '\n' {
				start = -1
				escape = false
				continue
			}

			// Advance past any non-delimiter characters
			if src[i] != MarkerDelimiter {
				continue
			}

			// Check if this is the start of a marker
			if start < 0 {
				// Capture the start index of the marker
				start = i
				continue
			}

			// Grab a slice representing the full marker string without start/end delimiters
			raw := src[start+1 : i]

			// Reset the start index to indicate we are not inside a marker stride
			start = -1

			// Skip parsing if this marker is a duplicate
			if _, ok := markers[raw]; ok {
				continue
			}

			// Parse the marker text into a Marker
			if m, ok := ParseMarker(raw); ok {
				markers[m.Raw] = m
			}
		}
	}

	return markers
}

// splitSegments splits raw on '|', honoring '\|' as an escaped, literal pipe rather than a delimiter.
func splitSegments(raw string) []string {
	// Pre-allocate space using pipe count
	segments := make([]string, 0, strings.Count(raw, string(SegmentDelimiter))+1)
	start := 0
	escape := false
	cleanSegment := func(in string) string {
		return strings.TrimSpace(strings.ReplaceAll(
			unescapeChars(in, EscapeChar, SegmentDelimiter),
			string([]rune{EscapeChar, 'n'}),
			"\n",
		))
	}
	for i := 0; i < len(raw); i++ {
		if escape {
			escape = false
			// If the previous character was an unescaped eascape character, skip an escape or segment delimiter
			if raw[i] == EscapeChar || raw[i] == SegmentDelimiter {
				continue
			}
		} else if raw[i] == EscapeChar {
			escape = true
			continue
		}

		// If this character is a delimiter, capture the segment
		if raw[i] == SegmentDelimiter {
			// Collect the string between the previous delimiter and this one (excludes delimiters)
			segments = append(segments, cleanSegment(raw[start:i]))

			// Move the start marker past this delimiter
			start = i + 1
		}
	}

	// Collect the trailing segment after the last delimiter (or the whole string, if there was none)
	segments = append(segments, cleanSegment(raw[start:]))

	return segments
}

func unescapeChars(in string, chars ...rune) string {
	for _, c := range chars {
		in = strings.ReplaceAll(in, string([]rune{EscapeChar, c}), string(c))
	}
	return in
}
