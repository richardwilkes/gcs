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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wswitch"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestWeaponBulk(t *testing.T) {
	c := check.New(t)
	for i, s := range []string{
		"",
		"-1",
		"-10",
		"-11/-14",
		"-3*",
	} {
		c.Equal(s, gurps.ParseWeaponBulk(s).String(), "test %d", i)
	}

	cases := []struct {
		input    string
		expected string
	}{
		{"-", ""},
		{"–", ""},
		{"—", ""},
		{"?", ""},
		{"0", ""},
	}
	for i, one := range cases {
		c.Equal(one.expected, gurps.ParseWeaponBulk(one.input).String(), "test %d", i)
	}
}

// TestWeaponBulkBonusResolution verifies that a bulk bonus doesn't materialize a giant bulk on a weapon that doesn't
// have one (a stored giant bulk of 0 means "no separate giant bulk"), while still adjusting one that is present.
func TestWeaponBulkBonusResolution(t *testing.T) {
	c := check.New(t)

	bonus := gurps.NewWeaponBulkBonus()
	bonus.Amount = fxp.NegOne
	w := newWeaponWithBonuses(false, bonus)

	w.Bulk = gurps.ParseWeaponBulk("-3")
	c.Equal("-4", w.Bulk.Resolve(w, nil).String(), "a weapon with no giant bulk should not gain one")

	w.Bulk = gurps.ParseWeaponBulk("-3*")
	c.Equal("-4*", w.Bulk.Resolve(w, nil).String(), "the retracting stock marker should be preserved")

	w.Bulk = gurps.ParseWeaponBulk("-3/-5")
	c.Equal("-4/-6", w.Bulk.Resolve(w, nil).String(), "an existing giant bulk should be adjusted")
}

// TestWeaponBulkRetractingStockSwitch verifies that the "retracting stock" weapon switch is actually resolved. It is
// offered by the weapon-switch feature editor, but WeaponBulk.Resolve never consulted it, so the feature did nothing.
func TestWeaponBulkRetractingStockSwitch(t *testing.T) {
	c := check.New(t)

	sw := gurps.NewWeaponSwitchBonus()
	sw.SwitchType = wswitch.RetractingStock
	sw.SwitchTypeValue = true
	w := newWeaponWithBonuses(false, sw)

	// Switched on, the stock marker appears, and with it the tooltip explaining the folded-stock stats.
	w.Bulk = gurps.ParseWeaponBulk("-3")
	resolved := w.Bulk.Resolve(w, nil)
	c.Equal("-3*", resolved.String(), "the switch adds a retracting stock")
	c.NotEqual("", resolved.Tooltip(w), "the retracting stock tooltip appears with it")

	w.Bulk = gurps.ParseWeaponBulk("-3*")
	c.Equal("-3*", w.Bulk.Resolve(w, nil).String(), "a weapon that already has one is unchanged")

	// Switched off, it is removed.
	sw.SwitchTypeValue = false
	w.Bulk = gurps.ParseWeaponBulk("-3*")
	resolved = w.Bulk.Resolve(w, nil)
	c.Equal("-3", resolved.String(), "the switch removes a retracting stock")
	c.Equal("", resolved.Tooltip(w), "and the tooltip goes with it")

	w.Bulk = gurps.ParseWeaponBulk("-3")
	c.Equal("-3", w.Bulk.Resolve(w, nil).String(), "a weapon that has none is unchanged")

	// A bulk bonus still applies alongside the switch.
	bonus := gurps.NewWeaponBulkBonus()
	bonus.Amount = fxp.NegOne
	sw.SwitchTypeValue = true
	w = newWeaponWithBonuses(false, sw, bonus)
	w.Bulk = gurps.ParseWeaponBulk("-3")
	c.Equal("-4*", w.Bulk.Resolve(w, nil).String(), "the bonus and the switch both apply")
}

// TestWeaponBulkPercentBonusResolution verifies that a percentage bulk bonus adjusts the existing values rather than
// replacing them, and likewise leaves an absent giant bulk absent.
func TestWeaponBulkPercentBonusResolution(t *testing.T) {
	c := check.New(t)

	bonus := gurps.NewWeaponBulkBonus()
	bonus.Percent = true
	bonus.Amount = fxp.FromInteger(50)
	w := newWeaponWithBonuses(false, bonus)

	w.Bulk = gurps.ParseWeaponBulk("-4")
	c.Equal("-6", w.Bulk.Resolve(w, nil).String())

	w.Bulk = gurps.ParseWeaponBulk("-4/-6")
	c.Equal("-6/-9", w.Bulk.Resolve(w, nil).String())
}
