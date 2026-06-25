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
	"github.com/richardwilkes/unison"
)

func equipmentLevelExtractor(amount fxp.Int) func(*gurps.Equipment) (*gurps.Equipment, bool) {
	return func(eqp *gurps.Equipment) (*gurps.Equipment, bool) {
		if eqp != nil && (amount > 0 || eqp.Level > 0) {
			return eqp, true
		}
		return nil, false
	}
}

func canAdjustEquipmentLevel(table *unison.Table[*Node[*gurps.Equipment]], amount fxp.Int) bool {
	return canAdjustSelection(table, equipmentLevelExtractor(amount))
}

func adjustEquipmentLevel(owner Rebuildable, table *unison.Table[*Node[*gurps.Equipment]], amount fxp.Int) {
	title := increaseEquipmentLevelAction.Title
	if amount < 0 {
		title = decreaseEquipmentLevelAction.Title
	}
	adjustSelection(title, owner, table, equipmentLevelExtractor(amount),
		func(e *gurps.Equipment) fxp.Int { return e.Level },
		func(e *gurps.Equipment, v fxp.Int) { e.Level = v.Max(0) },
		func(e *gurps.Equipment) { e.Level = (e.Level + amount).Max(0) },
		true, false)
}
