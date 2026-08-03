// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package emcost

import (
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
)

const (
	// multiplicationSign is the Unicode multiplication sign, which users may type in place of the ASCII "x".
	multiplicationSign = "×"
	// multiplierLeaders holds every rune that may lead a multiplier value. FromString lowercases before classifying
	// and also accepts the Unicode multiplication sign, so extraction must strip all of these forms.
	multiplierLeaders = "xX" + multiplicationSign
)

// Format returns a formatted version of the value.
func (enum Value) Format(value fxp.Int) string {
	switch enum {
	case Addition:
		return value.CommaWithSign()
	case Percentage:
		return value.CommaWithSign() + enum.String()
	case Multiplier:
		if value <= 0 {
			value = fxp.One
		}
		return enum.String() + value.Comma()
	case CostFactor:
		return value.CommaWithSign() + " " + enum.String()
	default:
		return Addition.Format(value)
	}
}

// ExtractValue from the string.
func (enum Value) ExtractValue(s string) fxp.Int {
	v, _ := fxp.Extract(strings.TrimLeft(strings.TrimSpace(s), multiplierLeaders))
	if enum.EnsureValid() == Multiplier && v <= 0 {
		v = fxp.One
	}
	return v
}

// FromString examines a string to determine what type it is.
func (enum Value) FromString(s string) Value {
	s = strings.ToLower(strings.TrimSpace(s))
	switch {
	case strings.HasSuffix(s, CostFactor.Key()):
		return CostFactor
	case strings.HasSuffix(s, Percentage.Key()):
		return Percentage
	case strings.HasPrefix(s, Multiplier.Key()) || strings.HasSuffix(s, Multiplier.Key()) ||
		strings.HasPrefix(s, multiplicationSign) || strings.HasSuffix(s, multiplicationSign):
		return Multiplier
	default:
		return Addition
	}
}
