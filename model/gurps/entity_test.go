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
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
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

// TestEntityEquipmentPrereqPenaltyHonorsAnyOfLists verifies that the equipment penalty follows the prerequisite list
// as a whole: a skill whose unmet equipment prerequisite sits in an "any of" list alongside a met skill prerequisite
// has satisfied prerequisites and so takes no penalty, and it takes the penalty once that skill prerequisite is no
// longer met. Whether the alternative is met depends on a skill level, so this also checks that the penalty settles
// correctly even though it is generated before the levels are updated.
func TestEntityEquipmentPrereqPenaltyHonorsAnyOfLists(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	guns := newSkillNeedingEquipment(e, "", "")
	guns.Prereq.All = false
	addSkillPrereq(guns.Prereq, "Broadsword", fxp.Ten)
	e.Skills = append(e.Skills, guns)
	control := addTestSkill(e, "Guns", "Pistol", "", fxp.One)
	broadsword := addTestSkill(e, "Broadsword", "", "", fxp.Four) // IQ+1, so 11
	guns.Difficulty = control.Difficulty
	guns.Points = control.Points

	e.Recalculate()
	c.Equal("", guns.UnsatisfiedReason, "the met skill prerequisite satisfies the list")
	c.Equal(fxp.Int(0), e.SkillBonusFor("Guns", "Longarm", "", nil, nil), "a satisfied list carries no penalty")
	c.Equal(control.LevelData.Level, guns.LevelData.Level, "the level is not penalized")

	broadsword.Points = fxp.One // IQ-1, so 9, which no longer meets the alternative
	e.Recalculate()
	c.NotEqual("", guns.UnsatisfiedReason, "with neither alternative met, the list is unsatisfied")
	c.Equal(-fxp.Five, e.SkillBonusFor("Guns", "Longarm", "", nil, nil), "an unsatisfied list carries the penalty")
	c.Equal(control.LevelData.Level-fxp.Five, guns.LevelData.Level, "the level is penalized")
	c.Equal(fxp.Int(0), e.SkillBonusFor("Guns", "Pistol", "", nil, nil), "an unrelated skill must not take the penalty")

	broadsword.Points = fxp.Four
	e.Recalculate()
	c.Equal("", guns.UnsatisfiedReason, "meeting the alternative again satisfies the list")
	c.Equal(control.LevelData.Level, guns.LevelData.Level, "and lifts the penalty")
}

// TestEntityEquipmentPrereqPenaltyHonorsNestedLists verifies that the equipment penalty follows nested lists the way
// PrereqList.Satisfied promises: an unmet equipment prerequisite reached through nested lists that are each
// unsatisfied earns the penalty, while one inside a nested list that is satisfied does not, however unsatisfied the
// list around it is.
func TestEntityEquipmentPrereqPenaltyHonorsNestedLists(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	control := addTestSkill(e, "Guns", "Rifle", "", fxp.One)
	addTestSkill(e, "Broadsword", "", "", fxp.Four) // IQ+1, so 11
	// The equipment prerequisite sits in an "any of" list nested within the skill's own "all of" list, alongside a
	// skill prerequisite that is not met either, so both lists are unsatisfied and the penalty follows.
	nestedUnmet := addTestSkill(e, "Guns", "Longarm", "", fxp.One)
	nestedUnmet.Prereq = NewPrereqList()
	inner := newPrereqListNeedingEquipment()
	inner.All = false
	inner.Parent = nestedUnmet.Prereq
	addSkillPrereq(inner, "Broadsword", fxp.Twelve)
	nestedUnmet.Prereq.Prereqs = append(nestedUnmet.Prereq.Prereqs, inner)
	// The same nested list, but with its skill prerequisite met, so the nested list is satisfied and the equipment
	// prerequisite within it earns no penalty, although the list around it is unsatisfied for a missing trait.
	nestedMet := addTestSkill(e, "Guns", "Pistol", "", fxp.One)
	nestedMet.Prereq = newPrereqListRequiringTrait("Combat Reflexes")
	inner = newPrereqListNeedingEquipment()
	inner.All = false
	inner.Parent = nestedMet.Prereq
	addSkillPrereq(inner, "Broadsword", fxp.Ten)
	nestedMet.Prereq.Prereqs = append(nestedMet.Prereq.Prereqs, inner)

	e.Recalculate()
	c.NotEqual("", nestedUnmet.UnsatisfiedReason, "the list with the unsatisfied nested list is unsatisfied")
	c.Equal(-fxp.Five, e.SkillBonusFor("Guns", "Longarm", "", nil, nil),
		"an unmet equipment prerequisite reached through unsatisfied lists carries the penalty")
	c.Equal(control.LevelData.Level-fxp.Five, nestedUnmet.LevelData.Level, "so the level is penalized")
	c.NotEqual("", nestedMet.UnsatisfiedReason, "the list with the satisfied nested list is unsatisfied all the same")
	c.Equal(fxp.Int(0), e.SkillBonusFor("Guns", "Pistol", "", nil, nil),
		"but an unmet equipment prerequisite inside a satisfied nested list carries no penalty")
	c.Equal(control.LevelData.Level, nestedMet.LevelData.Level, "so the level is not penalized")
	c.Equal(fxp.Int(0), e.SkillBonusFor("Guns", "Rifle", "", nil, nil), "an unrelated skill must not take the penalty")
	checkRecalculationConsistency(c, e, "nested lists")
}

// TestEntityEquipmentPrereqPenaltySeenByScriptsWhateverTheOrder verifies that a prerequisite script which reads the
// level of a skill taking the missing-equipment penalty sees the penalized level, whether the script's skill is listed
// before or after the penalized one. A script computes the level as it runs, from the penalties in place at the time,
// so were the penalties added as each skill is judged, a script listed before the penalized skill would see the level
// before the penalty, and the recalculation would settle with that verdict in place. "Alpha" takes the penalty, which
// brings it from 11 to 6, and "Zulu" requires Alpha at 10 or better by way of a script.
func TestEntityEquipmentPrereqPenaltySeenByScriptsWhateverTheOrder(t *testing.T) {
	c := check.New(t)
	for _, reversed := range []bool{false, true} {
		e := NewEntity()
		alpha := addTestSkill(e, "Alpha", "", "", fxp.Four) // IQ+1, so 11 on its own
		alpha.Prereq = newPrereqListNeedingEquipment()
		zulu := addTestSkill(e, "Zulu", "", "", fxp.One)
		zulu.Prereq = NewPrereqList()
		script := NewScriptPrereq()
		script.Parent = zulu.Prereq
		script.Script = `entity.skillLevel("Alpha") >= 10 ? "" : "Alpha is below 10"`
		zulu.Prereq.Prereqs = append(zulu.Prereq.Prereqs, script)
		if reversed {
			slices.Reverse(e.Skills)
		}

		e.Recalculate()
		c.Equal(fxp.Six, alpha.LevelData.Level, "reversed %t: Alpha takes the penalty", reversed)
		c.NotEqual("", zulu.UnsatisfiedReason, "reversed %t: the script sees the penalized level", reversed)
		checkRecalculationConsistency(c, e, fmt.Sprintf("reversed %t", reversed))
	}
}

// TestEntityEquipmentPrereqPenaltyHonorsTechLevelGate verifies that an unmet equipment prerequisite in a list that does
// not apply at the sheet's tech level neither flags the skill nor earns it the penalty, and that both follow once the
// sheet's tech level brings the list into play.
func TestEntityEquipmentPrereqPenaltyHonorsTechLevelGate(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	e.Profile.TechLevel = "3"
	guns := newSkillNeedingEquipment(e, "", "")
	guns.Prereq.WhenTL = numberCriteria(criteria.AtLeastNumber, fxp.Nine)
	e.Skills = append(e.Skills, guns)
	control := addTestSkill(e, "Guns", "Pistol", "", fxp.One)
	guns.Difficulty = control.Difficulty
	guns.Points = control.Points

	e.Recalculate()
	c.Equal("", guns.UnsatisfiedReason, "a list that does not apply at the sheet's tech level flags nothing")
	c.Equal(fxp.Int(0), e.SkillBonusFor("Guns", "Longarm", "", nil, nil), "and carries no penalty")
	c.Equal(control.LevelData.Level, guns.LevelData.Level, "so the level is not penalized")

	e.Profile.TechLevel = "9"
	e.Recalculate()
	c.NotEqual("", guns.UnsatisfiedReason, "the list applies once the sheet reaches its tech level")
	c.Equal(-fxp.Five, e.SkillBonusFor("Guns", "Longarm", "", nil, nil), "and carries the penalty")
	c.Equal(control.LevelData.Level-fxp.Five, guns.LevelData.Level, "so the level is penalized")
}

// TestEntityEquipmentPrereqOnTraitsAndEquipmentTakesNoPenalty verifies that an unmet equipped-equipment prerequisite on
// a trait or on a piece of equipment, neither of which has a level to penalize, is recorded as unsatisfied without
// generating a penalty for anything.
func TestEntityEquipmentPrereqOnTraitsAndEquipmentTakesNoPenalty(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	trait := NewTrait(e, nil, false)
	trait.Name = "Gunslinger"
	trait.Prereq = newPrereqListNeedingEquipment()
	e.Traits = append(e.Traits, trait)
	eqp := addCarriedEquipmentWithFeatures(e, "Bandolier")
	eqp.Prereq = newPrereqListNeedingEquipment()
	guns := addTestSkill(e, "Guns", "Longarm", "", fxp.One)

	e.Recalculate()
	c.NotEqual("", trait.UnsatisfiedReason, "the trait records its unmet equipment prerequisite")
	c.NotEqual("", eqp.UnsatisfiedReason, "the equipment records its unmet equipment prerequisite")
	c.Equal("", guns.UnsatisfiedReason, "the skill has no prerequisites of its own")
	c.Equal(fxp.Int(0), e.SkillBonusFor("Guns", "Longarm", "", nil, nil), "no skill takes a penalty")
	c.Equal(0, len(e.features.skillBonuses), "no skill penalty is generated at all")
	c.Equal(0, len(e.features.spellBonuses), "nor any spell penalty")
}

// newSkillNeedingEquipment creates a "Guns/Longarm" skill with an equipped-equipment prerequisite that no equipment
// can satisfy, so that a recalculation applies the missing-equipment penalty to it. It is not added to the entity.
func newSkillNeedingEquipment(e *Entity, optionalSpecialization, techLevel string) *Skill {
	s := NewSkill(e, nil, false)
	s.Name = "Guns"
	s.Specialization = "Longarm"
	s.OptionalSpecialization = optionalSpecialization
	if techLevel != "" {
		s.TechLevel = &techLevel
	}
	s.Prereq = newPrereqListNeedingEquipment()
	return s
}

// newPrereqListNeedingEquipment returns a prerequisite list with an equipped-equipment prerequisite that no equipment
// can satisfy.
func newPrereqListNeedingEquipment() *PrereqList {
	list := NewPrereqList()
	p := NewEquippedEquipmentPrereq()
	p.Parent = list
	p.NameCriteria.Qualifier = "Longarm"
	list.Prereqs = append(list.Prereqs, p)
	return list
}

func TestEntityAttributeBonus(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	c.Equal(fxp.Ten, e.Attributes.Current("st"), "ST default")
	c.Equal(e.SwingFor(10), e.Swing(), "Swing default")
	c.Equal(fxp.WeightFromInteger(20, fxp.Pound), e.BasicLift(), "Basic Lift default")
	c.Equal(fxp.Int(0), e.ThrowingStrengthBonus, "Throwing ST Bonus default")

	bonus := NewAttributeBonus("st")
	trait := addTraitWithFeatures(e, "", bonus)
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
	addCarriedEquipmentWithFeatures(e, "Mail Hauberk",
		newTestDRBonus(fxp.Four, AllID, TorsoID, "vitals"),
		newTestDRBonus(fxp.Two, AllID, "vitals"),  // repeats a location the bonus above already covers
		newTestDRBonus(fxp.Three, AllID, "Torso"), // repeats a location, but with a different case
		newTestDRBonus(fxp.Five, "piercing", "arm"),
		newTestDRBonus(fxp.One, AllID), // no locations, i.e. "this armor"
	)
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

// TestEntityThisArmorDRBonusIncludesModifierLocations verifies that the DR bonuses an equipment modifier contributes
// count toward the locations a "this armor" DR bonus expands onto, in both directions: a "this armor" bonus on the
// equipment picks up the locations its modifiers add, and one carried by a modifier picks up the equipment's own
// locations. Locations named by both are still only granted the bonus once, and a disabled modifier contributes
// nothing.
func TestEntityThisArmorDRBonusIncludesModifierLocations(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	eqp := addCarriedEquipmentWithFeatures(e, "Mail Hauberk",
		newTestDRBonus(fxp.Four, AllID, TorsoID),
		newTestDRBonus(fxp.One, AllID), // no locations, i.e. "this armor"
	)
	mod := NewEquipmentModifier(e, nil, false)
	mod.Name = "Sleeves"
	mod.Features = Features{
		newTestDRBonus(fxp.Two, AllID, "arm"),
		newTestDRBonus(fxp.Three, AllID, "Torso"), // repeats a location the equipment covers, with a different case
	}
	eqp.Modifiers = []*EquipmentModifier{mod}
	e.Recalculate()

	c.Equal(8, e.AddDRBonusesFor(TorsoID, nil, nil)[AllID],
		"torso: 4 + 3 base DR, plus the 'this armor' bonus applied once")
	c.Equal(3, e.AddDRBonusesFor("arm", nil, nil)[AllID],
		"arm: the modifier's 2 base DR, plus the 'this armor' bonus")

	mod.Disabled = true
	e.Recalculate()
	c.Equal(5, e.AddDRBonusesFor(TorsoID, nil, nil)[AllID],
		"torso: only the equipment's own 4 base DR remains, plus the 'this armor' bonus")
	c.Equal(0, e.AddDRBonusesFor("arm", nil, nil)[AllID],
		"arm: a disabled modifier grants no DR, so the 'this armor' bonus doesn't reach it")

	// The reverse direction: the "this armor" bonus lives on the modifier, so it must expand onto both the
	// equipment's own locations and those of the modifiers.
	e = NewEntity()
	eqp = addCarriedEquipmentWithFeatures(e, "Mail Hauberk", newTestDRBonus(fxp.Four, AllID, TorsoID))
	mod = NewEquipmentModifier(e, nil, false)
	mod.Name = "Sleeves"
	mod.Features = Features{
		newTestDRBonus(fxp.Two, AllID, "arm"),
		newTestDRBonus(fxp.One, AllID), // no locations, i.e. "this armor"
	}
	eqp.Modifiers = []*EquipmentModifier{mod}
	e.Recalculate()

	c.Equal(5, e.AddDRBonusesFor(TorsoID, nil, nil)[AllID],
		"torso: the modifier's 'this armor' bonus expands onto the equipment's own location")
	c.Equal(3, e.AddDRBonusesFor("arm", nil, nil)[AllID],
		"arm: the modifier's 'this armor' bonus expands onto the modifier's own location")
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
	addTraitWithFeatures(e, "", bonus)
}

// TestEntityHideZeroValueConditionalModifiersPrunesGroups verifies that hiding zero-value conditional modifiers is
// applied to the members of a group, and that a group left with no members disappears along with them.
func TestEntityHideZeroValueConditionalModifiersPrunesGroups(t *testing.T) {
	withGroupContainersOnSort(t, true)
	c := check.New(t)
	e := NewEntity()
	addGroupedConditionalModifier(e, "", "cancels out", "Mixed", fxp.One)
	addGroupedConditionalModifier(e, "", "cancels out", "Mixed", fxp.NegOne)
	addGroupedConditionalModifier(e, "", "still applies", "Mixed", fxp.Two)
	addGroupedConditionalModifier(e, "", "also cancels out", "Empty", fxp.One)
	addGroupedConditionalModifier(e, "", "also cancels out", "Empty", fxp.NegOne)
	addConditionalModifier(e, "ungrouped", fxp.Three)
	e.Recalculate()

	e.SheetSettings.HideZeroValueConditionalMods = false
	c.Equal([]string{"Empty[also cancels out]", "Mixed[cancels out,still applies]", "ungrouped"},
		rowNames(e.ConditionalModifiers()), "everything listed when zero-value modifiers are shown")

	e.SheetSettings.HideZeroValueConditionalMods = true
	c.Equal([]string{"Mixed[still applies]", "ungrouped"}, rowNames(e.ConditionalModifiers()),
		"zero-total members are omitted, and a group with none left is omitted with them")
}

// TestEntityProcessTraitPrereqsClearsUnsatisfiedReasonWhenDisabled verifies that disabling a trait (directly or by
// disabling one of its containers) clears any unsatisfied prerequisite reason it had accumulated, rather than leaving
// the stale reason behind until the sheet is reloaded.
func TestEntityProcessTraitPrereqsClearsUnsatisfiedReasonWhenDisabled(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	trait := newTraitNeedingMissingTrait(e, "Trained by a Master")
	e.Traits = append(e.Traits, trait)

	e.Recalculate()
	c.NotEqual("", trait.UnsatisfiedReason, "an enabled trait with an unmet prereq is flagged")

	trait.Disabled = true
	e.Recalculate()
	c.Equal("", trait.UnsatisfiedReason, "disabling the trait clears the stale reason")

	trait.Disabled = false
	e.Recalculate()
	c.NotEqual("", trait.UnsatisfiedReason, "re-enabling the trait flags it again")

	// The same must hold when the trait is only disabled by way of a container above it.
	e = NewEntity()
	container := NewTrait(e, nil, true)
	container.Name = "Martial Arts"
	nested := newTraitNeedingMissingTrait(e, "Weapon Master")
	nested.SetParent(container)
	container.Children = append(container.Children, nested)
	e.Traits = append(e.Traits, container)

	e.Recalculate()
	c.NotEqual("", nested.UnsatisfiedReason, "a trait inside an enabled container is flagged")

	container.Disabled = true
	e.Recalculate()
	c.Equal("", nested.UnsatisfiedReason, "disabling the container clears the nested trait's stale reason")

	// A trait exceeding a maximum level is flagged the same way, and must also be cleared when disabled.
	e = NewEntity()
	trait = NewTrait(e, nil, false)
	trait.Name = "Extra Arm"
	trait.CanLevel = true
	trait.Levels = fxp.Three
	trait.MaxLevels = "2"
	e.Traits = append(e.Traits, trait)

	e.Recalculate()
	c.NotEqual("", trait.UnsatisfiedReason, "a trait above its maximum level is flagged")

	trait.Disabled = true
	e.Recalculate()
	c.Equal("", trait.UnsatisfiedReason, "disabling the trait clears the maximum-level reason")
}

// newTraitNeedingMissingTrait creates a trait whose prerequisite requires another trait the entity doesn't have, so
// that processTraitPrereqs marks it unsatisfied.
func newTraitNeedingMissingTrait(e *Entity, name string) *Trait {
	return newTraitRequiring(e, name, "Combat Reflexes")
}

// newTraitRequiring creates a trait with the given name whose prerequisite requires a trait with the required name.
// It is not added to the entity.
func newTraitRequiring(e *Entity, name, required string) *Trait {
	t := NewTrait(e, nil, false)
	t.Name = name
	t.Prereq = newPrereqListRequiringTrait(required)
	return t
}

// newPrereqListRequiringTrait returns a prerequisite list satisfied only by a trait with the given name.
func newPrereqListRequiringTrait(name string) *PrereqList {
	return newPrereqListForTrait(name, true)
}

// newPrereqListForbiddingTrait returns a prerequisite list satisfied only by the absence of a trait with the given
// name.
func newPrereqListForbiddingTrait(name string) *PrereqList {
	return newPrereqListForTrait(name, false)
}

// newPrereqListForTrait returns a prerequisite list requiring the presence, or with has false the absence, of a trait
// with the given name.
func newPrereqListForTrait(name string, has bool) *PrereqList {
	list := NewPrereqList()
	p := NewTraitPrereq()
	p.Parent = list
	p.Has = has
	p.NameCriteria.Qualifier = name
	list.Prereqs = append(list.Prereqs, p)
	return list
}

// newPrereqListRequiringSkill returns a prerequisite list satisfied only by a skill with the given name at or above the
// given level.
func newPrereqListRequiringSkill(name string, level fxp.Int) *PrereqList {
	list := NewPrereqList()
	addSkillPrereq(list, name, level)
	return list
}

// recalculationState describes what a recalculation derives about the traits, skills, spells and equipment, covering
// everything Entity.derivedState fingerprints along with the reasons recorded and whether the data was found to
// settle, so that a test can check that recalculating again changes nothing.
func recalculationState(e *Entity) string {
	var buf strings.Builder
	fmt.Fprintf(&buf, "unsettled=%t\n", e.unsettled)
	Traverse(func(t *Trait) bool {
		fmt.Fprintf(&buf, "trait %s: enabled=%t contradicted=%t reason=%q\n", t.Name, t.Enabled(),
			t.ContradictedPrereqs(), t.UnsatisfiedReason)
		return false
	}, false, false, e.Traits...)
	Traverse(func(s *Skill) bool {
		fmt.Fprintf(&buf, "skill %s: level=%+v penalized=%t default=", s, s.LevelData, s.takesEquipmentPenalty)
		if s.DefaultedFrom == nil {
			buf.WriteString("none")
		} else {
			fmt.Fprintf(&buf, "%+v", *s.DefaultedFrom)
		}
		fmt.Fprintf(&buf, " reason=%q\n", s.UnsatisfiedReason)
		return false
	}, false, true, e.Skills...)
	Traverse(func(s *Spell) bool {
		fmt.Fprintf(&buf, "spell %s: level=%+v penalized=%t reason=%q\n", s.Name, s.LevelData,
			s.takesEquipmentPenalty, s.UnsatisfiedReason)
		return false
	}, false, true, e.Spells...)
	equipmentFunc := func(eqp *Equipment) bool {
		fmt.Fprintf(&buf, "equipment %s: reason=%q\n", eqp.Name, eqp.UnsatisfiedReason)
		return false
	}
	Traverse(equipmentFunc, false, false, e.CarriedEquipment...)
	Traverse(equipmentFunc, false, false, e.OtherEquipment...)
	return buf.String()
}

// checkRecalculationConsistency verifies what every recalculation guarantees, whether or not the data settles: the
// recorded skill and spell levels are what the features in effect produce, each trait records a reason exactly when
// its prerequisites are unmet against those levels or its level exceeds its maximum, a trait the sheet disabled has an
// unmet prerequisite to show for it, an enabled trait with an unmet prerequisite is one the sheet found caught in a
// contradiction while it enforces prerequisites, no trait is found so unless the sheet noted that its data never
// settles, and recalculating again changes nothing. That last check recalculates the entity once more. The label
// identifies the call site in the messages.
func checkRecalculationConsistency(c check.Checker, e *Entity, label string) {
	c.Helper()
	Traverse(func(s *Skill) bool {
		c.Helper()
		c.Equal(s.CalculateLevel(nil).Level, s.LevelData.Level, "%s: skill %s: the recorded level is current", label, s)
		return false
	}, false, true, e.Skills...)
	Traverse(func(s *Spell) bool {
		c.Helper()
		c.Equal(s.CalculateLevel().Level, s.LevelData.Level, "%s: spell %s: the recorded level is current", label,
			s.Name)
		return false
	}, false, true, e.Spells...)
	Traverse(func(t *Trait) bool {
		c.Helper()
		if t.Disabled || (t.parent != nil && !t.parent.Enabled()) {
			c.Equal("", t.UnsatisfiedReason,
				"%s: trait %s: a trait that is disabled, itself or by way of a container, has no reason", label, t.Name)
			c.False(t.DisabledByPrereqs(),
				"%s: trait %s: a trait that is disabled, itself or by way of a container, is not the one the sheet disabled",
				label, t.Name)
			return false
		}
		unmet := t.Prereq != nil && !t.Prereq.Satisfied(e, t, nil, "", nil)
		maximum := t.ResolvedMaxLevels()
		overMaximum := maximum > 0 && t.Levels > maximum
		c.Equal(unmet || overMaximum, t.UnsatisfiedReason != "",
			"%s: trait %s: a reason is recorded exactly when the prerequisites are unmet or the level exceeds its maximum",
			label, t.Name)
		c.Equal(overMaximum, strings.Contains(t.UnsatisfiedReason, levelExceedsMaximumReason(maximum)),
			"%s: trait %s: the reason names the maximum level exactly when the level exceeds it", label, t.Name)
		if t.DisabledByPrereqs() {
			c.NotEqual("", t.UnsatisfiedReason, "%s: trait %s: a trait the sheet disabled has a reason", label, t.Name)
			c.False(t.ContradictedPrereqs(), "%s: trait %s: a trait the sheet disabled is not one it left enabled",
				label, t.Name)
		} else if e.SheetSettings.EnforceTraitPrereqs && t.UnsatisfiedReason != "" {
			c.True(t.ContradictedPrereqs(),
				"%s: trait %s: an enabled trait with an unmet prerequisite is caught in a contradiction", label, t.Name)
		}
		if !e.unsettled {
			c.False(t.ContradictedPrereqs(),
				"%s: trait %s: no trait is caught in a contradiction when the data settles", label, t.Name)
		}
		return false
	}, false, false, e.Traits...)
	before := recalculationState(e)
	e.Recalculate()
	c.Equal(before, recalculationState(e), "%s: recalculating again changes nothing", label)
}

// TestEntityEnforceTraitPrereqsDisablesUnsatisfiedTraits verifies the sheet setting that disables traits whose
// prerequisites are unsatisfied: the trait keeps its unsatisfied reason but is treated as disabled, so its points,
// features and weapons drop out of the sheet, and it comes back on its own once the prerequisites are met or the
// setting is turned off. The user's own enabled state on the trait is never touched.
func TestEntityEnforceTraitPrereqsDisablesUnsatisfiedTraits(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	trait := newTraitNeedingMissingTrait(e, "Trained by a Master")
	trait.BasePoints = fxp.Twenty
	bonus := NewAttributeBonus(StrengthID)
	bonus.Amount = fxp.Two
	trait.Features = Features{bonus}
	trait.Weapons = append(trait.Weapons, NewWeapon(trait, true))
	e.Traits = append(e.Traits, trait)

	e.Recalculate()
	c.NotEqual("", trait.UnsatisfiedReason, "the unmet prereq is flagged")
	c.True(trait.Enabled(), "without the setting, the trait stays enabled")
	c.False(trait.DisabledByPrereqs(), "without the setting, the trait is not disabled by its prereqs")
	c.Equal(fxp.Twenty, trait.AdjustedPoints(nil), "without the setting, the trait's points count")
	c.Equal(fxp.Twenty, e.PointsBreakdown().Advantages, "without the setting, the trait's points are in the total")
	c.Equal(fxp.Two, e.AttributeBonusFor(StrengthID, stlimit.None, nil), "without the setting, its features apply")
	// The entity starts out with the natural attacks, so the weapon counts are relative to those.
	weaponCount := len(e.Weapons(true, false, false))
	c.True(weaponCount > 0, "without the setting, its weapons are in play")

	e.SheetSettings.EnforceTraitPrereqs = true
	e.Recalculate()
	c.NotEqual("", trait.UnsatisfiedReason, "the reason is kept so the sheet can show why the trait is disabled")
	c.False(trait.Enabled(), "the setting disables the trait")
	c.True(trait.EffectivelyDisabled(), "the setting disables the trait")
	c.True(trait.DisabledByPrereqs(), "the trait reports that its prereqs disabled it")
	c.False(trait.Disabled, "the user's own enabled state is left alone")
	c.Equal(fxp.Int(0), trait.AdjustedPoints(nil), "a disabled trait is worth no points")
	c.Equal(fxp.Int(0), e.PointsBreakdown().Advantages, "a disabled trait's points leave the total")
	c.Equal(fxp.Int(0), e.AttributeBonusFor(StrengthID, stlimit.None, nil), "its features are no longer active")
	c.Equal(weaponCount-1, len(e.Weapons(true, false, false)), "its weapons are no longer in play")

	// Satisfying the prerequisite brings the trait back without touching it.
	combatReflexes := NewTrait(e, nil, false)
	combatReflexes.Name = "Combat Reflexes"
	e.Traits = append(e.Traits, combatReflexes)
	e.Recalculate()
	c.Equal("", trait.UnsatisfiedReason, "the prereq is met")
	c.True(trait.Enabled(), "the trait is re-enabled once its prerequisite is met")
	c.False(trait.DisabledByPrereqs(), "the trait no longer reports being disabled by its prereqs")
	c.Equal(fxp.Twenty, trait.AdjustedPoints(nil), "its points count again")
	c.Equal(fxp.Two, e.AttributeBonusFor(StrengthID, stlimit.None, nil), "its features apply again")

	// Removing the prerequisite disables the trait again, and turning the setting off re-enables it.
	e.Traits = slices.DeleteFunc(e.Traits, func(t *Trait) bool { return t == combatReflexes })
	e.Recalculate()
	c.False(trait.Enabled(), "losing the prerequisite disables the trait again")
	e.SheetSettings.EnforceTraitPrereqs = false
	e.Recalculate()
	c.True(trait.Enabled(), "turning the setting off re-enables the trait")
	c.NotEqual("", trait.UnsatisfiedReason, "the unmet prereq is still flagged")
}

// TestEntityEnforceTraitPrereqsCascades verifies that a trait disabled for unsatisfied prerequisites no longer
// satisfies the prerequisites of other traits, which are disabled in turn within the same recalculation, and that
// all of them come back together once the root prerequisite is met.
func TestEntityEnforceTraitPrereqsCascades(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	e.SheetSettings.EnforceTraitPrereqs = true
	// The dependent trait is listed first, so a single pass over the traits would still see the trait it needs as
	// enabled; only a further pass notices that it has been disabled.
	dependent := newTraitRequiring(e, "Weapon Master", "Trained by a Master")
	needed := newTraitNeedingMissingTrait(e, "Trained by a Master")
	e.Traits = append(e.Traits, dependent, needed)

	e.Recalculate()
	c.False(needed.Enabled(), "the trait with the unmet prereq is disabled")
	c.False(dependent.Enabled(), "a trait whose prerequisite was disabled is disabled in turn")
	c.NotEqual("", dependent.UnsatisfiedReason, "the dependent trait records why")

	combatReflexes := NewTrait(e, nil, false)
	combatReflexes.Name = "Combat Reflexes"
	e.Traits = append(e.Traits, combatReflexes)
	e.Recalculate()
	c.True(needed.Enabled(), "meeting the root prerequisite re-enables the trait")
	c.True(dependent.Enabled(), "and the trait that depends on it")
	c.Equal("", dependent.UnsatisfiedReason, "the dependent trait's prereq is met")
}

// TestEntityEnforceTraitPrereqsLongCascadeSettles verifies that a long chain of dependent traits settles completely in
// a single recalculation, whatever order the traits are listed in. Each pass judges every trait against the traits as
// they stood when it began, so a pass cuts the chain by one link, whichever end of it is listed first, and the 7-link
// chain takes 9 passes to settle. Were the passes stopped short of that, the links at the head of the chain would be
// left enabled, and their features counted, on every recalculation.
func TestEntityEnforceTraitPrereqsLongCascadeSettles(t *testing.T) {
	c := check.New(t)
	const links = 7
	for _, reversed := range []bool{false, true} {
		e := NewEntity()
		e.SheetSettings.EnforceTraitPrereqs = true
		broadsword := addTestSkill(e, "Broadsword", "", "", fxp.Four) // IQ+1, so 11 on its own
		chain := make([]*Trait, links)
		for i := range chain {
			name := "Link " + strconv.Itoa(i+1)
			if i == links-1 {
				chain[i] = newTraitNeedingMissingTrait(e, name)
			} else {
				chain[i] = newTraitRequiring(e, name, "Link "+strconv.Itoa(i+2))
			}
			chain[i].Features = Features{newSkillBonusTo("Broadsword", fxp.One)}
		}
		e.Traits = append(e.Traits, chain...)
		if reversed {
			slices.Reverse(e.Traits)
		}

		e.Recalculate()
		for i, trait := range chain {
			c.False(trait.Enabled(), "reversed %t: link %d is disabled after one recalculation", reversed, i+1)
			c.NotEqual("", trait.UnsatisfiedReason, "reversed %t: link %d records why", reversed, i+1)
		}
		c.Equal(fxp.Int(0), e.SkillBonusFor("Broadsword", "", "", nil, nil),
			"reversed %t: no disabled link's bonus counts", reversed)
		c.Equal(fxp.FromInteger(11), broadsword.LevelData.Level, "reversed %t: the skill is unaffected", reversed)
		checkRecalculationConsistency(c, e, fmt.Sprintf("reversed %t, chain cut", reversed))

		// Meeting the root prerequisite brings the whole chain back in a single recalculation as well.
		combatReflexes := NewTrait(e, nil, false)
		combatReflexes.Name = "Combat Reflexes"
		e.Traits = append(e.Traits, combatReflexes)
		e.Recalculate()
		for i, trait := range chain {
			c.True(trait.Enabled(), "reversed %t: link %d is re-enabled after one recalculation", reversed, i+1)
			c.Equal("", trait.UnsatisfiedReason, "reversed %t: link %d's prerequisite is met", reversed, i+1)
		}
		c.Equal(fxp.FromInteger(links), e.SkillBonusFor("Broadsword", "", "", nil, nil),
			"reversed %t: every link's bonus counts", reversed)
		c.Equal(fxp.FromInteger(11+links), broadsword.LevelData.Level, "reversed %t: the skill is raised", reversed)
		checkRecalculationConsistency(c, e, fmt.Sprintf("reversed %t, chain restored", reversed))
	}
}

// TestEntityEnforceTraitPrereqsChainBeyondCapIsStoppedShort verifies what a recalculation does with a chain of
// dependent traits longer than the cap on its passes allows for: the passes are stopped at the cap, the sheet notes
// that its data never settles, and the links the passes reached are disabled while those they did not reach are left
// enabled, with none marked as caught in a contradiction, since each link flips only once. The chain is thus stopped
// short rather than given a wrong answer, whatever order the links are listed in.
func TestEntityEnforceTraitPrereqsChainBeyondCapIsStoppedShort(t *testing.T) {
	c := check.New(t)
	warnings := countLogs(t, slog.LevelWarn)
	const links = maxRecalculationPasses + 8
	for _, reversed := range []bool{false, true} {
		warnings.Store(0)
		e := NewEntity()
		e.SheetSettings.EnforceTraitPrereqs = true
		chain := make([]*Trait, links)
		for i := range chain {
			name := "Link " + strconv.Itoa(i+1)
			if i == links-1 {
				chain[i] = newTraitNeedingMissingTrait(e, name)
			} else {
				chain[i] = newTraitRequiring(e, name, "Link "+strconv.Itoa(i+2))
			}
		}
		e.Traits = append(e.Traits, chain...)
		if reversed {
			slices.Reverse(e.Traits)
		}

		e.Recalculate()
		c.True(e.unsettled, "reversed %t: the sheet notes that its data never settles", reversed)
		c.Equal(int32(1), warnings.Load(), "reversed %t: and logs that once", reversed)
		for i, trait := range chain {
			c.False(trait.ContradictedPrereqs(),
				"reversed %t: link %d flips only once, so is not taken to be caught in a contradiction", reversed, i+1)
			if i < links-maxRecalculationPasses {
				c.True(trait.Enabled(), "reversed %t: link %d is beyond the reach of the passes, so is left enabled",
					reversed, i+1)
			} else {
				c.False(trait.Enabled(), "reversed %t: link %d is within the reach of the passes, so is disabled",
					reversed, i+1)
				c.NotEqual("", trait.UnsatisfiedReason, "reversed %t: link %d records why", reversed, i+1)
			}
		}
	}
}

// TestEntityEnforceTraitPrereqsJudgedAgainstOwnFeatures verifies that a trait whose own features are what satisfy its
// prerequisites is judged against the levels those features produce, so it settles as enabled in a single recalculation
// and a further one leaves everything as it was, rather than oscillating between disabling it because the level is too
// low without its bonus and re-enabling it because the level is high enough with it. It also verifies that the verdict
// depends only on the sheet's data: once the trait has been disabled because the skill dropped too low even with the
// bonus, restoring the skill re-enables it, just as loading the sheet afresh would.
func TestEntityEnforceTraitPrereqsJudgedAgainstOwnFeatures(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	e.SheetSettings.EnforceTraitPrereqs = true
	broadsword := addTestSkill(e, "Broadsword", "", "", fxp.Four) // IQ+1, so 11 on its own
	trait := addTraitWithFeatures(e, "Weapon Master", newSkillBonusTo("Broadsword", fxp.Two))
	trait.Prereq = newPrereqListRequiringSkill("Broadsword", fxp.Twelve)

	for i := range 2 {
		e.Recalculate()
		c.True(trait.Enabled(), "recalculation %d: the trait's bonus lifts the skill past its own prerequisite", i+1)
		c.Equal("", trait.UnsatisfiedReason, "recalculation %d: the prerequisite is met", i+1)
		c.Equal(fxp.FromInteger(13), broadsword.LevelData.Level, "recalculation %d: the bonus is applied", i+1)
	}

	broadsword.Points = fxp.One // IQ-1, so 9 on its own and 11 with the bonus, which is still too low
	e.Recalculate()
	c.False(trait.Enabled(), "the trait is disabled once the skill is too low even with its bonus")
	c.NotEqual("", trait.UnsatisfiedReason, "the prerequisite is not met")
	c.Equal(fxp.FromInteger(9), broadsword.LevelData.Level, "the bonus is withdrawn")

	broadsword.Points = fxp.Four
	e.Recalculate()
	c.True(trait.Enabled(), "restoring the skill re-enables the trait, the same as on a fresh sheet")
	c.Equal("", trait.UnsatisfiedReason, "the prerequisite is met again")
	c.Equal(fxp.FromInteger(13), broadsword.LevelData.Level, "the bonus is applied again")
}

// TestEntityEnforceTraitPrereqsFeatureCascadeSettles verifies that a cascade carried by features settles completely
// in a single recalculation, whatever order the traits are listed in. Each link requires a skill level that only the
// next link's bonus provides, so a link's disabling is discovered only once the features have been collected without
// the link after it, a full pass later, and the 7-link chain takes 8 passes to settle.
func TestEntityEnforceTraitPrereqsFeatureCascadeSettles(t *testing.T) {
	c := check.New(t)
	const links = 7
	for _, reversed := range []bool{false, true} {
		e := NewEntity()
		e.SheetSettings.EnforceTraitPrereqs = true
		skills := make([]*Skill, links-1)
		for i := range skills {
			skills[i] = addTestSkill(e, "Skill "+strconv.Itoa(i+1), "", "", fxp.Four) // IQ+1, so 11 on its own
		}
		chain := make([]*Trait, links)
		for i := range chain {
			name := "Link " + strconv.Itoa(i+1)
			if i == links-1 {
				chain[i] = newTraitNeedingMissingTrait(e, name)
			} else {
				chain[i] = NewTrait(e, nil, false)
				chain[i].Name = name
				chain[i].Prereq = newPrereqListRequiringSkill(skills[i].Name, fxp.Twelve)
			}
			if i > 0 {
				chain[i].Features = Features{newSkillBonusTo(skills[i-1].Name, fxp.Two)}
			}
		}
		e.Traits = append(e.Traits, chain...)
		if reversed {
			slices.Reverse(e.Traits)
		}

		e.Recalculate()
		for i, trait := range chain {
			c.False(trait.Enabled(), "reversed %t: link %d is disabled after one recalculation", reversed, i+1)
			c.NotEqual("", trait.UnsatisfiedReason, "reversed %t: link %d records why", reversed, i+1)
		}
		for i, skill := range skills {
			c.Equal(fxp.FromInteger(11), skill.LevelData.Level,
				"reversed %t: skill %d counts no disabled link's bonus", reversed, i+1)
		}
		checkRecalculationConsistency(c, e, fmt.Sprintf("reversed %t, chain cut", reversed))

		// Meeting the root prerequisite brings the whole chain back in a single recalculation as well.
		combatReflexes := NewTrait(e, nil, false)
		combatReflexes.Name = "Combat Reflexes"
		e.Traits = append(e.Traits, combatReflexes)
		e.Recalculate()
		for i, trait := range chain {
			c.True(trait.Enabled(), "reversed %t: link %d is re-enabled after one recalculation", reversed, i+1)
			c.Equal("", trait.UnsatisfiedReason, "reversed %t: link %d's prerequisite is met", reversed, i+1)
		}
		for i, skill := range skills {
			c.Equal(fxp.FromInteger(13), skill.LevelData.Level,
				"reversed %t: skill %d is raised by the link after it", reversed, i+1)
		}
		checkRecalculationConsistency(c, e, fmt.Sprintf("reversed %t, chain restored", reversed))
	}
}

// TestEntityEnforceTraitPrereqsContradictionSettlesConsistently verifies what a sheet whose prerequisites contradict
// one another is left in. "Requires" requires "Excludes", which in turn requires the absence of "Requires", so no
// state satisfies both and enforcing them flips the pair back and forth forever. The recalculation must nonetheless
// finish, leave both traits enabled as the user set them with the contradiction pointed out on both, including the one
// whose own prerequisite is met only because the other is left enabled, keep disabling an unrelated trait whose
// prerequisite is plainly unmet, compute the levels from the traits actually in effect, and produce the same result
// each time, whatever order the traits are listed in.
func TestEntityEnforceTraitPrereqsContradictionSettlesConsistently(t *testing.T) {
	c := check.New(t)
	warnings := countLogs(t, slog.LevelWarn)
	for _, reversed := range []bool{false, true} {
		warnings.Store(0)
		e := NewEntity()
		e.SheetSettings.EnforceTraitPrereqs = true
		broadsword := addTestSkill(e, "Broadsword", "", "", fxp.Four) // IQ+1, so 11 on its own
		requires := newTraitRequiring(e, "Requires", "Excludes")
		excludes := NewTrait(e, nil, false)
		excludes.Name = "Excludes"
		excludes.Prereq = newPrereqListForbiddingTrait("Requires")
		unrelated := newTraitNeedingMissingTrait(e, "Unrelated")
		for _, trait := range []*Trait{requires, excludes, unrelated} {
			trait.Features = Features{newSkillBonusTo("Broadsword", fxp.One)}
			e.Traits = append(e.Traits, trait)
		}
		if reversed {
			slices.Reverse(e.Traits)
		}

		e.Recalculate()
		c.True(e.unsettled, "reversed %t: the sheet notes that its data never settles", reversed)
		c.Equal(int32(1), warnings.Load(), "reversed %t: and logs that once", reversed)
		c.True(requires.Enabled(), "reversed %t: a trait caught in the contradiction is left enabled", reversed)
		c.True(requires.ContradictedPrereqs(), "reversed %t: and is marked as caught in it", reversed)
		c.Equal("", requires.UnsatisfiedReason,
			"reversed %t: its prerequisite is met, since the trait it requires is enabled", reversed)
		c.True(excludes.Enabled(), "reversed %t: the other trait caught in the contradiction is left enabled",
			reversed)
		c.True(excludes.ContradictedPrereqs(), "reversed %t: and is marked as caught in it too", reversed)
		c.NotEqual("", excludes.UnsatisfiedReason, "reversed %t: with the contradiction visible on it", reversed)
		c.False(unrelated.Enabled(), "reversed %t: an unrelated trait with an unmet prerequisite is still disabled",
			reversed)
		c.False(unrelated.ContradictedPrereqs(), "reversed %t: and is not marked as caught in the contradiction",
			reversed)
		c.NotEqual("", unrelated.UnsatisfiedReason, "reversed %t: and records why", reversed)
		c.Equal(fxp.FromInteger(13), broadsword.LevelData.Level,
			"reversed %t: the level counts the bonuses of the enabled traits only", reversed)
		var data CellData
		requires.CellData(TraitDescriptionColumn, &data)
		c.NotEqual("", data.PrereqContradiction,
			"reversed %t: the trait whose prerequisite is met is nonetheless shown to be caught in the contradiction",
			reversed)
		c.Equal("", data.UnsatisfiedReason,
			"reversed %t: without claiming that its prerequisite is unmet, which is what the table flags", reversed)
		data = CellData{}
		excludes.CellData(TraitDescriptionColumn, &data)
		c.True(strings.HasPrefix(data.UnsatisfiedReason, excludes.UnsatisfiedReason),
			"reversed %t: the trait whose prerequisite is unmet is shown why", reversed)
		c.True(len(data.UnsatisfiedReason) > len(excludes.UnsatisfiedReason),
			"reversed %t: and that it is caught in the contradiction", reversed)
		c.Equal("", data.PrereqContradiction,
			"reversed %t: within the reason its prerequisite is unmet, not alongside it", reversed)
		data = CellData{}
		unrelated.CellData(TraitDescriptionColumn, &data)
		c.True(strings.HasPrefix(data.UnsatisfiedReason, unrelated.UnsatisfiedReason),
			"reversed %t: the trait the sheet disabled is shown why", reversed)
		c.True(len(data.UnsatisfiedReason) > len(unrelated.UnsatisfiedReason),
			"reversed %t: and that the sheet disabled it", reversed)
		c.Equal("", data.PrereqContradiction,
			"reversed %t: and is not shown as caught in the contradiction", reversed)
		checkRecalculationConsistency(c, e, fmt.Sprintf("reversed %t, enforced", reversed))
		c.Equal(int32(1), warnings.Load(),
			"reversed %t: recalculating again while the data still never settles logs nothing more", reversed)

		e.SheetSettings.EnforceTraitPrereqs = false
		e.Recalculate()
		c.False(e.unsettled, "reversed %t: without enforcement, there is nothing to contradict", reversed)
		c.True(unrelated.Enabled(), "reversed %t: without enforcement, every trait is enabled", reversed)
		c.False(excludes.ContradictedPrereqs(), "reversed %t: and no trait is marked as caught in a contradiction",
			reversed)
		c.NotEqual("", unrelated.UnsatisfiedReason, "reversed %t: and the unmet prerequisite is still flagged",
			reversed)
		c.Equal(fxp.FromInteger(14), broadsword.LevelData.Level, "reversed %t: every trait's bonus counts", reversed)
		checkRecalculationConsistency(c, e, fmt.Sprintf("reversed %t, unenforced", reversed))

		// Enforcing the prerequisites again finds the contradiction afresh, and says so again.
		e.SheetSettings.EnforceTraitPrereqs = true
		e.Recalculate()
		c.True(e.unsettled, "reversed %t: enforcing the prerequisites again finds the contradiction again", reversed)
		c.Equal(int32(2), warnings.Load(), "reversed %t: and logs it again", reversed)
	}
}

// TestEntityEnforceTraitPrereqsMutualExclusionIsSymmetric verifies that two traits which each require the absence of
// the other are treated alike, whichever is listed first: neither is disabled, since disabling one of them and not the
// other would be arbitrary, and both show the unmet prerequisite. The same holds for three traits in a ring, where the
// passes cycle through several states before repeating one.
func TestEntityEnforceTraitPrereqsMutualExclusionIsSymmetric(t *testing.T) {
	c := check.New(t)
	countLogs(t, slog.LevelWarn)
	for _, count := range []int{2, 3} {
		for _, reversed := range []bool{false, true} {
			e := NewEntity()
			e.SheetSettings.EnforceTraitPrereqs = true
			broadsword := addTestSkill(e, "Broadsword", "", "", fxp.Four) // IQ+1, so 11 on its own
			ring := make([]*Trait, count)
			for i := range ring {
				ring[i] = NewTrait(e, nil, false)
				ring[i].Name = "Ring " + strconv.Itoa(i+1)
				ring[i].Prereq = newPrereqListForbiddingTrait("Ring " + strconv.Itoa((i+1)%count+1))
				ring[i].Features = Features{newSkillBonusTo("Broadsword", fxp.One)}
			}
			e.Traits = append(e.Traits, ring...)
			if reversed {
				slices.Reverse(e.Traits)
			}

			e.Recalculate()
			c.True(e.unsettled, "ring of %d, reversed %t: the sheet notes that its data never settles", count,
				reversed)
			for i, trait := range ring {
				c.True(trait.Enabled(), "ring of %d, reversed %t: trait %d is left enabled", count, reversed, i+1)
				c.NotEqual("", trait.UnsatisfiedReason, "ring of %d, reversed %t: trait %d shows the contradiction",
					count, reversed, i+1)
			}
			c.Equal(fxp.FromInteger(11+count), broadsword.LevelData.Level,
				"ring of %d, reversed %t: every trait's bonus counts", count, reversed)
			checkRecalculationConsistency(c, e, fmt.Sprintf("ring of %d, reversed %t", count, reversed))
		}
	}
}

// TestEntityEnforceTraitPrereqsContradictionInContainer verifies what a contradiction involving a container does to
// the container's children. "Flipping" is a container that requires the absence of "Excludes", which in turn requires
// the absence of "Flipping", so the two flip together and are left enabled. The children of the container are judged
// only while it is enabled, so their verdicts are missing from the states in which it is not, which must not be
// mistaken for their flipping: a child whose prerequisite is plainly unmet is disabled just as the same trait at the
// top level is, a child whose prerequisite is met stays enabled, and a child whose prerequisite is the absence of the
// other trait caught in the contradiction is disabled, since that trait is left enabled. Unlike the same trait at the
// top level (see TestEntityEnforceTraitPrereqsContradictionSweepsUpDependents), that last child is not swept up, since
// while its container is disabled it is unjudged rather than enabled, and so is never seen to flip.
func TestEntityEnforceTraitPrereqsContradictionInContainer(t *testing.T) {
	c := check.New(t)
	countLogs(t, slog.LevelWarn)
	for _, reversed := range []bool{false, true} {
		e := NewEntity()
		e.SheetSettings.EnforceTraitPrereqs = true
		broadsword := addTestSkill(e, "Broadsword", "", "", fxp.Four) // IQ+1, so 11 on its own
		flipping := NewTrait(e, nil, true)
		flipping.Name = "Flipping"
		flipping.Prereq = newPrereqListForbiddingTrait("Excludes")
		excludes := NewTrait(e, nil, false)
		excludes.Name = "Excludes"
		excludes.Prereq = newPrereqListForbiddingTrait("Flipping")
		nestedUnmet := newTraitNeedingMissingTrait(e, "Nested Unmet")
		nestedMet := newTraitRequiring(e, "Nested Met", "Excludes")
		nestedForbidding := NewTrait(e, nil, false)
		nestedForbidding.Name = "Nested Forbidding"
		nestedForbidding.Prereq = newPrereqListForbiddingTrait("Excludes")
		for _, child := range []*Trait{nestedUnmet, nestedMet, nestedForbidding} {
			child.SetParent(flipping)
			flipping.Children = append(flipping.Children, child)
		}
		topLevelUnmet := newTraitNeedingMissingTrait(e, "Top Level Unmet")
		for _, trait := range []*Trait{excludes, nestedUnmet, nestedMet, nestedForbidding, topLevelUnmet} {
			trait.Features = Features{newSkillBonusTo("Broadsword", fxp.One)}
		}
		e.Traits = append(e.Traits, flipping, excludes, topLevelUnmet)
		if reversed {
			slices.Reverse(e.Traits)
			slices.Reverse(flipping.Children)
		}

		e.Recalculate()
		c.True(e.unsettled, "reversed %t: the sheet notes that its data never settles", reversed)
		c.True(flipping.Enabled(), "reversed %t: the container caught in the contradiction is left enabled", reversed)
		c.True(flipping.ContradictedPrereqs(), "reversed %t: and is marked as caught in it", reversed)
		c.True(excludes.Enabled(), "reversed %t: the trait caught in the contradiction with it is left enabled",
			reversed)
		c.True(excludes.ContradictedPrereqs(), "reversed %t: and is marked as caught in it", reversed)
		c.False(nestedUnmet.Enabled(), "reversed %t: a child with a plainly unmet prerequisite is disabled", reversed)
		c.True(nestedUnmet.DisabledByPrereqs(), "reversed %t: by the sheet", reversed)
		c.NotEqual("", nestedUnmet.UnsatisfiedReason, "reversed %t: and records why", reversed)
		c.False(topLevelUnmet.Enabled(), "reversed %t: just as the same trait at the top level is", reversed)
		c.True(nestedMet.Enabled(), "reversed %t: a child whose prerequisite is met stays enabled", reversed)
		c.Equal("", nestedMet.UnsatisfiedReason, "reversed %t: with nothing to show", reversed)
		c.False(nestedForbidding.Enabled(),
			"reversed %t: a child requiring the absence of a trait left enabled is disabled", reversed)
		c.True(nestedForbidding.DisabledByPrereqs(), "reversed %t: by the sheet", reversed)
		c.False(nestedForbidding.ContradictedPrereqs(),
			"reversed %t: a child judged only while its container is enabled is not seen to flip, so is not swept up",
			reversed)
		c.Equal(fxp.FromInteger(13), broadsword.LevelData.Level,
			"reversed %t: only the traits actually in effect contribute their bonuses", reversed)
		checkRecalculationConsistency(c, e, fmt.Sprintf("reversed %t", reversed))
	}
}

// TestEntityEnforceTraitPrereqsContradictionAmongChildren verifies that children of an enabled container which are
// caught in a contradiction among themselves are marked and left enabled, just as the same traits at the top level
// are, since the marking walks the children of every container. "Ring 1" and "Ring 2", each requiring the other's
// absence, sit inside "Martial Arts", which has no prerequisites of its own and so is never seen to flip, alongside a
// sibling whose prerequisite is plainly unmet, which is disabled as usual.
func TestEntityEnforceTraitPrereqsContradictionAmongChildren(t *testing.T) {
	c := check.New(t)
	countLogs(t, slog.LevelWarn)
	for _, reversed := range []bool{false, true} {
		e := NewEntity()
		e.SheetSettings.EnforceTraitPrereqs = true
		broadsword := addTestSkill(e, "Broadsword", "", "", fxp.Four) // IQ+1, so 11 on its own
		container := NewTrait(e, nil, true)
		container.Name = "Martial Arts"
		ring := make([]*Trait, 2)
		for i := range ring {
			ring[i] = NewTrait(e, nil, false)
			ring[i].Name = "Ring " + strconv.Itoa(i+1)
			ring[i].Prereq = newPrereqListForbiddingTrait("Ring " + strconv.Itoa((i+1)%len(ring)+1))
		}
		unmet := newTraitNeedingMissingTrait(e, "Unmet")
		for _, child := range []*Trait{ring[0], ring[1], unmet} {
			child.Features = Features{newSkillBonusTo("Broadsword", fxp.One)}
			child.SetParent(container)
			container.Children = append(container.Children, child)
		}
		if reversed {
			slices.Reverse(container.Children)
		}
		e.Traits = append(e.Traits, container)

		e.Recalculate()
		c.True(e.unsettled, "reversed %t: the sheet notes that its data never settles", reversed)
		c.True(container.Enabled(), "reversed %t: the container is enabled", reversed)
		c.False(container.ContradictedPrereqs(), "reversed %t: and is not caught in the contradiction", reversed)
		for i, trait := range ring {
			c.True(trait.Enabled(), "reversed %t: child %d caught in the contradiction is left enabled", reversed, i+1)
			c.True(trait.ContradictedPrereqs(), "reversed %t: child %d is marked as caught in it", reversed, i+1)
			c.NotEqual("", trait.UnsatisfiedReason, "reversed %t: child %d shows the contradiction", reversed, i+1)
		}
		c.False(unmet.Enabled(), "reversed %t: a child whose prerequisite is plainly unmet is disabled", reversed)
		c.True(unmet.DisabledByPrereqs(), "reversed %t: by the sheet", reversed)
		c.False(unmet.ContradictedPrereqs(), "reversed %t: and is not marked as caught in the contradiction", reversed)
		c.Equal(fxp.FromInteger(13), broadsword.LevelData.Level,
			"reversed %t: the bonuses of the children left enabled count", reversed)
		checkRecalculationConsistency(c, e, fmt.Sprintf("reversed %t", reversed))
	}
}

// TestEntityEnforceTraitPrereqsContradictionSweepsUpDependents verifies that a trait whose prerequisites turn on a
// trait caught in a contradiction, and which flips along with it as a result, is marked as caught in the contradiction
// and left enabled too, whether it requires that trait's presence or its absence, and although nothing turns on it in
// return; see Recalculate for why. "Ring 1" and "Ring 2" each require the other's absence. "Follower" requires the
// absence of "Ring 1" and "Requirer" its presence, and nothing depends on either. Both flip with the ring and both are
// marked and left enabled: Follower with its unmet prerequisite shown along with the contradiction, Requirer with the
// note that its prerequisite is met only because of it, and the bonuses of all four count. A trait whose prerequisite
// is unmet whatever the ring does is disabled as usual.
func TestEntityEnforceTraitPrereqsContradictionSweepsUpDependents(t *testing.T) {
	c := check.New(t)
	countLogs(t, slog.LevelWarn)
	for _, reversed := range []bool{false, true} {
		e := NewEntity()
		e.SheetSettings.EnforceTraitPrereqs = true
		broadsword := addTestSkill(e, "Broadsword", "", "", fxp.Four) // IQ+1, so 11 on its own
		ring1 := NewTrait(e, nil, false)
		ring1.Name = "Ring 1"
		ring1.Prereq = newPrereqListForbiddingTrait("Ring 2")
		ring2 := NewTrait(e, nil, false)
		ring2.Name = "Ring 2"
		ring2.Prereq = newPrereqListForbiddingTrait("Ring 1")
		follower := NewTrait(e, nil, false)
		follower.Name = "Follower"
		follower.Prereq = newPrereqListForbiddingTrait("Ring 1")
		requirer := newTraitRequiring(e, "Requirer", "Ring 1")
		unrelated := newTraitNeedingMissingTrait(e, "Unrelated")
		for _, trait := range []*Trait{ring1, ring2, follower, requirer, unrelated} {
			trait.Features = Features{newSkillBonusTo("Broadsword", fxp.One)}
			e.Traits = append(e.Traits, trait)
		}
		if reversed {
			slices.Reverse(e.Traits)
		}

		e.Recalculate()
		c.True(e.unsettled, "reversed %t: the sheet notes that its data never settles", reversed)
		for _, trait := range []*Trait{ring1, ring2, follower, requirer} {
			c.True(trait.Enabled(), "reversed %t: %s is left enabled", reversed, trait.Name)
			c.True(trait.ContradictedPrereqs(), "reversed %t: %s is marked as caught in the contradiction", reversed,
				trait.Name)
		}
		c.NotEqual("", follower.UnsatisfiedReason,
			"reversed %t: the trait requiring the absence of a trait left enabled shows its unmet prerequisite",
			reversed)
		c.Equal("", requirer.UnsatisfiedReason,
			"reversed %t: the trait requiring the presence of a trait left enabled has its prerequisite met", reversed)
		var data CellData
		requirer.CellData(TraitDescriptionColumn, &data)
		c.NotEqual("", data.PrereqContradiction, "reversed %t: but is shown to be caught in the contradiction",
			reversed)
		c.Equal("", data.UnsatisfiedReason, "reversed %t: without being flagged as unsatisfied", reversed)
		c.False(unrelated.Enabled(), "reversed %t: a trait whose prerequisite is unmet regardless is disabled", reversed)
		c.False(unrelated.ContradictedPrereqs(), "reversed %t: and is not marked as caught in the contradiction",
			reversed)
		c.Equal(fxp.FromInteger(15), broadsword.LevelData.Level,
			"reversed %t: the bonuses of the four traits left enabled count", reversed)
		checkRecalculationConsistency(c, e, fmt.Sprintf("reversed %t", reversed))
	}
}

// TestEntityEnforceTraitPrereqsNeverRepeatingStateStopsAtCap verifies what a recalculation does with data that never
// settles and never repeats a state either. Thirty-two traits each have a prerequisite script that consults a random
// number, so the set of traits enabled after a pass is effectively never seen twice and the passes have to be stopped
// at their cap. The recalculation must return, note that the data never settles, treat the traits that kept flipping
// as caught in a contradiction and leave them enabled, and still disable a trait whose prerequisite is plainly unmet
// and leave one whose prerequisite is plainly met alone. Two traits that each require the other's absence flip on
// every pass, so both must be left enabled, whichever of them the pass that reached the cap happened to leave
// disabled, and a trait requiring one of them flips along with it and is swept up with them.
func TestEntityEnforceTraitPrereqsNeverRepeatingStateStopsAtCap(t *testing.T) {
	c := check.New(t)
	countLogs(t, slog.LevelWarn)
	e := NewEntity()
	e.SheetSettings.EnforceTraitPrereqs = true
	const randomTraits = 32
	random := make([]*Trait, randomTraits)
	for i := range random {
		random[i] = NewTrait(e, nil, false)
		random[i].Name = "Random " + strconv.Itoa(i+1)
		list := NewPrereqList()
		script := NewScriptPrereq()
		script.Parent = list
		script.Script = `Math.random() < 0.5 ? "" : "The dice say no"`
		list.Prereqs = append(list.Prereqs, script)
		random[i].Prereq = list
		e.Traits = append(e.Traits, random[i])
	}
	ring := make([]*Trait, 2)
	for i := range ring {
		ring[i] = NewTrait(e, nil, false)
		ring[i].Name = "Ring " + strconv.Itoa(i+1)
		ring[i].Prereq = newPrereqListForbiddingTrait("Ring " + strconv.Itoa((i+1)%len(ring)+1))
		e.Traits = append(e.Traits, ring[i])
	}
	stable := NewTrait(e, nil, false)
	stable.Name = "Stable"
	unmet := newTraitNeedingMissingTrait(e, "Unmet")
	met := newTraitRequiring(e, "Met", "Stable")
	requirer := newTraitRequiring(e, "Requirer", "Ring 1")
	e.Traits = append(e.Traits, stable, unmet, met, requirer)

	e.Recalculate()
	c.True(e.unsettled, "the sheet notes that its data never settles")
	c.False(unmet.Enabled(), "a trait whose prerequisite is plainly unmet is disabled")
	c.NotEqual("", unmet.UnsatisfiedReason, "and records why")
	c.True(met.Enabled(), "a trait whose prerequisite is plainly met is left alone")
	c.False(met.ContradictedPrereqs(), "and is not marked as caught in the contradiction")
	c.Equal("", met.UnsatisfiedReason, "with nothing to show")
	for i, trait := range ring {
		c.True(trait.Enabled(), "ring trait %d is left enabled", i+1)
		c.True(trait.ContradictedPrereqs(), "ring trait %d is caught in the contradiction", i+1)
		c.NotEqual("", trait.UnsatisfiedReason, "ring trait %d shows the contradiction", i+1)
	}
	c.True(requirer.Enabled(), "a trait requiring a ring trait is left enabled")
	c.True(requirer.ContradictedPrereqs(), "and is swept up into the contradiction")
	// The verdict on a random trait cannot be predicted, and evaluating its prerequisite again would roll the dice
	// again, so what is checked is what the cap path guarantees whatever the dice said: a trait marked as caught in
	// the contradiction is left enabled, a trait the sheet disabled has a reason to show for it, and a trait it left
	// enabled without marking it has its prerequisite met as last judged. Nearly every random trait flips back and
	// forth over the passes the cap path judges and so is marked, but only what is certain is asserted.
	marked := 0
	for i, trait := range random {
		switch {
		case trait.ContradictedPrereqs():
			marked++
			c.True(trait.Enabled(), "random trait %d: a trait caught in the contradiction is left enabled", i+1)
		case trait.DisabledByPrereqs():
			c.NotEqual("", trait.UnsatisfiedReason, "random trait %d: disabled by the sheet only for a reason", i+1)
		default:
			c.True(trait.Enabled(), "random trait %d: a trait neither marked nor disabled is enabled", i+1)
			c.Equal("", trait.UnsatisfiedReason,
				"random trait %d: and is left enabled without being marked only with its prerequisite met", i+1)
		}
	}
	c.True(marked > 0, "some random trait is marked as caught in the contradiction, rather than %d", marked)
}

// newNestedContradictionRings adds depth rings of two traits that each require the other's absence to the entity, each
// ring after the first nested inside the first trait of the ring before it, and returns them outermost first. A nested
// ring is judged only while the container above it is enabled, so it is not seen to flip until the ring above it has
// been marked as caught in a contradiction and left enabled, and each ring costs the recalculation a round of its own;
// see Entity.Recalculate.
func newNestedContradictionRings(e *Entity, depth int) [][2]*Trait {
	rings := make([][2]*Trait, depth)
	var parent *Trait
	for i := range rings {
		for j := range rings[i] {
			trait := NewTrait(e, parent, j == 0 && i < depth-1)
			trait.Name = fmt.Sprintf("Ring %d-%d", i+1, j+1)
			trait.Prereq = newPrereqListForbiddingTrait(fmt.Sprintf("Ring %d-%d", i+1, 2-j))
			if parent == nil {
				e.Traits = append(e.Traits, trait)
			} else {
				parent.Children = append(parent.Children, trait)
			}
			rings[i][j] = trait
		}
		parent = rings[i][0]
	}
	return rings
}

// markedRings returns how many of the rings have both traits marked as caught in a contradiction.
func markedRings(rings [][2]*Trait) int {
	count := 0
	for _, ring := range rings {
		if ring[0].ContradictedPrereqs() && ring[1].ContradictedPrereqs() {
			count++
		}
	}
	return count
}

// TestEntityRecalculatePassBudgetSpansTheRounds verifies that the cap on the passes a recalculation makes is a total
// across its rounds rather than a cap on each round alone, so that data that keeps prompting further rounds cannot
// cost a round per trait. Three rings of two traits that each require the other's absence are nested one inside the
// next (see newNestedContradictionRings), so each ring costs a round of its own and the full recalculation takes four
// and marks the rings one round at a time. Given any smaller budget, the recalculation makes exactly that many passes
// and stops, however many rounds that spans, reports that the data never settles, has marked the rings the rounds it
// completed reached, and comes out the same for the same budget every time, since it starts from the sheet's data
// alone.
func TestEntityRecalculatePassBudgetSpansTheRounds(t *testing.T) {
	c := check.New(t)
	countLogs(t, slog.LevelWarn)
	e := NewEntity()
	e.SheetSettings.EnforceTraitPrereqs = true
	rings := newNestedContradictionRings(e, 3)

	total := e.recalculate(maxRecalculationPassesInAll)
	c.True(total < maxRecalculationPassesInAll, "the rings are unwound within the budget, rather than in %d passes",
		total)
	c.True(e.unsettled, "the sheet notes that its data never settles")
	for i, ring := range rings {
		for j, trait := range ring {
			c.True(trait.Enabled(), "ring %d, trait %d is left enabled", i+1, j+1)
			c.True(trait.ContradictedPrereqs(), "ring %d, trait %d is marked as caught in the contradiction", i+1, j+1)
		}
	}
	checkRecalculationConsistency(c, e, "full budget")

	marked := make([]int, 0, total)
	for budget := 1; budget < total; budget++ {
		c.Equal(budget, e.recalculate(budget), "budget %d: the rounds stop once the budget is spent", budget)
		c.True(e.unsettled, "budget %d: and the sheet notes that its data never settles", budget)
		marked = append(marked, markedRings(rings))
		state := recalculationState(e)
		c.Equal(budget, e.recalculate(budget), "budget %d: recalculating again makes the same passes", budget)
		c.Equal(state, recalculationState(e), "budget %d: and comes out the same", budget)
	}
	c.True(slices.IsSorted(marked), "a ring once marked stays marked as the budget grows: %v", marked)
	for count := range len(rings) + 1 {
		c.True(slices.Contains(marked, count),
			"some budget stops the rounds with exactly %d rings marked, so the rounds run within the budget: %v",
			count, marked)
	}
	c.Equal(total, e.recalculate(total+1), "a budget beyond what the rounds need leaves them to settle")
	c.True(e.unsettled, "and the sheet still notes that its data never settles")
}

// TestEntityRecalculateLevelCycleIsStable verifies that a recalculation that never settles because skill levels alone
// cycle, with no trait to mark as caught in a contradiction, leaves the sheet in the same state every time. Two skills
// each have an "any of" list of an equipped-equipment prerequisite no equipment satisfies and the other skill at a
// level of at most 8, and both are at 11 on their own: while the other is at 11 the list is unmet and the skill takes
// the equipment penalty, which brings it to 6, at which the other's list is met and its penalty lifts, and so on
// forever. Which state of that cycle the passes are found repeating in depends on the levels the previous
// recalculation left, so a recalculation that ended wherever that happened to be would leave the sheet penalized after
// one edit and not after the next. The state chosen must instead depend on the cycle alone: neither on what else is on
// the sheet nor on the order the skills are listed in.
func TestEntityRecalculateLevelCycleIsStable(t *testing.T) {
	c := check.New(t)
	warnings := countLogs(t, slog.LevelWarn)
	e := NewEntity()
	first := newSkillCappingOther(e, "First", "Second")
	second := newSkillCappingOther(e, "Second", "First")
	e.Skills = append(e.Skills, first, second)

	e.Recalculate()
	c.True(e.unsettled, "the sheet notes that its data never settles")
	c.Equal(int32(1), warnings.Load(), "and logs that once")
	level := first.LevelData.Level
	c.True(level == fxp.FromInteger(11) || level == fxp.FromInteger(6),
		"the sheet is left in one of the states of the cycle, not at %v", level)
	c.Equal(level, second.LevelData.Level, "the two skills are treated alike")
	checkRecalculationConsistency(c, e, "as listed")
	for i := range 3 {
		e.Recalculate()
		c.Equal(level, first.LevelData.Level, "further recalculation %d leaves the sheet in the same state", i+1)
		c.Equal(level, second.LevelData.Level, "further recalculation %d leaves the sheet in the same state", i+1)
	}
	c.Equal(int32(1), warnings.Load(), "recalculating again while the data still never settles logs nothing more")

	unrelated := addTestSkill(e, "Unrelated", "", "", fxp.One)
	e.Recalculate()
	c.Equal(level, first.LevelData.Level, "an unrelated skill does not change which state of the cycle is chosen")
	c.Equal(level, second.LevelData.Level, "an unrelated skill does not change which state of the cycle is chosen")
	unrelated.Points = fxp.Four
	e.Recalculate()
	c.Equal(level, first.LevelData.Level, "nor does editing it")
	c.Equal(level, second.LevelData.Level, "nor does editing it")
	checkRecalculationConsistency(c, e, "with an unrelated skill")

	slices.Reverse(e.Skills)
	e.Recalculate()
	c.Equal(level, first.LevelData.Level, "nor does the order the skills are listed in")
	c.Equal(level, second.LevelData.Level, "nor does the order the skills are listed in")
	checkRecalculationConsistency(c, e, "reversed")
}

// newSkillCappingOther creates an IQ/Average skill with 4 points, so at 11 on its own, whose prerequisites are met by
// either an equipped piece of equipment that no equipment can be, or the skill with the other name being at a level of
// at most 8. It is not added to the entity.
func newSkillCappingOther(e *Entity, name, other string) *Skill {
	s := NewSkill(e, nil, false)
	s.Name = name
	s.Difficulty.Attribute = IntelligenceID
	s.Difficulty.Difficulty = difficulty.Average
	s.Points = fxp.Four
	s.Prereq = NewPrereqList()
	s.Prereq.All = false
	p := NewEquippedEquipmentPrereq()
	p.Parent = s.Prereq
	p.NameCriteria.Qualifier = "Nothing"
	s.Prereq.Prereqs = append(s.Prereq.Prereqs, p)
	addSkillPrereq(s.Prereq, other, fxp.Eight).LevelCriteria.Compare = criteria.AtMostNumber
	return s
}

// TestEntityRecalculateSeesLevelsScriptsRecompute verifies that a level a script recomputes in the middle of a pass
// is not missed. A prerequisite script that reads a skill's level recomputes the level as it reads it, so by the time
// the pass recomputes the levels itself, a change to that one is already in place and the pass finds nothing to
// change. "Dependent" requires Broadsword at 12 or better and is judged before "Scripted", whose script reads
// Broadsword's level, so on the pass in which the bonus that brought Broadsword to 12 is withdrawn, Dependent is
// judged against 12 while Scripted's script brings the level to 11, and a further pass is needed to judge Dependent
// against it. The recalculation must make that pass rather than stop with Dependent judged against the level it no
// longer has.
func TestEntityRecalculateSeesLevelsScriptsRecompute(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	broadsword := addTestSkill(e, "Broadsword", "", "", fxp.Four) // IQ+1, so 11 on its own
	dependent := addTestSkill(e, "Dependent", "", "", fxp.One)
	dependent.Prereq = newPrereqListRequiringSkill("Broadsword", fxp.Twelve)
	scripted := addTestSkill(e, "Scripted", "", "", fxp.One)
	scripted.Prereq = NewPrereqList()
	script := NewScriptPrereq()
	script.Parent = scripted.Prereq
	script.Script = `entity.skillLevel("Broadsword") >= 12 ? "" : "Broadsword is below 12"`
	scripted.Prereq.Prereqs = append(scripted.Prereq.Prereqs, script)
	bonus := NewTrait(e, nil, false)
	bonus.Name = "Bonus"
	bonus.Features = Features{newSkillBonusTo("Broadsword", fxp.One)}
	e.Traits = append(e.Traits, bonus)

	e.Recalculate()
	c.Equal(fxp.Twelve, broadsword.LevelData.Level, "the bonus brings Broadsword to 12")
	c.Equal("", dependent.UnsatisfiedReason, "the skill prerequisite is met")
	c.Equal("", scripted.UnsatisfiedReason, "the script prerequisite is met")
	checkRecalculationConsistency(c, e, "bonus in effect")

	bonus.Disabled = true
	e.Recalculate()
	c.Equal(fxp.Eleven, broadsword.LevelData.Level, "without the bonus, Broadsword is at 11")
	c.NotEqual("", scripted.UnsatisfiedReason, "the script prerequisite is unmet")
	c.NotEqual("", dependent.UnsatisfiedReason,
		"the skill prerequisite judged before the script recomputed the level is unmet too")
	checkRecalculationConsistency(c, e, "bonus withdrawn")
}

// TestEntityDiscardCachesKeepsAbandonedScriptsBetweenPasses verifies that discarding the caches between the passes
// of a recalculation keeps the result recorded for a script that was stopped before it could produce an answer, so
// that a runaway script costs its permitted execution time once per recalculation rather than once per pass, while
// every completed result is discarded as usual, and that a recalculation begins by discarding both.
func TestEntityDiscardCachesKeepsAbandonedScriptsBetweenPasses(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	completed := scriptResolveKey{id: "completed", text: "1 + 1"}
	abandoned := scriptResolveKey{id: "abandoned", text: "for (;;);"}
	seed := func() {
		e.scriptCache[completed] = scriptResolveResult{text: "2"}
		e.scriptCache[abandoned] = scriptResolveResult{text: "script execution timed out", abandoned: true}
	}

	seed()
	e.discardCaches(true)
	_, exists := e.scriptCache[completed]
	c.False(exists, "a completed result is discarded between passes")
	result, exists := e.scriptCache[abandoned]
	c.True(exists, "an abandoned result is kept between passes")
	c.True(result.abandoned, "and is still marked as abandoned, so its readers still know not to trust it")

	seed()
	e.DiscardCaches()
	c.Equal(0, len(e.scriptCache), "discarding the caches outright discards both")
}

// TestEntityEnforceTraitPrereqsHonorsUserDisabling verifies that the setting only acts on traits the user has left
// enabled: a trait the user disabled has no prerequisites to enforce and no stale flag is left behind, and a trait
// inside a container the user disabled is left alone too.
func TestEntityEnforceTraitPrereqsHonorsUserDisabling(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	e.SheetSettings.EnforceTraitPrereqs = true
	trait := newTraitNeedingMissingTrait(e, "Trained by a Master")
	e.Traits = append(e.Traits, trait)

	e.Recalculate()
	c.True(trait.DisabledByPrereqs(), "an enabled trait with an unmet prereq is disabled by the setting")

	trait.Disabled = true
	e.Recalculate()
	c.Equal("", trait.UnsatisfiedReason, "a trait the user disabled has no prerequisites to enforce")
	c.Equal(prereqsNotJudged, trait.prereqVerdict, "disabling the trait clears the verdict the setting had reached")
	c.False(trait.DisabledByPrereqs(), "a trait the user disabled is not reported as disabled by its prereqs")

	trait.Disabled = false
	e.Recalculate()
	c.True(trait.DisabledByPrereqs(), "re-enabling the trait lets the setting disable it again")

	e = NewEntity()
	e.SheetSettings.EnforceTraitPrereqs = true
	container := NewTrait(e, nil, true)
	container.Name = "Martial Arts"
	container.Disabled = true
	nested := newTraitNeedingMissingTrait(e, "Weapon Master")
	nested.SetParent(container)
	container.Children = append(container.Children, nested)
	e.Traits = append(e.Traits, container)

	e.Recalculate()
	c.Equal("", nested.UnsatisfiedReason, "a trait inside a disabled container has no prerequisites to enforce")
	c.Equal(prereqsNotJudged, nested.prereqVerdict, "a trait inside a disabled container is not judged by the setting")
}

// TestEntityEnforceTraitPrereqsContainers verifies that a container disabled for unsatisfied prerequisites takes
// its children out of play with it, that those children have no prerequisites of their own to enforce while it is
// disabled, and that they are checked again once the container's prerequisites are met.
func TestEntityEnforceTraitPrereqsContainers(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	e.SheetSettings.EnforceTraitPrereqs = true
	container := NewTrait(e, nil, true)
	container.Name = "Martial Arts"
	container.Prereq = newPrereqListRequiringTrait("Combat Reflexes")
	plain := NewTrait(e, container, false)
	plain.Name = "Weapon Bond"
	plain.BasePoints = fxp.One
	nested := newTraitRequiring(e, "Weapon Master", "Trained by a Master")
	nested.SetParent(container)
	container.Children = append(container.Children, plain, nested)
	e.Traits = append(e.Traits, container)

	e.Recalculate()
	c.True(container.DisabledByPrereqs(), "the container with the unmet prereq is disabled")
	c.False(plain.Enabled(), "a child of the disabled container is disabled with it")
	c.False(plain.DisabledByPrereqs(), "the child itself is not the one the setting disabled")
	c.Equal(fxp.Int(0), plain.AdjustedPoints(nil), "the child's points do not count")
	c.Equal(fxp.Int(0), container.AdjustedPoints(nil), "the container's points do not count")
	c.Equal(fxp.Int(0), e.PointsBreakdown().Total(), "nothing in the container counts toward the total")
	c.Equal("", nested.UnsatisfiedReason, "a child of the disabled container has no prerequisites to enforce")
	c.Equal(prereqsNotJudged, nested.prereqVerdict, "a child of the disabled container is not judged by the setting")

	combatReflexes := NewTrait(e, nil, false)
	combatReflexes.Name = "Combat Reflexes"
	e.Traits = append(e.Traits, combatReflexes)
	e.Recalculate()
	c.False(container.DisabledByPrereqs(), "meeting the prereq re-enables the container")
	c.True(plain.Enabled(), "and its child")
	c.Equal(fxp.One, plain.AdjustedPoints(nil), "the child's points count again")
	c.True(nested.DisabledByPrereqs(), "the child with its own unmet prereq is checked again and disabled")
	c.NotEqual("", nested.UnsatisfiedReason, "the child with its own unmet prereq records why")
}

// TestEntityEnforceTraitPrereqsMaxLevel verifies that a trait above its maximum level, which the sheet flags the same
// way as an unmet prerequisite, is disabled by the setting as well, and re-enabled once its level is brought back
// within the maximum.
func TestEntityEnforceTraitPrereqsMaxLevel(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	e.SheetSettings.EnforceTraitPrereqs = true
	trait := NewTrait(e, nil, false)
	trait.Name = "Extra Arm"
	trait.CanLevel = true
	trait.Levels = fxp.Three
	trait.MaxLevels = "2"
	e.Traits = append(e.Traits, trait)

	e.Recalculate()
	c.NotEqual("", trait.UnsatisfiedReason, "a trait above its maximum level is flagged")
	c.True(trait.DisabledByPrereqs(), "a trait above its maximum level is disabled by the setting")
	checkRecalculationConsistency(c, e, "above the maximum")

	trait.Levels = fxp.Two
	e.Recalculate()
	c.Equal("", trait.UnsatisfiedReason, "a trait within its maximum level is not flagged")
	c.False(trait.DisabledByPrereqs(), "a trait within its maximum level is enabled again")
	checkRecalculationConsistency(c, e, "within the maximum")
}

// TestEntityReactionsUseResolvedSelfControl verifies that the reaction penalty derived from a trait's self-control
// roll honors selector overrides on both the roll and the adjustment, matching what the sheet displays for the trait.
func TestEntityReactionsUseResolvedSelfControl(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	trait := NewTrait(e, nil, false)
	trait.Name = "Greed"
	trait.SelfControl = selfctrl.CR12
	trait.SelfControlAdj = selfctrl.ReactionPenalty
	e.Traits = append(e.Traits, trait)
	e.Recalculate()
	reactions := e.Reactions()
	c.Equal(1, len(reactions), "the trait's self-control roll produces a reaction")
	if len(reactions) == 1 {
		c.Equal(fxp.FromInteger(selfctrl.CR12.Penalty()), reactions[0].Total(), "the penalty derives from CR12")
	}

	// Override the roll to CR6; the reaction penalty must track the overridden roll.
	trait.Features = append(trait.Features, newTraitSelectorOverride(selector.TraitSelfControlRoll, "Greed",
		strconv.Itoa(int(selfctrl.CR6))))
	e.Recalculate()
	reactions = e.Reactions()
	c.Equal(1, len(reactions), "the overridden roll still produces a reaction")
	if len(reactions) == 1 {
		c.Equal(fxp.FromInteger(selfctrl.CR6.Penalty()), reactions[0].Total(), "the penalty tracks the overridden CR6 roll")
	}

	// Override the adjustment away from a reaction penalty; the reaction must disappear.
	trait.Features = append(trait.Features, newTraitSelectorOverride(selector.TraitSelfControlAdjustment, "Greed",
		selfctrl.MajorCostOfLivingIncrease.Key()))
	e.Recalculate()
	c.Equal(0, len(e.Reactions()), "overriding the adjustment away from a reaction penalty removes the reaction")

	// The converse: a trait with a non-reaction adjustment that is overridden into one must gain the reaction.
	e = NewEntity()
	trait = addTraitWithFeatures(e, "Bad Temper", newTraitSelectorOverride(selector.TraitSelfControlAdjustment,
		"Bad Temper", selfctrl.ReactionPenalty.Key()))
	trait.SelfControl = selfctrl.CR9
	trait.SelfControlAdj = selfctrl.NoAdjustment
	e.Recalculate()
	reactions = e.Reactions()
	c.Equal(1, len(reactions), "overriding the adjustment into a reaction penalty adds the reaction")
	if len(reactions) == 1 {
		c.Equal(fxp.FromInteger(selfctrl.CR9.Penalty()), reactions[0].Total(), "the added penalty derives from CR9")
	}
}

func newTraitSelectorOverride(field selector.Field, traitName, value string) *SelectorOverride {
	override := NewSelectorOverride(field)
	override.Value = value
	override.NameCriteria.Qualifier = traitName
	return override
}

// TestEntityTraitLevels verifies that TraitLevels sums the current levels of every enabled, leveled trait with the
// given name, ignoring case, and reports whether any such trait exists. Disabled traits and containers contribute
// nothing, while the leveled children of a container do, and the callers that need a "not found" sentinel (the
// scripting traitLevel binding) and those that don't (TelekineticStrength) both derive from the same walk.
func TestEntityTraitLevels(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	levels, found := e.TraitLevels("Telekinesis")
	c.False(found, "no trait at all")
	c.Equal(fxp.Int(0), levels, "no trait at all contributes no levels")
	c.Equal("-1", ResolveScript(e, ScriptSelfProvider{}, `entity.traitLevel("Telekinesis")`),
		"the script binding reports -1 when the trait is absent")

	newLeveledTrait := func(name string, levels fxp.Int, parent *Trait) *Trait {
		trait := NewTrait(e, parent, false)
		trait.Name = name
		trait.CanLevel = true
		trait.Levels = levels
		return trait
	}
	e.Traits = append(e.Traits, newLeveledTrait("Telekinesis", fxp.Three, nil))
	container := NewTrait(e, nil, true)
	container.Name = "Telekinesis"
	container.Children = append(container.Children, newLeveledTrait("telekinesis", fxp.Two, container))
	e.Traits = append(e.Traits, container)
	disabled := newLeveledTrait("Telekinesis", fxp.Ten, nil)
	disabled.Disabled = true
	e.Traits = append(e.Traits, disabled)
	unleveled := NewTrait(e, nil, false)
	unleveled.Name = "Telekinesis"
	e.Traits = append(e.Traits, unleveled)
	e.Recalculate()

	levels, found = e.TraitLevels("TELEKINESIS")
	c.True(found, "a matching trait exists")
	c.Equal(fxp.Five, levels, "levels sum across the enabled leveled traits, including a container's children")
	c.Equal(fxp.Five, e.TelekineticStrength(), "TelekineticStrength reports the same sum")
	c.Equal("5", ResolveScript(e, ScriptSelfProvider{}, `entity.traitLevel("Telekinesis")`),
		"the script binding reports the same sum")

	levels, found = e.TraitLevels("Telekinesis (Reach)")
	c.False(found, "the name must match in full")
	c.Equal(fxp.Int(0), levels, "a non-matching name contributes no levels")
}
