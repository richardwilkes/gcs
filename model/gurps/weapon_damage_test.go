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
