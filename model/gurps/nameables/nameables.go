/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package nameables

import (
	"strings"
)

// Nameables defines methods types that want to participate the nameable adjustments should implement.
type Nameables interface {
	// FillWithNameableKeys fills the map with nameable keys.
	FillWithNameableKeys(m map[string]string)
	// ApplyNameableKeys applies the nameable keys to this object.
	ApplyNameableKeys(m map[string]string)
}

// Extract the nameable sections of the string into the set.
func Extract(str string, set map[string]string) {
	count := strings.Count(str, "@")
	if count > 1 {
		parts := strings.Split(str, "@")
		for i, one := range parts {
			if i%2 == 1 && i < count {
				set[one] = one
			}
		}
	}
}

// Apply replaces the matching nameable sections with the values from the set.
func Apply(str string, set map[string]string) string {
	for k, v := range set {
		str = strings.ReplaceAll(str, "@"+k+"@", v)
	}
	return str
}
