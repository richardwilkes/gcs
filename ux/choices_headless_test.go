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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// TestChoiceQualifierUndoAfterTypeChange verifies that undoing an edit made to a choice's qualifier before its type was
// changed can't leave the choice asking for something its new type forbids. The edit belongs to the field the type
// change replaced, whose own limits were those of the old type.
func TestChoiceQualifierUndoAfterTypeChange(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	trait := gurps.NewTraitChoiceContainer(nil, nil)
	trait.TemplatePicker.Type = picker.Points
	trait.TemplatePicker.Qualifier.Compare = criteria.AtMostNumber
	trait.TemplatePicker.Qualifier.Qualifier = fxp.FromInteger(-10)
	data := gurps.NewTemplate()
	data.Traits = []*gurps.Trait{trait}
	var e *editor[*gurps.Trait, *gurps.TraitEditData]
	screen.Do(func() {
		template := newTestTemplateDockable("Template", data)
		DisplayNewDockable(template)
		e = EditTrait(template, trait)
	})
	c.NotNil(e)
	var field *DecimalField
	var typePopup *unison.PopupMenu[picker.Type]
	screen.Do(func() {
		for _, one := range panelsOfType[*DecimalField](e.AsPanel()) {
			if one.Min() == fxp.Min {
				field = one
			}
		}
		for _, one := range panelsOfType[*unison.PopupMenu[picker.Type]](e.AsPanel()) {
			typePopup = one
		}
	})
	c.NotNil(field, "a points choice must offer a qualifier that may be less than nothing")
	c.NotNil(typePopup)
	screen.Do(func() { field.SetText("-5") })
	c.Equal(fxp.FromInteger(-5), e.editorData.TemplatePicker.Qualifier.Qualifier)
	screen.Do(func() { typePopup.Select(picker.Count) })
	c.Equal(fxp.Int(0), e.editorData.TemplatePicker.Qualifier.Qualifier, "a count may not be less than nothing")
	screen.Do(func() { e.UndoManager().Undo() })
	c.Equal(picker.Count, e.editorData.TemplatePicker.Type)
	c.Equal(fxp.Int(0), e.editorData.TemplatePicker.Qualifier.Qualifier,
		"undoing the earlier edit must not make the count less than nothing")
}
