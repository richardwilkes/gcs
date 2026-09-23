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

// copyDestination is what copySelectionTo asks of a sheet or template it copies rows onto: the page list for a block
// key, so the type of the rows being copied can pick the list they land in without naming the destination's fields.
// Both *Sheet and *Template satisfy it.
type copyDestination interface {
	FileBackedDockable
	list(key string) sheetList
}

// blockKeyForRow returns the key of the block that holds rows of the given type on a sheet or template, or "" when rows
// of that type can't be copied onto one. Equipment goes to the carried list; a sheet's other-equipment list only takes
// rows by drag and drop or by the move commands.
func blockKeyForRow(data any) string {
	switch data.(type) {
	case *gurps.Trait:
		return gurps.BlockTraitsKey
	case *gurps.Skill:
		return gurps.BlockSkillsKey
	case *gurps.Spell:
		return gurps.BlockSpellsKey
	case *gurps.Equipment:
		return gurps.BlockEquipmentKey
	case *gurps.Note:
		return gurps.BlockNotesKey
	default:
		return ""
	}
}

// canCopySelectionTo returns true if the table has a selection whose rows can be copied onto a sheet or template and
// there is at least one destination to copy them to.
func canCopySelectionTo[T gurps.Node[T], D copyDestination](table *unison.Table[*Node[T]], destinations []D) bool {
	var t T
	return table.HasSelection() && len(destinations) > 0 && blockKeyForRow(t) != ""
}

// copySelectionTo copies the table's selected rows onto each of the destinations the user picks from those given (see
// PromptForDestination), landing them in the destination's list for the rows' type and resolving them the way a drop
// onto a sheet would (see processCopiedRows).
func copySelectionTo[T gurps.Node[T], D copyDestination](table *unison.Table[*Node[T]], destinations []D) {
	if !table.HasSelection() {
		return
	}
	destinations = PromptForDestination(destinations)
	if len(destinations) == 0 {
		return
	}
	sel := table.SelectedRows(true)
	key := blockKeyForRow(sel[0].Data())
	for _, d := range destinations {
		// The assertion fails for a key that isn't a block key of this destination, since its list then comes back as
		// an untyped nil, and for a destination that hasn't built the list yet, whose list is a typed nil.
		target, ok := d.list(key).(*PageList[T])
		if !ok || target == nil {
			continue
		}
		// All processing must happen inside the postProcessor so it is captured by the undo edit's after-state
		// (CopyRowsTo records that after the postProcessor runs); otherwise redo would not restore the resolved tech
		// levels, nameables, or the merged points.
		CopyRowsTo(target.Table, sel, func(rows []*Node[T]) {
			target.provider.ProcessDropData(nil, target.Table)
			processCopiedRows(table, target.Table)
			maybeClearPreconfiguredFlag(target.Table, rows)
		}, true)
	}
}

// processCopiedRows resolves the just-copied, currently-selected rows of a sheet's or template's table the same way a
// drop onto one does: prompting for the modifiers and nameables of rows that arrived from somewhere other than a
// sheet, then folding the points of rows that duplicate ones already present into those rows. Does nothing when the
// destination isn't a character sheet, loot sheet or template.
func processCopiedRows[T gurps.Node[T]](source, target *unison.Table[*Node[T]]) {
	if shouldProcessModifiersAndNameablesTo(target) {
		if shouldProcessModifiersAndNameablesFrom(source) {
			// Answering the modifier prompt rebuilds the owner, and that rebuild can replace the table underneath us:
			// only enabled modifiers count toward a row having switchable features, so toggling one can add or take
			// away the switch column, and a list can only change its columns by building a new table. An orphaned table
			// has no Rebuildable above it and reports its own rows as selected rather than the ones the user is now
			// looking at, both of which the steps below depend upon. Applying nameable substitutions rebuilds as well,
			// so look it up again afterwards too.
			ProcessModifiersForSelection(target, true)
			target = liveTable(target)
			ProcessNameablesForSelection(target, true)
			target = liveTable(target)
		}
		// The copy always adds rows to a different sheet, so merge points into identical existing rows even when
		// copying from another sheet.
		MergeAddedRows(target)
	}
}
