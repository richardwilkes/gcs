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
	"strings"

	"github.com/richardwilkes/toolbox/v2/fixed/fixed64"
)

// NumberFormat describes how a fixed-point value should be rendered for display. The zero value renders values
// exactly as Int.Comma() does, so it is safe to use anywhere no formatting preference has been expressed.
//
// This is a display-only concern: nothing produced by a NumberFormat should ever be parsed back into a stored value,
// since the rounding it performs is lossy.
type NumberFormat struct {
	Places       DecimalPlace `json:"decimal_places,omitzero"`
	PadWithZeros bool         `json:"pad_with_zeros,omitzero"`
}

// EnsureValid returns a copy with any unknown Places replaced by the default.
func (f NumberFormat) EnsureValid() NumberFormat {
	f.Places = f.Places.EnsureValid()
	return f
}

// places returns the number of decimal places this format rounds and pads to, or -1 when every decimal place the value
// has is to be kept, which is what AsNeeded asks for and what an unrecognized Places is treated as. This is the one
// place the DecimalPlace choices are mapped to counts. A count is capped at the number of decimal places an Int can
// hold, so that a choice beyond that neither rounds nor pads with zeros for precision the value cannot have.
func (f NumberFormat) places() int {
	var places int
	switch f.Places {
	case ZeroPlaces:
		places = 0
	case OnePlace:
		places = 1
	case TwoPlaces:
		places = 2
	case ThreePlaces:
		places = 3
	case FourPlaces:
		places = 4
	default:
		return -1
	}
	return min(places, fixed64.MaxDecimalDigits[DP]())
}

// roundingUnit returns the raw fixed-point unit that Round works in, or 0 if no rounding is to be done. The unit is
// derived from the fixed-point multiplier rather than written out, so that it stays right if the number of decimal
// places an Int holds is ever changed.
func (f NumberFormat) roundingUnit() int64 {
	places := f.places()
	if places < 0 || places >= fixed64.MaxDecimalDigits[DP]() {
		// Every decimal place the value has is kept: either that is what was asked for, or the value cannot hold more
		// than was asked for.
		return 0
	}
	unit := fixed64.Multiplier[DP]()
	for range places {
		unit /= 10
	}
	return unit
}

// Round returns the value rounded, half away from zero, to this format's number of decimal places. AsNeeded returns
// the value unchanged, as does a choice of as many decimal places as an Int holds, since there is nothing to drop.
// Values whose rounded result would not be representable are returned unchanged rather than overflowing.
func (f NumberFormat) Round(value Int) Int {
	unit := f.roundingUnit()
	if unit == 0 {
		return value
	}
	raw := int64(value)
	half := unit / 2
	if raw >= 0 {
		if raw > int64(Max)-half {
			return value
		}
		return Int((raw + half) / unit * unit)
	}
	if raw < int64(Min)+half {
		return value
	}
	return Int((raw - half) / unit * unit)
}

// Format returns the value rounded to this format's number of decimal places and rendered with thousands separators.
// When PadWithZeros is set and an explicit, non-zero number of decimal places has been chosen, the fractional part is
// padded with trailing zeros to that many places, so that e.g. 7.5 becomes "7.50" at two places. Padding has nothing
// to add at zero places, and AsNeeded shows only the places the value has, so PadWithZeros has no effect for either.
func (f NumberFormat) Format(value Int) string {
	text := f.Round(value).Comma()
	places := f.places()
	if !f.PadWithZeros || places < 1 {
		return text
	}
	existing := 0
	if i := strings.IndexByte(text, '.'); i != -1 {
		existing = len(text) - i - 1
	} else {
		text += "."
	}
	if existing < places {
		text += strings.Repeat("0", places-existing)
	}
	return text
}
