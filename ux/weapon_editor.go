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
	"fmt"

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
	displayEditor[*gurps.Weapon, *gurps.Weapon](owner, w, w.Type.SVG(), help, nil, initWeaponEditor, preApply)
}

func initWeaponEditor(e *editor[*gurps.Weapon, *gurps.Weapon], content *unison.Panel) func() {
	addLabelAndStringField(content, i18n.Text("Usage"), "", &e.editorData.Usage)
	addNotesLabelAndField(content, &e.editorData.UsageNotes)
	addLabelAndDecimalField(content, nil, "", i18n.Text("Minimum ST"), "", &e.editorData.StrengthParts.Minimum, 0, fxp.Max)
	if e.editorData.Type == gurps.RangedWeaponType {
		content.AddChild(unison.NewPanel())
		addCheckBox(content, i18n.Text("Has bipod"), &e.editorData.StrengthParts.Bipod)
		content.AddChild(unison.NewPanel())
		addCheckBox(content, i18n.Text("Mounted"), &e.editorData.StrengthParts.Mounted)
		content.AddChild(unison.NewPanel())
		addCheckBox(content, i18n.Text("Uses a musket rest"), &e.editorData.StrengthParts.MusketRest)
	}
	content.AddChild(unison.NewPanel())
	addCheckBox(content, i18n.Text("Two-handed"), &e.editorData.StrengthParts.TwoHanded)
	content.AddChild(unison.NewPanel())
	addCheckBox(content, i18n.Text("Two-handed and unready after attack"), &e.editorData.StrengthParts.TwoHandedUnready)
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
		addLabelAndDecimalField(content, nil, "", i18n.Text("Minimum Reach"), "", &e.editorData.MinReach, 0, fxp.Max)
		addLabelAndDecimalField(content, nil, "", i18n.Text("Maximum Reach"), "", &e.editorData.MaxReach, 0, fxp.Max)
		content.AddChild(unison.NewPanel())
		addCheckBox(content, i18n.Text("Close Combat"), &e.editorData.CloseCombat)
		content.AddChild(unison.NewPanel())
		addCheckBox(content, i18n.Text("Reach Change Requires Ready"), &e.editorData.ReachChangeRequiresReady)
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
		addRateOfFireBlock(content, &e.editorData.RateOfFireMode1, 1)
		addRateOfFireBlock(content, &e.editorData.RateOfFireMode2, 2)
		addLabelAndDecimalField(content, nil, "", i18n.Text("Half-Damage Range"), "", &e.editorData.HalfDamageRange, 0, fxp.Max)
		addLabelAndDecimalField(content, nil, "", i18n.Text("Minimum Range"), "", &e.editorData.MinRange, 0, fxp.Max)
		addLabelAndDecimalField(content, nil, "", i18n.Text("Maximum Range"), "", &e.editorData.MaxRange, 0, fxp.Max)
		content.AddChild(unison.NewPanel())
		addCheckBox(content, i18n.Text("Muscle Powered"), &e.editorData.MusclePowered)
		content.AddChild(unison.NewPanel())
		addCheckBox(content, i18n.Text("Range in Miles"), &e.editorData.RangeInMiles)
		addLabelAndDecimalField(content, nil, "", i18n.Text("Shot Recoil"), "", &e.editorData.ShotRecoil, 0, fxp.Max)
		addLabelAndDecimalField(content, nil, "", i18n.Text("Slug Recoil"), "", &e.editorData.SlugRecoil, 0, fxp.Max)
		addLabelAndDecimalField(content, nil, "", i18n.Text("Shots"), "", &e.editorData.NonChamberShots, 0, fxp.Max)
		addLabelAndDecimalField(content, nil, "", i18n.Text("Chamber Shots"), "", &e.editorData.ChamberShots, 0, fxp.Max)
		addLabelAndDecimalField(content, nil, "", i18n.Text("Shot Duration"), "", &e.editorData.ShotDuration, 0, fxp.Max)
		addLabelAndDecimalField(content, nil, "", i18n.Text("Reload Time"), "", &e.editorData.ChamberShots, 0, fxp.Max)
		content.AddChild(unison.NewPanel())
		addCheckBox(content, i18n.Text("Reload Time is Per Shot"), &e.editorData.ReloadTimeIsPerShot)
		content.AddChild(unison.NewPanel())
		addCheckBox(content, i18n.Text("Thrown Weapon"), &e.editorData.Thrown)
		addLabelAndDecimalField(content, nil, "", i18n.Text("Normal Bulk"), "", &e.editorData.NormalBulk, -fxp.Max, 0)
		giant := i18n.Text("Giant Bulk")
		wrapper := addFlowWrapper(content, giant, 2)
		addDecimalField(wrapper, nil, "", giant, "", &e.editorData.GiantBulk, -fxp.Max, 0)
		wrapper.AddChild(NewFieldTrailingLabel(i18n.Text("(only needed if different from normal bulk)"), true))
		content.AddChild(unison.NewPanel())
		addCheckBox(content, i18n.Text("Retracting Stock"), &e.editorData.RetractingStock)
	default:
	}
	content.AddChild(newDefaultsPanel(e.editorData.Entity(), &e.editorData.Defaults))
	return nil
}

func addRateOfFireBlock(content *unison.Panel, rof *gurps.RateOfFire, modeNum int) {
	wrapper := addFlowWrapper(content, fmt.Sprintf(i18n.Text("Rate of Fire Mode %d"), modeNum), 3)
	text := i18n.Text("Shots Per Attack")
	addDecimalField(wrapper, nil, "", text, "", &rof.ShotsPerAttack, 0, fxp.Max)
	label1 := NewFieldTrailingLabel(text, false)
	wrapper.AddChild(label1)
	addCheckBox(wrapper, i18n.Text("Fully Automatic Only"), &rof.FullAutoOnly)

	wrapper = addFlowWrapper(content, "", 3)
	text = i18n.Text("Secondary Projectiles")
	addDecimalField(wrapper, nil, "", text, "", &rof.SecondaryProjectiles, 0, fxp.Max)
	label2 := NewFieldTrailingLabel(text, false)
	wrapper.AddChild(label2)
	addCheckBox(wrapper, i18n.Text("High-cyclic Controlled Bursts"), &rof.HighCyclicControlledBursts)

	_, pref1, _ := label1.Sizes(unison.Size{})
	_, pref2, _ := label2.Sizes(unison.Size{})
	pref1 = pref1.Max(pref2)
	label1.LayoutData().(*unison.FlexLayoutData).SizeHint = pref1
	label2.LayoutData().(*unison.FlexLayoutData).SizeHint = pref1
}

func preApply(w *gurps.Weapon) {
	w.Reconcile()
}
