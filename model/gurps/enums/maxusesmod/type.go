// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package maxusesmod

import (
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
)

// FromString examines a string to determine which Type of adjustment it represents. A trailing "%" indicates a
// percentage, a leading or trailing "x" indicates a multiplier, and anything else is a plain addition.
func FromString(s string) Type {
	s = strings.ToLower(strings.TrimSpace(s))
	switch {
	case strings.HasSuffix(s, Percentage.Key()):
		return Percentage
	case strings.HasPrefix(s, Multiplier.Key()) || strings.HasSuffix(s, Multiplier.Key()):
		return Multiplier
	default:
		return Addition
	}
}

// ExtractValue extracts the numeric value from the string, interpreting it according to this Type. A non-positive
// multiplier is treated as 1.
func (enum Type) ExtractValue(s string) fxp.Int {
	v, _ := fxp.Extract(strings.TrimLeft(strings.TrimSpace(s), Multiplier.Key()))
	if enum.EnsureValid() == Multiplier && v <= 0 {
		v = fxp.One
	}
	return v
}

// Format returns a normalized string representation of the given value for this Type.
func (enum Type) Format(value fxp.Int) string {
	switch enum.EnsureValid() {
	case Percentage:
		return value.CommaWithSign() + enum.String()
	case Multiplier:
		if value <= 0 {
			value = fxp.One
		}
		return enum.String() + value.Comma()
	default: // Addition
		return value.CommaWithSign()
	}
}

// Normalize returns the canonical string form of the given input.
func Normalize(s string) string {
	t := FromString(s)
	return t.Format(t.ExtractValue(s))
}
