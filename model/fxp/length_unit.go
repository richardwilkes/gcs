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
		feet := inches.Div(Twelve).Trunc()
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
	case Inch:
		return inches.String() + " " + enum.Key()
	case Feet:
		return inches.Div(Twelve).Comma() + " " + enum.Key()
	case Yard, Meter:
		return inches.Div(ThirtySix).Comma() + " " + enum.Key()
	case Mile:
		return inches.Div(MileInInches).Comma() + " " + enum.Key()
	case Centimeter:
		return inches.Div(ThirtySix).Mul(Hundred).Comma() + " " + enum.Key()
	case Kilometer:
		return inches.Div(ThirtySixThousand).Comma() + " " + enum.Key()
	default:
		return FeetAndInches.Format(length)
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
		return length.Mul(ThirtySix).Div(Hundred)
	case Kilometer:
		return length.Mul(ThirtySixThousand)
	case Meter:
		return length.Mul(ThirtySix)
	default:
		return FeetAndInches.ToInches(length)
	}
}
