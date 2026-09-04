// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package fxp

import (
	"slices"
	"strings"
)

// lengthUnitsBySuffixLen holds the length units (excluding FeetAndInches, which is parsed specially) ordered by
// descending key length, so that suffix matching always prefers the most specific unit (e.g. "cm" before "m")
// regardless of the order in which the enum is declared.
var lengthUnitsBySuffixLen = func() []LengthUnit {
	units := make([]LengthUnit, 0, len(LengthUnits))
	for _, unit := range LengthUnits {
		if unit != FeetAndInches {
			units = append(units, unit)
		}
	}
	slices.SortStableFunc(units, func(a, b LengthUnit) int { return len(b.Key()) - len(a.Key()) })
	return units
}()

// Format the length for this LengthUnit, showing as many decimal places as the value has.
func (enum LengthUnit) Format(length Length) string {
	return enum.FormatWith(length, NumberFormat{})
}

// FormatWith formats the length for this LengthUnit, rendering the number according to the given NumberFormat. This is
// for display only: the rounding the NumberFormat may perform is lossy, so the result must never be parsed back into a
// stored value.
//
// For FeetAndInches, the rounding is applied to the total inches before they are split into feet and inches, so that a
// remainder which rounds up to a whole foot carries into the feet (e.g. 71.6 inches at zero places is `6'`, not
// `5'12"`). The NumberFormat's padding applies to the inches part when one is shown; a zero inches remainder is still
// omitted, just as it is for Format.
func (enum LengthUnit) FormatWith(length Length, format NumberFormat) string {
	inches := Int(length)
	switch enum {
	case FeetAndInches:
		inches = format.Round(inches)
		// The feet and the inches remainder are split off while the value still carries its sign, and each is then
		// negated on its own. Both are small enough for that, but the value as a whole may not be: Min has no positive
		// counterpart, so Abs() leaves it as it is, and taking the magnitude first rendered it as just "-", which
		// cannot be parsed back.
		negative := inches < 0
		feet := inches.Div(Twelve).Trunc()
		inches -= feet.Mul(Twelve)
		feet = feet.Abs()
		inches = inches.Abs()
		if feet == 0 && inches == 0 {
			return "0'"
		}
		var buffer strings.Builder
		if negative {
			buffer.WriteByte('-')
		}
		if feet > 0 {
			buffer.WriteString(feet.Comma())
			buffer.WriteByte('\'')
		}
		if inches > 0 {
			buffer.WriteString(format.Format(inches))
			buffer.WriteByte('"')
		}
		return buffer.String()

	default:
		return format.Format(enum.FromInches(inches)) + " " + enum.String()
	}
}

// FromInches converts inches to LengthUnit
func (enum LengthUnit) FromInches(inches Int) Int {
	switch enum {
	case Inch:
		return inches
	case Feet:
		return inches.Div(Twelve)
	case Yard:
		return inches.Div(ThirtySix)
	case Mile:
		return inches.Div(MileInInches)
	case Centimeter:
		// using 2.5cm per inch
		return inches.Mul(TwoAndAHalf)
	case Meter:
		// forty = 100 / 2.5 cm per inch
		return inches.Div(Forty)
	case Kilometer:
		// forty = 100 / 2.5 cm per inch
		return inches.Div(Forty).Div(Thousand)
	default:
		return inches
	}
}

// ToInches converts the length in this LengthUnit to inches.
func (enum LengthUnit) ToInches(length Int) Int {
	switch enum {
	case FeetAndInches, Inch:
		return length
	case Feet:
		return length.Mul(Twelve)
	case Yard:
		return length.Mul(ThirtySix)
	case Mile:
		return length.Mul(MileInInches)
	case Centimeter:
		// using 2.5cm per inch
		return length.Div(TwoAndAHalf)
	case Meter:
		// forty = 100 / 2.5cm per inch
		return length.Mul(Forty)
	case Kilometer:
		// forty = 100 / 2.5cm per inch
		return length.Mul(Forty).Mul(Thousand)
	default:
		return FeetAndInches.ToInches(length)
	}
}
