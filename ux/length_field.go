// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
	"github.com/richardwilkes/gcs/v5/model/gurps"
)

// LengthField is field that holds a length value.
type LengthField = NumericField[fxp.Length]

// NewLengthField creates a new field that holds a fixed-point number.
func NewLengthField(targetMgr *TargetMgr, targetKey, undoTitle string, entity *gurps.Entity, get func() fxp.Length, set func(fxp.Length), minValue, maxValue fxp.Length, noMinWidth bool) *LengthField {
	var getPrototypes func(minValue, maxValue fxp.Length) []fxp.Length
	if !noMinWidth {
		getPrototypes = func(minValue, maxValue fxp.Length) []fxp.Length {
			if minValue == fxp.Length(fxp.Min) {
				minValue = fxp.Length(-fxp.One)
			}
			minValue = fxp.Length(fxp.Int(minValue).Trunc() + fxp.One - 1)
			if maxValue == fxp.Length(fxp.Max) {
				maxValue = fxp.Length(fxp.One)
			}
			maxValue = fxp.Length(fxp.Int(maxValue).Trunc() + fxp.One - 1)
			return []fxp.Length{minValue, fxp.Length(fxp.Two - 1), maxValue}
		}
	}
	format := func(value fxp.Length) string {
		return gurps.SheetSettingsFor(entity).DefaultLengthUnits.Format(value)
	}
	extract := func(s string) (fxp.Length, error) {
		return fxp.LengthFromString(s, gurps.SheetSettingsFor(entity).DefaultLengthUnits)
	}
	f := NewNumericField(targetMgr, targetKey, undoTitle, getPrototypes, get, set, format, extract, minValue, maxValue)
	f.RuneTypedCallback = f.DefaultRuneTyped
	return f
}
