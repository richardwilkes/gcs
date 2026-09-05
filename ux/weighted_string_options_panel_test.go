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
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// listPanelFor returns the panel editing the given weighted string list, failing the test if there is none.
func listPanelFor(t *testing.T, d structuralEditor, list *[]*gurps.WeightedStringOption) *weightedStringOptionsPanel {
	t.Helper()
	for _, p := range panelsOfType[*weightedStringOptionsPanel](d.AsPanel()) {
		if p.spec.list == list {
			return p
		}
	}
	t.Fatal("no panel edits the list")
	return nil
}

func optionValues(list []*gurps.WeightedStringOption) []string {
	values := make([]string, 0, len(list))
	for _, one := range list {
		values = append(values, one.Value)
	}
	return values
}

// TestWeightedStringOptionRowBindsFields verifies that a row's value and weight fields write to the model, that a
// weight below the minimum shows the field's validation tooltip and is held at the minimum, and that the explanatory
// tooltip returns once the weight is valid again. Zero is allowed, since the randomizers skip a zero-weight entry and a
// file that contains one must load without being altered.
func TestWeightedStringOptionRowBindsFields(t *testing.T) {
	c := check.New(t)
	a := gurps.NewAncestry()
	a.CommonOptions.HairOptions = []*gurps.WeightedStringOption{{Weight: 3, Value: "Black"}}
	d := newTestAncestryEditorDockable(a)
	option := a.CommonOptions.HairOptions[0]

	stringFieldFor(t, d, option.KeyPrefix+"value").SetText("Silver")
	c.Equal("Silver", option.Value)

	weightField := integerFieldFor(t, d, option.KeyPrefix+"weight")
	explanation := tooltipText(weightField.Tooltip)
	c.Contains(explanation, "relative likelihood", "the weight field explains itself")
	weightField.SetText("12")
	c.Equal(12, option.Weight)
	// While the field has the focus, as it does when the user is typing, a resync leaves its text as typed rather than
	// replacing it with the model's value, so the validation of what was typed can be seen.
	weightField.hasFocus = true
	weightField.SetText("0")
	c.Equal(0, option.Weight, "a zero weight is accepted, since the randomizers simply skip the entry")
	c.Equal(explanation, tooltipText(weightField.Tooltip), "a zero weight is not flagged")
	weightField.SetText("-1")
	c.Equal("Value must be at least 0", tooltipText(weightField.Tooltip), "a negative weight is flagged")
	c.Equal(0, option.Weight, "the field holds the model at its minimum meanwhile")
	weightField.SetText("5")
	c.Equal(5, option.Weight)
	c.Equal(explanation, tooltipText(weightField.Tooltip), "the explanation returns once the weight is valid")
	weightField.hasFocus = false
	c.True(d.Modified())
}

// TestWeightedStringOptionsAddAndRemove verifies that an empty list shows no rows container, that adding an option
// appends a weight-1 option with its own key prefix and builds a row for it, that removing the last option removes the
// rows container again, and that both are undoable.
func TestWeightedStringOptionsAddAndRemove(t *testing.T) {
	c := check.New(t)
	a := gurps.NewAncestry()
	d := newTestAncestryEditorDockable(a)
	list := listPanelFor(t, d, &a.CommonOptions.HairOptions)
	c.Nil(list.rows, "an empty list has no rows container")
	c.Equal(1, len(list.Children()), "only the header is shown for an empty list")

	list.addOption()
	c.Equal(1, len(d.model.CommonOptions.HairOptions))
	added := d.model.CommonOptions.HairOptions[0]
	c.Equal(1, added.Weight)
	c.Equal("", added.Value)
	c.NotEqual("", added.KeyPrefix, "the new option has a key prefix")
	list = listPanelFor(t, d, &d.model.CommonOptions.HairOptions)
	c.NotNil(list.rows, "the rows container appears with the first option")
	c.Equal(1, len(list.rows.Children()))
	c.NotNil(d.targetMgr.Find(added.KeyPrefix+"value"), "the new option's value field exists")
	c.True(d.Modified())

	row, ok := list.rows.Children()[0].Self.(*weightedStringOptionPanel)
	c.True(ok, "the rows container holds option rows")
	c.True(row.deleteButton.Enabled(), "the last option may still be removed")
	list.removeOption(row.option)
	c.Equal(0, len(d.model.CommonOptions.HairOptions), "the last option is removed")
	list = listPanelFor(t, d, &d.model.CommonOptions.HairOptions)
	c.Nil(list.rows, "the rows container is gone again")
	c.False(d.Modified())

	d.undoMgr.Undo()
	c.Equal(1, len(d.model.CommonOptions.HairOptions), "undo restores the removed option")
	d.undoMgr.Undo()
	c.Equal(0, len(d.model.CommonOptions.HairOptions), "undo removes the added option")
	c.False(d.undoMgr.CanUndo(), "each edit is its own undo")
}

// TestWeightedStringOptionsRemoveFromMiddle verifies that an option is removed by identity, not by position, so the
// other options keep their order.
func TestWeightedStringOptionsRemoveFromMiddle(t *testing.T) {
	c := check.New(t)
	a := gurps.NewAncestry()
	a.CommonOptions.EyeOptions = []*gurps.WeightedStringOption{
		{Weight: 1, Value: "Brown"},
		{Weight: 1, Value: "Blue"},
		{Weight: 1, Value: "Green"},
	}
	d := newTestAncestryEditorDockable(a)
	list := listPanelFor(t, d, &a.CommonOptions.EyeOptions)
	list.removeOption(a.CommonOptions.EyeOptions[1])
	c.Equal([]string{"Brown", "Green"}, optionValues(d.model.CommonOptions.EyeOptions))
	c.Equal(2, len(listPanelFor(t, d, &d.model.CommonOptions.EyeOptions).rows.Children()))
}

// TestWeightedStringOptionsPerGenderListsBindToGender verifies that each gender's lists edit that gender's options, not
// the common ones, and that every one of the four lists is present for each block.
func TestWeightedStringOptionsPerGenderListsBindToGender(t *testing.T) {
	c := check.New(t)
	a := gurps.NewAncestry()
	d := newTestAncestryEditorDockable(a)
	g := a.GenderOptions[1].Value
	for _, list := range []*[]*gurps.WeightedStringOption{&g.HairOptions, &g.EyeOptions, &g.SkinOptions, &g.HandednessOptions} {
		listPanelFor(t, d, list).addOption()
	}
	g = d.model.GenderOptions[1].Value
	c.Equal(1, len(g.HairOptions))
	c.Equal(1, len(g.EyeOptions))
	c.Equal(1, len(g.SkinOptions))
	c.Equal(1, len(g.HandednessOptions))
	common := d.model.CommonOptions
	c.Equal(0, len(common.HairOptions)+len(common.EyeOptions)+len(common.SkinOptions)+len(common.HandednessOptions),
		"the common options are untouched")
	other := d.model.GenderOptions[0].Value
	c.Equal(0, len(other.HairOptions)+len(other.EyeOptions)+len(other.SkinOptions)+len(other.HandednessOptions),
		"the other gender is untouched")
}

// TestWeightedStringOptionsSelection verifies how clicks build up a selection -- a plain click selects one row, a
// shift-click extends from the last plain click in either direction, and a command-click toggles a row without
// touching the others -- that the header's remove button follows the selection, that removing the selection takes out
// exactly the selected options in one undoable edit, and that the selection does not outlive the rows it was made on.
func TestWeightedStringOptionsSelection(t *testing.T) {
	c := check.New(t)
	a := gurps.NewAncestry()
	a.CommonOptions.HairOptions = []*gurps.WeightedStringOption{
		{Weight: 1, Value: "Black"},
		{Weight: 1, Value: "Brown"},
		{Weight: 1, Value: "Blond"},
		{Weight: 1, Value: "Red"},
		{Weight: 1, Value: "Gray"},
	}
	d := newTestAncestryEditorDockable(a)
	list := listPanelFor(t, d, &a.CommonOptions.HairOptions)
	options := a.CommonOptions.HairOptions
	c.False(list.removeSelectedButton.Enabled(), "with nothing selected there is nothing to remove")
	c.Equal(0, len(list.selectedOptions()))

	list.selectRow(options[1], mod.None)
	c.Equal([]string{"Brown"}, optionValues(list.selectedOptions()), "a plain click selects the row alone")
	c.True(list.removeSelectedButton.Enabled())
	list.selectRow(options[3], mod.Shift)
	c.Equal([]string{"Brown", "Blond", "Red"}, optionValues(list.selectedOptions()), "shift extends to the clicked row")
	list.selectRow(options[0], mod.Shift)
	c.Equal([]string{"Black", "Brown", "Blond", "Red"}, optionValues(list.selectedOptions()),
		"and extends the other way from the same anchor, keeping what was selected")
	list.selectRow(options[2], mod.OSMenuCommand())
	c.Equal([]string{"Black", "Brown", "Red"}, optionValues(list.selectedOptions()), "command-click takes a row out")
	list.selectRow(options[4], mod.OSMenuCommand())
	c.Equal([]string{"Black", "Brown", "Red", "Gray"}, optionValues(list.selectedOptions()), "or adds one")
	list.selectRow(options[4], mod.OSMenuCommand())
	list.selectRow(options[4], mod.OSMenuCommand())
	list.selectRow(options[4], mod.OSMenuCommand())
	c.Equal([]string{"Black", "Brown", "Red"}, optionValues(list.selectedOptions()))
	c.True(list.isSelected(options[0]))
	c.False(list.isSelected(options[2]))
	list.selectRow(options[3], mod.None)
	c.Equal([]string{"Red"}, optionValues(list.selectedOptions()), "a plain click starts over")
	list.selectRow(options[4], mod.Shift)
	c.Equal([]string{"Red", "Gray"}, optionValues(list.selectedOptions()))

	list.removeSelected()
	c.Equal([]string{"Black", "Brown", "Blond"}, optionValues(d.model.CommonOptions.HairOptions),
		"exactly the selected options are removed")
	c.True(d.Modified())
	list = listPanelFor(t, d, &d.model.CommonOptions.HairOptions)
	c.Equal(0, len(list.selectedOptions()), "the rebuilt rows start with nothing selected")
	c.False(list.removeSelectedButton.Enabled())
	list.removeSelected()
	c.Equal(3, len(d.model.CommonOptions.HairOptions), "removing nothing changes nothing")

	d.undoMgr.Undo()
	c.Equal([]string{"Black", "Brown", "Blond", "Red", "Gray"}, optionValues(d.model.CommonOptions.HairOptions),
		"one undo brings every removed option back")
	c.False(d.undoMgr.CanUndo(), "the removal was a single edit")
}

// TestWeightedStringOptionRowClicksSelect verifies that a left click on a row, or on its drag handle, selects the row
// with the click's modifiers, while a click with another button is left for whatever else wants it.
func TestWeightedStringOptionRowClicksSelect(t *testing.T) {
	c := check.New(t)
	a := gurps.NewAncestry()
	a.CommonOptions.SkinOptions = []*gurps.WeightedStringOption{
		{Weight: 1, Value: "Pale"},
		{Weight: 1, Value: "Dark"},
	}
	d := newTestAncestryEditorDockable(a)
	list := listPanelFor(t, d, &a.CommonOptions.SkinOptions)
	rows := list.rows.Children()
	c.False(rows[0].MouseDownCallback(geom.Point{}, unison.ButtonRight, 1, mod.None),
		"a right click is not a selection click")
	c.Equal(0, len(list.selectedOptions()))
	c.True(rows[0].MouseDownCallback(geom.Point{}, unison.ButtonLeft, 1, mod.None), "a left click is consumed")
	c.Equal([]string{"Pale"}, optionValues(list.selectedOptions()))
	handle := panelsOfType[*DragHandle](rows[1])[0]
	c.True(handle.MouseDownCallback(geom.Point{}, unison.ButtonLeft, 1, mod.Shift),
		"the handle still takes the press, so a drag can follow")
	c.Equal([]string{"Pale", "Dark"}, optionValues(list.selectedOptions()), "pressing the handle selects too")
	c.Contains(tooltipText(rows[0].Tooltip), "select", "the row says how it is selected")
}
