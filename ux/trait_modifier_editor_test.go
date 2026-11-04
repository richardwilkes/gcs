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
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	uncheck "github.com/richardwilkes/unison/enums/check"
)

// TestModifierEditorFollowsTheChoiceRules verifies that enabling an option in its own editor makes it the pick,
// turning the one that was picked off, and that undo puts both back. On a sheet, the pick of a mandatory choice can't
// be turned off there either, though it can be elsewhere.
func TestModifierEditorFollowsTheChoiceRules(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	traitEditor, _, _ := traitEditorOnSheet(t, sheet, newTraitModifierChoiceFor(sheet.Entity(), true, []string{"A", "B"}, "A"))
	options := traitEditor.editorData.Modifiers[0].Children

	e, _ := buildEditorContent(traitEditor, options[1], initTraitModifierEditor)
	e.editorData.Disabled = false
	e.applyEdits()
	c.False(options[0].Enabled(), "enabling B in its editor turns A off")
	c.True(options[1].Enabled())
	mgr := unison.UndoManagerFor(traitEditor)
	mgr.Undo()
	c.True(options[0].Enabled(), "undo turns A back on")
	c.False(options[1].Enabled(), "and B back off")
	mgr.Redo()
	c.False(options[0].Enabled(), "redo turns A off again")
	c.True(options[1].Enabled(), "and B back on")
	mgr.Undo()

	e, _ = buildEditorContent(traitEditor, options[0], initTraitModifierEditor)
	e.editorData.Disabled = true
	e.applyEdits()
	c.True(options[0].Enabled(), "the pick of a mandatory choice on a sheet can't be turned off in its editor")

	offSheet := newTraitModifierChoiceFor(nil, true, []string{"A"}, "A")
	e, _ = buildEditorContent(nil, offSheet.Children[0], initTraitModifierEditor)
	e.editorData.Disabled = true
	e.applyEdits()
	c.False(offSheet.Children[0].Enabled(), "off a sheet a mandatory choice may be left without its pick")
}

// TestEditorKeepsAPickMadeSinceItOpened verifies that applying an option's editor, opened while it was the pick, after
// another option has been picked in the list leaves that later pick alone, since the editor didn't change whether its
// option is enabled.
func TestEditorKeepsAPickMadeSinceItOpened(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	traitEditor, _, table := traitEditorOnSheet(t, sheet, newTraitModifierChoiceFor(sheet.Entity(), true, []string{"A", "B"}, "A"))
	options := traitEditor.editorData.Modifiers[0].Children
	aEditor, _ := buildEditorContent(traitEditor, options[0], initTraitModifierEditor)
	adjustModifierEnabled(traitEditor, table, options[1], true)
	c.True(options[1].Enabled())
	aEditor.editorData.Name = "Renamed"
	aEditor.applyEdits()
	c.Equal("Renamed", options[0].Name)
	c.False(options[0].Enabled(), "the editor didn't change whether A is enabled, so A stays off")
	c.True(options[1].Enabled(), "and B stays the pick")
}

// TestChoiceMadeMandatoryInItsEditorOnASheet verifies that applying a choice's editor on a sheet without changing what
// the choice asks for leaves its options as they are, so a mandatory choice flagged for want of a pick isn't given one,
// while a choice made mandatory there gets its first option picked, as Convert to Choice does, undo and redo covering
// it.
func TestChoiceMadeMandatoryInItsEditorOnASheet(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	traitEditor, _, _ := traitEditorOnSheet(t, sheet, newTraitModifierChoiceFor(sheet.Entity(), true, []string{"A", "B"}))
	choice := traitEditor.editorData.Modifiers[0]
	choiceEditor, _ := buildEditorContent(traitEditor, choice, initTraitModifierEditor)
	choiceEditor.editorData.Name = "Renamed"
	choiceEditor.applyEdits()
	c.Equal("Renamed", choice.Name)
	c.False(choice.Children[0].Enabled(), "the edit doesn't pick the first option")
	c.False(choice.Children[1].Enabled())

	choice.SetMandatoryChoice(false)
	choiceEditor, _ = buildEditorContent(traitEditor, choice, initTraitModifierEditor)
	choiceEditor.editorData.SetMandatoryChoice(true)
	choiceEditor.applyEdits()
	c.True(gurps.IsMandatoryModifierChoice(choice))
	c.True(choice.Children[0].Enabled(), "the first option is picked")
	mgr := unison.UndoManagerFor(traitEditor)
	mgr.Undo()
	c.False(gurps.IsMandatoryModifierChoice(choice))
	c.False(choice.Children[0].Enabled(), "undo takes the pick back")
	mgr.Redo()
	c.True(choice.Children[0].Enabled(), "redo picks it again")
}

// TestLockedPickHasItsEnabledBoxDisabled verifies that the editor of the pick of a mandatory choice on a sheet doesn't
// let its Enabled box be unchecked, and that the editor of an option that isn't the pick does.
func TestLockedPickHasItsEnabledBoxDisabled(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	traitEditor, _, _ := traitEditorOnSheet(t, sheet, newTraitModifierChoiceFor(entity, true, []string{"A", "B"}, "A"))
	options := traitEditor.editorData.Modifiers[0].Children
	_, content := buildEditorContent(traitEditor, options[0], initTraitModifierEditor)
	box := findCheckBoxTitled(content, "Enabled")
	c.NotNil(box)
	c.False(box.Enabled(), "the pick can't be turned off")
	c.NotNil(box.Tooltip, "and the box says why")
	_, content = buildEditorContent(traitEditor, options[1], initTraitModifierEditor)
	c.True(findCheckBoxTitled(content, "Enabled").Enabled(), "another option can be turned on")

	pick := newEquipmentModifierChoiceFor(entity, []string{"A"}, "A").Children[0]
	_, content = buildEditorContent(nil, pick, initEquipmentModifierEditor)
	c.False(findCheckBoxTitled(content, "Enabled").Enabled(), "the same goes for equipment modifiers")
}

// modifierShowTitles are the titles of the checkboxes that say where a modifier is shown, in the order they appear: the
// first goes with the name, the other two with the notes.
var modifierShowTitles = []string{"Show in title notes", "Show notes in owner's notes", "Show notes in attack usage"}

// checkModifierShowCheckBoxes checks that the content has the modifier's show checkboxes in order, that "Show notes in
// owner's notes" starts checked and clearing it hides the notes until it is checked again, and that the other two set
// their flags.
func checkModifierShowCheckBoxes(c check.Checker, content *unison.Panel, showInTitle, hideNotes, showOnWeapon *bool) {
	c.Helper()
	var titles []string
	for _, box := range uxtest.PanelsOfType[*CheckBox](content) {
		for _, want := range modifierShowTitles {
			if box.Text.String() == want {
				titles = append(titles, want)
			}
		}
	}
	c.Equal(modifierShowTitles, titles, "the show checkboxes appear once each, in order")

	c.False(*showInTitle, "precondition: the modifier starts out in the notes")
	c.False(*hideNotes, "precondition: the notes start out shown")
	c.False(*showOnWeapon, "precondition: the notes start out off attack usage")
	title := findCheckBoxTitled(content, modifierShowTitles[0])
	owner := findCheckBoxTitled(content, modifierShowTitles[1])
	attack := findCheckBoxTitled(content, modifierShowTitles[2])
	if title == nil || owner == nil || attack == nil {
		return
	}
	c.Equal(uncheck.Off, title.State, "the title notes box starts clear")
	c.Equal(uncheck.On, owner.State, "the owner's notes box starts checked")
	c.Equal(uncheck.Off, attack.State, "the attack usage box starts clear")
	clickCheckBox(title, true)
	c.True(*showInTitle, "checking the title notes box shows the modifier in the title")
	clickCheckBox(owner, false)
	c.True(*hideNotes, "clearing the owner's notes box hides the notes")
	clickCheckBox(owner, true)
	c.False(*hideNotes, "checking the owner's notes box again shows the notes")
	clickCheckBox(attack, true)
	c.True(*showOnWeapon, "checking the attack usage box shows the notes in attack usage")
}

// TestModifierEditorsShowCheckBoxes checks the show checkboxes of the trait and equipment modifier editors.
func TestModifierEditorsShowCheckBoxes(t *testing.T) {
	c := check.New(t)
	trait, content := buildEditorContent(nil, gurps.NewTraitModifier(nil, nil, false), initTraitModifierEditor)
	checkModifierShowCheckBoxes(c, content, &trait.editorData.ShowInTitle, &trait.editorData.HideNotes,
		&trait.editorData.ShowNotesOnWeapon)
	eqp, content := buildEditorContent(nil, gurps.NewEquipmentModifier(nil, nil, false), initEquipmentModifierEditor)
	checkModifierShowCheckBoxes(c, content, &eqp.editorData.ShowInTitle, &eqp.editorData.HideNotes,
		&eqp.editorData.ShowNotesOnWeapon)
}
