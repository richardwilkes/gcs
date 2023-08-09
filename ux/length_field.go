/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package ux

import (
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
)

// LengthField is field that holds a length value.
type LengthField = NumericField[gurps.Length]

// NewLengthField creates a new field that holds a fixed-point number.
func NewLengthField(targetMgr *TargetMgr, targetKey, undoTitle string, entity *gurps.Entity, get func() gurps.Length, set func(gurps.Length), minValue, maxValue gurps.Length, noMinWidth bool) *LengthField {
	var getPrototypes func(minValue, maxValue gurps.Length) []gurps.Length
	if !noMinWidth {
		getPrototypes = func(minValue, maxValue gurps.Length) []gurps.Length {
			if minValue == gurps.Length(fxp.Min) {
				minValue = gurps.Length(-fxp.One)
			}
			minValue = gurps.Length(fxp.Int(minValue).Trunc() + fxp.One - 1)
			if maxValue == gurps.Length(fxp.Max) {
				maxValue = gurps.Length(fxp.One)
			}
			maxValue = gurps.Length(fxp.Int(maxValue).Trunc() + fxp.One - 1)
			return []gurps.Length{minValue, gurps.Length(fxp.Two - 1), maxValue}
		}
	}
	format := func(value gurps.Length) string {
		return gurps.SheetSettingsFor(entity).DefaultLengthUnits.Format(value)
	}
	extract := func(s string) (gurps.Length, error) {
		return gurps.LengthFromString(s, gurps.SheetSettingsFor(entity).DefaultLengthUnits)
	}
	f := NewNumericField[gurps.Length](targetMgr, targetKey, undoTitle, getPrototypes, get, set, format, extract, minValue, maxValue)
	f.RuneTypedCallback = f.DefaultRuneTyped
	return f
}
