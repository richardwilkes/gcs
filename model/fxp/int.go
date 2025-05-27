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

	"github.com/richardwilkes/toolbox/xmath"
	"github.com/richardwilkes/toolbox/xmath/fixed"
	"github.com/richardwilkes/toolbox/xmath/fixed/f64"
)

// Common values that can be reused.
var (
	Min                 = Int(f64.Min)
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
	One                 = From(1)
	OnePointOne         = FromStringForced("1.1")
	OnePointTwo         = FromStringForced("1.2")
	OneAndAQuarter      = FromStringForced("1.25")
	OneAndAHalf         = FromStringForced("1.5")
	Two                 = From(2)
	TwoAndAHalf         = FromStringForced("2.5")
	Three               = From(3)
	ThreeAndAHalf       = FromStringForced("3.5")
	Four                = From(4)
	Five                = From(5)
	Six                 = From(6)
	Seven               = From(7)
	Eight               = From(8)
	Nine                = From(9)
	Ten                 = From(10)
	Eleven              = From(11)
	Twelve              = From(12)
	Thirteen            = From(13)
	Fifteen             = From(15)
	Sixteen             = From(16)
	Nineteen            = From(19)
	Twenty              = From(20)
	TwentyFour          = From(24)
	TwentyFive          = From(25)
	ThirtySix           = From(36)
	Thirty              = From(30)
	Forty               = From(40)
	Fifty               = From(50)
	Sixty               = From(60)
	Seventy             = From(70)
	Eighty              = From(80)
	NinetyNine          = From(99)
	Hundred             = From(100)
	OneHundredFifty     = From(150)
	FiveHundred         = From(500)
	SixHundred          = From(600)
	ThousandMinusOne    = From(999)
	Thousand            = From(1000)
	TwoThousand         = From(2000)
	ThirtySixHundred    = From(3600)
	TenThousandMinusOne = From(9999)
	ThirtySixThousand   = From(36000)
	MileInInches        = From(63360)
	MillionMinusOne     = From(999999)
	TenMillionMinusOne  = From(9999999)
	BillionMinusOne     = From(999999999)
	MaxBasePoints       = MillionMinusOne
	Max                 = Int(f64.Max)
	// MaxSafeMultiply returns the maximum value that can be safely multiplied without overflow.
	MaxSafeMultiply = f64.MaxSafeMultiply[DP]()
)

// DP is an alias for the fixed-point decimal places configuration we are using.
type DP = fixed.D4

// Int is an alias for the fixed-point type we are using.
type Int = f64.Int[DP]

// From creates an Int from a numeric value.
func From[T xmath.Numeric](value T) Int {
	return f64.From[DP](value)
}

// FromString creates an Int from a string.
func FromString(value string) (Int, error) {
	return f64.FromString[DP](value)
}

// FromStringForced creates an Int from a string, ignoring any conversion inaccuracies.
func FromStringForced(value string) Int {
	return f64.FromStringForced[DP](value)
}

// As returns the equivalent value in the destination type.
func As[T xmath.Numeric](value Int) T {
	return f64.As[DP, T](value)
}

// ApplyRounding rounds in the positive direction if roundDown is false, or in the negative direction if roundDown is
// true.
func ApplyRounding(value Int, roundDown bool) Int {
	if truncated := value.Trunc(); value != truncated {
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
func ResetIfOutOfRange[T xmath.Numeric | Int](value, minValue, maxValue, defValue T) T {
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
	return time.Duration(As[int64](value.Mul(Thousand))) * time.Millisecond
}

// IntLessFromString compares two strings as Ints.
func IntLessFromString(a, b string) bool {
	return FromStringForced(a) < FromStringForced(b)
}
