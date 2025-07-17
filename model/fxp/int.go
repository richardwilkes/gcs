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
	"time"

	"github.com/richardwilkes/toolbox/v2/fixed"
	"github.com/richardwilkes/toolbox/v2/fixed/fixed64"
	"golang.org/x/exp/constraints"
)

// Common values that can be reused.
var (
	Min                 = fixed64.Minimum[DP]()
	NegPointEight       = FromStringForced("-0.8")
	Twentieth           = FromStringForced("0.05")
	PointZeroSix        = FromStringForced("0.06")
	PointZeroSeven      = FromStringForced("0.07")
	PointZeroEight      = FromStringForced("0.08")
	PointZeroNine       = FromStringForced("0.09")
	Tenth               = FromStringForced("0.1")
	PointOneTwo         = FromStringForced("0.12")
	Eighth              = FromStringForced("0.125")
	PointOneFive        = FromStringForced("0.15")
	Fifth               = FromStringForced("0.2")
	Quarter             = FromStringForced("0.25")
	ThreeTenths         = FromStringForced("0.3")
	TwoFifths           = FromStringForced("0.4")
	Half                = FromStringForced("0.5")
	ThreeFifths         = FromStringForced("0.6")
	SevenTenths         = FromStringForced("0.7")
	ThreeQuarters       = FromStringForced("0.75")
	FourFifths          = FromStringForced("0.8")
	One                 = FromInteger(1)
	OnePointOne         = FromStringForced("1.1")
	OnePointTwo         = FromStringForced("1.2")
	OneAndAQuarter      = FromStringForced("1.25")
	OneAndAHalf         = FromStringForced("1.5")
	Two                 = FromInteger(2)
	TwoAndAHalf         = FromStringForced("2.5")
	Three               = FromInteger(3)
	ThreeAndAHalf       = FromStringForced("3.5")
	Four                = FromInteger(4)
	Five                = FromInteger(5)
	Six                 = FromInteger(6)
	Seven               = FromInteger(7)
	Eight               = FromInteger(8)
	Nine                = FromInteger(9)
	Ten                 = FromInteger(10)
	Eleven              = FromInteger(11)
	Twelve              = FromInteger(12)
	Thirteen            = FromInteger(13)
	Fifteen             = FromInteger(15)
	Sixteen             = FromInteger(16)
	Nineteen            = FromInteger(19)
	Twenty              = FromInteger(20)
	TwentyFour          = FromInteger(24)
	TwentyFive          = FromInteger(25)
	ThirtySix           = FromInteger(36)
	Thirty              = FromInteger(30)
	Forty               = FromInteger(40)
	Fifty               = FromInteger(50)
	Sixty               = FromInteger(60)
	Seventy             = FromInteger(70)
	Eighty              = FromInteger(80)
	NinetyNine          = FromInteger(99)
	Hundred             = FromInteger(100)
	OneHundredFifty     = FromInteger(150)
	FiveHundred         = FromInteger(500)
	SixHundred          = FromInteger(600)
	ThousandMinusOne    = FromInteger(999)
	Thousand            = FromInteger(1000)
	TwoThousand         = FromInteger(2000)
	ThirtySixHundred    = FromInteger(3600)
	TenThousandMinusOne = FromInteger(9999)
	TenThousand         = FromInteger(10000)
	ThirtySixThousand   = FromInteger(36000)
	MileInInches        = FromInteger(63360)
	MillionMinusOne     = FromInteger(999999)
	TenMillionMinusOne  = FromInteger(9999999)
	BillionMinusOne     = FromInteger(999999999)
	MaxBasePoints       = MillionMinusOne
	Max                 = fixed64.Maximum[DP]()
)

// DP is an alias for the fixed-point decimal places configuration we are using.
type DP = fixed.D4

// Int is an alias for the fixed-point type we are using.
type Int = fixed64.Int[DP]

// FromInteger creates an Int from a numeric value.
func FromInteger[T constraints.Integer](value T) Int {
	return fixed64.FromInteger[DP](value)
}

// FromFloat creates an Int from a numeric value.
func FromFloat[T constraints.Float](value T) Int {
	return fixed64.FromFloat[DP](value)
}

// FromString creates an Int from a string.
func FromString(value string) (Int, error) {
	return fixed64.FromString[DP](value)
}

// FromStringForced creates an Int from a string, ignoring any conversion inaccuracies.
func FromStringForced(value string) Int {
	return fixed64.FromStringForced[DP](value)
}

// AsInteger returns the equivalent value in the destination type.
func AsInteger[T constraints.Integer](value Int) T {
	return fixed64.AsInteger[DP, T](value)
}

// AsFloat returns the equivalent value in the destination type.
func AsFloat[T constraints.Float](value Int) T {
	return fixed64.AsFloat[DP, T](value)
}

// ApplyRounding rounds in the positive direction if roundDown is false, or in the negative direction if roundDown is
// true.
func ApplyRounding(value Int, roundDown bool) Int {
	if truncated := value.Floor(); value != truncated {
		if roundDown {
			if value < 0 {
				return truncated - One
			}
		} else {
			if value > 0 {
				return truncated + One
			}
		}
		return truncated
	}
	return value
}

// ResetIfOutOfRange checks the value and if it is lower than minValue or greater than maxValue, returns defValue,
// otherwise returns value.
func ResetIfOutOfRange[T constraints.Integer | constraints.Float | Int](value, minValue, maxValue, defValue T) T {
	if value < minValue || value > maxValue {
		return defValue
	}
	return value
}

// Extract a leading value from a string. If a value is found, it is returned along with the portion of the string that
// was unused. If a value is not found, then 0 is returned along with the original string.
func Extract(in string) (value Int, remainder string) {
	last := 0
	maximum := len(in)
	if last < maximum && in[last] == ' ' {
		last++
	}
	if last >= maximum {
		return 0, in
	}
	ch := in[last]
	found := false
	decimal := false
	start := last
	for (start == last && (ch == '-' || ch == '+')) || ch == ',' || (!decimal && ch == '.') || (ch >= '0' && ch <= '9') {
		if ch >= '0' && ch <= '9' {
			found = true
		}
		if ch == '.' {
			decimal = true
		}
		last++
		if last >= maximum {
			break
		}
		ch = in[last]
	}
	if !found {
		return 0, in
	}
	var err error
	if value, err = FromString(in[start:last]); err != nil {
		return 0, in
	}
	return value, in[last:]
}

// SecondsToDuration converts a fixed-point value in seconds to a time.Duration.
func SecondsToDuration(value Int) time.Duration {
	return time.Duration(AsInteger[int64](value.Mul(Thousand))) * time.Millisecond
}

// IntLessFromString compares two strings as Ints.
func IntLessFromString(a, b string) bool {
	return FromStringForced(a) < FromStringForced(b)
}
