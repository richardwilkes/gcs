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

// Unset is a sentinel value used in the nameables map produced by Extract to indicate that a marker has no recorded
// replacement yet. It is bracketed with NUL bytes so it won't collide with a value a user typed or that came from file
// content. Callers that read Extract's output must treat it the same as the key being absent.
const Unset = "\x00unset\x00"

// Filler defines the method for filling the nameable key map.
type Filler interface {
	FillWithNameableKeys(m, existing map[string]string)
}

// Accesser defines the method for retrieving the nameable replacements.
type Accesser interface {
	NameableReplacements() map[string]string
}

// Setter defines the method for setting the nameable replacements.
type Setter interface {
	SetNameableReplacements(replacements map[string]string)
}

// Applier is implemented by types that participate in the nameable adjustments.
type Applier interface {
	Accesser
	Filler
	ApplyNameableKeys(m map[string]string)
}

// Extract adds a key for each nameable marker found in the provided strings to nameables, allocating that map if it is
// nil, and returns it. Each key's value comes from replacements, or is Unset when replacements has no entry for it.
func Extract(nameables, replacements map[string]string, in ...string) map[string]string {
	if nameables == nil {
		nameables = make(map[string]string)
	}
	for _, src := range in {
		for _, part := range ExtractParts(src, MarkerDelimiter, MarkerDelimiter) {
			if part.Placeholder {
				if m, ok := NewMarker(UnescapeRunes(part.Value, MarkerDelimiter)); ok {
					// Duplicate markers are fine, since they collapse to a single entry.
					if v, exists := replacements[m.Key()]; exists {
						nameables[m.Key()] = v
					} else {
						nameables[m.Key()] = Unset
					}
				}
			}
		}
	}
	return nameables
}

// Normalize returns a replacements map with each key rewritten to its normalized marker form, or nil if there is
// nothing to normalize. Keys that fail to parse as markers are retained unchanged. Source keys are processed in sorted
// order, so two source keys that normalize to the same form always resolve the same way.
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

// Apply replaces nameable markers with their replacement values in a single string. An unresolved marker is rendered
// in its compact form, "`@Label@`", keeping the '@' wrapper so displayed text (sheet rows, table columns, tooltips)
// still visibly flags it as unresolved.
func Apply(str string, replacements map[string]string) string {
	if !strings.ContainsRune(str, MarkerDelimiter) {
		return str
	}
	return ApplyToList([]string{str}, replacements)[0]
}

// ApplyToList is Apply for a slice of strings. Returns nil if the slice is empty.
func ApplyToList(in []string, replacements map[string]string) []string {
	if len(in) == 0 {
		return nil
	}

	out := make([]string, len(in))

	for i, str := range in {
		if !strings.ContainsRune(str, MarkerDelimiter) {
			out[i] = str
			continue
		}

		var sb strings.Builder

		for _, part := range ExtractParts(str, MarkerDelimiter, MarkerDelimiter) {
			if part.Placeholder {
				if m, ok := NewMarker(UnescapeRunes(part.Value, MarkerDelimiter)); ok {
					if r, hasReplacement := replacements[m.Key()]; hasReplacement {
						sb.WriteString(r)
					} else {
						sb.WriteRune(MarkerDelimiter)
						sb.WriteString(m.Label) // We do not escape markers here by design
						sb.WriteRune(MarkerDelimiter)
					}
				} else {
					// Marker parsing failed, so write the original text back out with its delimiters.
					sb.WriteRune(MarkerDelimiter)
					sb.WriteString(part.Value)
					sb.WriteRune(MarkerDelimiter)
				}
			} else {
				// Unescape any escaped marker delimiters, so a literal "\@" the user typed to keep an '@' out of marker
				// detection displays as a plain '@' rather than surfacing the escape itself.
				sb.WriteString(UnescapeRunes(part.Value, MarkerDelimiter))
			}
		}
		out[i] = sb.String()
	}

	return out
}

// Reduce returns a map of the replacements which exist in nameables.
//
// Both maps must already be keyed by normalized marker key: nameables from Extract, and replacements from load-time
// normalization (see Normalize) or the output of a previous Reduce call. Reduce does not itself normalize either map's
// keys.
//
// An entry still holding the Unset sentinel (i.e. the substitutions dialog was shown but the user never chose a value
// for that marker) is dropped, so an untouched marker is never persisted as if it had been resolved.
//
// Returns nil, not an empty map, when there is nothing to keep -- callers that assign the result directly to a stored
// Replacements field and later write into it (e.g. `x.Replacements[k] = v`) must guard for nil first.
func Reduce(nameables, replacements map[string]string) map[string]string {
	if len(nameables) == 0 || len(replacements) == 0 {
		return nil
	}

	ret := make(map[string]string, min(len(nameables), len(replacements)))
	for k, v := range replacements {
		if v == Unset {
			continue
		}
		if _, found := nameables[k]; found {
			ret[k] = v
		}
	}

	return ret
}

// Missing returns the nameable keys that have no replacement, in no particular order.
//
// Both maps must already be keyed by normalized marker key: nameables from Extract, and replacements from load-time
// normalization (see Normalize) or the output of a previous Reduce call. Missing does not itself normalize either
// map's keys. The returned keys are normalized, since they are drawn from nameables.
func Missing(nameables, replacements map[string]string) []string {
	if len(nameables) == 0 {
		return nil
	}

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
