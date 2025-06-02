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
	"fmt"
	"strings"

	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/errs"
	"golang.org/x/exp/constraints"
)

// Length contains a fixed-point value in inches. Conversions to/from metric are done using the simplified Length metric
// conversion of 1 yd = 1 meter. For consistency, all metric lengths are converted to meters, then to yards, rather than
// the variations at different lengths that the Length rules suggest.
type Length Int

// LengthFromInteger creates a new Length.
func LengthFromInteger[T constraints.Integer](value T, unit LengthUnit) Length {
	return Length(unit.ToInches(From(value)))
}

// LengthFromFixed creates a new Length.
func LengthFromFixed(value Int, unit LengthUnit) Length {
	return Length(unit.ToInches(value))
}

// LengthFromStringForced creates a new Length. May have any of the known Units suffixes, a feet and inches format (e.g.
// 6'2"), or no notation at all, in which case defaultUnits is used.
func LengthFromStringForced(text string, defaultUnits LengthUnit) Length {
	length, err := LengthFromString(text, defaultUnits)
	if err != nil {
		return 0
	}
	return length
}

// LengthFromString creates a new Length. May have any of the known Units suffixes, a feet and inches format (e.g.
// 6'2"), or no notation at all, in which case defaultUnits is used.
func LengthFromString(text string, defaultUnits LengthUnit) (Length, error) {
	text = strings.TrimLeft(strings.TrimSpace(text), "+")
	for _, unit := range LengthUnits[1:] {
		if strings.HasSuffix(text, unit.Key()) {
			value, err := FromString(strings.TrimSpace(strings.TrimSuffix(text, unit.Key())))
			if err != nil {
				return 0, err
			}
			return Length(unit.ToInches(value)), nil
		}
	}
	// Didn't match any of the Units types, let's try feet & inches
	feetIndex := strings.Index(text, "'")
	inchIndex := strings.Index(text, `"`)
	if feetIndex == -1 && inchIndex == -1 {
		// Nope, so let's use our passed-in default units
		value, err := FromString(strings.TrimSpace(text))
		if err != nil {
			return 0, err
		}
		return Length(defaultUnits.ToInches(value)), nil
	}
	var feet, inches Int
	var err error
	if feetIndex != -1 {
		s := strings.TrimSpace(text[:feetIndex])
		feet, err = FromString(s)
		if err != nil {
			return 0, err
		}
	}
	if inchIndex != -1 {
		if feetIndex > inchIndex {
			return 0, errs.New(fmt.Sprintf("invalid format: %s", text))
		}
		s := strings.TrimSpace(text[feetIndex+1 : inchIndex])
		inches, err = FromString(s)
		if err != nil {
			return 0, err
		}
	}
	return Length(feet.Mul(Twelve) + inches), nil
}

func (l Length) String() string {
	return FeetAndInches.Format(l)
}

// MarshalJSON implements json.Marshaler.
func (l Length) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (l *Length) UnmarshalJSON(in []byte) error {
	var s string
	if err := json.Unmarshal(in, &s); err != nil {
		return err
	}
	var err error
	*l, err = LengthFromString(s, Inch)
	return err
}
