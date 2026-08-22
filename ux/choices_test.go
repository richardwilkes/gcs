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

// newChoices builds the "Choices" row an item editor shows for a picker, returning the two
// popups and the qualifier field it is made of.
func newChoices(pickerType picker.Type, compare criteria.NumericComparison) (typePopup *unison.PopupMenu[picker.Type], comparisonPopup *unison.PopupMenu[string], field unison.Paneler) {
	trait := gurps.NewTrait(nil, nil, true)
	trait.TemplatePicker.Type = pickerType
	trait.TemplatePicker.Qualifier.Compare = compare
	trait.TemplatePicker.Qualifier.Qualifier = fxp.One
	return addChoices(&editor[*gurps.Trait, *gurps.TraitData]{target: trait}, unison.NewPanel(), false)
}

// TestChoicesOpeningState verifies that a freshly opened editor blanks the picker's comparison popup
// and qualifier field under exactly the conditions its own selection callback uses. A picker set to Not Applicable is
// zero, so anything entered for it is dropped when the item is saved; leaving the comparison popup live let the user
// make a selection that set the comparison, un-blanked the qualifier and marked the item modified, all for edits that
// could never be kept. In the other direction, a saved picker whose comparison takes no qualifier opened with the
// qualifier editable, even though every later pass through the callback blanks it.
func TestChoicesOpeningState(t *testing.T) {
	c := check.New(t)

	// The default for every new trait, skill and spell container. Nothing about it can be saved, so nothing about it
	// may be edited.
	_, comparison, field := newChoices(picker.NotApplicable, criteria.AnyNumber)
	c.False(comparison.Enabled(), "a picker that isn't in use must not offer a comparison")
	c.False(field.AsPanel().Enabled(), "a picker that isn't in use must not offer a qualifier")

	// In use, but comparing against anything, which needs no number to compare with.
	_, comparison, field = newChoices(picker.Count, criteria.AnyNumber)
	c.True(comparison.Enabled(), "a picker in use must offer a comparison")
	c.False(field.AsPanel().Enabled(), "a comparison that takes no qualifier must not offer one")

	// In use with a comparison that needs a number, so everything is editable.
	_, comparison, field = newChoices(picker.Points, criteria.AtLeastNumber)
	c.True(comparison.Enabled(), "a picker in use must offer a comparison")
	c.True(field.AsPanel().Enabled(), "a comparison that takes a qualifier must offer one")
}

// TestChoicesFollowTheTypeSelection verifies that choosing a picker type updates the comparison popup
// and qualifier field, and that returning to Not Applicable blanks them both again.
func TestChoicesFollowTheTypeSelection(t *testing.T) {
	c := check.New(t)
	typePopup, comparison, field := newChoices(picker.NotApplicable, criteria.EqualsNumber)
	c.False(comparison.Enabled(), "a picker that isn't in use must not offer a comparison")
	c.False(field.AsPanel().Enabled(), "a picker that isn't in use must not offer a qualifier")

	typePopup.Select(picker.Count)
	c.True(comparison.Enabled(), "putting the picker to use must offer a comparison")
	c.True(field.AsPanel().Enabled(), "putting the picker to use must offer a qualifier")

	typePopup.Select(picker.NotApplicable)
	c.False(comparison.Enabled(), "taking the picker out of use must withdraw the comparison")
	c.False(field.AsPanel().Enabled(), "taking the picker out of use must withdraw the qualifier")
}
