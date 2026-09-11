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
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

func bonusDice(count, sides, modifier, multiplier int, sub bool) gurps.BonusDice {
	var b gurps.BonusDice
	b.Count = count
	b.Sides = sides
	b.Modifier = modifier
	b.Multiplier = multiplier
	b.Sub = sub
	return b
}

// TestParseWeaponDamageBonus verifies that the text of a weapon damage bonus is either a flat number, which may be
// fractional, or a complete dice specification with an optional leading sign, and that text in any other form is
// rejected.
func TestParseWeaponDamageBonus(t *testing.T) {
	c := check.New(t)
	gurps.GlobalSettings() // Initializes the dice rollers, which parse and format the dice
	for _, tc := range []struct {
		text   string
		dice   gurps.BonusDice
		amount fxp.Int
		ok     bool
	}{
		{text: "+2", amount: fxp.Two, ok: true},
		{text: "2", amount: fxp.Two, ok: true},
		{text: "-1", amount: -fxp.One, ok: true},
		{text: "+0.5", amount: fxp.Half, ok: true},
		{text: "+", ok: true},
		{text: "0", ok: true},
		{text: "+1d", dice: bonusDice(1, 6, 0, 1, false), ok: true},
		{text: "1d", dice: bonusDice(1, 6, 0, 1, false), ok: true},
		{text: "-1d", dice: bonusDice(1, 6, 0, 1, true), ok: true},
		{text: "+2D6", dice: bonusDice(2, 6, 0, 1, false), ok: true},
		{text: "+1d3", dice: bonusDice(1, 3, 0, 1, false), ok: true},
		{text: "+1d+2", dice: bonusDice(1, 6, 2, 1, false), ok: true},
		{text: "1d-1", dice: bonusDice(1, 6, -1, 1, false), ok: true},
		{text: "-2d3-1", dice: bonusDice(2, 3, -1, 1, true), ok: true},
		{text: "2dx3", dice: bonusDice(2, 6, 0, 3, false), ok: true},
		{text: "+2d+1x3", dice: bonusDice(2, 6, 1, 3, false), ok: true},
		{text: "-2d+1x3", dice: bonusDice(2, 6, 1, 3, true), ok: true},
		{text: " +1d+2 ", dice: bonusDice(1, 6, 2, 1, false), ok: true},
		{text: "1d+", dice: bonusDice(1, 6, 0, 1, false), ok: true}, // A dangling sign is an empty modifier
		{text: ""},
		{text: "abc"},
		{text: "1d 2"},
		{text: "1d+abc"},
		{text: "1d+0.5"},
		{text: "2*self.level"},
	} {
		d, amount, ok := gurps.ParseWeaponDamageBonus(tc.text)
		c.Equal(tc.ok, ok, "%q: ok", tc.text)
		c.Equal(tc.dice, d, "%q: dice", tc.text)
		c.Equal(tc.amount, amount, "%q: amount", tc.text)
	}
}

// TestParseBonusDiceRejectsBareNumbers verifies that dice must have at least one die in them, so that a bare number is
// never carried as dice with only a modifier, which the editor could not show or accept, and that a zero
// specification is no dice.
func TestParseBonusDiceRejectsBareNumbers(t *testing.T) {
	c := check.New(t)
	gurps.GlobalSettings() // Initializes the dice rollers, which parse and format the dice
	for _, text := range []string{"2", "+2", "-2", "0d+2"} {
		d, ok := gurps.ParseBonusDice(text)
		c.False(ok, "%q: is rejected", text)
		c.Equal(gurps.BonusDice{}, d, "%q: yields no dice", text)
	}
	for _, text := range []string{"0", "+0", "-0", ""} {
		d, ok := gurps.ParseBonusDice(text)
		c.True(ok == (text != ""), "%q: ok", text)
		c.Equal(gurps.BonusDice{}, d, "%q: is no dice", text)
	}
	var b gurps.BonusDice
	c.HasError(b.UnmarshalText([]byte("2")), "a bare number is rejected as dice text")
	c.NoError(b.UnmarshalText([]byte("0")), "a zero specification is accepted as dice text")
	c.True(b.IsZero())
}

// TestFormatWeaponDamageBonus verifies that a weapon damage bonus formats with a leading sign, as the numeric bonus
// amounts always have, and that formatting round-trips through parsing.
func TestFormatWeaponDamageBonus(t *testing.T) {
	c := check.New(t)
	gurps.GlobalSettings() // Initializes the dice rollers, which parse and format the dice
	for _, tc := range []struct {
		dice   gurps.BonusDice
		amount fxp.Int
		want   string
	}{
		{want: "+0"},
		{amount: fxp.Two, want: "+2"},
		{amount: -fxp.Half, want: "-0.5"},
		{dice: bonusDice(1, 6, 0, 1, false), want: "+1d"},
		{dice: bonusDice(1, 6, 0, 1, true), want: "-1d"},
		{dice: bonusDice(2, 3, 0, 1, false), want: "+2d3"},
		{dice: bonusDice(1, 6, 2, 1, false), want: "+1d+2"},
		{dice: bonusDice(1, 6, -1, 1, true), want: "-1d-1"},
		{dice: bonusDice(2, 6, 0, 3, false), want: "+2dx3"},
		{dice: bonusDice(2, 6, 1, 3, false), want: "+2d+1x3"},
	} {
		got := gurps.FormatWeaponDamageBonus(tc.dice, tc.amount)
		c.Equal(tc.want, got)
		d, amount, ok := gurps.ParseWeaponDamageBonus(got)
		c.True(ok, "%s: parses", got)
		c.Equal(tc.dice, d, "%s: dice round-trip", got)
		c.Equal(tc.amount, amount, "%s: amount round-trip", got)
	}
	c.Equal("+1d +0.5", gurps.FormatWeaponDamageBonus(bonusDice(1, 6, 0, 1, false), fxp.Half),
		"both parts show when a hand-edited file carries both")
}

// TestWeaponBonusDiceJSON verifies that a weapon damage bonus carrying dice saves them as a dice string beside its flat
// amount, that a bonus without dice saves exactly as it did before dice existed, and that both load back.
func TestWeaponBonusDiceJSON(t *testing.T) {
	c := check.New(t)
	gurps.GlobalSettings() // Initializes the dice rollers, which parse and format the dice

	bonus := gurps.NewWeaponBonus(feature.WeaponBonus)
	bonus.Amount = 0
	bonus.Dice = bonusDice(2, 6, 1, 3, true)
	data, err := jio.Marshal(bonus)
	c.NoError(err)
	c.True(strings.Contains(string(data), `"dice":"-2d+1x3"`), "the dice are saved as a dice string: %s", data)
	var loaded gurps.WeaponBonus
	c.NoError(jio.Unmarshal(data, &loaded))
	c.Equal(bonus.Dice, loaded.Dice, "the dice load back")
	c.Equal(bonus.Amount, loaded.Amount, "the flat amount loads back")

	bonus.Dice = gurps.BonusDice{}
	bonus.Amount = fxp.Two
	data, err = jio.Marshal(bonus)
	c.NoError(err)
	c.False(strings.Contains(string(data), `"dice"`), "a bonus without dice saves no dice: %s", data)
	c.True(strings.Contains(string(data), `"amount":2`), "the flat amount is saved as a number: %s", data)
}

// TestWeaponBonusLoadsLegacyAndHandEditedDice verifies that the weapon bonus data written before dice existed, which
// carries a numeric amount alone, still loads unchanged; that the dice string is read strictly, so text that is not a
// dice specification is reported rather than silently dropped; and that dice on a bonus of any type other than the
// damage bonus are discarded, since nothing else can apply them.
func TestWeaponBonusLoadsLegacyAndHandEditedDice(t *testing.T) {
	c := check.New(t)

	var legacy gurps.WeaponBonus
	c.NoError(jio.Unmarshal([]byte(`{"type":"weapon_bonus","selection_type":"this_weapon","amount":2,"leveled":true}`),
		&legacy))
	c.Equal(fxp.Two, legacy.Amount)
	c.Equal(gurps.BonusDice{}, legacy.Dice, "a legacy bonus has no dice")
	c.True(legacy.PerLevel)

	var withDice gurps.WeaponBonus
	c.NoError(jio.Unmarshal([]byte(`{"type":"weapon_bonus","selection_type":"this_weapon","amount":0,"dice":"2d3"}`),
		&withDice))
	c.Equal(bonusDice(2, 3, 0, 1, false), withDice.Dice)

	var empty gurps.WeaponBonus
	c.NoError(jio.Unmarshal([]byte(`{"type":"weapon_bonus","selection_type":"this_weapon","amount":1,"dice":""}`),
		&empty))
	c.Equal(gurps.BonusDice{}, empty.Dice, "empty dice text is no dice")

	for _, bad := range []string{`"abc"`, `"1d 2"`, `"1d+abc"`, `"2*self.level"`, `"2"`, `"-2"`} {
		var b gurps.WeaponBonus
		c.HasError(jio.Unmarshal([]byte(`{"type":"weapon_bonus","selection_type":"this_weapon","amount":0,"dice":`+
			bad+`}`), &b), "%s: is rejected", bad)
	}

	var other gurps.WeaponBonus
	c.NoError(jio.Unmarshal([]byte(`{"type":"weapon_acc_bonus","selection_type":"this_weapon","amount":1,"dice":"1d"}`),
		&other))
	c.Equal(gurps.BonusDice{}, other.Dice, "dice on a bonus that is not a damage bonus are discarded")
	c.Equal(fxp.One, other.Amount)

	var percent gurps.WeaponBonus
	c.NoError(jio.Unmarshal([]byte(`{"type":"weapon_bonus","selection_type":"this_weapon","amount":10,"percent":true,"dice":"1d"}`),
		&percent))
	c.Equal(bonusDice(1, 6, 0, 1, false), percent.Dice, "the dice of a bonus marked as a percentage are kept")
	c.False(percent.Percent, "and the bonus is no longer a percentage, since dice cannot be one")
	c.Equal(fxp.Ten, percent.Amount)

	var percentOnly gurps.WeaponBonus
	c.NoError(jio.Unmarshal([]byte(`{"type":"weapon_bonus","selection_type":"this_weapon","amount":10,"percent":true}`),
		&percentOnly))
	c.True(percentOnly.Percent, "a percentage bonus without dice stays a percentage")
}
