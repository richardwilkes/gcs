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

// Applier defines methods types that want to participate the nameable adjustments should implement.
type Applier interface {
	Accesser
	Filler
	// ApplyNameableKeys applies the nameable keys to this object.
	ApplyNameableKeys(m map[string]string)
}

// Extract nameable markers from the provided strings
// Each extracted marker will be a key and it's value will come from existing or be the default for the marker.
func Extract(target, existing map[string]string, in ...string) map[string]string {
	if target == nil {
		target = make(map[string]string)
	}
	markers := ExtractMarkers(in...)
	for _, marker := range markers {
		if v, ok := existing[marker.Raw]; ok {
			target[marker.Raw] = v
		} else {
			target[marker.Raw] = marker.DefaultValue()
		}
	}
	return target
}

// Apply replaces nameable markers with their replacement values in a single string.
//
// Any unresolved markers are rendered in their compact form, "`@Label@`".
// The '@' wrapper is kept on unresolved markers so displayed text (sheet rows, table columns, tooltips) still
// visibly flags the value as an unresolved nameable marker.
func Apply(str string, replacements map[string]string) string {
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

	markers := ExtractMarkers(in...)

	delim := string(MarkerDelimiter)
	out := make([]string, len(in))
	for i, str := range in {
		if strings.ContainsRune(str, MarkerDelimiter) {
			for _, marker := range markers {
				if v, ok := replacements[marker.Raw]; ok {
					str = strings.ReplaceAll(str, delim+marker.Raw+delim, v)
				} else {
					str = strings.ReplaceAll(str, delim+marker.Raw+delim, delim+marker.Label+delim)
				}
			}
		}
		out[i] = str
	}
	return out
}

// Reduce returns a map of the replacements which exist in nameables. "Not set" and "empty" are not conflated here: an
// entry present in replacements with an empty value is a deliberate, resolved choice (only legal when the marker's
// AllowEmpty is set -- see ParseMarker) and is kept, exactly like any other value. "Not set" is represented purely
// by a key's absence from replacements in the first place, so it never reaches this function as an entry to
// consider -- there is nothing for Reduce to omit.
func Reduce(nameables, replacements map[string]string) map[string]string {
	ret := make(map[string]string)
	for k, v := range replacements {
		if _, ok := nameables[k]; ok {
			ret[k] = v
		}
	}
	return ret
}
