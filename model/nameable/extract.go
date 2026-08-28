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
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/richardwilkes/toolbox/v2/xmaps"
)

// EscapeRune is the fixed rune used to escape delimiters throughout this package.
const EscapeRune rune = '\\'

// List of reserved runes used to escape things like control characters
var reservedEscapes = map[rune]rune{'\n': 'n', '\t': 't'}

// reservedUnescapes is the reverse of reservedEscapes, mapping each reserved letter back to the rune it represents.
var reservedUnescapes = xmaps.Flip(reservedEscapes)

// checkReservedCollisions panics if chars requests escaping of both a reserved rune (e.g. '\n') and the plain letter
// used to represent it (e.g. 'n'), since that combination makes the two indistinguishable when unescaping.
func checkReservedCollisions(chars []rune) {
	for r, letter := range reservedEscapes {
		if slices.Contains(chars, r) && slices.Contains(chars, letter) {
			panic(fmt.Sprintf("cannot use both %q and its reserved letter %q in the same call", r, letter))
		}
	}
}

// Part is a segment of a string extracted by ExtractParts, either literal text or a placeholder.
type Part struct {
	Value       string
	Placeholder bool
}

// ExtractParts splits src into literal and placeholder parts, where a placeholder is text delimited by
// begin and end runes.
func ExtractParts(src string, begin, end rune) []Part {
	var parts []Part
	if begin == EscapeRune {
		panic("cannot use the escape rune as begin delimiter")
	}
	if end == EscapeRune {
		panic("cannot use the escape rune as end delimiter")
	}
	ol := utf8.RuneLen(begin)
	cl := utf8.RuneLen(end)
	start := 0
	isEscaping := false
	isPlaceholder := false
	for i, r := range src {
		// An escape rune toggles the escaping status. This means pairs of escape runes are escaped.
		if r == EscapeRune {
			isEscaping = !isEscaping
			continue
		}

		if isEscaping {
			isEscaping = false
			if r == begin || r == end {
				// The begin/end rune was escaped, so we advance past it.
				continue
			}
		}

		if r == begin && !isPlaceholder {
			// Collect previous non-empty part
			if start < i {
				parts = append(parts, Part{Value: src[start:i], Placeholder: false})
			}

			// Start new placeholder part
			isPlaceholder = true

			// Set the new start index after this begin delimiter
			start = i + ol

			// Advance to the next rune
			continue
		}

		if r == end && isPlaceholder {
			if start == i {
				// This is an empty placeholder and we treat this as a portion of a non-placeholder part

				// Clear the placeholder flag
				isPlaceholder = false

				// Move the start index backwards by the length of the begin delimiter
				start -= ol
			} else {
				// Collect the placeholder part
				parts = append(parts, Part{Value: src[start:i], Placeholder: true})

				// Start new non-placeholder part
				isPlaceholder = false

				// Set the new start index after this end delimiter
				start = i + cl
			}

			// Advance to the next rune
			continue
		}

		// A begin delimiter inside a placeholder is a special case.
		// It's treated like a misidentified placeholder start.
		if r == begin && isPlaceholder {
			// Move the start index backwards by the length of the begin delimiter
			start -= ol

			// Collect the non-placeholder part
			parts = append(parts, Part{Value: src[start:i], Placeholder: false})

			// Set the new start index after this begin delimiter
			start = i + ol

			// Advance to the next rune
			continue
		}

		// Any control byte breaks a placeholder part stride and any escaping
		if r < 32 {
			isEscaping = false
			isPlaceholder = false
		}
	}
	if isPlaceholder {
		// The final placeholder was never closed -- move the start index backwards by the length of the begin
		// delimiter so it's included in the trailing literal text instead of being silently dropped, even when
		// the begin delimiter is the last thing in the input.
		start -= ol
	}
	if start < len(src) {
		parts = append(parts, Part{Value: src[start:], Placeholder: false})
	}

	return parts
}

// ExtractSegments splits src on delimiter, honoring EscapeRune escapes of the delimiter.
func ExtractSegments(src string, delimiter rune) []string {
	if delimiter == EscapeRune {
		panic("cannot use the escape rune as delimiter")
	}
	dl := utf8.RuneLen(delimiter)

	var segments []string
	start := 0
	isEscaping := false
	for i, r := range src {
		// An escape rune toggles the escaping status. This means pairs of escape runes are escaped.
		if r == EscapeRune {
			isEscaping = !isEscaping
			continue
		}

		if isEscaping {
			isEscaping = false
			if r == delimiter {
				// The delimiter rune was escaped
				continue
			}
		}

		if r == delimiter {
			segments = append(segments, src[start:i])
			start = i + dl
		}
	}

	// Capture remaining runes (or all if there are no delimiters)
	if start < len(src) {
		segments = append(segments, src[start:])
	}

	return segments
}

// EscapeRunes prefixes each occurrence of any of chars in in with EscapeRune. A rune with a reserved letter
// (see reservedEscapes, e.g. a literal newline) is written as EscapeRune followed by its reserved letter
// instead of the literal rune itself.
func EscapeRunes(in string, chars ...rune) string {
	if len(chars) == 0 {
		return in
	}
	checkReservedCollisions(chars)

	var sb strings.Builder
	sb.Grow(len(in))
	for _, r := range in {
		if slices.Contains(chars, r) {
			sb.WriteRune(EscapeRune)
			if letter, ok := reservedEscapes[r]; ok {
				sb.WriteRune(letter)
				continue
			}
		}
		sb.WriteRune(r)
	}

	return sb.String()
}

// UnescapeRunes reverses EscapeRunes, removing EscapeRune prefixes before any of chars and translating a reserved
// letter (see reservedEscapes) back into the rune it represents.
func UnescapeRunes(in string, chars ...rune) string {
	if len(chars) == 0 {
		return in
	}
	checkReservedCollisions(chars)

	var sb strings.Builder
	sb.Grow(len(in))
	isEscaping := false
	for _, r := range in {
		if !isEscaping && r == EscapeRune {
			isEscaping = true
			continue
		}
		if isEscaping {
			if reservedRune, ok := reservedUnescapes[r]; ok && slices.Contains(chars, reservedRune) {
				isEscaping = false
				sb.WriteRune(reservedRune)
				continue
			}
			if !slices.Contains(chars, r) {
				sb.WriteRune(EscapeRune)
			}
		}
		isEscaping = false
		sb.WriteRune(r)
	}
	if isEscaping {
		sb.WriteRune(EscapeRune)
	}

	return sb.String()
}
