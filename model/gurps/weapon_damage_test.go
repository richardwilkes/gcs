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

// TestWeaponDRDivisorBonus verifies that a DR divisor bonus adjusts the armor divisor. A percentage bonus must add a
// percentage of the divisor rather than replacing the divisor with that percentage of itself, which is what the
// tooltip ("+10% to armor divisor") describes.
func TestWeaponDRDivisorBonus(t *testing.T) {
	c := check.New(t)
	for i, one := range []struct {
		amount   fxp.Int
		percent  bool
		divisor  fxp.Int
		expected string
	}{
		{amount: fxp.One, divisor: fxp.Two, expected: "1d(3) cr"},
		{amount: fxp.NegOne, divisor: fxp.Two, expected: "1d cr"}, // A divisor of 1 isn't shown
		{amount: fxp.FromInteger(10), percent: true, divisor: fxp.Two, expected: "1d(2.2) cr"},
		{amount: fxp.Hundred, percent: true, divisor: fxp.Two, expected: "1d(4) cr"},
		{amount: fxp.FromInteger(-50), percent: true, divisor: fxp.Two, expected: "1d cr"},
		{amount: fxp.Hundred, percent: true, divisor: fxp.Half, expected: "1d cr"},
		{amount: fxp.FromInteger(10), percent: true, divisor: fxp.Half, expected: "1d(0.55) cr"},
	} {
		bonus := gurps.NewWeaponDRDivisorBonus()
		bonus.Amount = one.amount
		bonus.Percent = one.percent
		w := newWeaponWithBonuses(false, bonus)
		w.Damage.ArmorDivisor = one.divisor
		c.Equal(one.expected, w.Damage.ResolvedDamage(nil), "test %d", i)
	}
}

// TestWeaponDamageSpecIsScriptOrCompleteDice verifies that a damage field is only treated as a plain dice
// specification when the dice grammar consumes all of it. The dice parser stops at the first character it cannot use
// and silently discards the remainder, so a script expression written without spaces (which is how anyone writing
// arithmetic naturally writes it) must not be truncated to the dice specification it happens to start with.
func TestWeaponDamageSpecIsScriptOrCompleteDice(t *testing.T) {
	c := check.New(t)
	for i, one := range []struct {
		base     string
		expected string
	}{
		// Script expressions that start with something the dice parser would happily consume. Truncating instead of
		// evaluating these would yield "2 cr", "9 cr", "4 cr" and "3 cr", respectively.
		{base: "2*3", expected: "6 cr"},
		{base: "10-1-1", expected: "8 cr"},
		{base: "4/2", expected: "2 cr"},
		{base: "1+2*3", expected: "7 cr"},
		// Complete dice specifications must still be taken as dice rather than handed to the script engine, which
		// would fail to parse them.
		{base: "1d", expected: "1d cr"},
		{base: "1d6", expected: "1d cr"},
		{base: "2d6+1", expected: "2d+1 cr"},
		{base: "1d4-1", expected: "1d4-1 cr"},
		{base: "2x3", expected: "2x3 cr"},
		{base: "1d+", expected: "1d cr"}, // A dangling sign is an empty modifier to the dice parser
		{base: "+1d6", expected: "1d cr"},
		{base: "3", expected: "3 cr"},
		{base: "-2", expected: "-2 cr"},
		{base: "0", expected: "cr"},
	} {
		w := newWeaponWithBonuses(false)
		w.Damage.Base = one.base
		c.Equal(one.expected, w.Damage.ResolvedDamage(nil), "test %d (%s)", i, one.base)
	}
}
