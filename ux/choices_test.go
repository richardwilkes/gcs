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

// newChoices builds the "Choices" row an item editor shows for the container, returning the two popups and the
// qualifier field it is made of, all nil when the editor shows no such row.
func newChoices(trait *gurps.Trait) (typePopup *unison.PopupMenu[picker.Type], comparisonPopup *unison.PopupMenu[string], field unison.Paneler) {
	e := &editor[*gurps.Trait, *gurps.TraitEditData]{target: trait, editorData: &gurps.TraitEditData{}}
	e.editorData.CopyFrom(trait)
	return addChoices(e, unison.NewPanel())
}

// newChoiceContainer returns a trait choice container whose picker is set as given.
func newChoiceContainer(pickerType picker.Type, compare criteria.NumericComparison) *gurps.Trait {
	trait := gurps.NewTraitChoiceContainer(nil, nil)
	trait.TemplatePicker.Type = pickerType
	trait.TemplatePicker.Qualifier.Compare = compare
	trait.TemplatePicker.Qualifier.Qualifier = fxp.One
	return trait
}

// TestChoicesOnlyForChoiceContainers verifies that only a choice container's editor offers choices, and that it can't
// take them out of use. A container becomes a choice container only by being created as one, and stays one.
func TestChoicesOnlyForChoiceContainers(t *testing.T) {
	c := check.New(t)
	typePopup, _, _ := newChoices(gurps.NewTrait(nil, nil, true))
	c.Nil(typePopup, "a plain container must not offer choices")

	typePopup, _, _ = newChoices(newChoiceContainer(picker.Count, criteria.EqualsNumber))
	c.NotNil(typePopup, "a choice container must offer choices")
	c.Equal(-1, typePopup.IndexOfItem(picker.NotApplicable), "a choice container must not offer to stop being one")
}

// TestChoicesOpeningState verifies that a freshly opened editor blanks the picker's qualifier field exactly when its
// comparison takes no qualifier.
func TestChoicesOpeningState(t *testing.T) {
	c := check.New(t)
	_, comparison, field := newChoices(newChoiceContainer(picker.Count, criteria.AnyNumber))
	c.True(comparison.Enabled(), "a choice container must offer a comparison")
	c.False(field.AsPanel().Enabled(), "a comparison that takes no qualifier must not offer one")

	_, comparison, field = newChoices(newChoiceContainer(picker.Points, criteria.AtLeastNumber))
	c.True(comparison.Enabled(), "a choice container must offer a comparison")
	c.True(field.AsPanel().Enabled(), "a comparison that takes a qualifier must offer one")

	_, comparison, field = newChoices(newChoiceContainer(picker.Count, criteria.AnyNumber))
	comparison.SelectIndex(int(criteria.EqualsNumber))
	c.True(field.AsPanel().Enabled(), "choosing a comparison that takes a qualifier must offer one")
}
