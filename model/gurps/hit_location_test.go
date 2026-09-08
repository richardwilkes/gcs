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

	"github.com/richardwilkes/gcs/v5/model/fxp"
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

// TestHitLocationArmorDRSeparatedFromInnate verifies that ArmorDR reports only the DR granted by worn equipment, while
// DR reports that plus the innate DR from the location itself and from a trait. The falling rules (B431) treat armor DR
// as flexible for blunt trauma and innate DR not at all, so the two have to be distinguishable.
func TestHitLocationArmorDRSeparatedFromInnate(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addCarriedEquipmentWithFeatures(e, "Mail Hauberk", newTestDRBonus(fxp.Three, AllID, TorsoID))
	addTraitWithFeatures(e, "Damage Resistance", newTestDRBonus(fxp.Two, AllID, TorsoID))
	e.Recalculate()

	torso := e.SheetSettings.BodyType.LookupLocationByID(e, TorsoID)
	c.NotNil(torso, "the default body has a torso")
	c.Equal(0, torso.DRBonus, "the default torso has no innate DR of its own")

	c.Equal(5, torso.DR(e, nil, nil)[AllID], "DR combines the armor's 3 with the trait's 2")
	c.Equal(3, torso.ArmorDR(e, nil)[AllID], "ArmorDR reports only the armor's 3")

	// A hit location's own DR bonus is innate, so it must not show up in the armor total either.
	torso.DRBonus = 4
	c.Equal(9, torso.DR(e, nil, nil)[AllID], "DR picks up the location's own innate DR")
	c.Equal(3, torso.ArmorDR(e, nil)[AllID], "ArmorDR still reports only the armor's 3")

	// Innate DR is what remains once the armor's share is removed, key by key.
	full := torso.DR(e, nil, nil)
	armor := torso.ArmorDR(e, nil)
	c.Equal(6, full[AllID]-armor[AllID], "innate DR is DR minus armor DR")

	// Equipment that isn't equipped grants no DR at all, so both totals drop by its contribution.
	addCarriedEquipmentWithFeatures(e, "Stowed Helm", newTestDRBonus(fxp.Eight, AllID, TorsoID)).Equipped = false
	e.Recalculate()
	c.Equal(9, torso.DR(e, nil, nil)[AllID], "unequipped armor contributes nothing to DR")
	c.Equal(3, torso.ArmorDR(e, nil)[AllID], "unequipped armor contributes nothing to armor DR")

	// A specialized armor bonus lands on its own key in both maps, leaving the shared keys aligned.
	addCarriedEquipmentWithFeatures(e, "Padding", newTestDRBonus(fxp.Five, "crushing", TorsoID))
	e.Recalculate()
	c.Equal(5, torso.DR(e, nil, nil)["crushing"], "the specialized bonus keeps its own key")
	c.Equal(5, torso.ArmorDR(e, nil)["crushing"], "the specialized bonus is armor, so it appears in both maps")
}
