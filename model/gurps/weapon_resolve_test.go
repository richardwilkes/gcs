// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/attack"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wswitch"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestWeaponAccuracyMixedBonusResolution verifies that when several bonus types are collected in one pass, each
// type's flat and percentage amounts are kept together and applied only to their own field, with the flat amounts
// applied before the percentage.
func TestWeaponAccuracyMixedBonusResolution(t *testing.T) {
	c := check.New(t)

	flatAcc := gurps.NewWeaponBonus(feature.WeaponAccBonus)
	flatAcc.Amount = fxp.Two
	percentAcc := gurps.NewWeaponBonus(feature.WeaponAccBonus)
	percentAcc.Percent = true
	percentAcc.Amount = fxp.FromInteger(50)
	percentScope := gurps.NewWeaponBonus(feature.WeaponScopeAccBonus)
	percentScope.Percent = true
	percentScope.Amount = fxp.Hundred
	w := newWeaponWithBonuses(false, flatAcc, percentAcc, percentScope)

	w.Accuracy = attack.ParseAccuracy("4+2")
	c.Equal("9+4", w.ResolvedAccuracy(nil).String(), "(4+2)*1.5 base, 2*2 scope")

	w.Accuracy = attack.ParseAccuracy("4")
	c.Equal("9", w.ResolvedAccuracy(nil).String(), "a percentage of an absent scope stays absent")
}

// TestWeaponBulkBonusResolution verifies that a flat or percentage bulk bonus adjusts the existing values rather than
// replacing them, and doesn't materialize a giant bulk on a weapon that doesn't have one (a stored giant bulk of 0
// means "no separate giant bulk"), while still adjusting one that is present.
func TestWeaponBulkBonusResolution(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		percent bool
		amount  fxp.Int
		bulk    string
		want    string
		msg     string
	}{
		{amount: fxp.NegOne, bulk: "-3", want: "-4", msg: "a weapon with no giant bulk should not gain one"},
		{amount: fxp.NegOne, bulk: "-3*", want: "-4*", msg: "the retracting stock marker should be preserved"},
		{amount: fxp.NegOne, bulk: "-3/-5", want: "-4/-6", msg: "an existing giant bulk should be adjusted"},
		{
			percent: true,
			amount:  fxp.FromInteger(50),
			bulk:    "-4",
			want:    "-6",
			msg:     "a percentage should adjust the bulk without adding a giant bulk",
		},
		{
			percent: true,
			amount:  fxp.FromInteger(50),
			bulk:    "-4/-6",
			want:    "-6/-9",
			msg:     "a percentage should adjust an existing giant bulk too",
		},
	} {
		bonus := gurps.NewWeaponBonus(feature.WeaponBulkBonus)
		bonus.Percent = tc.percent
		bonus.Amount = tc.amount
		w := newWeaponWithBonuses(false, bonus)
		w.Bulk = attack.ParseBulk(tc.bulk)
		c.Equal(tc.want, w.ResolvedBulk(nil).String(), tc.msg)
	}
}

// TestWeaponBulkRetractingStockSwitch verifies that the "retracting stock" weapon switch is actually resolved. It is
// offered by the weapon-switch feature editor, but resolving the bulk never consulted it, so the feature did nothing.
func TestWeaponBulkRetractingStockSwitch(t *testing.T) {
	c := check.New(t)

	sw := gurps.NewWeaponBonus(feature.WeaponSwitch)
	sw.SwitchType = wswitch.RetractingStock
	sw.SwitchTypeValue = true
	w := newWeaponWithBonuses(false, sw)

	w.Bulk = attack.ParseBulk("-3")
	resolved := w.ResolvedBulk(nil)
	c.Equal("-3*", resolved.String(), "the switch adds a retracting stock")
	c.NotEqual("", resolved.Tooltip(w.ResolvedAccuracy(nil), w.ResolvedRecoil(nil), w.ResolvedStrength(nil)), "the retracting stock tooltip appears with it")

	w.Bulk = attack.ParseBulk("-3*")
	c.Equal("-3*", w.ResolvedBulk(nil).String(), "a weapon that already has one is unchanged")

	sw.SwitchTypeValue = false
	w.Entity().DiscardCaches()
	w.Bulk = attack.ParseBulk("-3*")
	resolved = w.ResolvedBulk(nil)
	c.Equal("-3", resolved.String(), "the switch removes a retracting stock")
	c.Equal("", resolved.Tooltip(w.ResolvedAccuracy(nil), w.ResolvedRecoil(nil), w.ResolvedStrength(nil)), "and the tooltip goes with it")

	w.Bulk = attack.ParseBulk("-3")
	c.Equal("-3", w.ResolvedBulk(nil).String(), "a weapon that has none is unchanged")

	bonus := gurps.NewWeaponBonus(feature.WeaponBulkBonus)
	bonus.Amount = fxp.NegOne
	sw.SwitchTypeValue = true
	w = newWeaponWithBonuses(false, sw, bonus)
	w.Bulk = attack.ParseBulk("-3")
	c.Equal("-4*", w.ResolvedBulk(nil).String(), "the bonus and the switch both apply")
}

// TestWeaponParryResolve runs the resolution regression tests shared by both defenses against the parry and the block.
func TestWeaponParryResolve(t *testing.T) {
	for _, d := range []defenseUnderTest{parryUnderTest, blockUnderTest} {
		t.Run(d.name, func(t *testing.T) { runDefenseResolveTests(t, d) })
	}
}

// TestWeaponPercentBonusFloorsIncrement pins the documented rounding of weapon percentage bonuses: the increment is
// floored, i.e. rounded toward negative infinity, not toward zero. A -50% bonus on a recoil of 5 yields an increment
// of -2.5, which floors to -3, for a result of 2.
func TestWeaponPercentBonusFloorsIncrement(t *testing.T) {
	c := check.New(t)

	bonus := gurps.NewWeaponBonus(feature.WeaponRecoilBonus)
	bonus.Percent = true
	bonus.Amount = fxp.FromInteger(-50)
	w := newWeaponWithBonuses(false, bonus)

	w.Recoil = attack.ParseRecoil("5")
	c.Equal("2", w.ResolvedRecoil(nil).String())

	w.Recoil = attack.ParseRecoil("5/7")
	c.Equal("2/3", w.ResolvedRecoil(nil).String())
}
