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

func traitLevelExtractor(increment bool) func(*gurps.Trait) (*gurps.Trait, bool) {
	return func(t *gurps.Trait) (*gurps.Trait, bool) {
		if t != nil && t.IsLeveled() && (increment || t.Levels > 0) {
			return t, true
		}
		return nil, false
	}
}

func canAdjustTraitLevel(table *unison.Table[*Node[*gurps.Trait]], increment bool) bool {
	return canAdjustSelection(table, traitLevelExtractor(increment))
}

func adjustTraitLevel(owner Rebuildable, table *unison.Table[*Node[*gurps.Trait]], increment bool) {
	title := i18n.Text("Decrement Level")
	if increment {
		title = i18n.Text("Increment Level")
	}
	adjustSelection(title, owner, table, traitLevelExtractor(increment),
		func(t *gurps.Trait) fxp.Int { return t.Levels },
		func(t *gurps.Trait, v fxp.Int) { t.Levels = v },
		func(t *gurps.Trait) {
			original := t.Levels
			levels := original.Floor()
			if increment {
				levels += fxp.One
			} else if original == levels {
				levels -= fxp.One
			}
			t.Levels = levels.Max(0)
		},
		true, false)
}
