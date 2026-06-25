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

func disabledExtractor(t *gurps.Trait) (*gurps.Trait, bool) {
	return t, t != nil
}

func canToggleDisabled(table *unison.Table[*Node[*gurps.Trait]]) bool {
	return canAdjustSelection(table, disabledExtractor)
}

func toggleDisabled(owner Rebuildable, table *unison.Table[*Node[*gurps.Trait]]) {
	adjustSelection(i18n.Text("Toggle Enablement"), owner, table, disabledExtractor,
		func(t *gurps.Trait) bool { return t.Disabled },
		func(t *gurps.Trait, v bool) { t.Disabled = v },
		func(t *gurps.Trait) { t.Disabled = !t.Disabled },
		true, true)
}
