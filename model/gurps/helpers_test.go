// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

// addTraitWithFeatures creates a non-container trait with the given name and features, appends it to the entity's
// traits, and returns it. Callers that need more (levels, replacements, modifiers, weapons) set those on the result;
// the entity is not recalculated here.
func addTraitWithFeatures(e *Entity, name string, features ...Feature) *Trait {
	trait := NewTrait(e, nil, false)
	trait.Name = name
	trait.Features = features
	e.Traits = append(e.Traits, trait)
	return trait
}

// addCarriedEquipmentWithFeatures creates a non-container piece of equipment with the given name and features,
// appends it to the entity's carried equipment, and returns it. Callers that need more (replacements, modifiers,
// weapons, the equipped state) set those on the result; the entity is not recalculated here.
func addCarriedEquipmentWithFeatures(e *Entity, name string, features ...Feature) *Equipment {
	eqp := NewEquipment(e, nil, false)
	eqp.Name = name
	eqp.Features = features
	e.CarriedEquipment = append(e.CarriedEquipment, eqp)
	return eqp
}
