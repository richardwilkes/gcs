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
	"slices"
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

// TestEntityProcessPrereqsClearsUnsatisfiedReasonWhenDisabled verifies that disabling a trait (directly or by
// disabling one of its containers) clears any unsatisfied prerequisite reason it had accumulated, rather than leaving
// the stale reason behind until the sheet is reloaded.
func TestEntityProcessPrereqsClearsUnsatisfiedReasonWhenDisabled(t *testing.T) {
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
// that processPrereqs marks it unsatisfied.
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
	list := NewPrereqList()
	p := NewTraitPrereq()
	p.Parent = list
	p.NameCriteria.Qualifier = name
	list.Prereqs = append(list.Prereqs, p)
	return list
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
	c.Equal(fxp.Twenty, trait.AdjustedPoints(), "without the setting, the trait's points count")
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
	c.Equal(fxp.Int(0), trait.AdjustedPoints(), "a disabled trait is worth no points")
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
	c.Equal(fxp.Twenty, trait.AdjustedPoints(), "its points count again")
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
	c.True(trait.prereqDisabled, "an enabled trait with an unmet prereq is disabled by the setting")

	trait.Disabled = true
	e.Recalculate()
	c.Equal("", trait.UnsatisfiedReason, "a trait the user disabled has no prerequisites to enforce")
	c.False(trait.prereqDisabled, "disabling the trait clears the flag the setting had set")
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
	c.False(nested.prereqDisabled, "a trait inside a disabled container is not flagged by the setting")
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
	c.Equal(fxp.Int(0), plain.AdjustedPoints(), "the child's points do not count")
	c.Equal(fxp.Int(0), container.AdjustedPoints(), "the container's points do not count")
	c.Equal(fxp.Int(0), e.PointsBreakdown().Total(), "nothing in the container counts toward the total")
	c.Equal("", nested.UnsatisfiedReason, "a child of the disabled container has no prerequisites to enforce")
	c.False(nested.prereqDisabled, "a child of the disabled container is not flagged by the setting")

	combatReflexes := NewTrait(e, nil, false)
	combatReflexes.Name = "Combat Reflexes"
	e.Traits = append(e.Traits, combatReflexes)
	e.Recalculate()
	c.False(container.DisabledByPrereqs(), "meeting the prereq re-enables the container")
	c.True(plain.Enabled(), "and its child")
	c.Equal(fxp.One, plain.AdjustedPoints(), "the child's points count again")
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

	trait.Levels = fxp.Two
	e.Recalculate()
	c.Equal("", trait.UnsatisfiedReason, "a trait within its maximum level is not flagged")
	c.False(trait.DisabledByPrereqs(), "a trait within its maximum level is enabled again")
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
