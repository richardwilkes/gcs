/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package model

import (
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
)

// Prereq holds data necessary to track a prerequisite.
type Prereq interface {
	Nameables
	PrereqType() PrereqType
	// ParentList returns the owning parent list, if any.
	ParentList() *PrereqList
	// Clone creates a new copy of this Prereq.
	Clone(parent *PrereqList) Prereq
	// Satisfied returns true if this Prereq is satisfied by the specified Entity. 'buffer' will be used, if not nil, to
	// write a description of what was unsatisfied. 'prefix' will be appended to each line of the description.
	Satisfied(entity *Entity, exclude any, buffer *xio.ByteBuffer, prefix string, hasEquipmentPenalty *bool) bool
}

// HasText returns the appropriate text for has.
func HasText(has bool) string {
	if has {
		return i18n.Text("Has")
	}
	return i18n.Text("Does not have")
}
