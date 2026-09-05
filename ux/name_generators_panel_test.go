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

// nameGeneratorsPanelFor returns the panel editing the name generators of the given options block, failing the test if
// there is none.
func nameGeneratorsPanelFor(t *testing.T, d *ancestryEditorDockable, options *gurps.AncestryOptions) *nameGeneratorsPanel {
	t.Helper()
	for _, p := range panelsOfType[*nameGeneratorsPanel](d.AsPanel()) {
		if p.options == options {
			return p
		}
	}
	t.Fatal("no panel edits the options block's name generators")
	return nil
}

func selectedGenerator(p *Popup[string]) string {
	value, _ := p.Selected()
	return value
}

// TestNameGeneratorPopupItems verifies that the popup items are the available choices with an unknown current value
// prepended, that a known current value adds nothing, and that the result is always a fresh slice, so the cached
// choices are never appended into.
func TestNameGeneratorPopupItems(t *testing.T) {
	c := check.New(t)
	choices := []string{"Human First - Male", "Human Last"}

	items := nameGeneratorPopupItems(choices, "Human Last")
	c.Equal(choices, items, "a known current value adds nothing")
	items[0] = "changed"
	c.Equal("Human First - Male", choices[0], "the result is a copy")

	c.Equal([]string{"Elven", "Human First - Male", "Human Last"}, nameGeneratorPopupItems(choices, "Elven"),
		"an unknown current value comes first")
	c.Equal([]string{"Human First - Male", "Human Last"}, choices, "the choices are not appended into")

	c.Equal([]string{"Elven"}, nameGeneratorPopupItems(nil, "Elven"), "no choices: only the current value")
	c.Equal([]string{""}, nameGeneratorPopupItems(nil, ""), "an empty current value is still offered")
}

// TestNameGeneratorRowKeepsUnknownGenerator verifies that a generator missing from every library is still shown and
// selected, so that it can be kept or replaced, and that choosing another writes the model with exactly one undoable
// edit.
func TestNameGeneratorRowKeepsUnknownGenerator(t *testing.T) {
	c := check.New(t)
	a := gurps.NewAncestry()
	a.CommonOptions.NameGenerators = []string{"Missing Names"}
	d := newTestAncestryEditorDockable(a, "Human First - Male", "Human Last")
	rows := panelsOfType[*nameGeneratorRefPanel](d.AsPanel())
	c.Equal(1, len(rows), "one row per generator")
	row := rows[0]
	c.Equal("Missing Names", selectedGenerator(row.popup), "the unknown generator is selected")
	c.Equal(3, row.popup.ItemCount(), "the unknown generator is offered along with the known ones")

	row.popup.Select("Human Last")
	c.Equal([]string{"Human Last"}, d.model.CommonOptions.NameGenerators, "the selection writes the model")
	c.True(d.Modified())
	c.True(d.undoMgr.CanUndo())

	d.undoMgr.Undo()
	c.Equal([]string{"Missing Names"}, d.model.CommonOptions.NameGenerators, "undo restores the model")
	c.Equal("Missing Names", selectedGenerator(row.popup), "undo restores the selection")
	c.False(d.undoMgr.CanUndo(), "the selection produced exactly one edit")
}

// TestNameGeneratorRowEditButtonFollowsAvailability verifies that each row has a button to open the chosen generator in
// the name generator editor, enabled only while the chosen generator is one some library holds, and that it follows the
// popup's selection without the rows being rebuilt.
func TestNameGeneratorRowEditButtonFollowsAvailability(t *testing.T) {
	c := check.New(t)
	a := gurps.NewAncestry()
	a.CommonOptions.NameGenerators = []string{"Missing Names", "Human Last"}
	d := newTestAncestryEditorDockable(a, "Human First - Male", "Human Last")
	rows := panelsOfType[*nameGeneratorRefPanel](d.AsPanel())
	c.Equal(2, len(rows))
	c.Equal(4, len(rows[0].Children()), "a row holds the drag handle, the remove button, the popup and the edit button")
	c.False(rows[0].editButton.Enabled(), "a generator no library holds cannot be edited")
	c.True(rows[1].editButton.Enabled(), "a known generator can be")
	c.Equal("Edit this name generator", tooltipText(rows[1].editButton.Tooltip))

	rows[0].popup.Select("Human First - Male")
	c.True(rows[0].editButton.Enabled(), "choosing a known generator enables the button")
	rows[0].popup.Select("Missing Names")
	c.False(rows[0].editButton.Enabled(), "choosing the unknown one again disables it")
	rows[1].popup.Select("Human First - Male")
	c.True(rows[1].editButton.Enabled(), "a row that only offers known generators stays enabled")
}

// TestNameGeneratorsAddAndRemove verifies that an empty list shows no rows container, that adding a generator defaults
// to the first available one and builds a row with a popup for it, that removing the last generator removes the rows
// container again, and that both are undoable.
func TestNameGeneratorsAddAndRemove(t *testing.T) {
	c := check.New(t)
	a := gurps.NewAncestry()
	d := newTestAncestryEditorDockable(a, "Human First - Male", "Human Last")
	panel := nameGeneratorsPanelFor(t, d, a.CommonOptions)
	c.Nil(panel.rows, "an empty list has no rows container")
	c.Equal(1, len(panel.Children()), "only the header is shown for an empty list")

	panel.addGenerator()
	c.Equal([]string{"Human First - Male"}, d.model.CommonOptions.NameGenerators, "the first choice is the default")
	panel = nameGeneratorsPanelFor(t, d, d.model.CommonOptions)
	c.NotNil(panel.rows, "the rows container appears with the first generator")
	c.Equal(1, len(panel.rows.Children()))
	popup := widgetFor[*Popup[string]](t, d, nameGeneratorKey(d.model.CommonOptions, 0))
	c.Equal("Human First - Male", selectedGenerator(popup))
	c.True(d.Modified())

	panel.addGenerator()
	c.Equal([]string{"Human First - Male", "Human First - Male"}, d.model.CommonOptions.NameGenerators,
		"generators may repeat")
	panel = nameGeneratorsPanelFor(t, d, d.model.CommonOptions)
	panel.removeGenerator(0)
	c.Equal([]string{"Human First - Male"}, d.model.CommonOptions.NameGenerators)
	panel = nameGeneratorsPanelFor(t, d, d.model.CommonOptions)
	panel.removeGenerator(0)
	c.Equal(0, len(d.model.CommonOptions.NameGenerators), "the last generator may be removed")
	panel = nameGeneratorsPanelFor(t, d, d.model.CommonOptions)
	c.Nil(panel.rows, "the rows container is gone again")
	panel.removeGenerator(0)
	c.Equal(4, undoDepth(d), "removing from an empty list does nothing")

	d.undoMgr.Undo()
	c.Equal([]string{"Human First - Male"}, d.model.CommonOptions.NameGenerators, "undo restores the removed one")
}

// undoDepth returns the number of edits that can be undone.
func undoDepth(d *ancestryEditorDockable) int {
	depth := 0
	for d.undoMgr.CanUndo() {
		d.undoMgr.Undo()
		depth++
	}
	for range depth {
		d.undoMgr.Redo()
	}
	return depth
}

// TestNameGeneratorsAddWithoutChoices verifies that with no generators available at all, adding one still works and
// leaves an empty entry to be filled in later.
func TestNameGeneratorsAddWithoutChoices(t *testing.T) {
	c := check.New(t)
	d := newTestAncestryEditorDockable(gurps.NewAncestry())
	nameGeneratorsPanelFor(t, d, d.model.CommonOptions).addGenerator()
	c.Equal([]string{""}, d.model.CommonOptions.NameGenerators)
	c.Equal(1, len(panelsOfType[*nameGeneratorRefPanel](d.AsPanel())), "the empty entry still gets a row")
}

// TestNameGeneratorsPerGenderListBindsToGender verifies that a gender's name generator list edits that gender's options
// rather than the common ones.
func TestNameGeneratorsPerGenderListBindsToGender(t *testing.T) {
	c := check.New(t)
	a := gurps.NewAncestry()
	d := newTestAncestryEditorDockable(a, "Human First - Female")
	nameGeneratorsPanelFor(t, d, a.GenderOptions[1].Value).addGenerator()
	c.Equal([]string{"Human First - Female"}, d.model.GenderOptions[1].Value.NameGenerators)
	c.Equal(0, len(d.model.CommonOptions.NameGenerators), "the common options are untouched")
	c.Equal(0, len(d.model.GenderOptions[0].Value.NameGenerators), "the other gender is untouched")
}
