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

// Delimiter runes used in marker and segment extraction
const (
	MarkerDelimiter  rune = '@'
	SegmentDelimiter rune = '|'
)

// Tokens used for marker flags and tooltips
const (
	AllowEmptyToken    string = "?"
	FreeFormToken      string = "*"
	TooltipPrefixToken string = "tt("
	TooltipSuffixToken string = ")"
)

var (
	// legacyExampleListPattern matches an old-format marker shaped like "Label: item, item[, item...], etc." -- a
	// short leading label, then two or more comma-separated example items, ending in a trailing "etc."/"etc" that
	// signals the list isn't exhaustive. The item list itself is captured without the trailing ", etc." part.
	legacyExampleListPattern = regexp.MustCompile(`^([^:\n]{1,40}): ((?:[^,\n]+, )+[^,\n]+)(, ?etc\.?)$`)

	// legacyLabeledPattern matches an old-format marker shaped like "Label: anything" -- just a short leading label
	// before the first colon, with no other requirement placed on what follows.
	legacyLabeledPattern = regexp.MustCompile(`^([^:\n]{1,40}): (.+)$`)
)

// NewMarker parses a single nameable marker string into a Marker struct
func NewMarker(raw string) (Marker, bool) {
	segments := ExtractSegments(raw, SegmentDelimiter)
	if len(segments) == 0 {
		return Marker{Raw: raw}, false
	}

	cleanSegment := func(in string) string { return UnescapeRunes(in, EscapeRune, SegmentDelimiter) }

	label := cleanSegment(strings.TrimSpace(segments[0]))
	if label == "" {
		return Marker{Raw: raw}, false
	}

	if len(segments) == 1 {
		if m := legacyExampleListPattern.FindStringSubmatch(label); m != nil {
			var options []string
			for item := range strings.SplitSeq(m[2], ", ") {
				if item = strings.TrimSpace(item); item != "" {
					options = append(options, item)
				}
			}
			if len(options) >= 2 {
				return Marker{
					Raw:        raw,
					Label:      strings.TrimSpace(m[1]),
					Options:    options,
					FreeForm:   true,
					AllowEmpty: true,
					Tooltip:    strings.TrimSpace(m[2] + m[3]),
					Legacy:     true,
				}, true
			}
		}
		if m := legacyLabeledPattern.FindStringSubmatch(label); m != nil {
			return Marker{
				Raw:        raw,
				Label:      strings.TrimSpace(m[1]),
				FreeForm:   true,
				AllowEmpty: true,
				Tooltip:    strings.TrimSpace(m[2]),
				Legacy:     true,
			}, true
		}

		// The marker is a single segment and not known legacy format
		// In this case it's treated as having the FreeForm and AllowEmpty flags
		// This is an allowed and simplified form of a marker
		return Marker{
			Raw:        raw,
			Label:      label,
			FreeForm:   true,
			AllowEmpty: true,
		}, true
	}

	var options []string
	var freeForm bool
	var allowEmpty bool
	var tooltips []string

	for _, rawSegment := range segments[1:] {
		s := strings.TrimSpace(rawSegment)
		if strings.HasPrefix(s, TooltipPrefixToken) && strings.HasSuffix(s, TooltipSuffixToken) {
			tooltipSegment := s[len(TooltipPrefixToken) : len(s)-len(TooltipSuffixToken)]
			tooltipSegment = strings.TrimSpace(UnescapeRunes(tooltipSegment, EscapeRune, SegmentDelimiter, '\n'))
			tooltips = append(tooltips, strings.Split(tooltipSegment, "\n")...)
			continue
		}
		s = cleanSegment(s)
		if s == "" {
			continue
		}
		if s == FreeFormToken {
			freeForm = true
			continue
		}
		if s == AllowEmptyToken {
			allowEmpty = true
			continue
		}
		options = append(options, s)
	}

	// A marker with no literal options is treated as free-form, even if the FreeFormToken wasn't given.
	// This is a reasonable fallback to prevent an invalid marker config
	if len(options) == 0 {
		freeForm = true
	}

	return Marker{
		Raw:        raw,
		Label:      label,
		Options:    options,
		FreeForm:   freeForm,
		AllowEmpty: allowEmpty,
		Tooltip:    strings.Join(tooltips, "\n"),
	}, true
}

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
	Options    []string
	FreeForm   bool
	AllowEmpty bool
	Tooltip    string
	Legacy     bool
}

// Key returns a normalized marker key
func (m *Marker) Key() string {
	if m.Legacy {
		return m.Raw
	}
	var sb strings.Builder
	sb.Grow(len(m.Raw))
	sb.WriteString(EscapeRunes(m.Label, EscapeRune, SegmentDelimiter))
	if len(m.Options) == 0 && m.FreeForm && m.AllowEmpty && m.Tooltip == "" {
		// Handle an early return from simple markers to simplify the key
		return sb.String()
	}
	for _, o := range m.Options {
		sb.WriteRune(SegmentDelimiter)
		sb.WriteString(EscapeRunes(o, EscapeRune, SegmentDelimiter))
	}
	if m.FreeForm {
		sb.WriteRune(SegmentDelimiter)
		sb.WriteString(EscapeRunes(FreeFormToken, EscapeRune, SegmentDelimiter))
	}
	if m.AllowEmpty {
		sb.WriteRune(SegmentDelimiter)
		sb.WriteString(EscapeRunes(AllowEmptyToken, EscapeRune, SegmentDelimiter))
	}
	if m.Tooltip != "" {
		sb.WriteRune(SegmentDelimiter)
		sb.WriteString(TooltipPrefixToken)
		sb.WriteString(EscapeRunes(m.Tooltip, EscapeRune, SegmentDelimiter, '\n'))
		sb.WriteString(TooltipSuffixToken)
	}
	return sb.String()
}

// DefaultValue return the default value for this marker when no replacement is provided
func (m *Marker) DefaultValue() string {
	if !m.AllowEmpty && len(m.Options) > 0 {
		return m.Options[0]
	}
	if m.Legacy || (len(m.Options) == 0 && m.FreeForm && m.AllowEmpty && m.Tooltip == "") {
		// Legacy tiers 1/2 default to their own raw text for backward compatibility with sheets saved before this
		// feature existed. A simple current-form marker (see Key()'s short-circuit) isn't Legacy, but has nothing
		// sensible to offer either, so it gets the same treatment.
		return m.Raw
	}
	return ""
}

// ExtractMarkers extracts and parses markers from input strings
func ExtractMarkers(in ...string) map[string]Marker {
	markers := map[string]Marker{}

	for _, src := range in {
		for _, part := range ExtractParts(src, MarkerDelimiter, MarkerDelimiter) {
			if part.Placeholder {
				// Unescape any escaped marker delimiters and trim leading/trailing space
				raw := strings.TrimSpace(UnescapeRunes(part.Value, MarkerDelimiter))

				// Parse the marker text into a Marker
				if m, ok := NewMarker(raw); ok {
					// We may end up with duplicates and that's fine, they will collapse to one entry
					markers[m.Key()] = m
				}
			}
		}
	}

	return markers
}
