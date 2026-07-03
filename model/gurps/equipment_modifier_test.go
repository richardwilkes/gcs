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
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// TestEquipmentModifierSyncHashIncludesPerLevelAndPerPoundFlags verifies that the sync hash reflects the
// CostIsPerLevel, CostIsPerPound, and WeightIsPerLevel flags. These flags change the cost/weight math without
// necessarily changing CostAmount/WeightAmount, so if they were omitted from the hash (as they originally were),
// library-sync change detection would report a modified library modifier as Matched and never update the local copy.
// The TID is deliberately not part of the hash (sync matches by TID separately), so two modifiers differing only in a
// single flag must produce different hashes.
func TestEquipmentModifierSyncHashIncludesPerLevelAndPerPoundFlags(t *testing.T) {
	c := check.New(t)

	// hashWith builds a non-container modifier with a fixed cost and weight, applies the mutation, and returns its
	// sync hash.
	hashWith := func(mutate func(e *EquipmentModifier)) uint64 {
		e := NewEquipmentModifier(nil, nil, false)
		e.CostAmount = "-60%"
		e.WeightAmount = "2 lb"
		if mutate != nil {
			mutate(e)
		}
		return Hash64(e)
	}

	baseHash := hashWith(nil)

	// A flat "-60%" turning into "-60% per level" flips only CostIsPerLevel; the hash must change.
	c.NotEqual(baseHash, hashWith(func(e *EquipmentModifier) { e.CostIsPerLevel = true }),
		"CostIsPerLevel must affect the sync hash")
	c.NotEqual(baseHash, hashWith(func(e *EquipmentModifier) { e.CostIsPerPound = true }),
		"CostIsPerPound must affect the sync hash")
	c.NotEqual(baseHash, hashWith(func(e *EquipmentModifier) { e.WeightIsPerLevel = true }),
		"WeightIsPerLevel must affect the sync hash")
}
