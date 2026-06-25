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
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

func quantityExtractor(increment bool) func(*gurps.Equipment) (*gurps.Equipment, bool) {
	return func(eqp *gurps.Equipment) (*gurps.Equipment, bool) {
		if eqp != nil && (increment || eqp.Quantity > 0) {
			return eqp, true
		}
		return nil, false
	}
}

func canAdjustQuantity(table *unison.Table[*Node[*gurps.Equipment]], increment bool) bool {
	return canAdjustSelection(table, quantityExtractor(increment))
}

func adjustQuantity(owner Rebuildable, table *unison.Table[*Node[*gurps.Equipment]], increment bool) {
	title := i18n.Text("Decrement Quantity")
	if increment {
		title = i18n.Text("Increment Quantity")
	}
	adjustSelection(title, owner, table, quantityExtractor(increment),
		func(e *gurps.Equipment) fxp.Int { return e.Quantity },
		func(e *gurps.Equipment, v fxp.Int) { e.Quantity = v },
		func(e *gurps.Equipment) {
			original := e.Quantity
			qty := original.Floor()
			if increment {
				qty += fxp.One
			} else if original == qty {
				qty -= fxp.One
			}
			e.Quantity = qty.Max(0)
		},
		true, false)
}
