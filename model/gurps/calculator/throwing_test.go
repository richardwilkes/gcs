// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package calculator

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestThrow verifies the throwing rules of BX355 and the extra effort of BX357 for a character typed in: the worked
// example of ST 10 throwing a 1 lb object, the skill bonuses, the objects that cannot be thrown at all, a thrower with
// no ST, and the damage an object does by its weight.
func TestThrow(t *testing.T) {
	c := check.New(t)
	thrower := Thrower{ST: fxp.Ten, StrikingST: fxp.Ten, Throwing: ThrowingTiers()[0], ThrowingArt: ThrowingArtTiers()[0]}
	pound := fxp.Weight(fxp.One)

	r := thrower.Throw(pound, 0)
	c.Equal(fxp.FromInteger(1260), r.Distance, "1 lb is a twentieth of a Basic Lift of 20, for 3.5 x ST = 35 yards")
	c.Equal("1d-4", gurps.FormatDice(r.Damage, false), "a light object does thrust-2 per die: 1d-2 becomes 1d-4")
	c.False(r.TooHeavy, "it can be thrown")

	r = thrower.Throw(pound, -2)
	c.Equal(fxp.FromInteger(1386), r.Distance, "extra effort at -2 raises the ST to 11: 11 x 3.5 = 38.5 yards")
	c.Equal("1d-3", gurps.FormatDice(r.Damage, false), "and the thrust to that of ST 11, less 2 for the light object")

	art := thrower
	art.ThrowingArt = ThrowingArtTiers()[2]
	r = art.Throw(pound, 0)
	c.Equal(fxp.FromInteger(1512), r.Distance, "Throwing Art at DX+1 adds 2 to the ST for distance: 12 x 3.5 = 42 yards")
	c.Equal("1d-2", gurps.FormatDice(r.Damage, false), "and 2 per die to the damage, which the light object takes back")

	skilled := thrower
	skilled.Throwing = ThrowingTiers()[1]
	r = skilled.Throw(pound, 0)
	c.Equal(fxp.FromInteger(1386), r.Distance, "Throwing at DX+1 adds 1 to the ST for distance")
	c.Equal("1d-4", gurps.FormatDice(r.Damage, false), "but nothing to the damage")

	c.Equal(ThrowResult{}, thrower.Throw(0, 0), "nothing is thrown without a weight")
	r = thrower.Throw(fxp.Weight(fxp.FromInteger(300)), 0)
	c.True(r.TooHeavy, "15 times the Basic Lift is off the table and cannot be thrown")
	c.Equal(fxp.Int(0), r.Distance, "so it goes nowhere")

	weak := thrower
	weak.ST = 0
	weak.StrikingST = 0
	c.Equal(ThrowResult{TooHeavy: true}, weak.Throw(pound, 0), "without ST there is no Basic Lift, and nothing can be thrown")
	weak.Throwing = ThrowingTiers()[2]
	r = weak.Throw(pound, 0)
	c.Equal(fxp.FromInteger(28), r.Distance, "Throwing at DX+2 alone gives an ST of 2 to throw with: 2 x 0.4 = 0.8 yards")
	c.Equal("1d-8", gurps.FormatDice(r.Damage, false),
		"with no Basic Lift of the thrower's own to weigh it against, the object counts as the lightest: 1d-6 becomes 1d-8")

	// ST 20 has a Basic Lift of 80 lbs and a thrust of 2d-1, which is enough dice to tell the bands apart.
	strong := Thrower{ST: fxp.Twenty, StrikingST: fxp.Twenty, Throwing: ThrowingTiers()[0], ThrowingArt: ThrowingArtTiers()[0]}
	for _, tc := range []struct {
		name     string
		lightest int // The lightest whole number of pounds in the band.
		heaviest int // The heaviest.
		want     string
	}{
		{name: "up to BL/8 is thrust-2 per die", lightest: 1, heaviest: 10, want: "2d-5"},
		{name: "up to BL/4 is thrust-1 per die", lightest: 11, heaviest: 20, want: "2d-3"},
		{name: "up to BL/2 is thrust", lightest: 21, heaviest: 40, want: "2d-1"},
		{name: "up to BL is thrust+1 per die", lightest: 41, heaviest: 80, want: "2d+1"},
		{name: "up to 2xBL is thrust", lightest: 81, heaviest: 160, want: "2d-1"},
		{name: "up to 4xBL is thrust-1 per 2 dice", lightest: 161, heaviest: 320, want: "2d-2"},
		{name: "anything heavier is thrust-1 per die", lightest: 321, heaviest: 960, want: "2d-3"},
	} {
		for _, pounds := range []int{tc.lightest, tc.heaviest} {
			r = strong.Throw(fxp.Weight(fxp.FromInteger(pounds)), 0)
			c.Equal(tc.want, gurps.FormatDice(r.Damage, false), "%s: %d lbs", tc.name, pounds)
		}
	}
}

// TestThrowingDistanceModifier pins the Throwing Distance Table (BX355): a weight ratio reads the first row that
// reaches it, and anything past the last row cannot be thrown.
func TestThrowingDistanceModifier(t *testing.T) {
	c := check.New(t)
	table := []struct{ maxRatio, modifier string }{
		{"0.05", "3.5"},
		{"0.1", "2.5"},
		{"0.15", "2"},
		{"0.2", "1.5"},
		{"0.25", "1.2"},
		{"0.3", "1.1"},
		{"0.4", "1"},
		{"0.5", "0.8"},
		{"0.75", "0.7"},
		{"1", "0.6"},
		{"1.5", "0.4"},
		{"2", "0.3"},
		{"2.5", "0.25"},
		{"3", "0.2"},
		{"4", "0.15"},
		{"5", "0.12"},
		{"6", "0.1"},
		{"7", "0.09"},
		{"8", "0.08"},
		{"9", "0.07"},
		{"10", "0.06"},
		{"12", "0.05"},
	}
	for i, row := range table {
		ratio := fxp.FromStringForced(row.maxRatio)
		c.Equal(fxp.FromStringForced(row.modifier), throwingDistanceModifier(ratio), "a ratio of %s", row.maxRatio)
		var next fxp.Int // Past the last row, nothing can be thrown.
		if i+1 < len(table) {
			next = fxp.FromStringForced(table[i+1].modifier)
		}
		c.Equal(next, throwingDistanceModifier(ratio+1), "a ratio just over %s", row.maxRatio)
	}
}

// TestThrowingTiers verifies the tier tables, whose order the sheet reader's indexes depend on.
func TestThrowingTiers(t *testing.T) {
	c := check.New(t)
	tiers := ThrowingTiers()
	checkNamed(c, "Throwing tier", tiers, func(t ThrowingTier) string { return t.Name })
	c.Equal([]ThrowingTier{{Name: tiers[0].Name}, {Name: tiers[1].Name, Distance: 1}, {Name: tiers[2].Name, Distance: 2}}, tiers,
		"Throwing adds 1 to the ST at DX+1 and 2 at DX+2, and nothing to the damage")
	tiers = ThrowingArtTiers()
	checkNamed(c, "Throwing Art tier", tiers, func(t ThrowingTier) string { return t.Name })
	c.Equal([]ThrowingTier{{Name: tiers[0].Name}, {Name: tiers[1].Name, Distance: 1, Damage: 1}, {Name: tiers[2].Name, Distance: 2, Damage: 2}}, tiers,
		"Throwing Art adds 1 of each at DX and 2 of each at DX+1")
}

// TestThrowerStatsFromEntity verifies what a sheet supplies about a thrower: the lifting ST without its lifting-only
// bonus but with its lifting-only multipliers, the striking ST, and the tier each throwing skill's level falls in.
func TestThrowerStatsFromEntity(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	s := ThrowerStatsFromEntity(e)
	c.Equal(e.LiftingStrength(), s.ST, "the ST is the lifting ST without a lifting-only bonus")
	c.Equal(e.StrikingStrength(), s.StrikingST, "the striking ST is the sheet's")
	lifter := gurps.NewTrait(e, nil, false)
	lifter.Name = "Lifting ST 4"
	bonus := gurps.NewAttributeBonus(gurps.StrengthID)
	bonus.Limitation = stlimit.LiftingOnly
	bonus.Amount = fxp.Four
	lifter.Features = gurps.Features{bonus}
	e.Traits = append(e.Traits, lifter)
	e.Attributes.Set[gurps.FatiguePointsID].Damage = fxp.Seven
	e.Recalculate()
	c.Equal(fxp.Seven, e.LiftingStrength(), "precondition: ST 10 plus the bonus, halved while Tired")
	c.Equal(fxp.Five, ThrowerStatsFromEntity(e).ST,
		"the ST is halved while Tired, but the lifting-only bonus is left out")
	e.Attributes.Set[gurps.FatiguePointsID].Damage = 0
	c.Equal(0, s.ThrowingIndex, "without Throwing, the first tier")
	c.Equal(0, s.ThrowingArtIndex, "without Throwing Art, the first tier")

	throwing := addDXSkill(e, "Throwing", difficulty.Average, fxp.Four)
	throwingArt := addDXSkill(e, "Throwing Art", difficulty.Hard, fxp.Four)
	e.Recalculate()
	s = ThrowerStatsFromEntity(e)
	c.Equal(1, s.ThrowingIndex, "4 points in the DX/Average skill is DX+1, the second tier")
	c.Equal(1, s.ThrowingArtIndex, "4 points in the DX/Hard skill is DX, the second tier")

	throwing.Points = fxp.Eight
	throwingArt.Points = fxp.Eight
	e.Recalculate()
	s = ThrowerStatsFromEntity(e)
	c.Equal(2, s.ThrowingIndex, "8 points is DX+2, the third tier")
	c.Equal(2, s.ThrowingArtIndex, "8 points is DX+1, the third tier")

	throwing.Points = fxp.One
	throwingArt.Points = fxp.One
	e.Recalculate()
	s = ThrowerStatsFromEntity(e)
	c.Equal(0, s.ThrowingIndex, "1 point is below DX+1, the first tier")
	c.Equal(0, s.ThrowingArtIndex, "1 point is below DX, the first tier")
}
