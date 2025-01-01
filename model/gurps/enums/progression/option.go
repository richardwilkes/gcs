// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package progression

import (
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/xmath"
)

// Thrust returns the thrust damage for the given strength.
func (enum Option) Thrust(strength int) *dice.Dice {
	switch enum {
	case BasicSet:
		if strength < 19 {
			return &dice.Dice{
				Count:      1,
				Sides:      6,
				Modifier:   -(6 - (strength-1)/2),
				Multiplier: 1,
			}
		}
		value := strength - 11
		if strength > 50 {
			value--
			if strength > 79 {
				value -= 1 + (strength-80)/5
			}
		}
		return &dice.Dice{
			Count:      value/8 + 1,
			Sides:      6,
			Modifier:   value%8/2 - 1,
			Multiplier: 1,
		}
	case KnowingYourOwnStrength:
		if strength < 12 {
			return &dice.Dice{
				Count:      1,
				Sides:      6,
				Modifier:   strength - 12,
				Multiplier: 1,
			}
		}
		return &dice.Dice{
			Count:      (strength - 7) / 4,
			Sides:      6,
			Modifier:   (strength+1)%4 - 1,
			Multiplier: 1,
		}
	case NoSchoolGrognardDamage:
		if strength < 11 {
			return &dice.Dice{
				Count:      1,
				Sides:      6,
				Modifier:   -(14 - strength) / 2,
				Multiplier: 1,
			}
		}
		strength -= 11
		return &dice.Dice{
			Count:      strength/8 + 1,
			Sides:      6,
			Modifier:   (strength%8)/2 - 1,
			Multiplier: 1,
		}
	case ThrustEqualsSwingMinus2:
		thr := BasicSet.Swing(strength)
		thr.Modifier -= 2
		return thr
	case SwingEqualsThrustPlus2:
		return BasicSet.Thrust(strength)
	case PhoenixFlameD3:
		if strength < 7 {
			if strength < 1 {
				strength = 1
			}
			return &dice.Dice{
				Count:      1,
				Sides:      6,
				Modifier:   ((strength + 1) / 2) - 7,
				Multiplier: 1,
			}
		}
		if strength < 10 {
			return &dice.Dice{
				Count:      1,
				Sides:      3,
				Modifier:   ((strength + 1) / 2) - 5,
				Multiplier: 1,
			}
		}
		strength -= 8
		return &dice.Dice{
			Count:      strength / 2,
			Sides:      3,
			Modifier:   strength % 2,
			Multiplier: 1,
		}
	case Tbone1:
		if strength < 10 {
			return &dice.Dice{
				Count:    1,
				Sides:    6,
				Modifier: -(6 - (strength+2)/2),
			}
		}
		d := &dice.Dice{
			Count: strength / 10,
			Sides: 6,
		}
		switch strength - (strength/10)*10 {
		case 0, 1:
		case 2, 3:
			d.Modifier = 1
		case 4:
			d.Modifier = -2
			d.Count++
		case 5, 6:
			d.Modifier = 2
		case 7:
			d.Modifier = -1
			d.Count++
		case 8, 9:
			d.Modifier = 3
		}
		return d
	case Tbone1Clean:
		if strength < 10 {
			return Tbone1.Thrust(strength)
		}
		d := &dice.Dice{
			Count: strength / 10,
			Sides: 6,
		}
		switch strength - (strength/10)*10 {
		case 0, 1:
		case 2, 3, 4:
			d.Modifier = 1
		case 5, 6:
			d.Modifier = 2
		case 7, 8, 9:
			d.Modifier = -1
			d.Count++
		}
		return d
	case Tbone2:
		return Tbone2.Swing(int(xmath.Ceil(float64(strength) * 2 / 3)))
	case Tbone2Clean:
		return Tbone2Clean.Swing(int(xmath.Ceil(float64(strength) * 2 / 3)))
	default:
		return BasicSet.Thrust(strength)
	}
}

// Swing returns the swing damage for the given strength.
func (enum Option) Swing(strength int) *dice.Dice {
	switch enum {
	case BasicSet:
		if strength < 10 {
			return &dice.Dice{
				Count:      1,
				Sides:      6,
				Modifier:   -(5 - (strength-1)/2),
				Multiplier: 1,
			}
		}
		if strength < 28 {
			strength -= 9
			return &dice.Dice{
				Count:      strength/4 + 1,
				Sides:      6,
				Modifier:   strength%4 - 1,
				Multiplier: 1,
			}
		}
		value := strength
		if strength > 40 {
			value -= (strength - 40) / 5
		}
		if strength > 59 {
			value++
		}
		value += 9
		return &dice.Dice{
			Count:      value/8 + 1,
			Sides:      6,
			Modifier:   value%8/2 - 1,
			Multiplier: 1,
		}
	case KnowingYourOwnStrength:
		if strength < 10 {
			return &dice.Dice{
				Count:      1,
				Sides:      6,
				Modifier:   strength - 10,
				Multiplier: 1,
			}
		}
		return &dice.Dice{
			Count:      (strength - 5) / 4,
			Sides:      6,
			Modifier:   (strength-1)%4 - 1,
			Multiplier: 1,
		}
	case NoSchoolGrognardDamage:
		return NoSchoolGrognardDamage.Thrust(strength + 3)
	case ThrustEqualsSwingMinus2:
		return BasicSet.Swing(strength)
	case SwingEqualsThrustPlus2:
		sw := BasicSet.Thrust(strength)
		sw.Modifier += 2
		return sw
	case PhoenixFlameD3:
		return PhoenixFlameD3.Thrust(strength)
	case Tbone1:
		return Tbone1.Thrust(int(xmath.Ceil(float64(strength) * 1.5)))
	case Tbone1Clean:
		return Tbone1Clean.Thrust(int(xmath.Ceil(float64(strength) * 1.5)))
	case Tbone2:
		return Tbone1.Thrust(strength)
	case Tbone2Clean:
		return Tbone1Clean.Thrust(strength)
	default:
		return BasicSet.Swing(strength)
	}
}
