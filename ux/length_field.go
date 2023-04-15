/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
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
func NewLengthField(targetMgr *TargetMgr, targetKey, undoTitle string, entity *gurps.Entity, get func() gurps.Length, set func(gurps.Length), min, max gurps.Length, noMinWidth bool) *LengthField {
	var getPrototypes func(min, max gurps.Length) []gurps.Length
	if !noMinWidth {
		getPrototypes = func(min, max gurps.Length) []gurps.Length {
			if min == gurps.Length(fxp.Min) {
				min = gurps.Length(-fxp.One)
			}
			min = gurps.Length(fxp.Int(min).Trunc() + fxp.One - 1)
			if max == gurps.Length(fxp.Max) {
				max = gurps.Length(fxp.One)
			}
			max = gurps.Length(fxp.Int(max).Trunc() + fxp.One - 1)
			return []gurps.Length{min, gurps.Length(fxp.Two - 1), max}
		}
	}
	format := func(value gurps.Length) string {
		return gurps.SheetSettingsFor(entity).DefaultLengthUnits.Format(value)
	}
	extract := func(s string) (gurps.Length, error) {
		return gurps.LengthFromString(s, gurps.SheetSettingsFor(entity).DefaultLengthUnits)
	}
	f := NewNumericField[gurps.Length](targetMgr, targetKey, undoTitle, getPrototypes, get, set, format, extract, min, max)
	f.RuneTypedCallback = f.DefaultRuneTyped
	return f
}
