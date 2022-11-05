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

package ux

import (
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// ContextMenuItem holds the title and ID of a context menu item that is derived from a command ID.
type ContextMenuItem struct {
	Title string
	ID    int
}

// DefaultContextMenuItems holds the default set of context menu items for lists.
var DefaultContextMenuItems = []ContextMenuItem{
	{"", -1},
	{i18n.Text("Open Detail Editor"), OpenEditorItemID},
	{"", -1},
	{i18n.Text("Duplicate"), DuplicateItemID},
	{i18n.Text("Delete"), unison.DeleteItemID},
	{"", -1},
	{i18n.Text("Copy to Character Sheet"), CopyToSheetItemID},
	{i18n.Text("Copy to Template"), CopyToTemplateItemID},
	{i18n.Text("Apply Template to Character Sheet"), ApplyTemplateItemID},
	{"", -1},
	{i18n.Text("Increment"), IncrementItemID},
	{i18n.Text("Decrement"), DecrementItemID},
	{i18n.Text("Increase Uses"), IncrementUsesItemID},
	{i18n.Text("Decrease Uses"), DecrementUsesItemID},
	{i18n.Text("Increase Skill Level"), IncrementSkillLevelItemID},
	{i18n.Text("Decrease Skill Level"), DecrementSkillLevelItemID},
	{i18n.Text("Increase Tech Level"), IncrementTechLevelItemID},
	{i18n.Text("Decrease Tech Level"), DecrementTechLevelItemID},
	{"", -1},
	{i18n.Text("Toggle State"), ToggleStateItemID},
	{i18n.Text("Swap Defaults"), SwapDefaultsItemID},
	{i18n.Text("Convert to Container"), ConvertToContainerItemID},
	{"", -1},
	{i18n.Text("Open Page Reference"), OpenOnePageReferenceItemID},
	{i18n.Text("Open Each Page Reference"), OpenEachPageReferenceItemID},
}

// CarriedEquipmentExtraContextMenuItems holds context menu items specific to the carried equipment list.
var CarriedEquipmentExtraContextMenuItems = []ContextMenuItem{
	{i18n.Text("New Carried Equipment"), NewCarriedEquipmentItemID},
	{i18n.Text("New Carried Equipment Container"), NewCarriedEquipmentContainerItemID},
}

// EquipmentModifierExtraContextMenuItems holds context menu items specific to the equipment modifier list.
var EquipmentModifierExtraContextMenuItems = []ContextMenuItem{
	{i18n.Text("New Equipment Modifier"), NewEquipmentModifierItemID},
	{i18n.Text("New Equipment Modifier Container"), NewEquipmentContainerModifierItemID},
}

// MeleeWeaponExtraContextMenuItems holds context menu items specific to the melee weapon list.
var MeleeWeaponExtraContextMenuItems = []ContextMenuItem{
	{i18n.Text("New Melee Weapon"), NewMeleeWeaponItemID},
}

// NoteExtraContextMenuItems holds context menu items specific to the note list.
var NoteExtraContextMenuItems = []ContextMenuItem{
	{i18n.Text("New Note"), NewNoteItemID},
	{i18n.Text("New Note Container"), NewNoteContainerItemID},
}

// OtherEquipmentExtraContextMenuItems holds context menu items specific to the other equipment list.
var OtherEquipmentExtraContextMenuItems = []ContextMenuItem{
	{i18n.Text("New Other Equipment"), NewOtherEquipmentItemID},
	{i18n.Text("New Other Equipment Container"), NewOtherEquipmentContainerItemID},
}

// RangedWeaponExtraContextMenuItems holds context menu items specific to the ranged weapon list.
var RangedWeaponExtraContextMenuItems = []ContextMenuItem{
	{i18n.Text("New Ranged Weapon"), NewRangedWeaponItemID},
}

// SkillExtraContextMenuItems holds context menu items specific to the skill list.
var SkillExtraContextMenuItems = []ContextMenuItem{
	{i18n.Text("New Skill"), NewSkillItemID},
	{i18n.Text("New Skill Container"), NewSkillContainerItemID},
	{i18n.Text("New Technique"), NewTechniqueItemID},
}

// SpellExtraContextMenuItems holds context menu items specific to the spell list.
var SpellExtraContextMenuItems = []ContextMenuItem{
	{i18n.Text("New Spell"), NewSpellItemID},
	{i18n.Text("New Spell Container"), NewSpellContainerItemID},
	{i18n.Text("New Ritual Magic Spell"), NewRitualMagicSpellItemID},
}

// TraitExtraContextMenuItems holds context menu items specific to the trait list.
var TraitExtraContextMenuItems = []ContextMenuItem{
	{i18n.Text("New Trait"), NewTraitItemID},
	{i18n.Text("New Trait Container"), NewTraitContainerItemID},
	{i18n.Text("Add Natural Attacks"), AddNaturalAttacksItemID},
}

// TraitModifierExtraContextMenuItems holds context menu items specific to the trait modifier list.
var TraitModifierExtraContextMenuItems = []ContextMenuItem{
	{i18n.Text("New Trait Modifier"), NewTraitModifierItemID},
	{i18n.Text("New Trait Modifier Container"), NewTraitContainerModifierItemID},
}
