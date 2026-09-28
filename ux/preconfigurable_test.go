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

// TestPreconfiguredCheckBoxHiddenOnSheets verifies that the Preconfigured check box is offered wherever an item's
// modifiers and nameables are still to be settled, a template or a library, and nowhere on a sheet, where they already
// have been.
func TestPreconfiguredCheckBoxHiddenOnSheets(t *testing.T) {
	c := check.New(t)
	shown := func(owner Rebuildable) int {
		_, content := buildEditorContent(owner, gurps.NewSkill(nil, nil, false), initSkillEditor)
		return len(checkBoxesTitled(content, "Preconfigured"))
	}
	c.Equal(1, shown(newTestTemplateDockable("Template", gurps.NewTemplate())), "a template offers it")
	c.Equal(1, shown(NewSkillTableDockable("skills"+gurps.SkillsExt, nil)), "a library offers it")
	c.Equal(0, shown(newTestSheetForTemplate(t)), "a character sheet must not offer it")

	_, content := buildEditorContent(Rebuildable(newTestLootSheet(t)), gurps.NewEquipment(nil, nil, false),
		initEquipmentEditor(false))
	c.Equal(0, len(checkBoxesTitled(content, "Preconfigured")), "a loot sheet must not offer it")
}

// TestPreconfiguredFlagKeptOutsideSheets verifies that only a sheet clears the Preconfigured flag of rows arriving in
// it; a template and a library keep it, so that it is honored when their rows are copied onward.
func TestPreconfiguredFlagKeptOutsideSheets(t *testing.T) {
	c := check.New(t)
	newTrait := func() *gurps.Trait {
		trait := gurps.NewTrait(nil, nil, false)
		trait.Preconfigured = true
		return trait
	}
	libraryTrait := newTrait()
	library := newLibraryStyleTraitsTable(libraryTrait)
	c.False(clearPreconfiguredFlag(library, library.RootRows()), "a library must keep the flag")
	c.True(libraryTrait.Preconfigured)

	templateTrait := newTrait()
	template := newTestTemplateWithTraits(templateTrait)
	c.False(clearPreconfiguredFlag(template.Traits.Table, template.Traits.Table.RootRows()),
		"a template must keep the flag")
	c.True(templateTrait.Preconfigured)

	sheetTrait := newTrait()
	sheet := newTestSheetForTemplate(t)
	sheet.Entity().Traits = []*gurps.Trait{sheetTrait}
	sheet.Rebuild(true)
	c.True(clearPreconfiguredFlag(sheet.Traits.Table, sheet.Traits.Table.RootRows()), "a sheet must clear the flag")
	c.False(sheetTrait.Preconfigured)
}

// TestCopyFromLibraryToTemplateHonorsPreconfigured verifies that a row marked preconfigured in a library isn't prompted
// for when it is copied into a template, and keeps its mark there, while an unmarked row is prompted for as usual.
func TestCopyFromLibraryToTemplateHonorsPreconfigured(t *testing.T) {
	c := check.New(t)
	preconfigured := gurps.NewTrait(nil, nil, false)
	preconfigured.Name = "Settled"
	preconfigured.Preconfigured = true
	preconfigured.Modifiers = []*gurps.TraitModifier{newSwitchableTraitModifier("Retractable")}
	plain := gurps.NewTrait(nil, nil, false)
	plain.Name = "Unsettled"
	plain.Modifiers = []*gurps.TraitModifier{newSwitchableTraitModifier("Retractable")}
	library := NewTraitTableDockable("traits"+gurps.TraitsExt, []*gurps.Trait{preconfigured, plain})
	library.table.SelectAll()
	destinationData := gurps.NewTemplate()
	destination := newTestTemplateDockable("Destination", destinationData)
	prompts := captureModifierPrompts(t)

	copySelectionTo(library.table, []*Template{destination})

	c.Equal(1, len(*prompts), "only the row that isn't preconfigured must be prompted for")
	c.Equal("Unsettled", (*prompts)[0].title)
	c.Equal(2, len(destinationData.Traits))
	c.True(destinationData.Traits[0].Preconfigured, "the preconfigured row must keep its mark in the template")
	c.False(destinationData.Traits[1].Preconfigured)
}
