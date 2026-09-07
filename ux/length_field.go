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

// LengthField is a field that holds a length value.
type LengthField = NumericField[fxp.Length]

// NewLengthField creates a new field that holds a length value, shown in the entity's default length units.
func NewLengthField(targetMgr *TargetMgr, targetKey, undoTitle string, entity *gurps.Entity, get func() fxp.Length, set func(fxp.Length), minValue, maxValue fxp.Length, noMinWidth bool) *LengthField {
	return newUnitsField(targetMgr, targetKey, undoTitle, get, set,
		func(value fxp.Length) string { return gurps.SheetSettingsFor(entity).DefaultLengthUnits.Format(value) },
		func(s string) (fxp.Length, error) {
			return fxp.LengthFromString(s, gurps.SheetSettingsFor(entity).DefaultLengthUnits)
		}, minValue, maxValue, noMinWidth)
}
