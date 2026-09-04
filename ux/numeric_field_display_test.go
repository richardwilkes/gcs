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

// The stored weight used throughout these tests, along with its exact and two-decimal-place renderings. It carries more
// decimal places than the sheet is told to show, so that rounding for display is visible and any rounding that leaked
// back into the stored value would be, too.
const (
	exactWeightText   = "7.5127 lb"
	roundedWeightText = "7.51 lb"
)

// pageUndoRoot is an undo root whose MarkModified behaves as Sheet.MarkModified does for the Description block's
// fields: those are flagged SkipDeepSync, so no deep sync follows an edit made through them. The tests use it so that
// they assert what happens on a real sheet rather than what a deep sync after every edit would paper over.
type pageUndoRoot struct {
	unison.Panel
	mgr *unison.UndoManager
}

func newPageUndoRoot() *pageUndoRoot {
	r := &pageUndoRoot{mgr: unison.NewUndoManager(100, func(err error) { panic(err) })}
	r.Self = r
	return r
}

func (r *pageUndoRoot) UndoManager() *unison.UndoManager { return r.mgr }

// MarkModified implements ModifiableRoot.
func (r *pageUndoRoot) MarkModified(src unison.Paneler) {
	if _, skip := src.AsPanel().ClientData()[SkipDeepSync]; !skip {
		DeepSync(r)
	}
}

// newDisplayFormatWeightField returns a weight field backed by a local variable, holding exactWeightText, whose entity
// rounds body weights to two decimal places for display. The field is built exactly as the sheet's Description block
// builds its weight field, through NewWeightPageField, so that the production wiring of the DisplayFormat -- including
// the extra Sync() there, needed because the field was synced with the exact text before the format was installed --
// is what the tests exercise.
func newDisplayFormatWeightField(t *testing.T) (field *WeightField, value *fxp.Weight) {
	t.Helper()
	c := check.New(t)
	entity := gurps.NewEntity()
	entity.SheetSettings.BodyWeightFormat = fxp.NumberFormat{Places: fxp.TwoPlaces}
	stored, err := fxp.WeightFromString(exactWeightText, entity.SheetSettings.DefaultWeightUnits)
	c.NoError(err, "the test weight must be parsable")
	field = NewWeightPageField(nil, "", "Weight", entity,
		func() fxp.Weight { return stored },
		func(v fxp.Weight) { stored = v },
		0, fxp.Weight(fxp.Max), true)
	field.ClientData()[SkipDeepSync] = true
	return field, &stored
}

// TestNumericFieldSyncShowsDisplayText verifies that a page field with a DisplayFormat shows the rounded text from the
// moment it is built, and that syncing it shows that text without writing the rounding back into the value. Sync()
// previously went through setWithoutUndo, which parses whatever text it is handed and stores the result, so showing
// rounded text there would have silently truncated the character's weight.
func TestNumericFieldSyncShowsDisplayText(t *testing.T) {
	c := check.New(t)
	f, value := newDisplayFormatWeightField(t)

	c.Equal(roundedWeightText, f.Text(), "a freshly built page field shows the display rendering of the value")
	c.Equal(exactWeightText, f.Format(*value), "the stored value keeps every decimal place it had")
	c.NotEqual(f.Format(*value), f.Text(), "the test is only meaningful if the two renderings differ")

	f.Sync()
	c.Equal(roundedWeightText, f.Text(), "syncing again leaves the display rendering in place")
	c.Equal(exactWeightText, f.Format(*value), "syncing must not round the stored value")
}

// TestNumericFieldFocusRestoresExactText verifies that the exact value is put back into the field the moment it is
// focused, so that the user always edits what is actually stored, and that the rounded text returns once the focus is
// lost. The field cannot be truly focused headlessly, so the focus callbacks are invoked directly.
func TestNumericFieldFocusRestoresExactText(t *testing.T) {
	c := check.New(t)
	f, value := newDisplayFormatWeightField(t)

	f.gainedFocus()
	c.Equal(exactWeightText, f.Text(), "focusing the field shows the exact value")
	c.Equal(exactWeightText, f.Format(*value), "focusing the field must not change the stored value")

	f.lostFocus()
	c.Equal(roundedWeightText, f.Text(), "losing the focus returns to the display rendering")
	c.Equal(exactWeightText, f.Format(*value), "losing the focus must not round the stored value")
}

// TestNumericFieldSyncWhileFocusedKeepsEdit verifies that a Sync() arriving while the field has the focus leaves the
// text the user is working on alone rather than replacing it with the rounded rendering. The field decides this from
// its own focus callbacks, not from Panel.Focused(), because the latter is false whenever the window is inactive even
// though the field still holds the focus: a sync in that state used to swap in the rounded text, and the next keystroke
// then parsed it and committed the rounded value to the character with no undo.
func TestNumericFieldSyncWhileFocusedKeepsEdit(t *testing.T) {
	c := check.New(t)
	f, value := newDisplayFormatWeightField(t)

	f.gainedFocus()
	f.SetText("9")
	c.Equal("9 lb", f.Format(*value), "the in-progress edit is applied as it is typed")

	f.Sync()
	c.Equal("9", f.Text(), "syncing while focused must not replace the text being edited")
	c.Equal("9 lb", f.Format(*value), "syncing while focused must not alter the value")

	f.lostFocus()
	c.Equal("9 lb", f.Text(), "losing the focus shows the display rendering of the edited value")
}

// TestNumericFieldSetTextWhileShowingDisplayText verifies that replacing the text of a field that is showing rounded
// display text applies the new value and records an undo whose "before" state holds the exact value. Recording the
// rounded text instead would make undo quietly round the character's weight. After the undo, the field holds the
// exact text and is focused for editing, and the rounded text returns once the focus is lost.
func TestNumericFieldSetTextWhileShowingDisplayText(t *testing.T) {
	c := check.New(t)
	f, value := newDisplayFormatWeightField(t)
	root := newPageUndoRoot()
	root.AddChild(f)

	var beforeText string
	modified := f.ModifiedCallback
	f.ModifiedCallback = func(before, after *unison.FieldState) {
		beforeText = before.Text
		modified(before, after)
	}

	f.SetText("8 lb")
	c.Equal(exactWeightText, beforeText, `the undo "before" state must hold the exact text, not the rounded text`)
	c.Equal("8 lb", f.Format(*value), "the typed value must be applied")
	c.True(root.mgr.CanUndo(), "replacing the text must record an undo edit")

	root.mgr.Undo()
	c.Equal(exactWeightText, f.Format(*value), "undo must restore every decimal place the value had")
	c.Equal(exactWeightText, f.Text(), "undo puts the exact text into the field, which it focuses for editing")
	f.lostFocus()
	c.Equal(roundedWeightText, f.Text(), "the display rendering returns once the focus is lost")
}

// TestNumericFieldSetTextWithIdenticalText covers the two ways SetText can be handed text the field already holds.
// Text identical to the exact value behind the display rendering is not a change, so nothing may be recorded. Text
// identical to the display rendering itself is a change -- the caller asked for that value -- and the undo it records
// must still hold the exact value as its "before" state.
func TestNumericFieldSetTextWithIdenticalText(t *testing.T) {
	c := check.New(t)
	f, value := newDisplayFormatWeightField(t)
	root := newPageUndoRoot()
	root.AddChild(f)

	f.SetText(exactWeightText)
	c.Equal(exactWeightText, f.Text(), "SetText restores the exact text before applying the new text")
	c.Equal(exactWeightText, f.Format(*value), "the value is unchanged")
	c.False(root.mgr.CanUndo(), "handing the field the text it already holds must not record an undo edit")

	f.Sync()
	c.Equal(roundedWeightText, f.Text(), "syncing the unfocused field shows the display rendering again")

	f.SetText(roundedWeightText)
	c.Equal("7.51 lb", f.Format(*value), "the rounded text, when actually asked for, becomes the value")
	c.True(root.mgr.CanUndo(), "that is a change, so it must record an undo edit")
	root.mgr.Undo()
	c.Equal(exactWeightText, f.Format(*value), "undo must restore every decimal place the value had")
}

// TestNumericFieldSetTextAfterModelChangedRecordsUndo covers the shape of the Description block's height and weight
// randomizers: they assign the new value to the profile and then hand the field its new text. The field is showing
// rounded display text at that point, and restoring the exact text has to restore the text for the value the field was
// showing rather than for the value the model now holds -- otherwise the field would already hold the very text it is
// about to be given, the assignment would be a no-op, and the randomization would go unrecorded and be impossible to
// undo.
func TestNumericFieldSetTextAfterModelChangedRecordsUndo(t *testing.T) {
	c := check.New(t)
	f, value := newDisplayFormatWeightField(t)
	root := newPageUndoRoot()
	root.AddChild(f)

	randomized := fxp.WeightFromStringForced("200 lb", fxp.Pound)
	*value = randomized
	SetTextAndMarkModified(f, f.Format(randomized))

	c.Equal("200 lb", f.Format(*value), "the randomized value must be in place")
	c.True(root.mgr.CanUndo(), "randomizing must record an undo edit")

	root.mgr.Undo()
	c.Equal(exactWeightText, f.Format(*value), "undo must restore the prior weight, with every decimal place")
	c.Equal(exactWeightText, f.Text(), "undo puts the exact text into the field, which it focuses for editing")
	f.lostFocus()
	c.Equal(roundedWeightText, f.Text(), "the display rendering returns once the focus is lost")
}

// newEquipmentTable returns a table holding a single piece of carried equipment whose weight has more decimal places
// than the sheet is told to show, along with the node for that row and the index of the weight column. forPage says
// whether the table is one of a sheet's page lists or a library-style table.
func newEquipmentTable(t *testing.T, forPage bool) (node *Node[*gurps.Equipment], weightCol int) {
	t.Helper()
	c := check.New(t)
	entity := gurps.NewEntity()
	entity.SheetSettings.EquipmentWeightFormat = fxp.NumberFormat{Places: fxp.TwoPlaces}
	e := gurps.NewEquipment(entity, nil, false)
	e.Name = "Heavy Thing"
	e.BaseWeight = exactWeightText
	entity.CarriedEquipment = []*gurps.Equipment{e}

	provider := NewEquipmentProvider(entity, true, forPage)
	table := unison.NewTable(provider)
	provider.SetTable(table)
	ids := provider.ColumnIDs()
	table.Columns = make([]unison.ColumnInfo, len(ids))
	for i, id := range ids {
		table.Columns[i].ID = id
	}
	rows := provider.RootRows()
	c.Equal(1, len(rows), "the table must hold the single piece of equipment")
	weightCol = slices.Index(ids, gurps.EquipmentWeightColumn)
	c.True(weightCol != -1, "the table must have a weight column")
	return rows[0], weightCol
}

// TestEquipmentPageCellDataRoundsOnlyForDisplay verifies that a page table's weight cells show the rounded weight, with
// the exact weight available as a tooltip, while the text the rows are sorted by stays exact. Sorting on the rounded
// text would tie rows whose weights actually differ, leaving them in an arbitrary order.
func TestEquipmentPageCellDataRoundsOnlyForDisplay(t *testing.T) {
	c := check.New(t)
	node, weightCol := newEquipmentTable(t, true)

	displayed := node.cellData(weightCol, true)
	c.Equal(roundedWeightText, displayed.Primary, "a page cell shows the rounded weight")
	c.Equal(exactWeightText, displayed.Tooltip, "the exact weight is available as a tooltip")

	exact := node.cellData(weightCol, false)
	c.Equal(exactWeightText, exact.Primary, "a cell that isn't for display holds the exact weight")
	c.Equal("", exact.Tooltip, "there is nothing for a tooltip to add when the exact weight is shown")

	c.Equal(exactWeightText, node.CellDataForSort(weightCol), "sorting must order by the exact weight")
}

// TestEquipmentRowsMatchDisplayedAndExactText verifies that searching finds a page row by the rounded weight the user
// can see as well as by the exact weight behind it, for both of the matchers the searches go through, while a row off
// the page -- where the two renderings are the same -- matches only the exact text. The rounded text is chosen so that
// it is not simply a prefix of the exact text.
func TestEquipmentRowsMatchDisplayedAndExactText(t *testing.T) {
	c := check.New(t)
	node, _ := newEquipmentTable(t, true)
	c.True(node.Match(roundedWeightText), "the sheet search must find the weight as displayed")
	c.True(node.Match("7.5127"), "the sheet search must find the exact weight")
	c.True(node.PartialMatchExceptTag(roundedWeightText), "the filter must find the weight as displayed")
	c.True(node.PartialMatchExceptTag("7.5127"), "the filter must find the exact weight")
	c.False(node.Match("7.52"), "a value that is neither rendering must not match")

	node, _ = newEquipmentTable(t, false)
	c.False(node.Match(roundedWeightText), "off the page, there is no rounded rendering to match")
	c.True(node.Match("7.5127"), "off the page, the exact weight still matches")
	c.False(node.PartialMatchExceptTag(roundedWeightText))
	c.True(node.PartialMatchExceptTag("7.5127"))
}

// TestEquipmentPageHeaderSyncRefreshesTotalsTooltip verifies that syncing a page's equipment list refreshes the
// header's tooltip along with the totals in its title. The tooltip is where the exact totals are offered once the
// display formats round them, and the formats can change after the headers were built -- which is exactly what
// changing them in the sheet settings does -- so refreshing only the title used to leave the tooltip missing, or stale.
func TestEquipmentPageHeaderSyncRefreshesTotalsTooltip(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	e := gurps.NewEquipment(entity, nil, false)
	e.Name = "Heavy Thing"
	e.BaseWeight = exactWeightText
	entity.CarriedEquipment = []*gurps.Equipment{e}
	provider := NewEquipmentProvider(entity, true, true)
	table := unison.NewTable(provider)
	provider.SetTable(table)
	ids := provider.ColumnIDs()
	table.Columns = make([]unison.ColumnInfo, len(ids))
	for i, id := range ids {
		table.Columns[i].ID = id
	}
	headers := provider.Headers()
	header, ok := headers[slices.Index(ids, gurps.EquipmentDescriptionColumn)].(*PageTableColumnHeader[*gurps.Equipment])
	c.True(ok, "the page's description column must have a page header")
	c.Equal("CARRIED EQUIPMENT (7.5127 LB; $0)", header.Text.String())
	c.Equal("", header.TooltipText(), "with no display preference, there are no exact totals to offer")
	c.Nil(header.Tooltip)

	entity.SheetSettings.EquipmentWeightFormat = fxp.NumberFormat{Places: fxp.TwoPlaces}
	provider.SyncHeader(headers)
	c.Equal("CARRIED EQUIPMENT (7.51 LB; $0)", header.Text.String())
	c.Equal("7.5127 lb; $0", header.TooltipText(), "syncing must offer the exact totals once the formats round them")
	c.NotNil(header.Tooltip)

	entity.SheetSettings.EquipmentWeightFormat = fxp.NumberFormat{}
	provider.SyncHeader(headers)
	c.Equal("CARRIED EQUIPMENT (7.5127 LB; $0)", header.Text.String())
	c.Equal("", header.TooltipText(), "syncing must remove the tooltip once there is nothing to add")
	c.Nil(header.Tooltip)
}
