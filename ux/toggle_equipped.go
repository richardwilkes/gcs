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
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

func equippedExtractor(eqp *gurps.Equipment) (*gurps.Equipment, bool) {
	return eqp, eqp != nil
}

func canToggleEquipped(table *unison.Table[*Node[*gurps.Equipment]]) bool {
	return canAdjustSelection(table, equippedExtractor)
}

// toggleEquipped flips the equipped state of each selected piece of equipment. The owner is rebuilt rather than merely
// marked as modified, since an item that starts or stops contributing takes its weapons, reactions and conditional
// modifiers into or out of play with it, and whether the lists showing those appear on the page at all -- along with
// which columns they hold -- is decided only when the owner creates its lists.
func toggleEquipped(owner Rebuildable, table *unison.Table[*Node[*gurps.Equipment]]) {
	adjustSelection(i18n.Text("Toggle Equipped"), owner, table, equippedExtractor,
		func(e *gurps.Equipment) bool { return e.Equipped },
		func(e *gurps.Equipment, v bool) { e.Equipped = v },
		func(e *gurps.Equipment) { e.Equipped = !e.Equipped },
		true, true)
}

// adjustEquipped sets the equipped state of a single piece of equipment. This is the form the checkmark cell uses, so
// that a click and the command register the same kind of undoable edit and, as above, rebuild the owner.
func adjustEquipped(owner Rebuildable, undoSource unison.Paneler, eqp *gurps.Equipment, equipped bool) {
	adjustTargets(i18n.Text("Toggle Equipped"), owner, undoSource, gurps.EntityFromNode(eqp),
		[]*gurps.Equipment{eqp},
		func(e *gurps.Equipment) bool { return e.Equipped },
		func(e *gurps.Equipment, v bool) { e.Equipped = v },
		func(e *gurps.Equipment) { e.Equipped = equipped },
		true)
}
