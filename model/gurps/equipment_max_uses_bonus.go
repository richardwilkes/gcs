// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"hash"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/equipmentsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Bonus = &EquipmentMaxUsesBonus{}

// EquipmentMaxUsesBonus holds an adjustment to a piece of equipment's maximum uses: that of the equipment it is
// attached to, or of the equipment whose name and tags match its criteria. See MaxUsesModAmount for how the adjustment
// is encoded.
type EquipmentMaxUsesBonus struct {
	maxAdjustmentBonusData[equipmentsel.Type]
}

// NewEquipmentMaxUsesBonus creates a new EquipmentMaxUsesBonus.
func NewEquipmentMaxUsesBonus() *EquipmentMaxUsesBonus {
	return &EquipmentMaxUsesBonus{
		maxAdjustmentBonusData: newMaxAdjustmentBonusData(feature.EquipmentMaxUsesBonus, equipmentsel.ThisEquipment),
	}
}

// Clone implements Feature.
func (e *EquipmentMaxUsesBonus) Clone() Feature {
	return clonePtr(e)
}

// FillWithNameableKeys implements Feature.
func (e *EquipmentMaxUsesBonus) FillWithNameableKeys(m, existing map[string]string) {
	e.fillWithNameableKeysWhen(m, existing, equipmentsel.EquipmentWithName)
}

// Hash writes this object's contents into the hasher.
func (e *EquipmentMaxUsesBonus) Hash(h hash.Hash) {
	if e == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	e.maxAdjustmentBonusData.Hash(h)
}
