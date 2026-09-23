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

// newTestSheetForTemplate returns a character sheet that can be built and rebuilt without a window. Both the toolbar
// and the rebuild path reach for global state, so the key bindable actions are registered and a document dock is
// installed for the test.
func newTestSheetForTemplate(t *testing.T) *Sheet {
	t.Helper()
	registerKeyBindingsOnce.Do(func() { registerActions() })
	swapForTest(t, &Workspace.DocumentDock, NewDocumentDock())
	return NewSheet("test"+gurps.SheetExt, gurps.NewEntity())
}

// newTestTemplateWithBodyType returns a template that supplies both a body type of the given name and a trait, so that
// applying it to a sheet alters the sheet settings as well as the traits table.
func newTestTemplateWithBodyType(bodyTypeName string) *Template {
	data := gurps.NewTemplate()
	body := gurps.FactoryBody()
	body.Name = bodyTypeName
	data.BodyType = body
	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Template Trait"
	data.Traits = []*gurps.Trait{trait}
	return NewTemplate("test"+gurps.TemplatesExt, data)
}

// The body type used to be replaced (and any existing ancestry traits disabled) before the pickers ran, so canceling
// one left the character with the template's hit locations, none of the template's content, and no undo edit with
// which to get the original back.
func TestApplyTemplateCanceledPickerLeavesSheetUntouched(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	originalBody := entity.SheetSettings.BodyType
	originalTraitCount := len(entity.Traits)
	template := newTestTemplateWithBodyType("Template Body")

	c.False(ApplyTemplateToSheetWithPickers(template, sheet, true, func(_ *templateParts) bool { return false }),
		"a canceled picker must report that the template was not applied")
	c.Equal(originalBody, entity.SheetSettings.BodyType, "the body type must not have been replaced")
	c.Equal(originalTraitCount, len(entity.Traits), "no traits must have been added")
	c.Equal(originalTraitCount, len(sheet.Traits.Table.RootRows()), "no rows must have been added to the traits table")
	c.False(sheet.Modified(), "the sheet must not have been marked as modified")
	c.False(sheet.undoMgr.CanUndo(), "nothing was changed, so there must be nothing to undo")
}

// The undo data used to preserve only the profile randomizer fields and the five table data sets, leaving the
// character permanently stuck with the template's body type.
func TestApplyTemplateUndoRestoresBodyType(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	originalBodyName := entity.SheetSettings.BodyType.Name
	c.NotEqual("Template Body", originalBodyName, "the test requires the template's body type to be distinguishable")
	template := newTestTemplateWithBodyType("Template Body")

	c.True(ApplyTemplateToSheet(template, sheet, true), "the template must be applied")
	c.Equal("Template Body", entity.SheetSettings.BodyType.Name, "the template's body type must be applied")
	c.NotEqual(template.template.BodyType, entity.SheetSettings.BodyType,
		"the sheet must get a copy of the template's body type, not the template's own")
	c.True(sheet.undoMgr.CanUndo(), "applying a template must be undoable")

	sheet.undoMgr.Undo()
	c.Equal(originalBodyName, entity.SheetSettings.BodyType.Name, "undo must restore the original body type")

	sheet.undoMgr.Redo()
	c.Equal("Template Body", entity.SheetSettings.BodyType.Name, "redo must reapply the template's body type")

	// Editing the body type in place must not corrupt the copies the undo edit is holding onto.
	entity.SheetSettings.BodyType.Name = "Edited"
	sheet.undoMgr.Undo()
	c.Equal(originalBodyName, entity.SheetSettings.BodyType.Name, "a second undo must still restore the original")
	sheet.undoMgr.Redo()
	c.Equal("Template Body", entity.SheetSettings.BodyType.Name, "a second redo must still reapply the template's")
}

func TestApplyTemplateWithoutBodyTypeLeavesBodyTypeAlone(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	originalBody := entity.SheetSettings.BodyType
	template := newTestTemplateWithBodyType("Template Body")
	template.template.BodyType = nil

	c.True(ApplyTemplateToSheet(template, sheet, true), "the template must be applied")
	c.Equal(originalBody, entity.SheetSettings.BodyType, "the body type must have been left alone")

	sheet.undoMgr.Undo()
	c.Equal(originalBody.Name, entity.SheetSettings.BodyType.Name, "undo must leave the body type alone as well")
}

// TestApplyTemplateAsksForTheNameablesOfAModifierEnabledDuringTheApply verifies that the two prompts an applied
// template presents run in the order that makes the second one complete: modifiers first, then nameables. Only an
// enabled modifier contributes nameable keys -- FillWithNameableKeys traverses a row's modifiers with onlyEnabled set
// -- so asking for the names first would never ask about a modifier the user turns on at the prompt, leaving raw
// @Key@ text on it with no way to fill it in afterwards.
func TestApplyTemplateAsksForTheNameablesOfAModifierEnabledDuringTheApply(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()

	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Trained By A Master"
	modifier := gurps.NewTraitModifier(nil, nil, false)
	modifier.Name = "@Style@ Training"
	modifier.Disabled = true // Contributes no nameable keys until the modifier prompt turns it on.
	trait.Modifiers = []*gurps.TraitModifier{modifier}
	templateData := gurps.NewTemplate()
	templateData.SetTraitList([]*gurps.Trait{trait})
	template := newTestTemplateDockable("Source", templateData)

	modifiersShown := stubTraitModifierPrompt(t, enableAllModifiers)
	nameablesShown := stubNameablesPrompt(t, fillNameables("Style", "Karate"))

	c.True(ApplyTemplateToSheet(template, sheet, true), "the template must be applied")
	c.Equal(1, *modifiersShown, "the modifier prompt must have been presented")
	c.Equal(1, *nameablesShown, "the nameables prompt must have been presented after it")

	applied := entity.Traits[len(entity.Traits)-1]
	c.Equal("Trained By A Master", applied.Name, "the template's trait must have been added")
	c.Equal(1, len(applied.Modifiers), "the trait must have kept its modifier")
	c.False(applied.Modifiers[0].Disabled, "the modifier prompt must have enabled the modifier")
	// The answers are reduced onto the row that owns the modifier, not onto the modifier itself, since a modifier's
	// markers are resolved through its owner's replacement map (Trait.ApplyNameableKeys).
	c.Equal("Karate", applied.Replacements["Style"],
		"the nameables of a modifier enabled during the apply must have been asked for and recorded")
}

// TestApplyTemplateAsksForTheNameablesOfANote verifies that a note arriving from a template is asked about like every
// other row. Notes were left out of the prompt for a while, which landed them on the sheet with their raw @Key@ text
// showing and no way to fill it in short of editing the note by hand.
func TestApplyTemplateAsksForTheNameablesOfANote(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()

	note := gurps.NewNote(nil, nil, false)
	note.MarkDown = "Sworn to @Patron@"
	templateData := gurps.NewTemplate()
	templateData.SetNoteList([]*gurps.Note{note})
	template := newTestTemplateDockable("Source", templateData)

	shown := stubNameablesPrompt(t, fillNameables("Patron", "The Duke"))

	c.True(ApplyTemplateToSheet(template, sheet, true), "the template must be applied")
	c.Equal(1, *shown, "the note's nameable key must have been prompted for")

	applied := entity.Notes[len(entity.Notes)-1]
	c.Equal("Sworn to The Duke", applied.TextWithReplacements(),
		"the answer must have been recorded on the note")
}
