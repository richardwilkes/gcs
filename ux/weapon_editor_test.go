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

// TestWeaponEditorFragmentationFieldsFollowTheDice verifies that the fragmentation armor divisor and type fields, whose
// values are dropped unless the weapon has fragmentation dice of its own, can only be used while it has some.
func TestWeaponEditorFragmentationFieldsFollowTheDice(t *testing.T) {
	for _, melee := range []bool{true, false} {
		c := check.New(t)
		sheet := newTestSheetForTemplate(t)
		trait := gurps.NewTrait(sheet.Entity(), nil, false)
		weapon := gurps.NewWeapon(trait, melee)
		trait.Weapons = []*gurps.Weapon{weapon}
		var we weaponEditor
		e, content := buildEditorContent(sheet, weapon, we.initWeaponEditor)
		var dice, fragType *StringField
		for _, field := range panelsOfType[*StringField](content) {
			switch field.undoTitle {
			case "Fragmentation Base Damage":
				dice = field
			case "Fragmentation Type":
				fragType = field
			}
		}
		var divisor *DecimalField
		for _, field := range panelsOfType[*DecimalField](content) {
			if field.undoTitle == "Fragmentation Armor Divisor" {
				divisor = field
			}
		}
		if dice == nil || fragType == nil || divisor == nil {
			t.Fatal("the weapon editor is missing one of its fragmentation fields")
		}
		c.False(divisor.Enabled(), "without fragmentation dice, there is nothing for an armor divisor to apply to")
		c.False(fragType.Enabled(), "nor a type")

		dice.SetText("2d")
		c.Equal("2d", e.editorData.Damage.Fragmentation, "precondition: the field edits the fragmentation dice")
		c.True(divisor.Enabled(), "fragmentation dice can have an armor divisor")
		c.True(fragType.Enabled(), "and a type")
		divisor.SetText("2")
		fragType.SetText("cut")
		c.Equal("cut", e.editorData.Damage.FragmentationType, "precondition: the type field edits the type")
		c.True(e.isModified())

		dice.SetText(" ")
		c.Equal("", e.editorData.Damage.Fragmentation, "blank fragmentation dice are none at all")
		c.False(divisor.Enabled(), "taking the dice away takes away the use of the armor divisor")
		c.False(fragType.Enabled(), "and the type")
		c.False(e.isModified(), "and leaves the weapon as it started")
	}
}
