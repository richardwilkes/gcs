/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package ux

import (
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// EditWeapon displays the editor for a weapon.
func EditWeapon(owner Rebuildable, w *gurps.Weapon) {
	displayEditor[*gurps.Weapon, *gurps.Weapon](owner, w, w.Type.SVG(), initWeaponEditor)
}

func initWeaponEditor(e *editor[*gurps.Weapon, *gurps.Weapon], content *unison.Panel) func() {
	addLabelAndStringField(content, i18n.Text("Usage"), "", &e.editorData.Usage)
	addNotesLabelAndField(content, &e.editorData.UsageNotes)
	addLabelAndStringField(content, i18n.Text("Minimum ST"), "", &e.editorData.MinimumStrength)
	addLabelAndPopup(content, i18n.Text("Base Damage"), "", gurps.AllStrengthDamage, &e.editorData.Damage.StrengthType)
	addLabelAndNullableDice(content, i18n.Text("Damage Modifier"), "", &e.editorData.Damage.Base)
	addLabelAndDecimalField(content, nil, "", i18n.Text("Damage Modifier Per Die"), "", &e.editorData.Damage.ModifierPerDie,
		fxp.Min, fxp.Max)
	addLabelAndDecimalField(content, nil, "", i18n.Text("Armor Divisor"), "", &e.editorData.Damage.ArmorDivisor, 0, fxp.Max)
	addLabelAndStringField(content, i18n.Text("Damage WeaponType"), "", &e.editorData.Damage.Type)
	addLabelAndNullableDice(content, i18n.Text("Fragmentation Base Damage"), "", &e.editorData.Damage.Fragmentation)
	addLabelAndDecimalField(content, nil, "", i18n.Text("Fragmentation Armor Divisor"), "",
		&e.editorData.Damage.FragmentationArmorDivisor, 0, fxp.Max)
	addLabelAndStringField(content, i18n.Text("Fragmentation WeaponType"), "", &e.editorData.Damage.FragmentationType)
	switch e.editorData.Type {
	case gurps.MeleeWeaponType:
		addLabelAndStringField(content, i18n.Text("Reach"), "", &e.editorData.Reach)
		addLabelAndStringField(content, i18n.Text("Parry Modifier"), "", &e.editorData.Parry)
		addLabelAndStringField(content, i18n.Text("Block Modifier"), "", &e.editorData.Block)
	case gurps.RangedWeaponType:
		addLabelAndStringField(content, i18n.Text("Accuracy"), "", &e.editorData.Accuracy)
		addLabelAndStringField(content, i18n.Text("Rate of Fire"), "", &e.editorData.RateOfFire)
		addLabelAndStringField(content, i18n.Text("Range"), "", &e.editorData.Range)
		addLabelAndStringField(content, i18n.Text("Recoil"), "", &e.editorData.Recoil)
		addLabelAndStringField(content, i18n.Text("Shots"), "", &e.editorData.Shots)
		addLabelAndStringField(content, i18n.Text("Bulk"), "", &e.editorData.Bulk)
	}
	content.AddChild(newDefaultsPanel(e.editorData.Entity(), &e.editorData.Defaults))
	return nil
}
