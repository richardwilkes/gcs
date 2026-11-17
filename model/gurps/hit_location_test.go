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
	"testing/fstest"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/jio"
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
// DR reports that plus the innate DR from the location itself and from a trait. The falling rules (BX431) treat armor DR
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

// TestHitLocationReservedIDRefusedOnLoad pins that a hit location whose ID is one of the ReservedIDs, which only a
// hand-edited file can hold since the body settings never give one, fails to load, on its own and within a body.
func TestHitLocationReservedIDRefusedOnLoad(t *testing.T) {
	c := check.New(t)
	for _, id := range ReservedIDs {
		loc := `{"id":"` + id + `","choice_name":"X","table_name":"X","slots":1}`
		c.HasError(jio.Unmarshal([]byte(loc), &HitLocation{}), "%s: alone", id)
		data := `{"version":5,"name":"Test","roll":"3d","locations":[` + loc + `]}`
		_, err := NewBodyFromFile(fstest.MapFS{"Test.body": {Data: []byte(data)}}, "Test.body")
		c.HasError(err, "%s: in a body", id)
		if err != nil {
			c.Contains(err.Error(), `"`+id+`" uses a reserved ID`, "%s: names the ID", id)
		}
	}
	c.NoError(jio.Unmarshal([]byte(`{"id":"sm_","choice_name":"X","table_name":"X","slots":1}`), &HitLocation{}),
		"an ID the sanitizer would turn a reserved one into loads")
}

// TestHitLocationSetIDAvoidsReservedIDs verifies that SetID, which the body settings editor uses, never stores one of
// the ReservedIDs.
func TestHitLocationSetIDAvoidsReservedIDs(t *testing.T) {
	c := check.New(t)
	for _, id := range ReservedIDs {
		var loc HitLocation
		loc.SetID(id)
		c.NotEqual(id, loc.LocID, "%s: SetID changes it", id)
	}
}

// armAndHandBody is a body file whose arm, with DR 2, owns a sub-table holding a hand, with DR 1.
const armAndHandBody = `{"version":5,"name":"Test","roll":"3d","locations":[{"id":"arm","choice_name":"Arm",` +
	`"table_name":"Arm","slots":1,"dr_bonus":2,"sub_table":{"roll":"1d","locations":[{"id":"hand",` +
	`"choice_name":"Hand","table_name":"Hand","slots":1,"dr_bonus":1}]}}]}`

// loadTestBody loads a body from the given file content, reporting a failure to load.
func loadTestBody(c check.Checker, data string) *Body {
	c.Helper()
	body, err := NewBodyFromFile(fstest.MapFS{"Test.body": {Data: []byte(data)}}, "Test.body")
	c.NoError(err, "the body loads")
	c.NotNil(body, "a body is returned")
	return body
}

// TestHitLocationSubTableOwnership verifies that a location in a sub-table adds the DR and armor DR of the location
// owning the sub-table to its own, both in a body loaded from a file and in a clone of it, whose sub-tables belong to
// the clone alone.
func TestHitLocationSubTableOwnership(t *testing.T) {
	c := check.New(t)
	body := loadTestBody(c, armAndHandBody)
	if body == nil {
		return
	}
	e := NewEntity()
	addCarriedEquipmentWithFeatures(e, "Bracer", newTestDRBonus(fxp.Three, AllID, "arm"))
	e.Recalculate()
	hand := body.LookupLocationByID(e, "hand")
	c.Equal(1+2+3, hand.DR(e, nil, nil)[AllID], "the hand adds the arm's innate and armor DR")
	c.Equal(3, hand.ArmorDR(e, nil)[AllID], "the hand adds the arm's armor DR")

	clone := body.Clone(e, nil)
	clone.LookupLocationByID(e, "arm").DRBonus = 5
	c.Equal(1+5+3, clone.LookupLocationByID(e, "hand").DR(e, nil, nil)[AllID], "the clone's hand adds the clone's arm")
	c.Equal(1+2+3, hand.DR(e, nil, nil)[AllID], "editing the clone leaves the original alone")
}

// TestHitLocationDisplayDR verifies that DisplayDR lists the DR against all attacks first, even when it is 0, then each
// specialized DR in natural order with the general DR added to it, as the tooltip's summary also does.
func TestHitLocationDisplayDR(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	torso := e.SheetSettings.BodyType.LookupLocationByID(e, TorsoID)
	c.Equal("0", torso.DisplayDR(e, nil), "no DR at all")

	addTraitWithFeatures(e, "Padding", newTestDRBonus(fxp.Three, "crushing", TorsoID))
	e.Recalculate()
	c.Equal("0/3", torso.DisplayDR(e, nil), "only specialized DR")

	addTraitWithFeatures(e, "Toughness", newTestDRBonus(fxp.Two, AllID, TorsoID),
		newTestDRBonus(fxp.One, "burning", TorsoID))
	e.Recalculate()
	var tooltip xbytes.InsertBuffer
	c.Equal("2/3/5", torso.DisplayDR(e, &tooltip), "general and specialized DR")
	c.Contains(tooltip.String(), "**DR 5** against **crushing** attacks", "the summary adds the general DR too")
	// The DR comes from a map, so an unsorted display could match by chance; repeat to make that vanishingly unlikely.
	for range 20 {
		if got := torso.DisplayDR(e, nil); got != "2/3/5" {
			c.Equal("2/3/5", got, "the specialized DR is in natural order")
			break
		}
	}
}
