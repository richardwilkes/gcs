/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package attribute

import (
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/i18n"
)

// Tooltip returns the tooltip for the DamageProgression.
func (enum DamageProgression) Tooltip() string {
	tooltip := i18n.Text("Determines the method used to calculate thrust and swing damage")
	if footnote := enum.AltString(); footnote != "" {
		return tooltip + ".\n" + footnote
	}
	return tooltip
}

// Thrust returns the thrust damage for the given strength.
func (enum DamageProgression) Thrust(strength int) *dice.Dice {
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
	default:
		return BasicSet.Thrust(strength)
	}
}

// Swing returns the swing damage for the given strength.
func (enum DamageProgression) Swing(strength int) *dice.Dice {
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
	default:
		return BasicSet.Swing(strength)
	}
}
