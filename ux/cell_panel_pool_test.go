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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// TestCellPanelPoolLetsGoOfPanelsNotUsedLately verifies that the rows of a list keep the panels of their cells only for
// as long as they are among those built or used in the last two turns of the table's pool, so that a list scrolled from
// end to end holds no more of them than that.
func TestCellPanelPoolLetsGoOfPanelsNotUsedLately(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(func() { registerActions() })
	uxtest.SwapForTest(t, &cellPanelPoolTurn, 2)
	traits := make([]*gurps.Trait, 5)
	for i := range traits {
		traits[i] = newFilterTestTrait("Trait")
	}
	d := NewTraitTableDockable("test"+gurps.TraitsExt, traits)
	col := d.table.ColumnIndexForID(gurps.TraitDescriptionColumn)
	rows := make([]*Node[*gurps.Trait], len(traits))
	panels := make([]unison.Paneler, len(traits))
	cell := func(i int) unison.Paneler {
		return rows[i].ColumnCell(i, col, unison.Black, unison.White, false, false, false)
	}
	held := func(i int) bool {
		kept := rows[i].cells[col].fitted[0]
		return kept != nil && kept.Panel == panels[i]
	}
	for i := range rows {
		rows[i] = d.table.RowFromIndex(i)
		panels[i] = cell(i)
	}
	c.False(held(0) || held(1), "the panels of the turn before last must be let go")
	c.True(held(2) && held(3) && held(4), "the panels of the last two turns must be kept")

	c.True(cell(2) == panels[2], "a panel still kept must be the one handed back")
	c.True(held(2), "a panel used in the turn just ended must be kept")
	c.False(held(3), "a panel not used since the turn before must be let go")
}
