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

// WeightField is field that holds a weight value.
type WeightField = NumericField[model.Weight]

// NewWeightField creates a new field that holds a fixed-point number.
func NewWeightField(targetMgr *TargetMgr, targetKey, undoTitle string, entity *model.Entity, get func() model.Weight, set func(model.Weight), min, max model.Weight, noMinWidth bool) *WeightField {
	var getPrototypes func(min, max model.Weight) []model.Weight
	if !noMinWidth {
		getPrototypes = func(min, max model.Weight) []model.Weight {
			if min == model.Weight(fxp.Min) {
				min = model.Weight(-fxp.One)
			}
			min = model.Weight(fxp.Int(min).Trunc() + fxp.One - 1)
			if max == model.Weight(fxp.Max) {
				max = model.Weight(fxp.One)
			}
			max = model.Weight(fxp.Int(max).Trunc() + fxp.One - 1)
			return []model.Weight{min, model.Weight(fxp.Two - 1), max}
		}
	}
	format := func(value model.Weight) string {
		return model.SheetSettingsFor(entity).DefaultWeightUnits.Format(value)
	}
	extract := func(s string) (model.Weight, error) {
		return model.WeightFromString(s, model.SheetSettingsFor(entity).DefaultWeightUnits)
	}
	f := NewNumericField[model.Weight](targetMgr, targetKey, undoTitle, getPrototypes, get, set, format, extract, min, max)
	f.RuneTypedCallback = f.DefaultRuneTyped
	return f
}
