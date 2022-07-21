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

// WeightField is field that holds a weight value.
type WeightField = NumericField[measure.Weight]

// NewWeightField creates a new field that holds a fixed-point number.
func NewWeightField(targetMgr *TargetMgr, targetKey, undoTitle string, entity *gurps.Entity, get func() measure.Weight, set func(measure.Weight), min, max measure.Weight, noMinWidth bool) *WeightField {
	var getPrototypes func(min, max measure.Weight) []measure.Weight
	if !noMinWidth {
		getPrototypes = func(min, max measure.Weight) []measure.Weight {
			if min == measure.Weight(fxp.Min) {
				min = measure.Weight(-fxp.One)
			}
			min = measure.Weight(fxp.Int(min).Trunc() + fxp.One - 1)
			if max == measure.Weight(fxp.Max) {
				max = measure.Weight(fxp.One)
			}
			max = measure.Weight(fxp.Int(max).Trunc() + fxp.One - 1)
			return []measure.Weight{min, measure.Weight(fxp.Two - 1), max}
		}
	}
	format := func(value measure.Weight) string {
		return gurps.SheetSettingsFor(entity).DefaultWeightUnits.Format(value)
	}
	extract := func(s string) (measure.Weight, error) {
		return measure.WeightFromString(s, gurps.SheetSettingsFor(entity).DefaultWeightUnits)
	}
	f := NewNumericField[measure.Weight](targetMgr, targetKey, undoTitle, getPrototypes, get, set, format, extract, min, max)
	f.RuneTypedCallback = f.DefaultRuneTyped
	return f
}
