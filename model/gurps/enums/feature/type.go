// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package feature

import "strings"

// SelectableTypes holds the possible Type values that may be chosen by the user. Unknown is excluded, since it exists
// only to hold onto feature data this version of GCS doesn't understand.
var SelectableTypes []Type

// SelectableTypesWithoutContainedWeightReduction holds the same values as SelectableTypes, minus the
// ContainedWeightReduction.
var SelectableTypesWithoutContainedWeightReduction []Type

func init() {
	SelectableTypes = make([]Type, 0, len(Types)-1)
	SelectableTypesWithoutContainedWeightReduction = make([]Type, 0, len(Types)-2)
	for _, one := range Types {
		if one == Unknown {
			continue
		}
		SelectableTypes = append(SelectableTypes, one)
		if one != ContainedWeightReduction {
			SelectableTypesWithoutContainedWeightReduction = append(SelectableTypesWithoutContainedWeightReduction, one)
		}
	}
}

// ExtractKnownType extracts the value from a string, reporting whether the string was actually recognized. Unlike
// ExtractType, which quietly maps anything it doesn't recognize onto the first value, this permits a caller that is
// dispatching on the type to detect data it has no knowledge of.
func ExtractKnownType(str string) (value Type, known bool) {
	for _, enum := range Types {
		if enum != Unknown && strings.EqualFold(enum.Key(), str) {
			return enum, true
		}
	}
	return Unknown, false
}
