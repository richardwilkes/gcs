// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package nameable

import "strings"

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

// Extract the nameable sections of the strings into the target.
func Extract(target, existing map[string]string, in ...string) {
	for _, str := range in {
		count := strings.Count(str, "@")
		if count > 1 {
			parts := strings.Split(str, "@")
			for i, one := range parts {
				if i%2 == 1 && i < count {
					if value, ok := existing[one]; ok {
						target[one] = value
					} else {
						target[one] = DefaultValue(one)
					}
				}
			}
		}
	}
}

// Apply replaces the matching nameable sections with the values from the set. Any section left unresolved -- i.e.
// it has no entry in m at all -- is rendered in its compact form, "@Label@", rather than the full raw
// "@Label|...@" markup: the '@' wrapper is kept so displayed text (sheet rows, table columns, tooltips) still
// visibly flags the value as an unresolved nameable, while collapsing away the option list, tooltip, and flags that
// are only meaningful while editing. The label used here comes from ParseMarker, so an old-format marker that
// matches its fallback cascade (see ParseMarker) gets the same compact treatment as a real pipe-delimited key,
// rather than dumping its full raw text into the display.
func Apply(str string, m map[string]string) string {
	count := strings.Count(str, "@")
	if count > 1 {
		for k, v := range m {
			str = strings.ReplaceAll(str, "@"+k+"@", v)
		}
		count = strings.Count(str, "@")
		if count > 1 {
			parts := strings.Split(str, "@")
			var b strings.Builder
			for i, one := range parts {
				if i%2 == 1 && i < count {
					b.WriteByte('@')
					b.WriteString(ParseMarker(one).Label)
					b.WriteByte('@')
					continue
				}
				b.WriteString(one)
			}
			str = b.String()
		}
	}
	return str
}

// ApplyToList replaces the matching nameable sections with the values from the set.
func ApplyToList(in []string, m map[string]string) []string {
	if len(in) == 0 {
		return nil
	}
	list := make([]string, len(in))
	for i := range list {
		list[i] = Apply(in[i], m)
	}
	return list
}

// Reduce returns a map of the replacements which exist in needed. "Not set" and "empty" are not conflated here: an
// entry present in replacements with an empty value is a deliberate, resolved choice (only legal when the marker's
// AllowEmpty is set -- see ParseSyntax) and is kept, exactly like any other value. "Not set" is represented purely
// by a key's absence from replacements in the first place, so it never reaches this function as an entry to
// consider -- there is nothing for Reduce to omit.
func Reduce(needed, replacements map[string]string) map[string]string {
	ret := make(map[string]string)
	for k, v := range replacements {
		if _, ok := needed[k]; ok {
			ret[k] = v
		}
	}
	return ret
}
