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
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestWeaponAccuracy(t *testing.T) {
	c := check.New(t)
	for i, s := range []string{
		"0",
		"1",
		"1+3",
		"0+3",
		"Jet",
	} {
		c.Equal(s, gurps.ParseWeaponAccuracy(s).String(), "test %d", i)
	}

	cases := []struct {
		input    string
		expected string
	}{
		{"", "0"},
		{"-", "0"},
		{"?", "0"},
		{"+0", "0"},
		{"+1", "1"},
		{"+0+3", "0+3"},
		{"+1+3", "1+3"},
		{"+1+0", "1"},
		{"1+0", "1"},
		{"51,", "51"},
	}
	for i, one := range cases {
		c.Equal(one.expected, gurps.ParseWeaponAccuracy(one.input).String(), "test %d", i)
	}
}

// TestWeaponAccuracyMixedBonusResolution verifies that when several bonus types are collected in one pass, each
// type's flat and percentage amounts are kept together and applied only to their own field, with the flat amounts
// applied before the percentage.
func TestWeaponAccuracyMixedBonusResolution(t *testing.T) {
	c := check.New(t)

	flatAcc := gurps.NewWeaponAccBonus()
	flatAcc.Amount = fxp.Two
	percentAcc := gurps.NewWeaponAccBonus()
	percentAcc.Percent = true
	percentAcc.Amount = fxp.FromInteger(50)
	percentScope := gurps.NewWeaponScopeAccBonus()
	percentScope.Percent = true
	percentScope.Amount = fxp.Hundred
	w := newWeaponWithBonuses(false, flatAcc, percentAcc, percentScope)

	w.Accuracy = gurps.ParseWeaponAccuracy("4+2")
	c.Equal("9+4", w.Accuracy.Resolve(w, nil).String(), "(4+2)*1.5 base, 2*2 scope")

	w.Accuracy = gurps.ParseWeaponAccuracy("4")
	c.Equal("9", w.Accuracy.Resolve(w, nil).String(), "a percentage of an absent scope stays absent")
}
