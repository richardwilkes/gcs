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
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xbytes"
)

// TestDRBonusFillWithNameableKeys verifies that a DR bonus contributes the markers in its specialization to the
// nameable key map, so that the owning item's "Set Substitutions" prompt asks for the damage type.
func TestDRBonusFillWithNameableKeys(t *testing.T) {
	c := check.New(t)

	m := make(map[string]string)
	newTestDRBonus(fxp.One, "@Damage Type@", TorsoID).FillWithNameableKeys(m, nil)
	c.Equal(map[string]string{"Damage Type": nameable.Unset}, m,
		"a marker in the specialization becomes an unanswered nameable key")

	m = make(map[string]string)
	newTestDRBonus(fxp.One, AllID, TorsoID).FillWithNameableKeys(m, nil)
	c.Equal(map[string]string{}, m, "a specialization without markers contributes no nameable keys")
}

// TestDRBonusSpecializationWithReplacements verifies that the specialization resolves through the owning item's
// replacements and is normalized the same way the stored value is, and that an unowned bonus (or one whose marker
// hasn't been answered) keeps the marker visible rather than silently resolving to "all".
func TestDRBonusSpecializationWithReplacements(t *testing.T) {
	c := check.New(t)
	bonus := newTestDRBonus(fxp.Three, "@Damage Type@", TorsoID)
	_, trait := newTestEntityWithDRTrait(bonus, map[string]string{"Damage Type": "Cold"})
	c.Equal("Cold", bonus.SpecializationWithReplacements(), "the owning trait's replacement resolves the marker")

	c.Equal("@Damage Type@", newTestDRBonus(fxp.Three, "@Damage Type@", TorsoID).SpecializationWithReplacements(),
		"an unowned bonus leaves the marker standing")

	trait.Replacements = map[string]string{"Damage Type": ""}
	c.Equal(AllID, bonus.SpecializationWithReplacements(), "an empty replacement normalizes to 'all'")

	trait.Replacements = map[string]string{"Damage Type": "All"}
	c.Equal(AllID, bonus.SpecializationWithReplacements(), "a replacement of 'All' normalizes to 'all'")
}

// TestEntityDRBonusSpecializationUsesReplacements verifies that the DR map an entity builds is keyed by the resolved
// damage type, and that an unanswered marker shows up as its own key rather than being folded into "all".
func TestEntityDRBonusSpecializationUsesReplacements(t *testing.T) {
	c := check.New(t)
	e, _ := newTestEntityWithDRTrait(newTestDRBonus(fxp.Three, "@Damage Type@", TorsoID),
		map[string]string{"Damage Type": "Cold"})

	drMap := e.AddDRBonusesFor(TorsoID, nil, nil)
	c.Equal(3, drMap["cold"], "the DR lands under the resolved damage type")
	c.Equal(0, drMap[AllID], "nothing lands under 'all'")

	// A second trait carrying the same marker, but with no replacement to answer it, must remain distinguishable.
	unresolved := NewTrait(e, nil, false)
	unresolved.Name = "Damage Resistance"
	unresolved.Features = Features{newTestDRBonus(fxp.Two, "@Damage Type@", TorsoID)}
	e.Traits = append(e.Traits, unresolved)
	e.Recalculate()

	drMap = e.AddDRBonusesFor(TorsoID, nil, nil)
	c.Equal(3, drMap["cold"], "the resolved bonus is unaffected by the unresolved one")
	c.Equal(2, drMap["@damage type@"], "an unanswered marker keys the map visibly rather than becoming 'all'")
	c.Equal(0, drMap[AllID], "an unanswered marker doesn't become DR against everything")
}

// TestDRBonusTooltipUsesResolvedSpecialization verifies that the tooltip generated while collecting DR names the
// resolved damage type rather than the raw marker.
func TestDRBonusTooltipUsesResolvedSpecialization(t *testing.T) {
	c := check.New(t)
	e, _ := newTestEntityWithDRTrait(newTestDRBonus(fxp.Three, "@Damage Type@", TorsoID),
		map[string]string{"Damage Type": "Cold"})

	var tooltip xbytes.InsertBuffer
	e.AddDRBonusesFor(TorsoID, &tooltip, nil)
	c.Contains(tooltip.String(), "against Cold attacks", "the tooltip names the resolved damage type")
}

// TestThisArmorDRBonusResolvesSpecialization verifies that the synthetic copy made for a "this armor" DR bonus (one
// that names no locations) still resolves its specialization through the owning equipment's replacements.
func TestThisArmorDRBonusResolvesSpecialization(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	eqp := NewEquipment(e, nil, false)
	eqp.Name = "Cloak of Winter"
	eqp.Replacements = map[string]string{"Damage Type": "Cold"}
	eqp.Features = Features{
		newTestDRBonus(fxp.Four, AllID, TorsoID),
		newTestDRBonus(fxp.One, "@Damage Type@"), // no locations, i.e. "this armor"
	}
	e.CarriedEquipment = append(e.CarriedEquipment, eqp)
	e.Recalculate()

	drMap := e.AddDRBonusesFor(TorsoID, nil, nil)
	c.Equal(4, drMap[AllID], "torso: the located bonus grants DR against everything")
	c.Equal(1, drMap["cold"], "torso: the 'this armor' copy resolves its marker through the equipment")
}

// newTestEntityWithDRTrait creates an entity holding a single non-container trait that carries the given DR bonus and
// the given nameable replacements, then recalculates so the bonus is collected and its owner set. Both the entity and
// the trait are returned, since callers need the trait to alter its replacements.
func newTestEntityWithDRTrait(bonus *DRBonus, replacements map[string]string) (*Entity, *Trait) {
	e := NewEntity()
	trait := NewTrait(e, nil, false)
	trait.Name = "Damage Resistance"
	trait.Replacements = replacements
	trait.Features = Features{bonus}
	e.Traits = append(e.Traits, trait)
	e.Recalculate()
	return e, trait
}
