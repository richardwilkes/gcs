// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package prereq

// TypesForEquipment holds the types that can be used for equipment.
var TypesForEquipment = []Type{
	Trait,
	Attribute,
	ContainedQuantity,
	ContainedWeight,
	EquippedEquipment,
	Skill,
	Spell,
}

// TypesForNonEquipment holds the types that can be used for things other than equipment.
var TypesForNonEquipment = []Type{
	Trait,
	Attribute,
	EquippedEquipment,
	Skill,
	Spell,
}
