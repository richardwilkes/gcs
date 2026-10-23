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
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// TestOptionalDecimalFieldBlankMeansNone verifies that zero is a value of its own, and that text which isn't a number is
// flagged rather than taken as one. Blank meaning none is covered through the template's target fields.
func TestOptionalDecimalFieldBlankMeansNone(t *testing.T) {
	c := check.New(t)
	var value *fxp.Int
	f := NewOptionalDecimalField(nil, "", "Value",
		func() *fxp.Int { return value },
		func(v *fxp.Int) { value = v })
	c.Equal("", f.Text(), "no value shows as a blank field")

	f.SetText("0")
	c.NotNil(value, "zero is a value of its own")
	c.Equal(fxp.Int(0), *value)

	f.SetText("abc")
	c.False(f.validate(), "text that isn't a number must be flagged")
}

// TestTemplateTopBlockEditsTheTemplate verifies that the fields at the top of a template edit the template's own
// values, with each end of the target range settable on its own and an inverted range flagged.
func TestTemplateTopBlockEditsTheTemplate(t *testing.T) {
	c := check.New(t)
	data := gurps.NewTemplate()
	template := newTestTemplateDockable("Top", data)

	name, ok := template.targetMgr.Find(templatePanelNameFieldRefKey).Self.(*StringField)
	c.True(ok, "the template must have a name field")
	name.SetText("Knight")
	c.Equal("Knight", data.Name)

	notes, ok := template.targetMgr.Find(templatePanelLocalNotesFieldRefKey).Self.(*StringField)
	c.True(ok, "the template must have a notes field")
	notes.SetText("line one\nline two")
	c.Equal("line one\nline two", data.LocalNotes, "the notes must keep every line")

	minField, ok := template.targetMgr.Find(templatePanelTargetMinFieldRefKey).Self.(*OptionalDecimalField)
	c.True(ok, "the template must have a target minimum field")
	maxField, ok := template.targetMgr.Find(templatePanelTargetMaxFieldRefKey).Self.(*OptionalDecimalField)
	c.True(ok, "the template must have a target maximum field")

	minField.SetText("100")
	c.Equal("100+", data.TargetPoints.String(), "a minimum alone leaves the top of the range open")
	maxField.SetText("150")
	c.Equal("100~150", data.TargetPoints.String())
	maxField.SetText("100")
	c.Equal("100", data.TargetPoints.String(), "the same value at both ends is an exact target")

	maxField.SetText("50")
	c.False(maxField.validate(), "a maximum below the minimum must be flagged")
	c.False(minField.validate(), "a minimum above the maximum must be flagged")
	minField.SetText("")
	c.True(maxField.validate(), "a maximum with no minimum can't be inverted")
	c.Equal("≤50", data.TargetPoints.String())
}

// TestTemplateTopBlockFlagsALoadedInvertedRange verifies that a target range that is already inverted when the
// template is opened is flagged straight away, rather than only once a later edit syncs the fields.
func TestTemplateTopBlockFlagsALoadedInvertedRange(t *testing.T) {
	c := check.New(t)
	data := gurps.NewTemplate()
	data.TargetPoints = gurps.NumericRange{Min: new(fxp.FromInteger(150)), Max: new(fxp.FromInteger(100))}
	template := newTestTemplateDockable("Inverted", data)

	minField, ok := template.targetMgr.Find(templatePanelTargetMinFieldRefKey).Self.(*OptionalDecimalField)
	c.True(ok, "the template must have a target minimum field")
	maxField, ok := template.targetMgr.Find(templatePanelTargetMaxFieldRefKey).Self.(*OptionalDecimalField)
	c.True(ok, "the template must have a target maximum field")
	c.True(minField.Invalid(), "a minimum loaded above the maximum must be flagged on opening")
	c.True(maxField.Invalid(), "a maximum loaded below the minimum must be flagged on opening")
	c.NotNil(minField.Tooltip, "the flagged minimum must say why")
	c.NotNil(maxField.Tooltip, "the flagged maximum must say why")
}

// TestTargetPointsMet verifies that a template only meets its target when every cost it may end up with falls within
// that target, and that an open end of either range is treated as reaching as far as it can.
func TestTargetPointsMet(t *testing.T) {
	c := check.New(t)
	of := func(value int) *fxp.Int { return new(fxp.FromInteger(value)) }
	target := gurps.NumericRange{Min: of(100), Max: of(150)}
	c.True(targetPointsMet(gurps.NumericRangeOf(fxp.FromInteger(120)), gurps.NumericRange{}), "no target is always met")
	c.True(targetPointsMet(gurps.NumericRangeOf(fxp.FromInteger(120)), target))
	c.True(targetPointsMet(gurps.NumericRange{Min: of(100), Max: of(150)}, target), "the ends are inclusive")
	c.False(targetPointsMet(gurps.NumericRange{Min: of(90), Max: of(120)}, target), "an outcome below must miss")
	c.False(targetPointsMet(gurps.NumericRange{Min: of(120), Max: of(160)}, target), "an outcome above must miss")
	c.False(targetPointsMet(gurps.NumericRange{Min: of(120)}, target), "an open top can't stay under a maximum")
	c.True(targetPointsMet(gurps.NumericRange{Min: of(120)}, gurps.NumericRange{Min: of(100)}),
		"an open top is fine when the target has none either")
}

// TestTemplatePageRefButtonFollowsTheField verifies that the button beside the page reference is only enabled while the
// field holds a reference to open, including after an undo.
func TestTemplatePageRefButtonFollowsTheField(t *testing.T) {
	c := check.New(t)
	data := gurps.NewTemplate()
	template := newTestTemplateDockable("PageRef", data)
	field, ok := template.targetMgr.Find(templatePanelPageRefFieldRefKey).Self.(*StringField)
	c.True(ok, "the template must have a page reference field")
	siblings := field.Parent().Children()
	index := slices.Index(siblings, field.AsPanel())
	c.True(index > 0, "the page reference field must have a label before it")
	button, ok := siblings[index-1].Children()[1].Self.(*unison.Button)
	c.True(ok, "the page reference label must be followed by its button")
	c.False(button.Enabled(), "with no reference there is nothing to open")

	field.SetText("DF1:5")
	c.Equal("DF1:5", data.PageRef)
	c.True(button.Enabled(), "a reference enables the button")
	template.undoMgr.Undo()
	c.Equal("", data.PageRef)
	c.False(button.Enabled(), "undoing the reference disables the button again")
	field.SetText(" , ")
	c.False(button.Enabled(), "separators alone are no reference")
}

// TestTemplateNewCharacterStartsAtAnExactTarget verifies that a sheet made from a template aimed at an exact number of
// points starts with that many, and that any other target leaves the default from the settings alone.
func TestTemplateNewCharacterStartsAtAnExactTarget(t *testing.T) {
	c := check.New(t)
	defaultPoints := gurps.GlobalSettings().GeneralSettings().InitialPoints
	c.NotEqual(fxp.FromInteger(250), defaultPoints, "the test requires the target to differ from the default")
	data := gurps.NewTemplate()
	template := newTestTemplateDockable("Exact", data)

	data.TargetPoints = gurps.NumericRangeOf(fxp.FromInteger(250))
	e := template.newCharacter()
	c.Equal(fxp.FromInteger(250), e.TotalPoints)
	c.Equal(1, len(e.PointsRecord), "the target replaces the initial points rather than adding to them")
	c.Equal(fxp.FromInteger(250), e.PointsRecord[0].Points)

	for _, target := range []gurps.NumericRange{
		{},
		{Min: new(fxp.FromInteger(100)), Max: new(fxp.FromInteger(150))},
		{Min: new(fxp.FromInteger(100))},
	} {
		data.TargetPoints = target
		c.Equal(defaultPoints, template.newCharacter().TotalPoints, "a target of %s must leave the default", target)
	}
}
