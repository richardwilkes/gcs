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
	"maps"
	"slices"
	"strings"
)

// Filler defines the method for filling the nameable key map.
type Filler interface {
	// FillWithNameableKeys fills the map with nameable keys.
	FillWithNameableKeys(m, existing map[string]string)
}

// Accesser defines the method for retrieving the nameable replacements.
type Accesser interface {
	// NameableReplacements returns the replacements to be used with Nameables.
	NameableReplacements() map[string]string
}

// Setter defines the method for setting the nameable replacements.
type Setter interface {
	SetNameableReplacements(replacements map[string]string)
}

// Applier defines methods types that want to participate the nameable adjustments should implement.
type Applier interface {
	Accesser
	Filler
	// ApplyNameableKeys applies the nameable keys to this object.
	ApplyNameableKeys(m map[string]string)
}

// Extract nameable markers from the provided strings
// Each extracted marker will be a key and it's value will come from existing or be the default for the marker.
func Extract(nameables, replacements map[string]string, in ...string) map[string]string {
	if nameables == nil {
		nameables = make(map[string]string)
	}
	for _, src := range in {
		for _, part := range ExtractParts(src, MarkerDelimiter, MarkerDelimiter) {
			if part.Placeholder {
				// Unescape any escaped marker delimiters and trim leading/trailing space
				raw := strings.TrimSpace(UnescapeRunes(part.Value, MarkerDelimiter))
				// Parse the marker text into a Marker
				if m, ok := NewMarker(raw); ok {
					// We may end up with duplicates and that's fine, they will collapse to one entry
					if v, exists := replacements[m.Key()]; exists {
						nameables[m.Key()] = v
					} else {
						nameables[m.Key()] = m.DefaultValue()
					}
				}
			}
		}
	}
	return nameables
}

// Normalize returns a replacements map with each key rewritten to its normalized marker form.
//
// Keys that fail to parse as markers are retained unchanged.
// Map is normalized in sorted source key order.
func Normalize(replacements map[string]string) map[string]string {
	if len(replacements) == 0 {
		return nil
	}
	out := make(map[string]string, len(replacements))
	for _, k := range slices.Sorted(maps.Keys(replacements)) {
		if m, ok := NewMarker(k); ok {
			out[m.Key()] = replacements[k]
		} else {
			out[k] = replacements[k]
		}
	}
	return out
}

// Apply replaces nameable markers with their replacement values in a single string.
//
// Any unresolved markers are rendered in their compact form, "`@Label@`".
// The '@' wrapper is kept on unresolved markers so displayed text (sheet rows, table columns, tooltips) still
// visibly flags the value as an unresolved nameable marker.
func Apply(str string, replacements map[string]string) string {
	if !strings.ContainsRune(str, MarkerDelimiter) {
		return str
	}
	return ApplyToList([]string{str}, replacements)[0]
}

// ApplyToList replaces nameable markers with their replacement values in a slice of strings.
//
// Any unresolved markers are rendered in their compact form, "`@Label@`".
// The '@' wrapper is kept on unresolved markers so displayed text (sheet rows, table columns, tooltips) still
// visibly flags the value as an unresolved nameable marker.
func ApplyToList(in []string, replacements map[string]string) []string {
	if len(in) == 0 {
		return nil
	}

	out := make([]string, len(in))

	for i, str := range in {
		if !strings.ContainsRune(str, MarkerDelimiter) {
			// Skip processing a string with no markers
			out[i] = str
			continue
		}

		// Allocate a string builder to collect the processed string as we go
		var sb strings.Builder

		for _, part := range ExtractParts(str, MarkerDelimiter, MarkerDelimiter) {
			if part.Placeholder {
				// Unescape any escaped marker delimiters and trim leading/trailing space
				raw := strings.TrimSpace(UnescapeRunes(part.Value, MarkerDelimiter))

				// Parse the marker text into a Marker
				if m, ok := NewMarker(raw); ok {
					if r, hasReplacement := replacements[m.Key()]; hasReplacement {
						sb.WriteString(r)
					} else {
						sb.WriteRune(MarkerDelimiter)
						sb.WriteString(m.Label) // We do not escape markers here by design
						sb.WriteRune(MarkerDelimiter)
					}
				} else {
					// Since marker parsing failed we write the exact marker back out, with delimiters
					sb.WriteRune(MarkerDelimiter)
					sb.WriteString(part.Value)
					sb.WriteRune(MarkerDelimiter)
				}
			} else {
				// Unescape any escaped marker delimiters so a literal "\@" the user typed to keep an '@' out of
				// marker detection displays as a plain '@' rather than surfacing the escape itself.
				sb.WriteString(UnescapeRunes(part.Value, MarkerDelimiter))
			}
		}
		out[i] = sb.String()
	}

	return out
}

// Reduce returns a map of the replacements which exist in nameables.
//
// Both maps must already be keyed by normalized marker key: nameables from Extract, and replacements from
// load-time normalization (see Normalize) or the output of a previous Reduce call. Reduce does not itself
// normalize either map's keys.
//
// Returns nil, not an empty map, when there is nothing to keep -- callers that assign the result directly to a
// stored Replacements field and later write into it without a nil check (e.g. `x.Replacements[k] = v`) must guard
// for nil first.
func Reduce(nameables, replacements map[string]string) map[string]string {
	// We can return early if there are no known namables or no known replacements
	if len(nameables) == 0 || len(replacements) == 0 {
		return nil
	}

	ret := make(map[string]string, min(len(nameables), len(replacements)))
	for k, v := range replacements {
		if _, found := nameables[k]; found {
			ret[k] = v
		}
	}

	return ret
}

// Missing returns a list of nameable keys without replacements
//
// Both maps must already be keyed by normalized marker key: nameables from Extract, and replacements from
// load-time normalization (see Normalize) or the output of a previous Reduce call. Missing does not itself
// normalize either map's keys. The returned keys are normalized, since they are drawn from nameables.
func Missing(nameables, replacements map[string]string) []string {
	// We can return early if there are no known namables
	if len(nameables) == 0 {
		return nil
	}

	// If there are no replacements, everything is missing
	if len(replacements) == 0 {
		return slices.Collect(maps.Keys(nameables))
	}

	missing := make([]string, 0, min(len(nameables), len(replacements)))
	for k := range nameables {
		if _, exists := replacements[k]; !exists {
			missing = append(missing, k)
		}
	}

	return missing
}
