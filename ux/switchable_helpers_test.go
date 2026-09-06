// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/gurps"
)

// switchableSTBonus returns an attribute bonus of +1 ST that only applies while its owner's switch is on. A nil owner
// leaves the bonus unowned, which suits features that belong to a modifier or that a test attaches to an item later.
func switchableSTBonus(owner fmt.Stringer) *gurps.AttributeBonus {
	bonus := gurps.NewAttributeBonus(gurps.StrengthID)
	bonus.SetSwitchable(true)
	if owner != nil {
		bonus.SetOwner(owner)
	}
	return bonus
}

// newSwitchableTrait returns a non-container trait carrying a single switchable +1 ST bonus, so that a sheet's traits
// list needs the switch column while the trait is in it.
func newSwitchableTrait(entity *gurps.Entity, name string) *gurps.Trait {
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = name
	trait.Features = gurps.Features{switchableSTBonus(trait)}
	return trait
}

// newSwitchableTraitModifier returns a disabled trait modifier carrying a switchable +1 ST bonus, so that enabling it
// is what gives its owner switchable features.
func newSwitchableTraitModifier(name string) *gurps.TraitModifier {
	modifier := gurps.NewTraitModifier(nil, nil, false)
	modifier.Name = name
	modifier.Disabled = true
	modifier.Features = gurps.Features{switchableSTBonus(nil)}
	return modifier
}

// newSwitchableSkill returns a non-container skill carrying a single switchable +1 ST bonus, so that a sheet's skills
// list needs the switch column while the skill is in it.
func newSwitchableSkill(entity *gurps.Entity, name string) *gurps.Skill {
	skill := gurps.NewSkill(entity, nil, false)
	skill.Name = name
	skill.Features = gurps.Features{switchableSTBonus(skill)}
	return skill
}

// newSwitchableSpell returns a non-container spell carrying a single switchable +1 ST bonus, so that a sheet's spells
// list needs the switch column while the spell is in it.
func newSwitchableSpell(entity *gurps.Entity, name string) *gurps.Spell {
	spell := gurps.NewSpell(entity, nil, false)
	spell.Name = name
	spell.Features = gurps.Features{switchableSTBonus(spell)}
	return spell
}

// newSwitchableEquipment returns a non-container piece of equipment carrying a single switchable +1 ST bonus, so that
// the sheet's list holding it needs the switch column while it is there.
func newSwitchableEquipment(entity *gurps.Entity, name string) *gurps.Equipment {
	eqp := gurps.NewEquipment(entity, nil, false)
	eqp.Name = name
	eqp.Features = gurps.Features{switchableSTBonus(eqp)}
	return eqp
}
