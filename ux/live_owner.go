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
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
)

// liveOwner returns the panel that is currently showing the data the given panel was created for. It is the type-erased
// counterpart of liveTable, for callers that only need something to ask for a rebuild through and can't name the row
// type of the table they were handed -- a caller whose type parameter comes from the rows it was given rather than from
// the table those rows live in, for one. An owner that has to alter its set of columns can only do so by replacing its
// table entirely, which leaves any table captured earlier orphaned, and an orphan has no Rebuildable above it, so a
// rebuild asked for through one would silently be skipped. Anything that isn't a table its owner may have replaced is
// returned as-is, as is a nil panel, since callers that may not have a panel at all pass one through here.
func liveOwner(owner unison.Paneler) unison.Paneler {
	if xreflect.IsNil(owner) {
		return owner
	}
	panel := owner.AsPanel()
	if panel == nil || panel.RefKey == "" {
		return owner
	}
	rebuildable, found := panel.ClientData()[TableOwnerClientKey].(Rebuildable)
	if !found || xreflect.IsNil(rebuildable) {
		return owner
	}
	current := rebuildable.AsPanel().FindRefKey(panel.RefKey)
	if current == nil || xreflect.IsNil(current.Self) {
		return owner
	}
	return current.Self
}
