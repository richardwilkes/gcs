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

package widget

import (
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/measure"
)

// LengthField is field that holds a length value.
type LengthField = NumericField[measure.Length]

// NewLengthField creates a new field that holds a fixed-point number.
func NewLengthField(targetMgr *TargetMgr, targetKey, undoTitle string, entity *gurps.Entity, get func() measure.Length, set func(measure.Length), min, max measure.Length, noMinWidth bool) *LengthField {
	var getPrototypes func(min, max measure.Length) []measure.Length
	if !noMinWidth {
		getPrototypes = func(min, max measure.Length) []measure.Length {
			if min == measure.Length(fxp.Min) {
				min = measure.Length(-fxp.One)
			}
			min = measure.Length(fxp.Int(min).Trunc() + fxp.One - 1)
			if max == measure.Length(fxp.Max) {
				max = measure.Length(fxp.One)
			}
			max = measure.Length(fxp.Int(max).Trunc() + fxp.One - 1)
			return []measure.Length{min, measure.Length(fxp.Two - 1), max}
		}
	}
	format := func(value measure.Length) string {
		return gurps.SheetSettingsFor(entity).DefaultLengthUnits.Format(value)
	}
	extract := func(s string) (measure.Length, error) {
		return measure.LengthFromString(s, gurps.SheetSettingsFor(entity).DefaultLengthUnits)
	}
	f := NewNumericField[measure.Length](targetMgr, targetKey, undoTitle, getPrototypes, get, set, format, extract, min, max)
	f.RuneTypedCallback = f.DefaultRuneTyped
	return f
}
