// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

// ListProvider defines the methods needed to access list data.
type ListProvider interface {
	TraitListProvider
	EquipmentListProvider
	NoteListProvider
	SkillListProvider
	SpellListProvider
}

// TraitListProvider defines the methods needed to access the trait list data.
type TraitListProvider interface {
	DataOwnerProvider
	TraitList() []*Trait
	SetTraitList(list []*Trait)
}

// EquipmentListProvider defines the methods needed to access the equipment list data.
type EquipmentListProvider interface {
	DataOwnerProvider
	CarriedEquipmentList() []*Equipment
	SetCarriedEquipmentList(list []*Equipment)
	OtherEquipmentList() []*Equipment
	SetOtherEquipmentList(list []*Equipment)
}

// NoteListProvider defines the methods needed to access the note list data.
type NoteListProvider interface {
	DataOwnerProvider
	NoteList() []*Note
	SetNoteList(list []*Note)
}

// SkillListProvider defines the methods needed to access the skill list data.
type SkillListProvider interface {
	DataOwnerProvider
	SkillList() []*Skill
	SetSkillList(list []*Skill)
}

// SpellListProvider defines the methods needed to access the spell list data.
type SpellListProvider interface {
	DataOwnerProvider
	SpellList() []*Spell
	SetSpellList(list []*Spell)
}

// TraitModifierListProvider defines the methods needed to access the trait modifier list data.
type TraitModifierListProvider interface {
	DataOwnerProvider
	TraitModifierList() []*TraitModifier
	SetTraitModifierList(list []*TraitModifier)
}

// EquipmentModifierListProvider defines the methods needed to access the equipment modifier list data.
type EquipmentModifierListProvider interface {
	DataOwnerProvider
	EquipmentModifierList() []*EquipmentModifier
	SetEquipmentModifierList(list []*EquipmentModifier)
}

// ConditionalModifierListProvider defines the methods needed to access the conditional modifier list data.
type ConditionalModifierListProvider interface {
	DataOwnerProvider
	ConditionalModifiers() []*ConditionalModifier
}

// ReactionModifierListProvider defines the methods needed to access the reaction modifier list data.
type ReactionModifierListProvider interface {
	DataOwnerProvider
	Reactions() []*ConditionalModifier
}

// WeaponListProvider defines the methods needed to access the weapon list data.
type WeaponListProvider interface {
	DataOwnerProvider
	WeaponOwner() WeaponOwner
	Weapons(melee, excludeHidden bool) []*Weapon
	SetWeapons(melee bool, list []*Weapon)
}
