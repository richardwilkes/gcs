// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/unison"
)

// DecimalField is a field that holds a decimal (fixed-point) number.
type DecimalField = NumericField[fxp.Int]

// NewDecimalField creates a new field that holds a fixed-point number.
func NewDecimalField(targetMgr *TargetMgr, targetKey, undoTitle string, get func() fxp.Int, set func(fxp.Int), minValue, maxValue fxp.Int, forceSign, noMinWidth bool) *DecimalField {
	var getPrototypes func(minValue, maxValue fxp.Int) []fxp.Int
	if !noMinWidth {
		getPrototypes = fixedPointPrototypes[fxp.Int]
	}
	format := func(value fxp.Int) string {
		if forceSign {
			return value.StringWithSign()
		}
		return value.String()
	}
	return NewNumericField(targetMgr, targetKey, undoTitle, getPrototypes, get, set, format, fxp.FromString,
		minValue, maxValue)
}

// refuseFractions has the field refuse, with a beep, a typed decimal point, and treat text holding a fractional part,
// such as pasted text, as invalid, for a quantity that counts whole things, such as the copies of a container. Text
// that spells a whole number with a decimal point, such as pasted "3.0", is valid.
func refuseFractions(f *DecimalField) {
	extract := f.extract
	f.extract = func(s string) (fxp.Int, error) {
		v, err := extract(s)
		if err == nil && v.Floor() != v {
			return 0, errs.New("not a whole number")
		}
		return v, err
	}
	runeTyped := f.RuneTypedCallback
	f.RuneTypedCallback = func(ch rune) bool {
		if ch == '.' {
			unison.Beep()
			return false
		}
		return runeTyped(ch)
	}
}
