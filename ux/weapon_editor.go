// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stdmg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/check"
)

type weaponEditor struct {
	jetCheckBox           *CheckBox
	panelsControlledByJet []unison.Paneler
}

// EditWeapon displays the editor for a weapon.
func EditWeapon(owner Rebuildable, w *gurps.Weapon) {
	var we weaponEditor
	var part string
	if w.IsMelee() {
		part = "Melee"
	} else {
		part = "Ranged"
	}
	displayEditor(owner, w, gurps.WeaponSVG(w.IsMelee()), "md:Help/Interface/"+part+" Weapon Usage", nil,
		we.initWeaponEditor, we.preApply)
}

func (we *weaponEditor) initWeaponEditor(e *editor[*gurps.Weapon, *gurps.Weapon], content *unison.Panel) func() {
	w := e.editorData
	empty := unison.NewLabel()
	empty.SetTitle(" ")
	content.AddChild(empty)
	addCheckBox(content, i18n.Text("Hide"), &w.Hide)
	we.addUsageBlock(w, content)
	if w.IsMelee() {
		we.addParryBlock(w, content)
		we.addBlockBlock(w, content)
		we.addDamageBlock(w, content)
		we.addReachBlock(w, content)
	} else {
		we.addAccuracyBlock(w, content)
		we.addDamageBlock(w, content)
		we.addRangeBlock(w, content)
		we.addRateOfFireBlock(w, content)
		we.addShotsBlock(w, content)
		we.addBulkBlock(w, content)
		we.addRecoilBlock(w, content)
	}
	we.addStrengthBlock(w, content)
	content.AddChild(newDefaultsPanel(gurps.EntityFromNode(w), &w.Defaults))
	if w.IsRanged() {
		we.jetCheckBox.OnSet = func() {
			state := we.jetCheckBox.State == check.Off
			for _, p := range we.panelsControlledByJet {
				p.AsPanel().SetEnabled(state)
			}
		}
		we.jetCheckBox.OnSet()
	}
	return nil
}

func (we *weaponEditor) addUsageBlock(w *gurps.Weapon, content *unison.Panel) {
	addLabelAndStringField(content, i18n.Text("Usage"), "", &w.Usage)
	addLabelAndMultiLineStringField(content, i18n.Text("Notes"), "", &w.UsageNotes)
}

func (we *weaponEditor) addParryBlock(w *gurps.Weapon, content *unison.Panel) {
	parry := &w.Parry
	on := addCheckBox(content, i18n.Text("Parry"), &parry.CanParry)
	on.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.End,
		VAlign: align.Middle,
	})
	wrapper := unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  3,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		VAlign:   align.Middle,
	})
	content.AddChild(wrapper)
	text := i18n.Text("Parry Modifier")
	field := addDecimalFieldWithSign(wrapper, nil, "", text, text, &parry.Modifier, -fxp.Thousand, fxp.Thousand)
	fencing := addCheckBox(wrapper, i18n.Text("Fencing"), &parry.Fencing)
	unbalanced := addCheckBox(wrapper, i18n.Text("Unbalanced"), &parry.Unbalanced)
	on.OnSet = func() {
		field.SetEnabled(on.State == check.On)
		fencing.SetEnabled(on.State == check.On)
		unbalanced.SetEnabled(on.State == check.On)
	}
	on.OnSet()
}

func (we *weaponEditor) addBlockBlock(w *gurps.Weapon, content *unison.Panel) {
	block := &w.Block
	on := addCheckBox(content, i18n.Text("Block"), &block.CanBlock)
	on.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.End,
		VAlign: align.Middle,
	})
	text := i18n.Text("Block Modifier")
	blockField := addDecimalFieldWithSign(content, nil, "", text, text, &block.Modifier, -fxp.Thousand, fxp.Thousand)
	on.OnSet = func() {
		blockField.SetEnabled(on.State == check.On)
	}
	on.OnSet()
}

func (we *weaponEditor) addDamageBlock(w *gurps.Weapon, content *unison.Panel) {
	damage := &w.Damage
	wrapper := addFillWrapper(content, i18n.Text("Damage"), 5)
	addPopup(wrapper, filteredDamageOptions(), &damage.StrengthType)
	addCheckBox(wrapper, i18n.Text("(leveled)"), &damage.Leveled)
	text := i18n.Text("Damage Modifier")
	addNullableDice(wrapper, text, text, &damage.Base, true)
	text = i18n.Text("Damage Modifier Per Die")
	addDecimalFieldWithSign(wrapper, nil, "", text, text, &damage.ModifierPerDie, -fxp.BillionMinusOne, fxp.BillionMinusOne)
	wrapper.AddChild(NewFieldTrailingLabel(i18n.Text("per die"), false))

	wrapper = addFillWrapper(content, "", 4)
	armorDivisor := i18n.Text("Armor Divisor")
	addLabelAndDecimalField(wrapper, nil, "", armorDivisor, armorDivisor, &damage.ArmorDivisor, 0, fxp.MillionMinusOne)
	typeText := i18n.Text("Type")
	text = i18n.Text("Damage Type")
	wrapper.AddChild(NewFieldTrailingLabel(typeText, false))
	addStringField(wrapper, text, text, &damage.Type)

	wrapper = addFillWrapper(content, "", 2)
	addLabelAndDecimalField(wrapper, nil, "", i18n.Text("Multiply ST used for sw or thr damage calculation by"), "",
		&damage.StrengthMultiplier, fxp.Tenth, fxp.BillionMinusOne)

	wrapper = addFillWrapper(content, i18n.Text("Fragmentation"), 5)
	text = i18n.Text("Fragmentation Base Damage")
	addNullableDice(wrapper, text, text, &damage.Fragmentation, false)
	wrapper.AddChild(NewFieldTrailingLabel(armorDivisor, false))
	text = i18n.Text("Fragmentation Armor Divisor")
	addDecimalField(wrapper, nil, "", text, text, &damage.FragmentationArmorDivisor, 0, fxp.MillionMinusOne)
	wrapper.AddChild(NewFieldTrailingLabel(typeText, false))
	text = i18n.Text("Fragmentation Type")
	addStringField(wrapper, text, text, &damage.FragmentationType)
}

func filteredDamageOptions() []stdmg.Option {
	options := stdmg.Options
	filtered := make([]stdmg.Option, 0, len(options))
	for _, opt := range options {
		if opt != stdmg.OldLeveledThrust && opt != stdmg.OldLeveledSwing {
			filtered = append(filtered, opt)
		}
	}
	return filtered
}

func (we *weaponEditor) addReachBlock(w *gurps.Weapon, content *unison.Panel) {
	reach := &w.Reach
	wrapper := addFlowWrapper(content, i18n.Text("Reach"), 3)
	text := i18n.Text("Maximum Reach")
	addDecimalField(wrapper, nil, "", text, text, &reach.Max, 0, fxp.Thousand)
	wrapper.AddChild(NewFieldInteriorLeadingLabel(i18n.Text("Minimum"), false))
	text = i18n.Text("Minimum Reach")
	addDecimalField(wrapper, nil, "", text, text, &reach.Min, 0, fxp.Thousand)
	wrapper = addFlowWrapper(content, "", 2)
	addCheckBox(wrapper, i18n.Text("Close Combat"), &reach.CloseCombat)
	addCheckBox(wrapper, i18n.Text("Reach Change Requires Ready"), &reach.ChangeRequiresReady)
}

func (we *weaponEditor) addAccuracyBlock(w *gurps.Weapon, content *unison.Panel) {
	accuracy := &w.Accuracy
	wrapper := addFlowWrapper(content, i18n.Text("Accuracy"), 4)
	text := i18n.Text("Weapon Accuracy")
	base := addDecimalField(wrapper, nil, "", text, text, &accuracy.Base, 0, fxp.MillionMinusOne)
	wrapper.AddChild(NewFieldInteriorLeadingLabel(i18n.Text("Scope"), false))
	text = i18n.Text("Scope Accuracy")
	scope := addDecimalField(wrapper, nil, "", text, text, &accuracy.Scope, 0, fxp.MillionMinusOne)
	we.jetCheckBox = addCheckBox(wrapper, i18n.Text("Jet"), &accuracy.Jet)
	we.panelsControlledByJet = append(we.panelsControlledByJet, base, scope)
}

func (we *weaponEditor) addRangeBlock(w *gurps.Weapon, content *unison.Panel) {
	weaponRange := &w.Range
	wrapper := addFlowWrapper(content, i18n.Text("Range"), 5)
	text := i18n.Text("Maximum Range")
	addDecimalField(wrapper, nil, "", text, text, &weaponRange.Max, 0, fxp.BillionMinusOne)
	wrapper.AddChild(NewFieldInteriorLeadingLabel(i18n.Text("½ Damage"), false))
	text = i18n.Text("½ Damage Range")
	addDecimalField(wrapper, nil, "", text, text, &weaponRange.HalfDamage, 0, fxp.BillionMinusOne)
	text = i18n.Text("Minimum Range")
	wrapper.AddChild(NewFieldInteriorLeadingLabel(i18n.Text("Minimum"), false))
	addDecimalField(wrapper, nil, "", text, text, &weaponRange.Min, 0, fxp.BillionMinusOne)
	wrapper = addFlowWrapper(content, "", 2)
	addCheckBox(wrapper, i18n.Text("Muscle-powered"), &weaponRange.MusclePowered)
	addCheckBox(wrapper, i18n.Text("Ranges are in miles"), &weaponRange.InMiles)
}

func (we *weaponEditor) addRateOfFireBlock(w *gurps.Weapon, content *unison.Panel) {
	rof := &w.RateOfFire
	we.addRateOfFireModeBlock(content, &rof.Mode1, 1)
	we.addRateOfFireModeBlock(content, &rof.Mode2, 2)
}

func (we *weaponEditor) addRateOfFireModeBlock(content *unison.Panel, mode *gurps.WeaponRoFMode, modeNum int) {
	var wrapper *unison.Panel
	if modeNum == 1 {
		wrapper = addFlowWrapper(content, i18n.Text("Rate of Fire"), 5)
	} else {
		wrapper = addFlowWrapper(content, "", 5)
	}
	wrapper.AddChild(NewFieldLeadingLabel(fmt.Sprintf(i18n.Text("Mode %d"), modeNum), false))
	text := i18n.Text("Shots Per Attack")
	spa := addDecimalField(wrapper, nil, "", text, text, &mode.ShotsPerAttack, 0, fxp.MillionMinusOne)
	wrapper.AddChild(NewFieldTrailingLabel(i18n.Text("per attack with"), false))
	text = i18n.Text("Secondary Projectiles")
	sp := addDecimalField(wrapper, nil, "", text, text, &mode.SecondaryProjectiles, 0, fxp.MillionMinusOne)
	wrapper.AddChild(NewFieldTrailingLabel(i18n.Text("secondary projectiles"), false))
	wrapper = addFlowWrapper(content, "", 2)
	auto := addCheckBox(wrapper, i18n.Text("Fully Automatic Only"), &mode.FullAutoOnly)
	hccb := addCheckBox(wrapper, i18n.Text("High-cyclic Controlled Bursts"), &mode.HighCyclicControlledBursts)
	we.panelsControlledByJet = append(we.panelsControlledByJet, spa, auto, sp, hccb)
}

func (we *weaponEditor) addShotsBlock(w *gurps.Weapon, content *unison.Panel) {
	shots := &w.Shots
	text := i18n.Text("Shots")
	wrapper := addFlowWrapper(content, text, 5)
	addDecimalField(wrapper, nil, "", text, text, &shots.Count, 0, fxp.MillionMinusOne)
	text = i18n.Text("In Chamber")
	wrapper.AddChild(NewFieldInteriorLeadingLabel(text, false))
	addDecimalField(wrapper, nil, "", text, text, &shots.InChamber, 0, fxp.Thousand)
	wrapper.AddChild(NewFieldInteriorLeadingLabel(i18n.Text("Duration"), false))
	text = i18n.Text("Shot Duration")
	addDecimalField(wrapper, nil, "", text, text, &shots.Duration, 0, fxp.Thousand)
	wrapper = addFlowWrapper(content, "", 4)
	text = i18n.Text("Reload Time")
	addLabelAndDecimalField(wrapper, nil, "", text, text, &shots.ReloadTime, 0, fxp.Thousand)
	addCheckBox(wrapper, i18n.Text("Per Shot"), &shots.ReloadTimeIsPerShot)
	addCheckBox(wrapper, i18n.Text("Thrown Weapon"), &shots.Thrown)
}

func (we *weaponEditor) addBulkBlock(w *gurps.Weapon, content *unison.Panel) {
	bulk := &w.Bulk
	wrapper := addFlowWrapper(content, i18n.Text("Bulk"), 4)
	text := i18n.Text("Normal Bulk")
	addDecimalField(wrapper, nil, "", text, text, &bulk.Normal, -fxp.Thousand, 0)
	wrapper.AddChild(NewFieldInteriorLeadingLabel(i18n.Text("For Giants"), false))
	text = i18n.Text("Giant Bulk")
	addDecimalField(wrapper, nil, "", text, text, &bulk.Giant, -fxp.Thousand, 0)
	addCheckBox(wrapper, i18n.Text("Retracting Stock"), &bulk.RetractingStock)
}

func (we *weaponEditor) addRecoilBlock(w *gurps.Weapon, content *unison.Panel) {
	recoil := &w.Recoil
	wrapper := addFlowWrapper(content, i18n.Text("Recoil"), 3)
	text := i18n.Text("Shot Recoil")
	addDecimalField(wrapper, nil, "", text, text, &recoil.Shot, 0, fxp.Thousand)
	wrapper.AddChild(NewFieldInteriorLeadingLabel(i18n.Text("For Slugs"), false))
	text = i18n.Text("Slug Recoil")
	addDecimalField(wrapper, nil, "", text, text, &recoil.Slug, 0, fxp.Thousand)
}

func (we *weaponEditor) addStrengthBlock(w *gurps.Weapon, content *unison.Panel) {
	strength := &w.Strength
	text := i18n.Text("Minimum ST")
	wrapper := addFlowWrapper(content, text, 3)
	addDecimalField(wrapper, nil, "", text, text, &strength.Min, 0, fxp.MillionMinusOne)
	addCheckBox(wrapper, i18n.Text("Two-handed"), &strength.TwoHanded)
	addCheckBox(wrapper, i18n.Text("Two-handed & unready"), &strength.TwoHandedUnready)
	if w.IsRanged() {
		wrapper = addFlowWrapper(content, "", 3)
		addCheckBox(wrapper, i18n.Text("Has bipod"), &strength.Bipod)
		addCheckBox(wrapper, i18n.Text("Mounted"), &strength.Mounted)
		addCheckBox(wrapper, i18n.Text("Uses a musket rest"), &strength.MusketRest)
	}
}

func (we *weaponEditor) preApply(w *gurps.Weapon) {
	w.RateOfFire.Jet = w.Accuracy.Jet
	w.Validate()
}
