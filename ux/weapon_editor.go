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
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/check"
)

// EditWeapon displays the editor for a weapon.
func EditWeapon(owner Rebuildable, w *gurps.Weapon) {
	var help string
	switch w.Type {
	case gurps.MeleeWeaponType:
		help = "md:Help/Interface/Melee Weapon Usage"
	case gurps.RangedWeaponType:
		help = "md:Help/Interface/Ranged Weapon Usage"
	default:
	}
	displayEditor[*gurps.Weapon, *gurps.Weapon](owner, w, w.Type.SVG(), help, nil, initWeaponEditor)
}

func initWeaponEditor(e *editor[*gurps.Weapon, *gurps.Weapon], content *unison.Panel) func() {
	addLabelAndStringField(content, i18n.Text("Usage"), "", &e.editorData.Usage)
	addNotesLabelAndField(content, &e.editorData.UsageNotes)
	addLabelAndDecimalField(content, nil, "", i18n.Text("Minimum ST"), "", &e.editorData.MinST, 0, fxp.Max)
	if e.editorData.Type == gurps.RangedWeaponType {
		content.AddChild(unison.NewPanel())
		addCheckBox(content, i18n.Text("Has bipod"), &e.editorData.Bipod)
		content.AddChild(unison.NewPanel())
		addCheckBox(content, i18n.Text("Mounted"), &e.editorData.Mounted)
		content.AddChild(unison.NewPanel())
		addCheckBox(content, i18n.Text("Uses a musket rest"), &e.editorData.MusketRest)
	}
	content.AddChild(unison.NewPanel())
	addCheckBox(content, i18n.Text("Two-handed"), &e.editorData.TwoHanded)
	content.AddChild(unison.NewPanel())
	addCheckBox(content, i18n.Text("Two-handed and unready after attack"), &e.editorData.UnreadyAfterAttack)
	addLabelAndPopup(content, i18n.Text("Base Damage"), "", gurps.AllStrengthDamage, &e.editorData.Damage.StrengthType)
	addLabelAndNullableDice(content, i18n.Text("Damage Modifier"), "", &e.editorData.Damage.Base)
	addLabelAndDecimalField(content, nil, "", i18n.Text("Damage Modifier Per Die"), "", &e.editorData.Damage.ModifierPerDie,
		fxp.Min, fxp.Max)
	addLabelAndDecimalField(content, nil, "", i18n.Text("Armor Divisor"), "", &e.editorData.Damage.ArmorDivisor, 0, fxp.Max)
	addLabelAndStringField(content, i18n.Text("Damage Type"), "", &e.editorData.Damage.Type)
	addLabelAndNullableDice(content, i18n.Text("Fragmentation Base Damage"), "", &e.editorData.Damage.Fragmentation)
	addLabelAndDecimalField(content, nil, "", i18n.Text("Fragmentation Armor Divisor"), "",
		&e.editorData.Damage.FragmentationArmorDivisor, 0, fxp.Max)
	addLabelAndStringField(content, i18n.Text("Fragmentation Type"), "", &e.editorData.Damage.FragmentationType)
	switch e.editorData.Type {
	case gurps.MeleeWeaponType:
		addLabelAndStringField(content, i18n.Text("Reach"), "", &e.editorData.Reach)
		parryCheckBox := addCheckBox(content, i18n.Text("Parry Modifier"), &e.editorData.CanParry)
		parryCheckBox.SetLayoutData(&unison.FlexLayoutData{
			HAlign: align.End,
			VAlign: align.Middle,
		})
		parryField := addDecimalFieldWithSign(content, nil, "", i18n.Text("Parry Modifier"), "", &e.editorData.ParryModifier, -fxp.Max, fxp.Max)
		parryCheckBox.OnSet = func() {
			parryField.SetEnabled(parryCheckBox.State == check.On)
		}
		parryCheckBox.OnSet()
		content.AddChild(unison.NewPanel())
		addCheckBox(content, i18n.Text("Fencing"), &e.editorData.Fencing)
		content.AddChild(unison.NewPanel())
		addCheckBox(content, i18n.Text("Unbalanced"), &e.editorData.Unbalanced)
		blockCheckBox := addCheckBox(content, i18n.Text("Block Modifier"), &e.editorData.CanBlock)
		blockCheckBox.SetLayoutData(&unison.FlexLayoutData{
			HAlign: align.End,
			VAlign: align.Middle,
		})
		blockField := addDecimalFieldWithSign(content, nil, "", i18n.Text("Block Modifier"), "", &e.editorData.BlockModifier, -fxp.Max, fxp.Max)
		blockCheckBox.OnSet = func() {
			blockField.SetEnabled(blockCheckBox.State == check.On)
		}
		blockCheckBox.OnSet()
	case gurps.RangedWeaponType:
		addLabelAndDecimalField(content, nil, "", i18n.Text("Weapon Accuracy"), "", &e.editorData.WeaponAcc, 0, fxp.Max)
		addLabelAndDecimalField(content, nil, "", i18n.Text("Scope Accuracy"), "", &e.editorData.ScopeAcc, 0, fxp.Max)
		content.AddChild(unison.NewPanel())
		addCheckBox(content, i18n.Text("Jet"), &e.editorData.Jet)
		addLabelAndStringField(content, i18n.Text("Rate of Fire"), "", &e.editorData.RateOfFire)
		addLabelAndStringField(content, i18n.Text("Range"), "", &e.editorData.Range)
		addLabelAndDecimalField(content, nil, "", i18n.Text("Shot Recoil"), "", &e.editorData.ShotRecoil, 0, fxp.Max)
		addLabelAndDecimalField(content, nil, "", i18n.Text("Slug Recoil"), "", &e.editorData.SlugRecoil, 0, fxp.Max)
		addLabelAndStringField(content, i18n.Text("Shots"), "", &e.editorData.Shots)
		addLabelAndDecimalField(content, nil, "", i18n.Text("Normal Bulk"), "", &e.editorData.NormalBulk, -fxp.Max, 0)
		giant := i18n.Text("Giant Bulk")
		wrapper := addFlowWrapper(content, giant, 2)
		addDecimalField(wrapper, nil, "", giant, "", &e.editorData.GiantBulk, -fxp.Max, 0)
		label := NewFieldTrailingLabel(i18n.Text("(only needed if different from normal bulk)"))
		fd := unison.DefaultLabelTheme.Font.Descriptor()
		fd.Size *= 0.8
		label.Font = fd.Font()
		wrapper.AddChild(label)
		content.AddChild(unison.NewPanel())
		addCheckBox(content, i18n.Text("Retracting Stock"), &e.editorData.RetractingStock)
	default:
	}
	content.AddChild(newDefaultsPanel(e.editorData.Entity(), &e.editorData.Defaults))
	return nil
}
