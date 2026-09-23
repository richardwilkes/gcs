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
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/i18n"
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
// PromptForDestination), landing them in the destination's list for the rows' type and applying them there the way any
// rows arriving in that destination are (see applyTransfer). Each destination is applied to independently, so canceling
// one of them leaves the others alone.
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
	editName := fmt.Sprintf(i18n.Text("Insert %s"), sel[0].Data().Kind())
	for _, d := range destinations {
		// The assertion fails for a key that isn't a block key of this destination, since its list then comes back as
		// an untyped nil, and for a destination that hasn't built the list yet, whose list is a typed nil.
		target, ok := d.list(key).(*PageList[T])
		if !ok || target == nil {
			continue
		}
		applyTransfer(target.Table, newApplyParts(newAppendPart(target.Table, sel)),
			applyOptionsFor(table, target.Table), editName)
	}
}
