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
	"strconv"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selector"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestEntitySelfControlOverride verifies that overriding a trait's self-control roll changes the Merchant penalty that
// the self-control machinery generates during processFeatures — not just the displayed roll. This exercises the
// deferred generation that waits for every selector override to be collected first.
func TestEntitySelfControlOverride(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	trait := NewTrait(e, nil, false)
	trait.Name = "Greed"
	trait.SelfControl = selfctrl.CR12
	trait.SelfControlAdj = selfctrl.MajorCostOfLivingIncrease
	e.Traits = append(e.Traits, trait)
	e.Recalculate()
	c.Equal(fxp.FromInteger(selfctrl.CR12.Penalty()), e.SkillBonusFor("Merchant", "", "", nil, nil),
		"Merchant penalty derives from CR12")

	// Override Greed's self-control roll to CR6; the generated penalty must track the overridden roll.
	override := NewSelectorOverride(selector.TraitSelfControlRoll)
	override.Value = strconv.Itoa(int(selfctrl.CR6))
	override.NameCriteria.Qualifier = "Greed"
	trait.Features = append(trait.Features, override)
	e.Recalculate()
	c.Equal(fxp.FromInteger(selfctrl.CR6.Penalty()), e.SkillBonusFor("Merchant", "", "", nil, nil),
		"Merchant penalty tracks the overridden CR6 roll")
}

// TestEntityEquipmentPrereqPenaltyOptionalSpecialization verifies that the penalty generated for a skill whose
// equipment prerequisite is unmet is scoped to that skill and doesn't bleed onto sibling skills that differ only in
// their optional specialization.
func TestEntityEquipmentPrereqPenaltyOptionalSpecialization(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	e.Skills = append(e.Skills, newSkillNeedingEquipment(e, "AK-47", ""))
	e.Recalculate()
	c.Equal(-fxp.Five, e.SkillBonusFor("Guns", "Longarm", "AK-47", nil, nil),
		"the skill carrying the unmet equipment prereq is penalized")
	c.Equal(fxp.Int(0), e.SkillBonusFor("Guns", "Longarm", "M16", nil, nil),
		"a sibling differing only in optional specialization is not penalized")
	c.Equal(fxp.Int(0), e.SkillBonusFor("Guns", "Longarm", "", nil, nil),
		"a sibling without an optional specialization is not penalized")

	// A skill with no optional specialization must still be penalized, and must not penalize one that has it.
	e = NewEntity()
	e.Skills = append(e.Skills, newSkillNeedingEquipment(e, "", ""))
	e.Recalculate()
	c.Equal(-fxp.Five, e.SkillBonusFor("Guns", "Longarm", "", nil, nil),
		"the skill carrying the unmet equipment prereq is penalized")
	c.Equal(fxp.Int(0), e.SkillBonusFor("Guns", "Longarm", "AK-47", nil, nil),
		"a sibling adding an optional specialization is not penalized")

	// The tech-level variant of the penalty is scoped the same way.
	e = NewEntity()
	e.Skills = append(e.Skills, newSkillNeedingEquipment(e, "AK-47", "8"))
	e.Recalculate()
	c.Equal(-fxp.Ten, e.SkillBonusFor("Guns", "Longarm", "AK-47", nil, nil),
		"the TL'd skill carrying the unmet equipment prereq is penalized")
	c.Equal(fxp.Int(0), e.SkillBonusFor("Guns", "Longarm", "M16", nil, nil),
		"a sibling differing only in optional specialization is not penalized")
}

// newSkillNeedingEquipment creates a skill with an equipped-equipment prerequisite that no equipment can satisfy, so
// that processPrereqs generates the equipment penalty for it.
func newSkillNeedingEquipment(e *Entity, optionalSpecialization, techLevel string) *Skill {
	s := NewSkill(e, nil, false)
	s.Name = "Guns"
	s.Specialization = "Longarm"
	s.OptionalSpecialization = optionalSpecialization
	if techLevel != "" {
		s.TechLevel = &techLevel
	}
	list := NewPrereqList()
	p := NewEquippedEquipmentPrereq()
	p.Parent = list
	p.NameCriteria.Qualifier = "Longarm"
	list.Prereqs = append(list.Prereqs, p)
	s.Prereq = list
	return s
}

func TestEntityAttributeBonus(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	c.Equal(fxp.Ten, e.Attributes.Current("st"), "ST default")
	c.Equal(e.SwingFor(10), e.Swing(), "Swing default")
	c.Equal(fxp.WeightFromInteger(20, fxp.Pound), e.BasicLift(), "Basic Lift default")
	c.Equal(fxp.Int(0), e.ThrowingStrengthBonus, "Throwing ST Bonus default")

	bonus := NewAttributeBonus("st")
	trait := NewTrait(e, nil, false)
	trait.Features = append(trait.Features, bonus)
	e.Traits = append(e.Traits, trait)
	e.Recalculate()
	c.Equal(fxp.Eleven, e.Attributes.Current("st"), "ST; simple +1 bonus")

	bonus.PerLevel = true
	e.Recalculate()
	c.Equal(fxp.Ten, e.Attributes.Current("st"), "ST; leveled +1 bonus, but no levels")

	trait.CanLevel = true
	trait.Levels = fxp.Three
	e.Recalculate()
	c.Equal(fxp.Thirteen, e.Attributes.Current("st"), "ST; leveled +1 bonus, with 3 levels")

	bonus.Limitation = stlimit.StrikingOnly
	e.Recalculate()
	c.Equal(fxp.Ten, e.Attributes.Current("st"), "ST; leveled +1 bonus, with 3 levels, for striking only")
	c.Equal(e.SwingFor(13), e.Swing(), "Swing; leveled +1 bonus, with 3 levels, for striking only")

	bonus.Limitation = stlimit.LiftingOnly
	e.Recalculate()
	c.Equal(fxp.Ten, e.Attributes.Current("st"), "ST; leveled +1 bonus, with 3 levels, for lifting only")
	c.Equal(fxp.WeightFromInteger(34, fxp.Pound), e.BasicLift(), "Basic Lift; leveled +1 bonus, with 3 levels, for lifting only")

	bonus.Limitation = stlimit.ThrowingOnly
	e.Recalculate()
	c.Equal(fxp.Ten, e.Attributes.Current("st"), "ST; leveled +1 bonus, with 3 levels, for throwing only")
	c.Equal(fxp.Three, e.ThrowingStrengthBonus, "Throwing ST Bonus; leveled +1 bonus, with 3 levels, for throwing only")
}

// TestEntityThisArmorDRBonus verifies that a "this armor" DR bonus (one that names no locations) is applied exactly
// once to each location the owning equipment already grants DR to, even when more than one of that equipment's other
// DR bonuses names the same location.
func TestEntityThisArmorDRBonus(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	eqp := NewEquipment(e, nil, false)
	eqp.Name = "Mail Hauberk"
	eqp.Features = Features{
		newTestDRBonus(fxp.Four, AllID, TorsoID, "vitals"),
		newTestDRBonus(fxp.Two, AllID, "vitals"),  // repeats a location the bonus above already covers
		newTestDRBonus(fxp.Three, AllID, "Torso"), // repeats a location, but with a different case
		newTestDRBonus(fxp.Five, "piercing", "arm"),
		newTestDRBonus(fxp.One, AllID), // no locations, i.e. "this armor"
	}
	e.CarriedEquipment = append(e.CarriedEquipment, eqp)
	e.Recalculate()

	drMap := e.AddDRBonusesFor(TorsoID, nil, nil)
	c.Equal(8, drMap[AllID], "torso: 4 + 3 base DR, plus the 'this armor' bonus applied once")

	drMap = e.AddDRBonusesFor("vitals", nil, nil)
	c.Equal(7, drMap[AllID], "vitals: 4 + 2 base DR, plus the 'this armor' bonus applied once")

	drMap = e.AddDRBonusesFor("arm", nil, nil)
	c.Equal(5, drMap["piercing"], "arm: the base DR retains its own specialization")
	c.Equal(1, drMap[AllID], "arm: covered by a differently specialized bonus, so still gets the 'this armor' bonus")

	drMap = e.AddDRBonusesFor("leg", nil, nil)
	c.Equal(0, drMap[AllID], "leg: not covered by the equipment, so gets nothing")
}

func newTestDRBonus(amount fxp.Int, specialization string, locations ...string) *DRBonus {
	bonus := NewDRBonus()
	bonus.Locations = locations
	bonus.Specialization = specialization
	bonus.Amount = amount
	return bonus
}

func TestEntityHideZeroValueConditionalModifiers(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	// A situation whose contributing bonuses cancel out to a total of zero.
	addConditionalModifier(e, "cancels out", fxp.One)
	addConditionalModifier(e, "cancels out", fxp.NegOne)
	// A situation with a non-zero total.
	addConditionalModifier(e, "still applies", fxp.Two)

	e.SheetSettings.HideZeroValueConditionalMods = false
	mods := e.ConditionalModifiers()
	c.Equal(2, len(mods), "both situations listed when zero-value modifiers are shown")

	e.SheetSettings.HideZeroValueConditionalMods = true
	mods = e.ConditionalModifiers()
	c.Equal(1, len(mods), "zero-total situation omitted when hidden")
	if len(mods) == 1 {
		c.Equal("still applies", mods[0].From, "the remaining modifier is the non-zero one")
		c.Equal(fxp.Two, mods[0].Total(), "the remaining modifier retains its total")
	}
}

func addConditionalModifier(e *Entity, situation string, amt fxp.Int) {
	bonus := NewConditionalModifierBonus()
	bonus.Situation = situation
	bonus.Amount = amt
	trait := NewTrait(e, nil, false)
	trait.Features = append(trait.Features, bonus)
	e.Traits = append(e.Traits, trait)
}
