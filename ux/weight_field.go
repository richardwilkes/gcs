// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
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

// WeightField is a field that holds a weight value.
type WeightField = NumericField[fxp.Weight]

// NewWeightField creates a new field that holds a weight value, shown in the entity's default weight units.
func NewWeightField(targetMgr *TargetMgr, targetKey, undoTitle string, entity *gurps.Entity, get func() fxp.Weight, set func(fxp.Weight), minValue, maxValue fxp.Weight, noMinWidth bool) *WeightField {
	return newUnitsField(targetMgr, targetKey, undoTitle, get, set,
		func(value fxp.Weight) string { return gurps.SheetSettingsFor(entity).DefaultWeightUnits.Format(value) },
		func(s string) (fxp.Weight, error) {
			return fxp.WeightFromString(s, gurps.SheetSettingsFor(entity).DefaultWeightUnits)
		}, minValue, maxValue, noMinWidth)
}
