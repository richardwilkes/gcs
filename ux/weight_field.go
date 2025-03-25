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

// WeightField is field that holds a weight value.
type WeightField = NumericField[fxp.Weight]

// NewWeightField creates a new field that holds a fixed-point number.
func NewWeightField(targetMgr *TargetMgr, targetKey, undoTitle string, entity *gurps.Entity, get func() fxp.Weight, set func(fxp.Weight), minValue, maxValue fxp.Weight, noMinWidth bool) *WeightField {
	var getPrototypes func(minValue, maxValue fxp.Weight) []fxp.Weight
	if !noMinWidth {
		getPrototypes = func(minValue, maxValue fxp.Weight) []fxp.Weight {
			if minValue == fxp.Weight(fxp.Min) {
				minValue = fxp.Weight(-fxp.One)
			}
			minValue = fxp.Weight(fxp.Int(minValue).Trunc() + fxp.One - 1)
			if maxValue == fxp.Weight(fxp.Max) {
				maxValue = fxp.Weight(fxp.One)
			}
			maxValue = fxp.Weight(fxp.Int(maxValue).Trunc() + fxp.One - 1)
			return []fxp.Weight{minValue, fxp.Weight(fxp.Two - 1), maxValue}
		}
	}
	format := func(value fxp.Weight) string {
		return gurps.SheetSettingsFor(entity).DefaultWeightUnits.Format(value)
	}
	extract := func(s string) (fxp.Weight, error) {
		return fxp.WeightFromString(s, gurps.SheetSettingsFor(entity).DefaultWeightUnits)
	}
	f := NewNumericField(targetMgr, targetKey, undoTitle, getPrototypes, get, set, format, extract, minValue, maxValue)
	f.RuneTypedCallback = f.DefaultRuneTyped
	return f
}
