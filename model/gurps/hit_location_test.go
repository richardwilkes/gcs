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
	"strings"
	"testing"

	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xbytes"
)

// TestHitLocationDRTooltipSummaryNotDuplicated verifies that the "DR n against x attacks" summary block appears exactly
// once in the tooltip for a location nested inside a sub-table. The DR walk recursed into the owning location before
// emitting the summary, and every level of that recursion inserted its own copy.
func TestHitLocationDRTooltipSummaryNotDuplicated(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	hand := NewHitLocation(e, "")
	hand.LocID = "hand"
	hand.ChoiceName = "Hand"
	hand.TableName = "Hand"
	hand.Slots = 1
	hand.DRBonus = 1

	sub := &Body{Roll: dice.Dice{Count: 3, Sides: 6}, Locations: []*HitLocation{hand}}
	sub.Update(e)

	arm := NewHitLocation(e, "")
	arm.LocID = "arm"
	arm.ChoiceName = "Arm"
	arm.TableName = "Arm"
	arm.Slots = 1
	arm.DRBonus = 2
	arm.SetSubTable(sub)

	body := &Body{Roll: dice.Dice{Count: 3, Sides: 6}, Locations: []*HitLocation{arm}}
	body.Update(e)
	e.SheetSettings.BodyType = body

	var tooltip xbytes.InsertBuffer
	drMap := hand.DR(e, &tooltip, nil)
	c.Equal(3, drMap[AllID], "DR accumulates through the owning location")
	c.Equal(1, strings.Count(tooltip.String(), "**DR 3** against"),
		"the summary is inserted once, not once per nesting level")

	// A top-level location still gets exactly one summary.
	tooltip = xbytes.InsertBuffer{}
	drMap = arm.DR(e, &tooltip, nil)
	c.Equal(2, drMap[AllID], "a top-level location only picks up its own bonus")
	c.Equal(1, strings.Count(tooltip.String(), "**DR 2** against"), "a top-level location gets one summary")
}
