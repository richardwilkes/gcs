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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/unison"
)

func usesExtractor(amount int) func(*gurps.Equipment) (*gurps.Equipment, bool) {
	return func(eqp *gurps.Equipment) (*gurps.Equipment, bool) {
		if eqp != nil {
			// Adjust relative to the displayed (capped) value so the action matches what the user sees, even when a
			// feature has reduced the maximum below the stored Uses value.
			if total := eqp.ResolvedUses() + amount; total >= 0 && total <= eqp.ResolvedMaxUses() {
				return eqp, true
			}
		}
		return nil, false
	}
}

func canAdjustUses(table *unison.Table[*Node[*gurps.Equipment]], amount int) bool {
	return canAdjustSelection(table, usesExtractor(amount))
}

func adjustUses(owner Rebuildable, table *unison.Table[*Node[*gurps.Equipment]], amount int) {
	title := increaseUsesAction.Title
	if amount < 0 {
		title = decreaseUsesAction.Title
	}
	adjustSelection(title, owner, table, usesExtractor(amount),
		func(e *gurps.Equipment) int { return e.Uses },
		func(e *gurps.Equipment, v int) { e.Uses = v },
		func(e *gurps.Equipment) { e.Uses = e.ResolvedUses() + amount },
		false, false)
}

func usesResetExtractor(eqp *gurps.Equipment) (*gurps.Equipment, bool) {
	if eqp != nil {
		// Compare against the displayed (capped) value so the action reflects what the user sees.
		if maxUses := eqp.ResolvedMaxUses(); maxUses > 0 && eqp.ResolvedUses() != maxUses {
			return eqp, true
		}
	}
	return nil, false
}

func canResetUsesToMax(table *unison.Table[*Node[*gurps.Equipment]]) bool {
	return canAdjustSelection(table, usesResetExtractor)
}

func resetUsesToMax(owner Rebuildable, table *unison.Table[*Node[*gurps.Equipment]]) {
	adjustSelection(resetUsesToMaxAction.Title, owner, table, usesResetExtractor,
		func(e *gurps.Equipment) int { return e.Uses },
		func(e *gurps.Equipment, v int) { e.Uses = v },
		func(e *gurps.Equipment) { e.Uses = e.ResolvedMaxUses() },
		false, false)
}
