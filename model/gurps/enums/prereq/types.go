// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package prereq

import "strings"

// TypesForEquipment holds the types that can be used for equipment.
var TypesForEquipment = []Type{
	Trait,
	Attribute,
	ContainedQuantity,
	ContainedWeight,
	EquippedEquipment,
	Skill,
	Spell,
	Script,
}

// TypesForNonEquipment holds the types that can be used for things other than equipment.
var TypesForNonEquipment = []Type{
	Trait,
	Attribute,
	EquippedEquipment,
	Skill,
	Spell,
	Script,
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
