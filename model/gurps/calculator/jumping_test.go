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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/encumbrance"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestJump verifies the jumping rules of BX352 and the extra effort of BX356 for a character typed in: the worked
// example of Basic Move 5 and ST 10, and each of the things that raise the Basic Move or scale the distance.
func TestJump(t *testing.T) {
	c := check.New(t)
	jumper := Jumper{BasicMove: fxp.Five, LiftingST: fxp.Ten, Weight: fxp.Weight(fxp.FromInteger(150))}
	c.Equal(fxp.Twenty, jumper.HighJump(0, 0), "a standing high jump is 6 x 5 - 10 = 20 inches")
	c.Equal(fxp.FromInteger(84), jumper.BroadJump(0, 0), "a standing broad jump is 2 x 5 - 3 = 7 feet")
	c.Equal(fxp.FromInteger(22), jumper.HighJump(0, -2), "extra effort at -2 adds 10%")
	c.Equal(fxp.FromInteger(92), jumper.BroadJump(0, -2), "to both jumps, rounded down to the inch")

	c.Equal(fxp.Forty, jumper.HighJump(fxp.Five, 0), "a running start adds to the Basic Move, but at most doubles the jump")
	c.Equal(fxp.FromInteger(168), jumper.BroadJump(fxp.Five, 0), "which is 14 feet for the broad jump")

	enhanced := jumper
	enhanced.EnhancedMove = fxp.One
	c.Equal(fxp.Twenty, enhanced.HighJump(0, 0), "Enhanced Move does nothing without a running start")
	c.Equal(fxp.Forty, enhanced.HighJump(fxp.One, 0), "with one, it doubles the Basic Move: 6 x 10 - 10 = 50, capped at double")

	strong := jumper
	strong.LiftingST = fxp.Forty
	c.Equal(fxp.FromInteger(50), strong.HighJump(0, 0), "a Basic Lift above the body weight allows ST/4 as the Basic Move")

	encumbered := jumper
	encumbered.Encumbrance = encumbrance.Light
	c.Equal(fxp.Sixteen, encumbered.HighJump(0, 0), "each level of encumbrance takes 20% off")

	skilled := jumper
	skilled.JumpingSkill = 12
	c.Equal(fxp.FromInteger(26), skilled.HighJump(0, 0), "half the Jumping skill replaces a lower Basic Move")

	super := jumper
	super.SuperJump = fxp.One
	c.Equal(fxp.Forty, super.HighJump(0, 0), "each level of Super Jump doubles the distance")
}

// TestJumperFromEntity verifies what a sheet supplies about a jumper: the Basic Move, lifting ST, weight and
// encumbrance, the Jumping skill when it is on the sheet, and the levels of the two traits that help.
func TestJumperFromEntity(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	j := JumperFromEntity(e)
	c.Equal(e, j.Entity, "the sheet's character supplies the Basic Lift")
	c.Equal(e.ResolveAttributeCurrent(gurps.BasicMoveID), j.BasicMove, "the Basic Move comes from the sheet")
	c.Equal(e.LiftingStrength(), j.LiftingST, "as does the lifting ST")
	c.Equal(e.Profile.Weight, j.Weight, "and the weight")
	c.Equal(encumbrance.No, j.Encumbrance, "an empty sheet carries nothing")
	c.Equal(0, j.JumpingSkill, "no Jumping skill is 0")
	c.Equal(fxp.Int(0), j.EnhancedMove, "no Enhanced Move is 0")
	c.Equal(fxp.Int(0), j.SuperJump, "no Super Jump is 0")

	addDXSkill(e, "Jumping", difficulty.Easy, fxp.One)
	addLeveledTrait(e, "Enhanced Move (Ground)", fxp.One)
	addLeveledTrait(e, "Super Jump", fxp.Two)
	e.Recalculate()
	j = JumperFromEntity(e)
	c.Equal(10, j.JumpingSkill, "one point in the DX/Easy skill is DX")
	c.Equal(fxp.One, j.EnhancedMove, "the trait's levels are read")
	c.Equal(fxp.Two, j.SuperJump, "for both traits")
}
