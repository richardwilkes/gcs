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
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestThrow verifies the throwing rules of BX355 and the extra effort of BX357 for a character typed in: the worked
// example of ST 10 throwing a 1 lb object, the skill bonuses, and the objects that cannot be thrown at all.
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
// bonus, the striking ST, and the tier each throwing skill's level falls in.
func TestThrowerStatsFromEntity(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	s := ThrowerStatsFromEntity(e)
	c.Equal(e.LiftingStrength()-e.LiftingStrengthBonus, s.ST, "the ST is the lifting ST without its lifting-only bonus")
	c.Equal(e.StrikingStrength(), s.StrikingST, "the striking ST is the sheet's")
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
