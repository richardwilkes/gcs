// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestContainedQuantityCountsWhatGroupsHold verifies that a contained quantity prerequisite counts what a group holds,
// in nested groups too, rather than the group itself, while a physical container inside counts as the pieces of it
// there are.
func TestContainedQuantityCountsWhatGroupsHold(t *testing.T) {
	c := check.New(t)
	quiver := NewEquipment(nil, nil, true)
	broadheads := NewEquipmentGroup(nil, quiver)
	arrows := NewEquipment(nil, broadheads, false)
	arrows.Quantity = fxp.FromInteger(30)
	spares := NewEquipmentGroup(nil, broadheads)
	shafts := NewEquipment(nil, spares, false)
	shafts.Quantity = fxp.FromInteger(5)
	spares.Children = []*Equipment{shafts}
	broadheads.Children = []*Equipment{arrows, spares}
	case1 := NewEquipment(nil, quiver, true)
	case1.Quantity = fxp.FromInteger(2)
	case1.Children = []*Equipment{NewEquipment(nil, case1, false)}
	quiver.Children = []*Equipment{broadheads, case1}
	c.Equal(fxp.FromInteger(37), containedQuantity(quiver.Children))

	p := NewContainedQuantityPrereq()
	p.QualifierCriteria.Qualifier = fxp.FromInteger(20)
	c.False(p.Satisfied(nil, quiver, nil, "", nil), "37 pieces must not pass a limit of 20")
	p.QualifierCriteria.Qualifier = fxp.FromInteger(40)
	c.True(p.Satisfied(nil, quiver, nil, "", nil), "37 pieces must pass a limit of 40")
}
