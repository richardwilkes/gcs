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

// WeightField is field that holds a weight value.
type WeightField = NumericField[gurps.Weight]

// NewWeightField creates a new field that holds a fixed-point number.
func NewWeightField(targetMgr *TargetMgr, targetKey, undoTitle string, entity *gurps.Entity, get func() gurps.Weight, set func(gurps.Weight), min, max gurps.Weight, noMinWidth bool) *WeightField {
	var getPrototypes func(min, max gurps.Weight) []gurps.Weight
	if !noMinWidth {
		getPrototypes = func(min, max gurps.Weight) []gurps.Weight {
			if min == gurps.Weight(fxp.Min) {
				min = gurps.Weight(-fxp.One)
			}
			min = gurps.Weight(fxp.Int(min).Trunc() + fxp.One - 1)
			if max == gurps.Weight(fxp.Max) {
				max = gurps.Weight(fxp.One)
			}
			max = gurps.Weight(fxp.Int(max).Trunc() + fxp.One - 1)
			return []gurps.Weight{min, gurps.Weight(fxp.Two - 1), max}
		}
	}
	format := func(value gurps.Weight) string {
		return gurps.SheetSettingsFor(entity).DefaultWeightUnits.Format(value)
	}
	extract := func(s string) (gurps.Weight, error) {
		return gurps.WeightFromString(s, gurps.SheetSettingsFor(entity).DefaultWeightUnits)
	}
	f := NewNumericField[gurps.Weight](targetMgr, targetKey, undoTitle, getPrototypes, get, set, format, extract, min, max)
	f.RuneTypedCallback = f.DefaultRuneTyped
	return f
}
