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

// addPickSeparately adds the Picked as a Unit check box, checked unless the group is picked from separately, for a group
// in a template that sits in a choice, the only place the flag has an effect.
func addPickSeparately[T gurps.Node[T]](parent *unison.Panel, target T, pickSeparately *bool) {
	if !HasOwner[*Template](parent) || !gurps.PickSeparatelyHasEffect(target) {
		return
	}
	// This panel only fills the space where a label would normally be
	parent.AddChild(unison.NewPanel())
	checkBox := addInvertedCheckBox(parent, i18n.Text("Picked as a Unit"), pickSeparately)
	checkBox.Tooltip = newWrappedTooltip(i18n.Text(
		"When off, the things inside are offered one by one in the choice, listed under this name.",
	))
}
