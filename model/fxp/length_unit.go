// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package fxp

import (
	"strings"
)

// Format the length for this LengthUnit.
func (enum LengthUnit) Format(length Length) string {
	inches := Int(length)
	switch enum {
	case FeetAndInches:
		feet := inches.Div(Twelve).Floor()
		inches -= feet.Mul(Twelve)
		if feet == 0 && inches == 0 {
			return "0'"
		}
		var buffer strings.Builder
		if feet > 0 {
			buffer.WriteString(feet.Comma())
			buffer.WriteByte('\'')
		}
		if inches > 0 {
			buffer.WriteString(inches.String())
			buffer.WriteByte('"')
		}
		return buffer.String()

	default:
		return enum.FromInches(inches).Comma() + " " + enum.String()
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
