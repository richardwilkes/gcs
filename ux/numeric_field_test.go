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
	"strconv"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
)

// newTestIntegerField returns an IntegerField backed by a local variable, along with a pointer to that variable, for
// use in headless tests.
func newTestIntegerField(initial, minValue, maxValue int) (field *IntegerField, value *int) {
	v := initial
	return NewIntegerField(nil, "", "Value",
		func() int { return v },
		func(newValue int) { v = newValue },
		minValue, maxValue, false, true), &v
}

// TestNumericFieldValidationPreservesTooltip verifies that validating a NumericField leaves any tooltip installed by
// the field's creator alone. Validation runs after every edit, so a validate() that unconditionally replaced the
// tooltip silently and permanently discarded explanatory tooltips — such as the Page Offset help on the Page Reference
// Mappings view — the first time the user typed in the field.
func TestNumericFieldValidationPreservesTooltip(t *testing.T) {
	c := check.New(t)

	t.Run("valid edits leave the creator's tooltip in place", func(_ *testing.T) {
		f, value := newTestIntegerField(0, -9999, 9999)
		tooltip := newWrappedTooltip("Page Offset explanation")
		f.Tooltip = tooltip
		f.SetText("3")
		c.Equal(3, *value, "the edit was applied")
		c.True(tooltip == f.Tooltip, "tooltip survives an edit")
		f.SetText("42")
		c.True(tooltip == f.Tooltip, "tooltip survives further edits")
	})

	t.Run("invalid input swaps in a validation tooltip, then restores", func(_ *testing.T) {
		f, _ := newTestIntegerField(5, 1, 10)
		tooltip := newWrappedTooltip("Page Offset explanation")
		f.Tooltip = tooltip

		f.SetText("not a number")
		c.NotNil(f.Tooltip, "an invalid value reports why via the tooltip")
		c.False(tooltip == f.Tooltip, "the validation tooltip replaces the creator's tooltip while invalid")

		f.SetText("20")
		c.NotNil(f.Tooltip, "an out-of-range value reports why via the tooltip")
		c.False(tooltip == f.Tooltip, "the validation tooltip is still in place while out of range")

		f.SetText("5")
		c.True(tooltip == f.Tooltip, "the creator's tooltip is restored once the value is valid again")
	})

	t.Run("no creator tooltip means no tooltip once valid", func(_ *testing.T) {
		f, _ := newTestIntegerField(5, 1, 10)
		c.Nil(f.Tooltip, "no tooltip to start with")

		f.SetText("20")
		c.NotNil(f.Tooltip, "an out-of-range value reports why via the tooltip")

		f.SetText("5")
		c.Nil(f.Tooltip, "the validation tooltip is removed once the value is valid again")
	})

	t.Run("a tooltip installed while invalid is not discarded", func(_ *testing.T) {
		f, _ := newTestIntegerField(5, 1, 10)
		f.SetText("20")
		c.NotNil(f.Tooltip, "an out-of-range value reports why via the tooltip")

		tooltip := newWrappedTooltip("Installed later")
		f.Tooltip = tooltip
		f.SetText("5")
		c.True(tooltip == f.Tooltip, "the tooltip installed while invalid is kept")
	})

	t.Run("a field created with an invalid value still keeps its tooltip", func(_ *testing.T) {
		// The initial value is out of range, so the field starts out invalid and the creator installs its tooltip on
		// top of the validation tooltip.
		f, _ := newTestIntegerField(20, 1, 10)
		c.NotNil(f.Tooltip, "the out-of-range initial value reports why via the tooltip")

		tooltip := newWrappedTooltip("Page Offset explanation")
		f.Tooltip = tooltip
		f.SetText("5")
		c.True(tooltip == f.Tooltip, "the creator's tooltip is in place once the value is valid")
	})

	t.Run("an exception value counts as valid and restores the tooltip", func(_ *testing.T) {
		value := 0
		f := NewNumericFieldWithException(nil, "", "Value", nil,
			func() int { return value },
			func(v int) { value = v },
			strconv.Itoa, strconv.Atoi, 1, 10, -1)
		tooltip := newWrappedTooltip("Page Offset explanation")
		f.Tooltip = tooltip

		f.SetText("20")
		c.False(tooltip == f.Tooltip, "the validation tooltip replaces the creator's tooltip while out of range")

		f.SetText("-1")
		c.True(tooltip == f.Tooltip, "the exception value is valid, so the creator's tooltip is restored")
	})
}

// TestFixedPointPrototypes verifies the values a fixed-point field is sized to fit, which the decimal, length and
// weight fields share: explicit bounds are widened to fill every decimal place of their integer part, while the
// unbounded extremes are stood in for by -1 and 1 rather than rendered, since a field sized to fxp.Min and fxp.Max
// would be enormous. The length and weight variants must agree exactly with the decimal one, since the conversions to
// and from fxp.Int inside the helper are the only thing that differs.
func TestFixedPointPrototypes(t *testing.T) {
	c := check.New(t)
	widest := fxp.Two - 1 // 1.9999
	c.Equal([]fxp.Int{fxp.FromInteger(-2) - 1, widest, fxp.FromInteger(13) - 1},
		fixedPointPrototypes(fxp.FromStringForced("-2.5"), fxp.FromStringForced("12.7")),
		"explicit bounds are floored and then filled out to the last decimal place")
	c.Equal([]fxp.Int{-1, widest, widest}, fixedPointPrototypes(fxp.Min, fxp.Max),
		"the unbounded extremes are replaced by -1 and 1 before being widened")
	c.Equal([]fxp.Int{fxp.FromInteger(6) - 1, widest, widest}, fixedPointPrototypes(fxp.FromInteger(5), fxp.Max),
		"a bound that is already integral is still filled out to the last decimal place")

	want := fixedPointPrototypes(fxp.Min, fxp.FromStringForced("500.25"))
	c.Equal(len(want), 3)
	gotLength := fixedPointPrototypes(fxp.Length(fxp.Min), fxp.Length(fxp.FromStringForced("500.25")))
	gotWeight := fixedPointPrototypes(fxp.Weight(fxp.Min), fxp.Weight(fxp.FromStringForced("500.25")))
	for i, v := range want {
		c.Equal(fxp.Length(v), gotLength[i], "length prototypes match the decimal ones")
		c.Equal(fxp.Weight(v), gotWeight[i], "weight prototypes match the decimal ones")
	}
}

// TestUnitsFieldAcceptsUnitsBeingTyped verifies that a units field lets a units suffix be typed after the number, one
// character at a time, rather than beeping and rejecting the keystroke when the partially typed suffix fails to parse.
// The other numeric fields validate each keystroke by parsing the text it would produce, which is exactly wrong here.
func TestUnitsFieldAcceptsUnitsBeingTyped(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	stored := fxp.Weight(fxp.Ten)
	f := NewWeightField(nil, "", "Weight", entity,
		func() fxp.Weight { return stored },
		func(v fxp.Weight) { stored = v },
		0, fxp.Weight(fxp.Max), true)
	f.gainedFocus()
	f.SetText("7 ")
	c.Equal("7 ", f.Text())
	for _, ch := range "kg" {
		c.True(f.RuneTypedCallback(ch), "typing %q into a units field is accepted", ch)
	}
	c.Equal("7 kg", f.Text(), "the suffix was typed into the field")
	f.lostFocus()
	c.Equal("7 kg", fxp.Kilogram.Format(stored), "the typed value and units were committed")
}

// TestDisabledFieldIsLeftAloneByTheFocus verifies that a field focused while disabled, which happens only for a screen
// reader, is not treated as being edited: Sync replaces its text and losing the focus does not reformat it. A field
// that could have been edited is still brought into range.
func TestDisabledFieldIsLeftAloneByTheFocus(t *testing.T) {
	c := check.New(t)

	f, value := newTestIntegerField(5, 0, 10)
	f.SetEnabled(false)
	f.gainedFocus()
	*value = 7
	f.Sync()
	c.Equal("7", f.Text(), "a disabled field holding the focus shows the value it is given")
	c.Equal(7, *value)
	*value = 50 // Out of the field's range, as a value from the model may be.
	f.Sync()
	c.Equal("50", f.Text(), "whatever the value is")
	f.lostFocus()
	c.Equal("50", f.Text(), "the focus leaving a field that was only read leaves its text alone")

	// A field enabled by the time the focus leaves it may have been typed into.
	f.gainedFocus()
	f.SetEnabled(true)
	f.lostFocus()
	c.Equal("10", f.Text(), "a field that could be typed into is brought into range")

	// So may one that gained the focus while enabled, even if disabled since.
	f, value = newTestIntegerField(5, 0, 10)
	f.gainedFocus()
	f.Field.SetText("50")
	f.SetEnabled(false)
	f.lostFocus()
	c.Equal("10", f.Text(), "a field that was being edited is brought into range")
	c.Equal(10, *value)
}

// TestNumericFieldRefusesValuesAboveItsMaximum verifies that a field told to refuse values above its maximum turns
// away typed and pasted text that would exceed it, rather than showing it as invalid, while one that isn't still
// accepts the text.
func TestNumericFieldRefusesValuesAboveItsMaximum(t *testing.T) {
	c := check.New(t)
	f, value := newTestIntegerField(0, 0, 11)
	f.refuseAboveMax()
	f.gainedFocus()
	f.SetText("1")
	f.SetSelectionToEnd()
	c.False(f.RuneTypedCallback('5'), "15 is above the maximum")
	c.Equal("1", f.Text())
	c.True(f.RuneTypedCallback('1'))
	c.Equal("11", f.Text())
	c.Equal(11, *value)
	c.True(f.wouldExceedMax("5"), "pasting at the end would make 115")
	f.SelectAll()
	c.False(f.wouldExceedMax("9"), "pasting over the selection would make 9")
	c.False(f.wouldExceedMax("x"), "text that isn't a number is left to validation")
	f.lostFocus()

	plain, _ := newTestIntegerField(1, 0, 11)
	plain.gainedFocus()
	plain.SetSelectionToEnd()
	c.True(plain.RuneTypedCallback('5'), "a field that doesn't refuse values above its maximum accepts the text")
	c.Equal("15", plain.Text())
	c.False(plain.wouldExceedMax("5"))
	plain.lostFocus()
}

// The stored weight used by the tests of rounding for display, along with its exact and two-decimal-place renderings.
// It carries more decimal places than the sheet is told to show, so that rounding for display is visible and any
// rounding that leaked back into the stored value would be, too.
const (
	exactWeightText   = "7.5127 lb"
	roundedWeightText = "7.51 lb"
)

// pageUndoRoot is an undo root whose MarkModified records its source and deep syncs unless the source is flagged
// SkipDeepSync, as the Description block's fields are. Sheet.MarkModified syncs the blocks even then, but the field
// edited keeps its text only because it has the focus, which without a window no field has, so this root syncs
// nothing and the tests assert what the fields do on their own.
type pageUndoRoot struct {
	unison.Panel
	mgr *unison.UndoManager
	// modifiedBy records, in order, each source handed to MarkModified.
	modifiedBy []unison.Paneler
}

func newPageUndoRoot() *pageUndoRoot {
	r := &pageUndoRoot{mgr: unison.NewUndoManager(100, func(err error) { panic(err) })}
	r.Self = r
	return r
}

func (r *pageUndoRoot) UndoManager() *unison.UndoManager { return r.mgr }

// MarkModified implements ModifiableRoot.
func (r *pageUndoRoot) MarkModified(src unison.Paneler) {
	r.modifiedBy = append(r.modifiedBy, src)
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
// display text, whether through SetText or by an assistive technology replacing the value or all of the text shown,
// applies the new value and records an undo whose "before" state holds the exact value. Recording the rounded text
// instead would make undo quietly round the character's weight. After the undo, the field holds the exact text and is
// focused for editing, and the rounded text returns once the focus is lost.
func TestNumericFieldSetTextWhileShowingDisplayText(t *testing.T) {
	for _, one := range []struct {
		name    string
		replace func(f *WeightField) bool
	}{
		{
			name: "SetText",
			replace: func(f *WeightField) bool {
				f.SetText("8 lb")
				return true
			},
		},
		{
			name: "assistive technology setting the value",
			replace: func(f *WeightField) bool {
				return f.PerformAccessibilityAction(accessibility.ActionRequest{
					Action: accessibility.SetValue,
					Value:  "8 lb",
				})
			},
		},
		{
			name: "assistive technology replacing all of the text shown",
			replace: func(f *WeightField) bool {
				return f.PerformAccessibilityAction(accessibility.ActionRequest{
					Action: accessibility.ReplaceText,
					End:    len([]rune(roundedWeightText)),
					Value:  "8 lb",
				})
			},
		},
	} {
		t.Run(one.name, func(t *testing.T) {
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

			c.True(one.replace(f), "the replacement must be carried out")
			c.Equal(exactWeightText, beforeText, `the undo "before" state must hold the exact text, not the rounded text`)
			c.Equal("8 lb", f.Format(*value), "the new value must be applied")
			c.True(root.mgr.CanUndo(), "replacing the text must record an undo edit")

			root.mgr.Undo()
			c.Equal(exactWeightText, f.Format(*value), "undo must restore every decimal place the value had")
			c.Equal(exactWeightText, f.Text(), "undo puts the exact text into the field, which it focuses for editing")
			f.lostFocus()
			c.Equal(roundedWeightText, f.Text(), "the display rendering returns once the focus is lost")
		})
	}
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
