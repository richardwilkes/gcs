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

	"github.com/richardwilkes/toolbox/v2/check"
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
