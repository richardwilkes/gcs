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
	"errors"

	"github.com/richardwilkes/rpgtools/dice"
)

type scriptDice struct{}

// From builds a dice specification from the supplied components. The components arrive as float64 rather than int so
// that intFromScript can narrow them; letting goja narrow them instead made dice.from(1e300, 6) yield "999999d" on
// arm64 and "0" on amd64. The Roller clamps whatever it is given to the configured maxima, so the saturated value it
// receives here lands on the same limit everywhere.
func (d scriptDice) From(count, sides, modifier, multiplier *float64) string {
	var result dice.Dice
	// Account for the dice(sides) shorthand
	if count != nil && sides == nil && modifier == nil && multiplier == nil {
		sides = count
		count = nil
	}
	if sides == nil {
		return ""
	}
	result.Sides = intFromScript(*sides)
	if count != nil {
		result.Count = intFromScript(*count)
	} else {
		result.Count = 1
	}
	if modifier != nil {
		result.Modifier = intFromScript(*modifier)
	}
	if multiplier != nil {
		result.Multiplier = intFromScript(*multiplier)
	} else {
		result.Multiplier = 1
	}
	return Roller.Format(result)
}

func (d scriptDice) Add(left, right string) (string, error) {
	d1 := Roller.Parse(left)
	d2 := Roller.Parse(right)
	if d1.Sides != d2.Sides {
		return "", errors.New("dice sides must match")
	}
	if d1.Multiplier > 0 {
		d1.Count *= d1.Multiplier
		d1.Modifier *= d1.Multiplier
		d1.Multiplier = 1
	}
	if d2.Multiplier > 0 {
		d2.Count *= d2.Multiplier
		d2.Modifier *= d2.Multiplier
		d2.Multiplier = 1
	}
	d1.Count += d2.Count
	d1.Modifier += d2.Modifier
	return Roller.Format(d1), nil
}

func (d scriptDice) Subtract(left, right string) (string, error) {
	d1 := Roller.Parse(left)
	d2 := Roller.Parse(right)
	if d1.Sides != d2.Sides {
		return "", errors.New("dice sides must match")
	}
	if d1.Multiplier > 0 {
		d1.Count *= d1.Multiplier
		d1.Modifier *= d1.Multiplier
		d1.Multiplier = 1
	}
	if d2.Multiplier > 0 {
		d2.Count *= d2.Multiplier
		d2.Modifier *= d2.Multiplier
		d2.Multiplier = 1
	}
	d1.Count -= d2.Count
	d1.Count = max(d1.Count, 0)
	d1.Modifier -= d2.Modifier
	return Roller.Format(d1), nil
}

func (d scriptDice) Count(diceSpec string) int {
	return Roller.Parse(diceSpec).Count
}

func (d scriptDice) Sides(diceSpec string) int {
	return Roller.Parse(diceSpec).Sides
}

func (d scriptDice) Modifier(diceSpec string) int {
	return Roller.Parse(diceSpec).Modifier
}

func (d scriptDice) Multiplier(diceSpec string) int {
	return Roller.Parse(diceSpec).Multiplier
}

func (d scriptDice) Roll(diceSpec string, extraDiceFromModifiers bool) int {
	return Roll(Roller.Parse(diceSpec), extraDiceFromModifiers)
}
