// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import "github.com/richardwilkes/gcs/v5/model/fxp"

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
