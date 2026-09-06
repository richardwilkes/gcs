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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestBlockKeyForRow verifies that each of the row types that can be copied onto a sheet or template maps to the block
// that holds it there, and that every other type maps to nothing, since that is what stops the copy commands from
// being offered for lists such as weapons or modifiers.
func TestBlockKeyForRow(t *testing.T) {
	c := check.New(t)
	c.Equal(gurps.BlockTraitsKey, blockKeyForRow((*gurps.Trait)(nil)))
	c.Equal(gurps.BlockSkillsKey, blockKeyForRow((*gurps.Skill)(nil)))
	c.Equal(gurps.BlockSpellsKey, blockKeyForRow((*gurps.Spell)(nil)))
	c.Equal(gurps.BlockEquipmentKey, blockKeyForRow((*gurps.Equipment)(nil)), "equipment lands in the carried list")
	c.Equal(gurps.BlockNotesKey, blockKeyForRow((*gurps.Note)(nil)))
	c.Equal("", blockKeyForRow((*gurps.TraitModifier)(nil)), "modifiers can't be copied onto a sheet or template")
	c.Equal("", blockKeyForRow((*gurps.Weapon)(nil)), "weapons can't be copied onto a sheet or template")
	c.Equal("", blockKeyForRow(nil))
}

// TestCanCopySelectionTo verifies the conditions under which the copy-to-sheet and copy-to-template commands are
// offered: the table must have a selection, and there must be somewhere open to copy it to.
func TestCanCopySelectionTo(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	template := newTestTemplateDockable("Destination", gurps.NewTemplate())
	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Claws"
	source := newLibraryStyleTraitsTable(trait)

	c.False(canCopySelectionTo(source, []*Sheet{sheet}), "nothing can be copied while nothing is selected")
	c.False(canCopySelectionTo(source, []*Template{template}), "nothing can be copied while nothing is selected")

	source.SelectAll()
	c.False(canCopySelectionTo(source, []*Sheet(nil)), "there has to be a sheet to copy to")
	c.False(canCopySelectionTo(source, []*Template(nil)), "there has to be a template to copy to")
	c.True(canCopySelectionTo(source, []*Sheet{sheet}))
	c.True(canCopySelectionTo(source, []*Template{template}))
}

// TestCopySelectionToLandsRowsInTheListForTheirType verifies that copying a selection onto a sheet and onto a template
// adds the rows to the list that holds rows of that type on each of them, resolving the rows as a drop onto either
// would along the way, so that the one copy routine serves both kinds of destination.
func TestCopySelectionToLandsRowsInTheListForTheirType(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	originalTraits := len(entity.Traits) // A new entity may come with traits of its own, such as the natural attacks.
	templateData := gurps.NewTemplate()
	template := newTestTemplateDockable("Destination", templateData)
	c.Equal(0, len(templateData.TraitList()), "a new template must start out without traits")

	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Claws"
	trait.Modifiers = []*gurps.TraitModifier{newSwitchableTraitModifier("Retractable")}
	source := newLibraryStyleTraitsTable(trait)
	source.SelectAll()
	shown := stubTraitModifierPrompt(t, enableAllModifiers)

	copySelectionTo(source, []*Sheet{sheet})
	c.Equal(1, *shown, "a copy onto a sheet from a library must prompt for the copied row's modifiers")
	c.Equal(originalTraits+1, len(entity.Traits), "the trait must have been copied onto the sheet")
	copied := entity.Traits[len(entity.Traits)-1]
	c.Equal("Claws", copied.Name)
	c.NotEqual(trait.ID(), copied.ID(), "the sheet must receive a copy rather than the library's row")
	c.False(copied.Modifiers[0].Disabled, "the answer to the modifier prompt must have been kept")
	c.True(sheet.Traits.Table.CopySelectionMap()[copied.ID()],
		"the copied row must be selected in the sheet's traits list")

	copySelectionTo(source, []*Template{template})
	c.Equal(2, *shown, "a copy onto a template from a library must prompt for the copied row's modifiers as well")
	c.Equal(1, len(templateData.TraitList()), "the trait must have been copied onto the template")
	c.Equal("Claws", templateData.TraitList()[0].Name)
	c.NotEqual(trait.ID(), templateData.TraitList()[0].ID(),
		"the template must receive a copy rather than the library's row")
	c.Equal(1, len(template.Traits.Table.RootRows()), "the template's traits list must show the copied row")
	c.Equal(0, len(templateData.Skills), "the trait must not have landed in any other list")
}
