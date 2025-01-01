// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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

// Extract the nameable sections of the string into the set.
func Extract(str string, m, existing map[string]string) {
	count := strings.Count(str, "@")
	if count > 1 {
		parts := strings.Split(str, "@")
		for i, one := range parts {
			if i%2 == 1 && i < count {
				if value, ok := existing[one]; ok {
					m[one] = value
				} else {
					m[one] = one
				}
			}
		}
	}
}

// Apply replaces the matching nameable sections with the values from the set.
func Apply(str string, m map[string]string) string {
	if strings.Count(str, "@") > 1 {
		for k, v := range m {
			str = strings.ReplaceAll(str, "@"+k+"@", v)
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

// Reduce returns a map of the replacements which exist in needed.
func Reduce(needed, replacements map[string]string) map[string]string {
	ret := make(map[string]string)
	for k, v := range replacements {
		if _, ok := needed[k]; ok {
			ret[k] = v
		}
	}
	return ret
}
