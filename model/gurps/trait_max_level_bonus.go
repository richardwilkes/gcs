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

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/traitsel"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Bonus = &TraitMaxLevelBonus{}

// TraitMaxLevelBonus holds an adjustment to a trait's maximum level: that of the trait it is attached to, or of the
// traits whose name and tags match its criteria. See MaxUsesModAmount for how the adjustment is encoded.
type TraitMaxLevelBonus struct {
	maxAdjustmentBonusData[traitsel.Type]
}

// NewTraitMaxLevelBonus creates a new TraitMaxLevelBonus.
func NewTraitMaxLevelBonus() *TraitMaxLevelBonus {
	return &TraitMaxLevelBonus{
		maxAdjustmentBonusData: newMaxAdjustmentBonusData(feature.TraitMaxLevelBonus, traitsel.ThisTrait),
	}
}

// Clone implements Feature.
func (t *TraitMaxLevelBonus) Clone() Feature {
	return clonePtr(t)
}

// FillWithNameableKeys implements Feature.
func (t *TraitMaxLevelBonus) FillWithNameableKeys(m, existing map[string]string) {
	t.fillWithNameableKeysWhen(m, existing, traitsel.TraitWithName)
}

// Hash writes this object's contents into the hasher.
func (t *TraitMaxLevelBonus) Hash(h hash.Hash) {
	if t == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	t.maxAdjustmentBonusData.Hash(h)
}
