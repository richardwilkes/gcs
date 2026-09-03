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

func hiddenExtractor(w *gurps.Weapon) (*gurps.Weapon, bool) {
	return w, w != nil
}

func canToggleHidden(table *unison.Table[*Node[*gurps.Weapon]]) bool {
	return canAdjustSelection(table, hiddenExtractor)
}

// toggleHidden flips the hidden state of each selected weapon. Hiding a weapon only keeps it off the owner's weapon
// lists; it doesn't change what the owner contributes, so the owner is merely marked as modified rather than rebuilt.
func toggleHidden(owner Rebuildable, table *unison.Table[*Node[*gurps.Weapon]]) {
	adjustSelection(i18n.Text("Toggle Hidden"), owner, table, hiddenExtractor,
		func(w *gurps.Weapon) bool { return w.Hide },
		func(w *gurps.Weapon, v bool) { w.Hide = v },
		func(w *gurps.Weapon) { w.Hide = !w.Hide },
		true, false)
}

// adjustHidden sets the hidden state of a single weapon. This is the form the checkmark cell uses, so that a click and
// the command register the same kind of undoable edit and report the change the same way.
func adjustHidden(owner Rebuildable, undoSource unison.Paneler, w *gurps.Weapon, hide bool) {
	adjustTargets(i18n.Text("Toggle Hidden"), owner, undoSource, gurps.EntityFromNode(w), []*gurps.Weapon{w},
		func(w *gurps.Weapon) bool { return w.Hide },
		func(w *gurps.Weapon, v bool) { w.Hide = v },
		func(w *gurps.Weapon) { w.Hide = hide },
		false)
}

func installToggleHiddenHandler(table *unison.Table[*Node[*gurps.Weapon]]) {
	table.InstallCmdHandlers(ToggleStateItemID,
		func(_ any) bool { return canToggleHidden(table) },
		func(_ any) { toggleHidden(table.AncestorOrSelf[Rebuildable](), table) })
}
