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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
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

	c.False(template.applyTemplateToSheetWithPickers(sheet, true, func(_ *templateRows) bool { return false }),
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

	c.True(template.applyTemplateToSheet(sheet, true), "the template must be applied")
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

	c.True(template.applyTemplateToSheet(sheet, true), "the template must be applied")
	c.Equal(originalBody, entity.SheetSettings.BodyType, "the body type must have been left alone")

	sheet.undoMgr.Undo()
	c.Equal(originalBody.Name, entity.SheetSettings.BodyType.Name, "undo must leave the body type alone as well")
}

// The picker replaces its choice group with the selected children inside the fixed-cost parent.
func TestApplyTemplateFixedCostSelection(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	originalCount := len(sheet.Entity().Traits)
	data := gurps.NewTemplate()
	parent := gurps.NewTrait(data, nil, true)
	parent.Name = "Fixed package"
	parent.ContainerType = container.FixedCost
	parent.FixedPoints = new(fxp.FromInteger(30))
	choices := gurps.NewTrait(data, parent, true)
	choices.TemplatePicker.Type = picker.Points
	choices.TemplatePicker.Qualifier.Compare = criteria.EqualsNumber
	choices.TemplatePicker.Qualifier.Qualifier = fxp.FromInteger(30)
	for _, cost := range []int{10, 20, 99} {
		child := gurps.NewTrait(data, choices, false)
		child.BasePoints = fxp.FromInteger(cost)
		choices.Children = append(choices.Children, child)
	}
	parent.Children = []*gurps.Trait{choices}
	data.Traits = []*gurps.Trait{parent}
	template := NewTemplate("test"+gurps.TemplatesExt, data)
	var validatedTotal fxp.Int
	applied := template.applyTemplateToSheetWithPickers(sheet, true, func(rows *templateRows) bool {
		// Supply the user's selection at the dialog boundary; cloning and sheet insertion use the production path.
		copied := ExtractNodeDataFromList(rows.traits)[0]
		c.Equal(container.Group, copied.ContainerType)
		c.True(copied.FixedPoints == nil)
		copiedChoices := copied.Children[0]
		selected := copiedChoices.Children[:2]
		total, matches := pickerSelectionState(&copiedChoices.TemplatePicker, selected)
		validatedTotal = total
		if !matches {
			return false
		}
		copied.SetChildren(selected)
		SetParents(selected, copied)
		return true
	})
	c.True(applied)
	if len(sheet.Entity().Traits) != originalCount+1 {
		t.Fatalf("expected one added container, got %d", len(sheet.Entity().Traits)-originalCount)
	}
	copied := sheet.Entity().Traits[originalCount]
	c.Equal(container.Group, copied.ContainerType)
	c.True(copied.FixedPoints == nil)
	c.Equal(2, len(copied.Children))
	c.Equal(fxp.FromInteger(30), validatedTotal)
	c.Equal(validatedTotal, copied.AdjustedPoints())
	for _, child := range copied.Children {
		c.True(child.Parent() == copied)
		c.True(child.DataOwner() == sheet.Entity())
	}
	c.Equal(container.FixedCost, parent.ContainerType)
	c.Equal(new(fxp.FromInteger(30)), parent.FixedPoints)
	c.Equal(3, len(choices.Children), "the template retains all choices")
}
