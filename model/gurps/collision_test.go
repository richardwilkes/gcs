// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"math"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/v2/check"
)

// printedFallingVelocities is the Falling Velocity Table exactly as it appears on B431, one entry per yard, written out
// independently of the row structure collision.go encodes so that a mistake in either one shows up as a disagreement.
var printedFallingVelocities = map[int]int{
	1: 5, 2: 7, 3: 8, 4: 9, 5: 10, 6: 11, 7: 12, 8: 13, 9: 14,
	10: 15, 11: 15,
	12: 16,
	13: 17, 14: 17,
	15: 18,
	16: 19, 17: 19,
	18: 20, 19: 20,
	20: 21, 21: 21,
	22: 22, 23: 22,
	24: 23, 25: 23,
	26: 24, 27: 24,
	28: 25, 29: 25,
	30: 26, 31: 26, 32: 26,
	33: 27, 34: 27,
	35: 28, 36: 28, 37: 28,
	38: 29, 39: 29,
	40: 30, 41: 30, 42: 30,
	43: 31, 44: 31, 45: 31,
	46: 32, 47: 32, 48: 32,
	49: 33, 50: 33, 51: 33,
	52: 34, 53: 34, 54: 34,
	55: 35, 56: 35, 57: 35,
	58: 36, 59: 36, 60: 36, 61: 36,
	62: 37, 63: 37, 64: 37,
	65: 38, 66: 38, 67: 38,
	68: 39, 69: 39, 70: 39, 71: 39,
	72: 40, 73: 40, 74: 40, 75: 40,
	76: 41, 77: 41, 78: 41, 79: 41,
	80: 42, 81: 42, 82: 42,
	83: 43, 84: 43, 85: 43, 86: 43,
	87: 44, 88: 44, 89: 44, 90: 44,
	91: 45, 92: 45, 93: 45, 94: 45, 95: 45,
	96: 46, 97: 46, 98: 46, 99: 46,
	100: 47, 101: 47, 102: 47, 103: 47,
	104: 48, 105: 48, 106: 48, 107: 48, 108: 48,
	109: 49, 110: 49, 111: 49, 112: 49,
}

// TestFallingVelocityTable verifies that FallingVelocity reproduces every entry of the printed Falling Velocity Table
// (B431) at one gravity, that it hands off to the formula past the end of the table without a step in the value, and
// that it uses the formula for any other gravity.
//
// The printed table cannot be replaced by the formula the same page offers as an alternative: 29 of its 112 entries are
// exactly one yard per second higher than round(sqrt(21.4 x distance)), so the table has to be encoded.
func TestFallingVelocityTable(t *testing.T) {
	c := check.New(t)
	formulaMismatches := 0
	for yards := 1; yards <= 112; yards++ {
		want := fxp.FromInteger(printedFallingVelocities[yards])
		c.Equal(want, FallingVelocity(fxp.FromInteger(yards), fxp.One), "%d yards at 1G", yards)
		if want != formulaFallingVelocity(float64(yards), 1) {
			formulaMismatches++
		}
	}
	c.Equal(29, formulaMismatches, "the printed table must differ from the formula, or it would not need encoding")

	// The worked example on B431: a 17 yard fall reaches 19 yards per second.
	c.Equal(fxp.Nineteen, FallingVelocity(fxp.FromInteger(17), fxp.One), "the B431 worked example")

	// The seam at the end of the table is clean: the last printed row and the first formula result agree.
	c.Equal(fxp.FromInteger(49), FallingVelocity(fxp.FromInteger(112), fxp.One), "the last printed row")
	c.Equal(fxp.FromInteger(49), FallingVelocity(fxp.FromInteger(113), fxp.One), "the first distance past the table")

	for _, tc := range []struct {
		name     string
		distance fxp.Int
		gravity  fxp.Int
		want     fxp.Int
	}{
		{name: "no distance", distance: 0, gravity: fxp.One, want: 0},
		{name: "negative distance", distance: fxp.FromInteger(-3), gravity: fxp.One, want: 0},
		{name: "no gravity", distance: fxp.FromInteger(17), gravity: 0, want: 0},
		{name: "negative gravity", distance: fxp.FromInteger(17), gravity: fxp.NegOne, want: 0},
		// Any gravity but one skips the table entirely: round(sqrt(21.4 x 2 x 17)) = 27.
		{name: "two gravities", distance: fxp.FromInteger(17), gravity: fxp.Two, want: fxp.FromInteger(27)},
		// A fractional distance rounds up to the next whole yard before the table is consulted, so 10.5 reads row 11.
		{name: "fractional distance", distance: fxp.FromStringForced("10.5"), gravity: fxp.One, want: fxp.Fifteen},
		// Well past the table, the formula alone answers: round(sqrt(21.4 x 1000)) = 146.
		{name: "beyond the table", distance: fxp.Thousand, gravity: fxp.One, want: fxp.FromInteger(146)},
	} {
		c.Equal(tc.want, FallingVelocity(tc.distance, tc.gravity), tc.name)
	}
}

// formulaFallingVelocity is the alternative formula from B431, expressed here so the test can show how far the printed
// table strays from it.
func formulaFallingVelocity(yards, gravity float64) fxp.Int {
	return fxp.FromFloat(math.Round(math.Sqrt(21.4 * gravity * yards)))
}

// TestTerminalVelocity verifies that a terminal velocity scales with the square root of the gravity and inversely with
// the square root of the atmospheric pressure, and that a vacuum reports an unlimited fall rather than a velocity.
func TestTerminalVelocity(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name      string
		base      fxp.Int
		gravity   fxp.Int
		pressure  fxp.Int
		want      fxp.Int
		unlimited bool
	}{
		{
			name: "spread-eagle human at 1G and 1atm", base: fxp.Sixty, gravity: fxp.One, pressure: fxp.One,
			want: fxp.Sixty,
		},
		{name: "swan dive at 1G and 1atm", base: fxp.Hundred, gravity: fxp.One, pressure: fxp.One, want: fxp.Hundred},
		{
			name: "twice the gravity", base: fxp.Sixty, gravity: fxp.Two, pressure: fxp.One,
			want: fxp.FromFloat(60 * math.Sqrt(2)),
		},
		{
			name: "half the pressure", base: fxp.Sixty, gravity: fxp.One, pressure: fxp.Half,
			want: fxp.FromFloat(60 / math.Sqrt(0.5)),
		},
		{name: "no base velocity", base: 0, gravity: fxp.One, pressure: fxp.One, want: 0},
		{name: "negative base velocity", base: fxp.NegOne, gravity: fxp.One, pressure: fxp.One, want: 0},
		{name: "no gravity", base: fxp.Sixty, gravity: 0, pressure: fxp.One, want: 0},
		{name: "vacuum", base: fxp.Sixty, gravity: fxp.One, pressure: 0, want: 0, unlimited: true},
		{name: "negative pressure", base: fxp.Sixty, gravity: fxp.One, pressure: fxp.NegOne, want: 0, unlimited: true},
	} {
		velocity, unlimited := TerminalVelocity(tc.base, tc.gravity, tc.pressure)
		c.Equal(tc.want, velocity, tc.name)
		c.Equal(tc.unlimited, unlimited, "%s: unlimited", tc.name)
	}

	// Both of the scaled cases above land on the same value, just under 84.8529 yards per second.
	doubled, _ := TerminalVelocity(fxp.Sixty, fxp.Two, fxp.One)
	thinned, _ := TerminalVelocity(fxp.Sixty, fxp.One, fxp.Half)
	c.Equal(doubled, thinned, "doubling the gravity and halving the pressure scale a fall the same way")
	c.Equal(fxp.FromStringForced("84.8528"), doubled, "60 x sqrt(2), truncated to four decimal places")
}

// TestCollisionDiceCount verifies that the fractional dice count is (HP x velocity)/100 and that a negative HP or
// velocity is treated as zero rather than producing negative damage.
func TestCollisionDiceCount(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name     string
		hp       fxp.Int
		velocity fxp.Int
		want     fxp.Int
	}{
		{name: "10 HP at 19", hp: fxp.Ten, velocity: fxp.Nineteen, want: fxp.FromStringForced("1.9")},
		{name: "20 HP at 19", hp: fxp.Twenty, velocity: fxp.Nineteen, want: fxp.FromStringForced("3.8")},
		{name: "60 HP at 20", hp: fxp.Sixty, velocity: fxp.Twenty, want: fxp.Twelve},
		{name: "1 HP at 1", hp: fxp.One, velocity: fxp.One, want: fxp.OneHundredth},
		{name: "no velocity", hp: fxp.Ten, velocity: 0, want: 0},
		{name: "no HP", hp: 0, velocity: fxp.Nineteen, want: 0},
		{name: "negative HP", hp: fxp.FromInteger(-10), velocity: fxp.Nineteen, want: 0},
		{name: "negative velocity", hp: fxp.Ten, velocity: fxp.FromInteger(-19), want: 0},
		{name: "both negative", hp: fxp.FromInteger(-10), velocity: fxp.FromInteger(-19), want: 0},
	} {
		c.Equal(tc.want, CollisionDiceCount(tc.hp, tc.velocity), tc.name)
	}
}

// TestCollisionDamageDice verifies the rounding of a fractional dice count into rolled dice: the 1d-3 / 1d-2 / 1d-1
// steps below a full die, and rounding to the nearest whole die from there, with a fraction of exactly a half rounding
// up. The dice are compared field by field because formatting them would require the global roller settings.
func TestCollisionDamageDice(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name  string
		count fxp.Int
		want  dice.Dice
	}{
		{
			name: "the B431 worked example", count: fxp.FromStringForced("3.8"),
			want: dice.Dice{Count: 4, Sides: 6, Multiplier: 1},
		},
		{
			name: "exactly a quarter die", count: fxp.Quarter,
			want: dice.Dice{Count: 1, Sides: 6, Modifier: -3, Multiplier: 1},
		},
		{
			name: "a hair over a quarter die", count: fxp.FromStringForced("0.2501"),
			want: dice.Dice{Count: 1, Sides: 6, Modifier: -2, Multiplier: 1},
		},
		{
			name: "exactly half a die", count: fxp.Half,
			want: dice.Dice{Count: 1, Sides: 6, Modifier: -2, Multiplier: 1},
		},
		{
			name: "a hair over half a die", count: fxp.FromStringForced("0.5001"),
			want: dice.Dice{Count: 1, Sides: 6, Modifier: -1, Multiplier: 1},
		},
		{
			name: "just short of a full die", count: fxp.FromStringForced("0.9999"),
			want: dice.Dice{Count: 1, Sides: 6, Modifier: -1, Multiplier: 1},
		},
		{name: "exactly one die", count: fxp.One, want: dice.Dice{Count: 1, Sides: 6, Multiplier: 1}},
		{name: "rounds down", count: fxp.FromStringForced("1.4"), want: dice.Dice{Count: 1, Sides: 6, Multiplier: 1}},
		{name: "half rounds up", count: fxp.OneAndAHalf, want: dice.Dice{Count: 2, Sides: 6, Multiplier: 1}},
		{name: "a dozen dice", count: fxp.Twelve, want: dice.Dice{Count: 12, Sides: 6, Multiplier: 1}},
		{name: "no dice", count: 0, want: dice.Dice{Sides: 6, Multiplier: 1}},
		{name: "negative", count: fxp.NegOne, want: dice.Dice{Sides: 6, Multiplier: 1}},
	} {
		c.Equal(tc.want, CollisionDamageDice(tc.count), tc.name)
	}
}

// TestCollisionVelocity verifies that each collision angle combines the two velocities the way B430 describes, that a
// rear-end collision in which the struck object is the faster of the two closes at zero rather than a negative
// velocity, and that negative velocities are treated as zero.
func TestCollisionVelocity(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name    string
		angle   CollisionAngle
		striker fxp.Int
		struck  fxp.Int
		want    fxp.Int
	}{
		{
			name: "head-on adds", angle: HeadOnCollision, striker: fxp.TwentyFive, struck: fxp.Five,
			want: fxp.Thirty,
		},
		{
			name: "head-on into a stationary object", angle: HeadOnCollision, striker: fxp.TwentyFive, struck: 0,
			want: fxp.TwentyFive,
		},
		{
			name: "rear-end subtracts", angle: RearEndCollision, striker: fxp.TwentyFive, struck: fxp.Five,
			want: fxp.Twenty,
		},
		{
			name: "rear-end at matched speeds", angle: RearEndCollision, striker: fxp.TwentyFive, struck: fxp.TwentyFive,
			want: 0,
		},
		{
			name: "rear-end where the struck object is faster", angle: RearEndCollision, striker: fxp.Five,
			struck: fxp.TwentyFive, want: 0,
		},
		{
			name: "side-on ignores the struck object", angle: SideOnCollision, striker: fxp.TwentyFive,
			struck: fxp.Fifty, want: fxp.TwentyFive,
		},
		{
			name: "side-on into a stationary object", angle: SideOnCollision, striker: fxp.TwentyFive, struck: 0,
			want: fxp.TwentyFive,
		},
		{
			name: "negative velocities clamp", angle: HeadOnCollision, striker: fxp.FromInteger(-25),
			struck: fxp.FromInteger(-5), want: 0,
		},
		{
			name: "a negative struck velocity does not add distance", angle: HeadOnCollision, striker: fxp.TwentyFive,
			struck: fxp.FromInteger(-5), want: fxp.TwentyFive,
		},
	} {
		c.Equal(tc.want, CollisionVelocity(tc.angle, tc.striker, tc.struck), tc.name)
	}
}

// TestCollisionRearEnd verifies the worked example on B430: a 60 HP car moving at 25 yards per second rear-ends a 10 HP
// pedestrian fleeing at 5, so the two meet at 20 yards per second, the car inflicts 12d and the pedestrian 2d. Neither
// side is capped, since the pedestrian's 2d is already well under the car's 12d.
func TestCollisionRearEnd(t *testing.T) {
	c := check.New(t)
	result := Collision(RearEndCollision,
		CollisionObject{HP: fxp.Sixty, Velocity: fxp.TwentyFive},
		CollisionObject{HP: fxp.Ten, Velocity: fxp.Five})
	c.Equal(fxp.Twenty, result.Velocity, "the two meet at 25 - 5 yards per second")
	c.Equal(fxp.Twelve, result.StrikerDice, "the car inflicts (60 x 20)/100 dice")
	c.Equal(fxp.Two, result.StruckDice, "the pedestrian inflicts (10 x 20)/100 dice")
	c.False(result.StrikerCapped, "the striker is never capped in a rear-end collision")
	c.False(result.StruckCapped, "the pedestrian was already under the car's dice")
	c.Equal(dice.Dice{Count: 12, Sides: 6, Multiplier: 1}, CollisionDamageDice(result.StrikerDice), "the car's damage")
	c.Equal(dice.Dice{Count: 2, Sides: 6, Multiplier: 1}, CollisionDamageDice(result.StruckDice),
		"the pedestrian's damage")
}

// TestCollisionCaps verifies that the "cannot inflict more dice than" limits of B430 hold the right side down: in a
// head-on collision the slower object is capped at the faster one's dice (and neither is capped when the two match),
// while in a rear-end or side-on collision the struck object is capped at the striker's. It also verifies that halving
// the damage of a bullet-shaped, sharp or spiked object happens before the cap is applied, so that such a striker
// cannot be out-damaged by an object capped at its unhalved count.
func TestCollisionCaps(t *testing.T) {
	c := check.New(t)

	// Head-on: the heavy, slow object would out-damage the light, fast one, so it is capped at the faster one's dice.
	result := Collision(HeadOnCollision,
		CollisionObject{HP: fxp.Hundred, Velocity: fxp.Five},
		CollisionObject{HP: fxp.Ten, Velocity: fxp.Twenty})
	c.Equal(fxp.TwentyFive, result.Velocity, "head-on velocities add")
	c.Equal(fxp.TwoAndAHalf, result.StrikerDice, "the slower 100 HP striker is held to the faster object's dice")
	c.True(result.StrikerCapped, "the striker was the slower of the two")
	c.Equal(fxp.TwoAndAHalf, result.StruckDice, "the faster 10 HP object inflicts (10 x 25)/100 dice")
	c.False(result.StruckCapped, "the faster object is never capped")

	// The same collision with the roles swapped caps the other side instead.
	result = Collision(HeadOnCollision,
		CollisionObject{HP: fxp.Ten, Velocity: fxp.Twenty},
		CollisionObject{HP: fxp.Hundred, Velocity: fxp.Five})
	c.Equal(fxp.TwoAndAHalf, result.StrikerDice, "the faster 10 HP striker keeps its own dice")
	c.False(result.StrikerCapped, "the faster object is never capped")
	c.Equal(fxp.TwoAndAHalf, result.StruckDice, "the slower 100 HP object is held to the striker's dice")
	c.True(result.StruckCapped, "the struck object was the slower of the two")

	// Head-on at matched speeds: neither is the slower one, so neither is capped.
	result = Collision(HeadOnCollision,
		CollisionObject{HP: fxp.Hundred, Velocity: fxp.Ten},
		CollisionObject{HP: fxp.Ten, Velocity: fxp.Ten})
	c.Equal(fxp.Twenty, result.Velocity, "head-on velocities add")
	c.Equal(fxp.Twenty, result.StrikerDice, "the 100 HP object inflicts (100 x 20)/100 dice")
	c.Equal(fxp.Two, result.StruckDice, "the 10 HP object inflicts (10 x 20)/100 dice")
	c.False(result.StrikerCapped, "matched speeds cap neither side")
	c.False(result.StruckCapped, "matched speeds cap neither side")

	// Side-on into a heavy stationary object: it would out-damage the striker, so it is held to the striker's dice.
	result = Collision(SideOnCollision,
		CollisionObject{HP: fxp.Ten, Velocity: fxp.Twenty},
		CollisionObject{HP: fxp.Hundred, Velocity: 0})
	c.Equal(fxp.Twenty, result.Velocity, "only the striker's velocity counts")
	c.Equal(fxp.Two, result.StrikerDice, "the striker inflicts (10 x 20)/100 dice")
	c.False(result.StrikerCapped, "the striker is never capped in a side-on collision")
	c.Equal(fxp.Two, result.StruckDice, "the stationary object's 20 dice are held to the striker's 2")
	c.True(result.StruckCapped, "the struck object cannot out-damage the striker")

	// Rear-end: the overtaken object is likewise held to the striker's dice.
	result = Collision(RearEndCollision,
		CollisionObject{HP: fxp.Ten, Velocity: fxp.TwentyFive},
		CollisionObject{HP: fxp.Hundred, Velocity: fxp.Five})
	c.Equal(fxp.Twenty, result.Velocity, "the two meet at 25 - 5 yards per second")
	c.Equal(fxp.Two, result.StrikerDice, "the striker inflicts (10 x 20)/100 dice")
	c.Equal(fxp.Two, result.StruckDice, "the overtaken object's 20 dice are held to the striker's 2")
	c.False(result.StrikerCapped, "the striker is never capped in a rear-end collision")
	c.True(result.StruckCapped, "the struck object cannot out-damage the striker")

	// A spiked striker is halved before the cap, so the struck object is held to the halved count, not the full one.
	result = Collision(SideOnCollision,
		CollisionObject{HP: fxp.Sixty, Velocity: fxp.TwentyFive, HalfDamage: true},
		CollisionObject{HP: fxp.Hundred, Velocity: 0})
	c.Equal(fxp.TwentyFive, result.Velocity, "only the striker's velocity counts")
	c.Equal(fxp.FromStringForced("7.5"), result.StrikerDice, "(60 x 25)/100 dice, halved for the spikes")
	c.False(result.StrikerCapped, "the striker is never capped in a side-on collision")
	c.Equal(fxp.FromStringForced("7.5"), result.StruckDice, "the stationary object's 25 dice are held to the halved 7.5")
	c.True(result.StruckCapped, "the struck object cannot out-damage the striker")

	// Halving applies to the struck object too, and can bring it under the cap on its own.
	result = Collision(SideOnCollision,
		CollisionObject{HP: fxp.Sixty, Velocity: fxp.TwentyFive},
		CollisionObject{HP: fxp.Twenty, Velocity: 0, HalfDamage: true})
	c.Equal(fxp.Fifteen, result.StrikerDice, "(60 x 25)/100 dice")
	c.Equal(fxp.TwoAndAHalf, result.StruckDice, "(20 x 25)/100 dice, halved")
	c.False(result.StruckCapped, "the halved count is already under the striker's")
}

// TestImmovableCollisionHP verifies that slamming into an immovable object doubles the mover's HP for the damage
// calculation when the obstacle is hard and leaves it alone when the obstacle is soft, and that negative HP is treated
// as zero.
func TestImmovableCollisionHP(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name string
		hp   fxp.Int
		hard bool
		want fxp.Int
	}{
		{name: "hard obstacle", hp: fxp.Ten, hard: true, want: fxp.Twenty},
		{name: "soft obstacle", hp: fxp.Ten, want: fxp.Ten},
		{name: "no HP against a hard obstacle", hp: 0, hard: true, want: 0},
		{name: "negative HP against a hard obstacle", hp: fxp.FromInteger(-10), hard: true, want: 0},
		{name: "negative HP against a soft obstacle", hp: fxp.FromInteger(-10), want: 0},
	} {
		c.Equal(tc.want, ImmovableCollisionHP(tc.hp, tc.hard), tc.name)
	}

	// The B431 worked example: 10 HP on hard ground at 19 yards per second is (2 x 10 x 19)/100 = 3.8d, which rounds
	// to 4d.
	count := CollisionDiceCount(ImmovableCollisionHP(fxp.Ten, true), FallingVelocity(fxp.FromInteger(17), fxp.One))
	c.Equal(fxp.FromStringForced("3.8"), count, "(2 x 10 x 19)/100 dice")
	c.Equal(dice.Dice{Count: 4, Sides: 6, Multiplier: 1}, CollisionDamageDice(count), "3.8d rounds to 4d")
}

// TestBluntTraumaFromFall verifies that armor which stops falling damage lets one point of injury through per five full
// points stopped, and that a negative amount is treated as zero.
func TestBluntTraumaFromFall(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		stopped fxp.Int
		want    int
	}{
		{stopped: 0, want: 0},
		{stopped: fxp.One, want: 0},
		{stopped: fxp.Two, want: 0},
		{stopped: fxp.Three, want: 0},
		{stopped: fxp.Four, want: 0},
		{stopped: fxp.FromStringForced("4.9999"), want: 0},
		{stopped: fxp.Five, want: 1},
		{stopped: fxp.Nine, want: 1},
		{stopped: fxp.Ten, want: 2},
		{stopped: fxp.FromInteger(24), want: 4},
		{stopped: fxp.FromInteger(25), want: 5},
		{stopped: fxp.NegOne, want: 0},
		{stopped: fxp.FromInteger(-10), want: 0},
	} {
		c.Equal(tc.want, BluntTraumaFromFall(tc.stopped), "%v stopped", tc.stopped)
	}
}
