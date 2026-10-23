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
)

// newTestSheetForTemplate returns a character sheet that can be built and rebuilt without a window. Both the toolbar
// and the rebuild path reach for global state, so the key bindable actions are registered and a document dock is
// installed for the test.
func newTestSheetForTemplate(t *testing.T) *Sheet {
	t.Helper()
	registerKeyBindingsOnce.Do(func() { registerActions() })
	uxtest.SwapForTest(t, &Workspace.DocumentDock, NewDocumentDock())
	return NewSheet("test"+gurps.SheetExt, gurps.NewEntity())
}

// newTestTemplateWithBodyType returns a template that supplies both a body type of the given name and a trait, so that
// applying it to a sheet alters the sheet settings as well as the traits table. The bindable actions its toolbar needs
// are registered first, so the template can be built before any sheet.
func newTestTemplateWithBodyType(bodyTypeName string) *Template {
	registerKeyBindingsOnce.Do(func() { registerActions() })
	data := gurps.NewTemplate()
	body := gurps.FactoryBody()
	body.Name = bodyTypeName
	data.BodyType = body
	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Template Trait"
	data.Traits = []*gurps.Trait{trait}
	return NewTemplate("test"+gurps.TemplatesExt, data)
}

// TestApplyTemplateCanceledPickerLeavesSheetUntouched verifies that canceling either the template choices or the
// nameables prompt, the last question put to the user before a template's rows and body type are applied, abandons the
// template, leaving the sheet exactly as it was and nothing to undo. The body type used to be replaced (and any
// existing ancestry traits disabled) before the pickers ran, so canceling one left the character with the template's
// hit locations, none of the template's content, and no undo edit with which to get the original back.
func TestApplyTemplateCanceledPickerLeavesSheetUntouched(t *testing.T) {
	for _, one := range []struct {
		name string
		// cancel substitutes a prompt that cancels, returning the count of times it is put to the user.
		cancel func(t *testing.T, template *Template) *int
	}{
		{
			name: "template choices",
			cancel: func(t *testing.T, _ *Template) *int {
				shown := 0
				uxtest.SwapForTest(t, &promptForPickers, func(_ promptOperation, _ *applyParts, _ bool) bool {
					shown++
					return false
				})
				return &shown
			},
		},
		{
			name: "nameables",
			cancel: func(t *testing.T, template *Template) *int {
				template.template.Traits[0].Name = "Phobia (@Subject@)"
				shown := 0
				uxtest.SwapForTest(t, &promptForNameables, func(_ promptOperation, _ []nameablesSection) bool {
					shown++
					return false
				})
				return &shown
			},
		},
	} {
		t.Run(one.name, func(t *testing.T) {
			c := check.New(t)
			sheet := newTestSheetForTemplate(t)
			entity := sheet.Entity()
			originalBody := entity.SheetSettings.BodyType
			originalTraitCount := len(entity.Traits)
			template := newTestTemplateWithBodyType("Template Body")
			shown := one.cancel(t, template)

			c.False(template.applyTemplateToSheet(sheet, promptOperation{}, true),
				"a canceled prompt must report that the template was not applied")
			c.Equal(1, *shown, "the prompt must have been put to the user")
			c.Equal(originalBody, entity.SheetSettings.BodyType, "the body type must not have been replaced")
			c.Equal(originalTraitCount, len(entity.Traits), "no traits must have been added")
			c.Equal(originalTraitCount, len(sheet.Traits.Table.RootRows()), "no rows must have been added to the traits table")
			c.False(sheet.Modified(), "the sheet must not have been marked as modified")
			c.False(sheet.undoMgr.CanUndo(), "nothing was changed, so there must be nothing to undo")
		})
	}
}

// Creating a character from a template used to leave the new, empty sheet open when a picker was canceled, even though
// canceling abandons the whole operation.
func TestNewSheetFromTemplateCanceledPickerClosesTheSheet(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	template := newTestTemplateWithBodyType("Template Body")
	uxtest.SwapForTest(t, &promptForPickers, func(_ promptOperation, _ *applyParts, _ bool) bool { return false })
	var open int
	screen.Do(func() {
		template.newSheetFromTemplate(nil)
		open = len(AllDockables())
	})
	c.Equal(0, open, "canceling a picker must close the new sheet")
}

// TestApplyTemplateUndoRestoresBodyType verifies that applying a template gives the sheet a copy of the template's body
// type, which undo and redo take away and put back, and that a template without a body type leaves the sheet's alone,
// both when applied and when undone. The undo data used to preserve only the profile randomizer fields and the five
// table data sets, leaving the character permanently stuck with the template's body type.
func TestApplyTemplateUndoRestoresBodyType(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	originalBodyName := entity.SheetSettings.BodyType.Name
	c.NotEqual("Template Body", originalBodyName, "the test requires the template's body type to be distinguishable")
	template := newTestTemplateWithBodyType("Template Body")

	c.True(template.applyTemplateToSheet(sheet, promptOperation{}, true), "the template must be applied")
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

	sheet = newTestSheetForTemplate(t)
	entity = sheet.Entity()
	originalBody := entity.SheetSettings.BodyType
	template = newTestTemplateWithBodyType("Template Body")
	template.template.BodyType = nil

	c.True(template.applyTemplateToSheet(sheet, promptOperation{}, true), "a template without a body type must be applied")
	c.Equal(originalBody, entity.SheetSettings.BodyType, "a template without a body type must leave the body type alone")

	sheet.undoMgr.Undo()
	c.Equal(originalBody.Name, entity.SheetSettings.BodyType.Name, "undo must leave the body type alone as well")
}
