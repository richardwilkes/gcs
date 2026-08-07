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
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/fxp"
)

type scriptMeasurement struct{}

func (s scriptMeasurement) FormatLength(inches float64, toUnits string) string {
	return fxp.ExtractLengthUnit(toUnits).Format(fxp.Length(fxp.FromFloat(inches)))
}

func (s scriptMeasurement) LengthToInches(value float64, fromUnits string) float64 {
	return fxp.AsFloat[float64](fxp.ExtractLengthUnit(fromUnits).ToInches(fxp.FromFloat(value)))
}

func (s scriptMeasurement) StringLengthToInches(str, defaultUnits string) float64 {
	length, err := fxp.LengthFromString(str, fxp.ExtractLengthUnit(defaultUnits))
	if err != nil {
		return 0
	}
	return fxp.AsFloat[float64](fxp.Int(length))
}

func (s scriptMeasurement) FormatWeight(pounds float64, toUnits string) string {
	return fxp.ExtractWeightUnit(toUnits).Format(fxp.Weight(fxp.FromFloat(pounds)))
}

func (s scriptMeasurement) WeightToPounds(value float64, fromUnits string) float64 {
	return fxp.AsFloat[float64](fxp.ExtractWeightUnit(fromUnits).ToPounds(fxp.FromFloat(value)))
}

func (s scriptMeasurement) StringWeightToPounds(str, defaultUnits string) float64 {
	weight, err := fxp.WeightFromString(str, fxp.ExtractWeightUnit(defaultUnits))
	if err != nil {
		return 0
	}
	return fxp.AsFloat[float64](fxp.Int(weight))
}

func (s scriptMeasurement) RangeModifier(yards float64) int {
	return -ssrtInchesToValue(fxp.Yard.ToInches(fxp.FromFloat(yards)), false)
}

func (s scriptMeasurement) SizeModifier(yards float64) int {
	return ssrtInchesToValue(fxp.Yard.ToInches(fxp.FromFloat(yards)), true)
}

func (s scriptMeasurement) Modifier(length float64, units string, forSize bool) int {
	l, err := fxp.LengthFromString(fmt.Sprintf("%v %s", length, units), fxp.Yard)
	if err != nil {
		return 0
	}
	result := ssrtInchesToValue(fxp.Int(l), forSize)
	if !forSize {
		result = -result
	}
	return result
}

func (s scriptMeasurement) ModifierToYards(ssrtValue int) float64 {
	return fxp.AsFloat[float64](ssrtToYards(ssrtValue))
}

func ssrtInchesToValue(inches fxp.Int, allowNegative bool) int {
	yards := inches.Div(fxp.ThirtySix)
	if allowNegative {
		switch {
		case inches <= fxp.One.Div(fxp.Five):
			return -15
		case inches <= fxp.One.Div(fxp.Three):
			return -14
		case inches <= fxp.Half:
			return -13
		case inches <= fxp.Two.Div(fxp.Three):
			return -12
		case inches <= fxp.One:
			return -11
		case inches <= fxp.OneAndAHalf:
			return -10
		case inches <= fxp.Two:
			return -9
		case inches <= fxp.Three:
			return -8
		case inches <= fxp.Five:
			return -7
		case inches <= fxp.Eight:
			return -6
		}
		feet := inches.Div(fxp.Twelve)
		switch {
		case feet <= fxp.One:
			return -5
		case feet <= fxp.OneAndAHalf:
			return -4
		case feet <= fxp.Two:
			return -3
		case yards <= fxp.One:
			return -2
		case yards <= fxp.OneAndAHalf:
			return -1
		}
	}
	if yards <= fxp.Two {
		return 0
	}
	amt := 0
	for yards > fxp.Ten {
		yards = yards.Div(fxp.Ten)
		amt += 6
	}
	switch {
	case yards > fxp.Seven:
		return amt + 4
	case yards > fxp.Five:
		return amt + 3
	case yards > fxp.Three:
		return amt + 2
	case yards > fxp.Two:
		return amt + 1
	case yards > fxp.OneAndAHalf:
		return amt
	default:
		return amt - 1
	}
}

// minSSRTValue and maxSSRTValue bound the Size/Speed/Range Table values ssrtToYards will act on. The table bottoms out
// at -15 (a fifth of an inch) and, above +3, steps up by a factor of ten every six entries, so its yardage grows
// without bound while fxp.Int tops out just above 9.2e14: value 87 is 700,000,000,000,000 yards, and 88 would need
// 1e15.
//
// Both ends are clamped rather than rejected. The low end has always been. The high end must be, because the value
// reaches ssrtToYards straight from a script (measure.modifierToYards) and the loop below runs (value-4)/6 times: an
// unclamped measure.modifierToYards(1e15) would spin here for weeks inside a Go function, where the per-script timeout
// cannot reach it — goja's Interrupt is only honored between VM instructions, so it cannot preempt a host call — and
// would then hand back a value saturated at fxp.Max regardless. With the clamp the loop runs at most 13 times and the
// multiplication can never overflow.
const (
	minSSRTValue = -15
	maxSSRTValue = 87
)

func ssrtToYards(value int) fxp.Int {
	value = min(max(value, minSSRTValue), maxSSRTValue)
	switch value {
	case minSSRTValue:
		return fxp.One.Div(fxp.Five).Div(fxp.ThirtySix)
	case -14:
		return fxp.One.Div(fxp.Three).Div(fxp.ThirtySix)
	case -13:
		return fxp.Half.Div(fxp.ThirtySix)
	case -12:
		return fxp.Two.Div(fxp.Three).Div(fxp.ThirtySix)
	case -11:
		return fxp.One.Div(fxp.ThirtySix)
	case -10:
		return fxp.OneAndAHalf.Div(fxp.ThirtySix)
	case -9:
		return fxp.Two.Div(fxp.ThirtySix)
	case -8:
		return fxp.Three.Div(fxp.ThirtySix)
	case -7:
		return fxp.Five.Div(fxp.ThirtySix)
	case -6:
		return fxp.Eight.Div(fxp.ThirtySix)
	case -5:
		return fxp.One.Div(fxp.Three)
	case -4:
		return fxp.OneAndAHalf.Div(fxp.Three)
	case -3:
		return fxp.Two.Div(fxp.Three)
	case -2:
		return fxp.One
	case -1:
		return fxp.OneAndAHalf
	case 0:
		return fxp.Two
	case 1:
		return fxp.Three
	case 2:
		return fxp.Five
	case 3:
		return fxp.Seven
	}
	value -= 4
	multiplier := fxp.One
	for range value / 6 {
		multiplier = multiplier.Mul(fxp.Ten)
	}
	var v fxp.Int
	switch value % 6 {
	case 0:
		v = fxp.Ten
	case 1:
		v = fxp.Fifteen
	case 2:
		v = fxp.Twenty
	case 3:
		v = fxp.Thirty
	case 4:
		v = fxp.Fifty
	case 5:
		v = fxp.Seventy
	}
	return v.Mul(multiplier)
}
