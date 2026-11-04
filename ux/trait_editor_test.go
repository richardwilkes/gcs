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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestTraitEditorShowsTheRangeOfAnOpenChoice verifies that a trait editor opened outside a sheet counts a mandatory
// modifier choice yet to be made in its Point Cost as the list does, as a range once its options cost different
// amounts, and as a single cost once the trait is marked preconfigured with its pick made, the editor's own pending
// state counting throughout.
func TestTraitEditorShowsTheRangeOfAnOpenChoice(t *testing.T) {
	c := check.New(t)
	trait := gurps.NewTrait(nil, nil, false)
	trait.BasePoints = fxp.FromInteger(10)
	choice := newTraitModifierChoiceFor(nil, true, []string{"Small", "Large"})
	choice.Children[0].CostAdj = "+5"
	choice.Children[1].CostAdj = "+5"
	trait.AddModifiers(choice)
	c.Equal(fxp.FromInteger(15), trait.AdjustedPoints(nil), "the list counts the choice")
	e, content := buildEditorContent(nil, trait, initTraitEditor)
	// The Point Cost field is the first of the editor's non-editable fields.
	pointCost := uxtest.PanelsOfType[*NonEditableField](content)[0]
	c.Equal("15", pointCost.String())
	options := e.editorData.Modifiers[0].Children
	options[1].CostAdj = "+10"
	DeepSync(e)
	c.Equal("15~20", pointCost.String())
	e.editorData.Preconfigured = true
	DeepSync(e)
	c.Equal("15~20", pointCost.String(), "a preconfigured trait still has a choice with no pick to make")
	options[1].SetDisabled(false)
	DeepSync(e)
	c.Equal("20", pointCost.String(), "a preconfigured trait takes the pick already made")
}

// TestTraitEditorLeavesOutInheritedModifiers verifies that the Point Cost of a trait with a modifier choice leaves out
// the modifiers of its container whether the choice is open or made.
func TestTraitEditorLeavesOutInheritedModifiers(t *testing.T) {
	c := check.New(t)
	parent := gurps.NewTrait(nil, nil, true)
	half := gurps.NewTraitModifier(nil, nil, false)
	half.CostAdj = "-50%"
	parent.AddModifiers(half)
	trait := gurps.NewTrait(nil, parent, false)
	trait.BasePoints = fxp.FromInteger(20)
	parent.Children = []*gurps.Trait{trait}
	choice := newTraitModifierChoiceFor(nil, true, []string{"+0%", "+100%"})
	for _, option := range choice.Children {
		option.CostAdj = option.Name
	}
	trait.AddModifiers(choice)
	e, content := buildEditorContent(nil, trait, initTraitEditor)
	pointCost := uxtest.PanelsOfType[*NonEditableField](content)[0]
	c.Equal("20~40", pointCost.String())
	e.editorData.Preconfigured = true
	e.editorData.Modifiers[0].Children[1].SetDisabled(false)
	DeepSync(e)
	c.Equal("40", pointCost.String())
	e.editorData.Modifiers[0].Children[1].SetDisabled(true)
	e.editorData.Modifiers[0].Children[0].SetDisabled(false)
	DeepSync(e)
	c.Equal("20", pointCost.String())
}
