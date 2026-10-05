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
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	uncheck "github.com/richardwilkes/unison/enums/check"
)

// modifierShowTitles are the titles of the checkboxes that say where a modifier is shown, in the order they appear: the
// first goes with the name fields, the other two with the notes.
var modifierShowTitles = []string{"Show in Title Notes", "Show in Owner's Notes", "Show in Weapon Usage"}

// checkModifierShowCheckBoxes checks that the content has the modifier's show checkboxes in order, that "Show in
// Owner's Notes" starts checked and clearing it hides the notes, and that the other two set their flags.
func checkModifierShowCheckBoxes(c check.Checker, content *unison.Panel, showInTitle, hideNotes, showOnWeapon *bool) {
	c.Helper()
	var titles []string
	for _, box := range panelsOfType[*CheckBox](content) {
		for _, want := range modifierShowTitles {
			if box.Text.String() == want {
				titles = append(titles, want)
			}
		}
	}
	c.Equal(modifierShowTitles, titles, "the show checkboxes appear once each, in order")

	owner := findCheckBoxTitled(content, i18n.Text("Show in Owner's Notes"))
	c.False(*hideNotes, "precondition: the notes start out shown")
	c.Equal(uncheck.On, owner.State, "the owner's notes box starts checked")
	clickCheckBox(owner, false)
	c.True(*hideNotes, "clearing the owner's notes box hides the notes")
	clickCheckBox(findCheckBoxTitled(content, i18n.Text("Show in Title Notes")), true)
	c.True(*showInTitle, "checking the title notes box shows the modifier in the title")
	clickCheckBox(findCheckBoxTitled(content, i18n.Text("Show in Weapon Usage")), true)
	c.True(*showOnWeapon, "checking the weapon usage box shows the notes in weapon usage")
}

func TestTraitModifierEditorShowCheckBoxes(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	e, content := buildEditorContent(sheet, gurps.NewTraitModifier(sheet.Entity(), nil, false),
		initTraitModifierEditor)
	checkModifierShowCheckBoxes(c, content, &e.editorData.ShowInTitle, &e.editorData.HideNotes,
		&e.editorData.ShowNotesOnWeapon)
}

func TestEquipmentModifierEditorShowCheckBoxes(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	e, content := buildEditorContent(sheet, gurps.NewEquipmentModifier(sheet.Entity(), nil, false),
		initEquipmentModifierEditor)
	checkModifierShowCheckBoxes(c, content, &e.editorData.ShowInTitle, &e.editorData.HideNotes,
		&e.editorData.ShowNotesOnWeapon)
}
