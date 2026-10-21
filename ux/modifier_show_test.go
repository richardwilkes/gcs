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
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	uncheck "github.com/richardwilkes/unison/enums/check"
)

// modifierShowTitles are the titles of the checkboxes that say where a modifier's notes are shown, in the order they
// appear.
var modifierShowTitles = []string{"Show notes in owner's notes", "Show notes in attack usage"}

// checkModifierShowCheckBoxes checks that the content has the modifier's show checkboxes in order, that "Show notes in
// owner's notes" starts checked and clearing it hides the notes until it is checked again, and that the other one sets
// its flag.
func checkModifierShowCheckBoxes(c check.Checker, content *unison.Panel, hideNotes, showOnWeapon *bool) {
	c.Helper()
	var titles []string
	for _, box := range uxtest.PanelsOfType[*CheckBox](content) {
		for _, want := range modifierShowTitles {
			if box.Text.String() == want {
				titles = append(titles, want)
			}
		}
	}
	c.Equal(modifierShowTitles, titles, "the show checkboxes appear once each, in order")

	c.False(*hideNotes, "precondition: the notes start out shown")
	c.False(*showOnWeapon, "precondition: the notes start out off attack usage")
	owner := findCheckBoxTitled(content, modifierShowTitles[0])
	attack := findCheckBoxTitled(content, modifierShowTitles[1])
	if owner == nil || attack == nil {
		return
	}
	c.Equal(uncheck.On, owner.State, "the owner's notes box starts checked")
	c.Equal(uncheck.Off, attack.State, "the attack usage box starts clear")
	clickCheckBox(owner, false)
	c.True(*hideNotes, "clearing the owner's notes box hides the notes")
	clickCheckBox(owner, true)
	c.False(*hideNotes, "checking the owner's notes box again shows the notes")
	clickCheckBox(attack, true)
	c.True(*showOnWeapon, "checking the attack usage box shows the notes in attack usage")
}

// TestModifierEditorsShowCheckBoxes checks the show checkboxes of the trait and equipment modifier editors.
func TestModifierEditorsShowCheckBoxes(t *testing.T) {
	c := check.New(t)
	trait, content := buildEditorContent(nil, gurps.NewTraitModifier(nil, nil, false), initTraitModifierEditor)
	checkModifierShowCheckBoxes(c, content, &trait.editorData.HideNotes, &trait.editorData.ShowNotesOnWeapon)
	eqp, content := buildEditorContent(nil, gurps.NewEquipmentModifier(nil, nil, false), initEquipmentModifierEditor)
	checkModifierShowCheckBoxes(c, content, &eqp.editorData.HideNotes, &eqp.editorData.ShowNotesOnWeapon)
}
