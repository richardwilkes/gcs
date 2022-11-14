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
	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/gcs/v5/model/fxp"
)

// LengthField is field that holds a length value.
type LengthField = NumericField[model.Length]

// NewLengthField creates a new field that holds a fixed-point number.
func NewLengthField(targetMgr *TargetMgr, targetKey, undoTitle string, entity *model.Entity, get func() model.Length, set func(model.Length), min, max model.Length, noMinWidth bool) *LengthField {
	var getPrototypes func(min, max model.Length) []model.Length
	if !noMinWidth {
		getPrototypes = func(min, max model.Length) []model.Length {
			if min == model.Length(fxp.Min) {
				min = model.Length(-fxp.One)
			}
			min = model.Length(fxp.Int(min).Trunc() + fxp.One - 1)
			if max == model.Length(fxp.Max) {
				max = model.Length(fxp.One)
			}
			max = model.Length(fxp.Int(max).Trunc() + fxp.One - 1)
			return []model.Length{min, model.Length(fxp.Two - 1), max}
		}
	}
	format := func(value model.Length) string {
		return model.SheetSettingsFor(entity).DefaultLengthUnits.Format(value)
	}
	extract := func(s string) (model.Length, error) {
		return model.LengthFromString(s, model.SheetSettingsFor(entity).DefaultLengthUnits)
	}
	f := NewNumericField[model.Length](targetMgr, targetKey, undoTitle, getPrototypes, get, set, format, extract, min, max)
	f.RuneTypedCallback = f.DefaultRuneTyped
	return f
}
