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

// WeightField is field that holds a weight value.
type WeightField = NumericField[gurps.Weight]

// NewWeightField creates a new field that holds a fixed-point number.
func NewWeightField(targetMgr *TargetMgr, targetKey, undoTitle string, entity *gurps.Entity, get func() gurps.Weight, set func(gurps.Weight), minValue, maxValue gurps.Weight, noMinWidth bool) *WeightField {
	var getPrototypes func(minValue, maxValue gurps.Weight) []gurps.Weight
	if !noMinWidth {
		getPrototypes = func(minValue, maxValue gurps.Weight) []gurps.Weight {
			if minValue == gurps.Weight(fxp.Min) {
				minValue = gurps.Weight(-fxp.One)
			}
			minValue = gurps.Weight(fxp.Int(minValue).Trunc() + fxp.One - 1)
			if maxValue == gurps.Weight(fxp.Max) {
				maxValue = gurps.Weight(fxp.One)
			}
			maxValue = gurps.Weight(fxp.Int(maxValue).Trunc() + fxp.One - 1)
			return []gurps.Weight{minValue, gurps.Weight(fxp.Two - 1), maxValue}
		}
	}
	format := func(value gurps.Weight) string {
		return gurps.SheetSettingsFor(entity).DefaultWeightUnits.Format(value)
	}
	extract := func(s string) (gurps.Weight, error) {
		return gurps.WeightFromString(s, gurps.SheetSettingsFor(entity).DefaultWeightUnits)
	}
	f := NewNumericField[gurps.Weight](targetMgr, targetKey, undoTitle, getPrototypes, get, set, format, extract, minValue, maxValue)
	f.RuneTypedCallback = f.DefaultRuneTyped
	return f
}
