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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stdmg"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wpn"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/check"
)

// EditWeapon displays the editor for a weapon.
func EditWeapon(owner Rebuildable, w *gurps.Weapon) {
	displayEditor[*gurps.Weapon, *gurps.Weapon](owner, w, w.Type.SVG(),
		"md:Help/Interface/"+txt.FirstToUpper(strings.Split(w.Type.Key(), "_")[0])+" Weapon Usage", nil,
		initWeaponEditor, preApply)
}

func initWeaponEditor(e *editor[*gurps.Weapon, *gurps.Weapon], content *unison.Panel) func() {
	w := e.editorData
	addUsageBlock(w, content)
	addStrengthBlock(w, content)
	addDamageBlock(w, content)
	switch w.Type {
	case wpn.Melee:
		addReachBlock(w, content)
		addParryBlock(w, content)
		addBlockBlock(w, content)
	case wpn.Ranged:
		addAccuracyBlock(w, content)
		addRateOfFireBlock(w, content)
		addRangeBlock(w, content)
		addRecoilBlock(w, content)
		addShotsBlock(w, content)
		addBulkBlock(w, content)
	}
	content.AddChild(newDefaultsPanel(w.Entity(), &w.Defaults))
	return nil
}

func addUsageBlock(w *gurps.Weapon, content *unison.Panel) {
	addLabelAndStringField(content, i18n.Text("Usage"), "", &w.Usage)
	addNotesLabelAndField(content, &w.UsageNotes)
}

func addStrengthBlock(w *gurps.Weapon, content *unison.Panel) {
	strength := &w.Strength
	addLabelAndDecimalField(content, nil, "", i18n.Text("Minimum ST"), "", &strength.Min, 0, fxp.Max)
	if w.Type == wpn.Ranged {
		content.AddChild(unison.NewPanel())
		addCheckBox(content, i18n.Text("Has bipod"), &strength.Bipod)
		content.AddChild(unison.NewPanel())
		addCheckBox(content, i18n.Text("Mounted"), &strength.Mounted)
		content.AddChild(unison.NewPanel())
		addCheckBox(content, i18n.Text("Uses a musket rest"), &strength.MusketRest)
	}
	content.AddChild(unison.NewPanel())
	addCheckBox(content, i18n.Text("Two-handed"), &strength.TwoHanded)
	content.AddChild(unison.NewPanel())
	addCheckBox(content, i18n.Text("Two-handed and unready after attack"), &strength.TwoHandedUnready)
}

func addDamageBlock(w *gurps.Weapon, content *unison.Panel) {
	damage := &w.Damage
	addLabelAndPopup(content, i18n.Text("Base Damage"), "", stdmg.Options, &damage.StrengthType)
	addLabelAndNullableDice(content, i18n.Text("Damage Modifier"), "", &damage.Base)
	addLabelAndDecimalField(content, nil, "", i18n.Text("Damage Modifier Per Die"), "", &damage.ModifierPerDie,
		fxp.Min, fxp.Max)
	addLabelAndDecimalField(content, nil, "", i18n.Text("Armor Divisor"), "", &damage.ArmorDivisor, 0, fxp.Max)
	addLabelAndStringField(content, i18n.Text("Damage Type"), "", &damage.Type)
	addLabelAndNullableDice(content, i18n.Text("Fragmentation Base Damage"), "", &damage.Fragmentation)
	addLabelAndDecimalField(content, nil, "", i18n.Text("Fragmentation Armor Divisor"), "",
		&damage.FragmentationArmorDivisor, 0, fxp.Max)
	addLabelAndStringField(content, i18n.Text("Fragmentation Type"), "", &damage.FragmentationType)
}

func addReachBlock(w *gurps.Weapon, content *unison.Panel) {
	reach := &w.Reach
	addLabelAndDecimalField(content, nil, "", i18n.Text("Minimum Reach"), "", &reach.Min, 0, fxp.Max)
	addLabelAndDecimalField(content, nil, "", i18n.Text("Maximum Reach"), "", &reach.Max, 0, fxp.Max)
	content.AddChild(unison.NewPanel())
	addCheckBox(content, i18n.Text("Close Combat"), &reach.CloseCombat)
	content.AddChild(unison.NewPanel())
	addCheckBox(content, i18n.Text("Reach Change Requires Ready"), &reach.ChangeRequiresReady)
}

func addParryBlock(w *gurps.Weapon, content *unison.Panel) {
	parry := &w.Parry
	parryCheckBox := addInvertedCheckBox(content, i18n.Text("Parry Modifier"), &parry.No)
	parryCheckBox.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.End,
		VAlign: align.Middle,
	})
	parryField := addDecimalFieldWithSign(content, nil, "", i18n.Text("Parry Modifier"), "", &parry.Modifier, -fxp.Max, fxp.Max)
	parryCheckBox.OnSet = func() {
		parryField.SetEnabled(parryCheckBox.State == check.On)
	}
	parryCheckBox.OnSet()
	content.AddChild(unison.NewPanel())
	addCheckBox(content, i18n.Text("Fencing"), &parry.Fencing)
	content.AddChild(unison.NewPanel())
	addCheckBox(content, i18n.Text("Unbalanced"), &parry.Unbalanced)
}

func addBlockBlock(w *gurps.Weapon, content *unison.Panel) {
	block := &w.Block
	blockCheckBox := addInvertedCheckBox(content, i18n.Text("Block Modifier"), &block.No)
	blockCheckBox.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.End,
		VAlign: align.Middle,
	})
	blockField := addDecimalFieldWithSign(content, nil, "", i18n.Text("Block Modifier"), "", &block.Modifier, -fxp.Max, fxp.Max)
	blockCheckBox.OnSet = func() {
		blockField.SetEnabled(blockCheckBox.State == check.On)
	}
	blockCheckBox.OnSet()
}

func addAccuracyBlock(w *gurps.Weapon, content *unison.Panel) {
	accuracy := &w.Accuracy
	addLabelAndDecimalField(content, nil, "", i18n.Text("Weapon Accuracy"), "", &accuracy.Base, 0, fxp.Max)
	addLabelAndDecimalField(content, nil, "", i18n.Text("Scope Accuracy"), "", &accuracy.Scope, 0, fxp.Max)
}

func addRateOfFireBlock(w *gurps.Weapon, content *unison.Panel) {
	rof := &w.RateOfFire
	content.AddChild(unison.NewPanel())
	addCheckBox(content, i18n.Text("Jet"), &rof.Jet)
	addRateOfFireModeBlock(content, &rof.Mode1, 1)
	addRateOfFireModeBlock(content, &rof.Mode2, 2)
}

func addRateOfFireModeBlock(content *unison.Panel, mode *gurps.WeaponRoFMode, modeNum int) {
	wrapper := addFlowWrapper(content, fmt.Sprintf(i18n.Text("Rate of Fire Mode %d"), modeNum), 3)
	text := i18n.Text("Shots Per Attack")
	addDecimalField(wrapper, nil, "", text, "", &mode.ShotsPerAttack, 0, fxp.Max)
	label1 := NewFieldTrailingLabel(text, false)
	wrapper.AddChild(label1)
	addCheckBox(wrapper, i18n.Text("Fully Automatic Only"), &mode.FullAutoOnly)

	wrapper = addFlowWrapper(content, "", 3)
	text = i18n.Text("Secondary Projectiles")
	addDecimalField(wrapper, nil, "", text, "", &mode.SecondaryProjectiles, 0, fxp.Max)
	label2 := NewFieldTrailingLabel(text, false)
	wrapper.AddChild(label2)
	addCheckBox(wrapper, i18n.Text("High-cyclic Controlled Bursts"), &mode.HighCyclicControlledBursts)

	_, pref1, _ := label1.Sizes(unison.Size{})
	_, pref2, _ := label2.Sizes(unison.Size{})
	pref1 = pref1.Max(pref2)
	label1.LayoutData().(*unison.FlexLayoutData).SizeHint = pref1
	label2.LayoutData().(*unison.FlexLayoutData).SizeHint = pref1
}

func addRangeBlock(w *gurps.Weapon, content *unison.Panel) {
	weaponRange := &w.Range
	addLabelAndDecimalField(content, nil, "", i18n.Text("Half-Damage Range"), "", &weaponRange.HalfDamage, 0, fxp.Max)
	addLabelAndDecimalField(content, nil, "", i18n.Text("Minimum Range"), "", &weaponRange.Min, 0, fxp.Max)
	addLabelAndDecimalField(content, nil, "", i18n.Text("Maximum Range"), "", &weaponRange.Max, 0, fxp.Max)
	content.AddChild(unison.NewPanel())
	addCheckBox(content, i18n.Text("Muscle Powered"), &weaponRange.MusclePowered)
	content.AddChild(unison.NewPanel())
	addCheckBox(content, i18n.Text("Range in Miles"), &weaponRange.InMiles)
}

func addRecoilBlock(w *gurps.Weapon, content *unison.Panel) {
	recoil := &w.Recoil
	addLabelAndDecimalField(content, nil, "", i18n.Text("Shot Recoil"), "", &recoil.Shot, 0, fxp.Max)
	addLabelAndDecimalField(content, nil, "", i18n.Text("Slug Recoil"), "", &recoil.Slug, 0, fxp.Max)
}

func addShotsBlock(w *gurps.Weapon, content *unison.Panel) {
	shots := &w.Shots
	addLabelAndDecimalField(content, nil, "", i18n.Text("Shots"), "", &shots.Count, 0, fxp.Max)
	addLabelAndDecimalField(content, nil, "", i18n.Text("Chamber Shots"), "", &shots.InChamber, 0, fxp.Max)
	addLabelAndDecimalField(content, nil, "", i18n.Text("Shot Duration"), "", &shots.Duration, 0, fxp.Max)
	addLabelAndDecimalField(content, nil, "", i18n.Text("Reload Time"), "", &shots.InChamber, 0, fxp.Max)
	content.AddChild(unison.NewPanel())
	addCheckBox(content, i18n.Text("Reload Time is Per Shot"), &shots.ReloadTimeIsPerShot)
	content.AddChild(unison.NewPanel())
	addCheckBox(content, i18n.Text("Thrown Weapon"), &shots.Thrown)
}

func addBulkBlock(w *gurps.Weapon, content *unison.Panel) {
	bulk := &w.Bulk
	addLabelAndDecimalField(content, nil, "", i18n.Text("Normal Bulk"), "", &bulk.Normal, -fxp.Max, 0)
	giant := i18n.Text("Giant Bulk")
	wrapper := addFlowWrapper(content, giant, 2)
	addDecimalField(wrapper, nil, "", giant, "", &bulk.Giant, -fxp.Max, 0)
	wrapper.AddChild(NewFieldTrailingLabel(i18n.Text("(only needed if different from normal bulk)"), true))
	content.AddChild(unison.NewPanel())
	addCheckBox(content, i18n.Text("Retracting Stock"), &bulk.RetractingStock)
}

func preApply(w *gurps.Weapon) {
	w.Validate()
}
