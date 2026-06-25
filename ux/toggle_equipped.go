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

func toggleEquipped(owner Rebuildable, table *unison.Table[*Node[*gurps.Equipment]]) {
	adjustSelection(i18n.Text("Toggle Equipped"), owner, table, equippedExtractor,
		func(e *gurps.Equipment) bool { return e.Equipped },
		func(e *gurps.Equipment, v bool) { e.Equipped = v },
		func(e *gurps.Equipment) { e.Equipped = !e.Equipped },
		true, false)
}
