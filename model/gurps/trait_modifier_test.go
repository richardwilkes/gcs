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
)

// TestTraitModifierCloneDoesNotShareReplacements verifies that a copy of a modifier still carrying the legacy
// replacements map gets its own map. SetTarget() installs a modifier's replacements directly into the trait it is
// attached to when that trait has none of its own, so sharing the map let a later merge into the trait's map write into
// the library row the copy came from.
func TestTraitModifierCloneDoesNotShareReplacements(t *testing.T) {
	c := check.New(t)

	// A library row with legacy replacements, as it arrives from an .adm file written by an older version.
	library := NewTraitModifier(nil, nil, false)
	library.Name = "@Element@ Attunement"
	library.Replacements = map[string]string{"Element": "Fire"}

	// Dropping it onto a sheet's trait row clones it...
	dup := library.Clone(LibraryFile{}, nil, nil, Reference)
	c.Equal(map[string]string{"Element": "Fire"}, dup.Replacements, "the copy carries the replacements")

	// ...and the next attachment pass migrates the copy's map into the trait, which has none of its own.
	trait := NewTrait(nil, nil, false)
	dup.SetTarget(trait)
	c.Nil(dup.Replacements, "the copy hands its map off to the trait")
	c.Equal(map[string]string{"Element": "Fire"}, trait.Replacements, "the trait picks up the replacements")

	// A second modifier attaching to the same trait takes SetTarget's merge branch, writing into the map the trait now
	// holds. That must not reach back into the library row.
	second := NewTraitModifier(nil, nil, false)
	second.Replacements = map[string]string{"Aspect": "Heat"}
	second.SetTarget(trait)
	c.Equal(map[string]string{"Element": "Fire", "Aspect": "Heat"}, trait.Replacements,
		"the trait merges in the second modifier's replacements")
	c.Equal(map[string]string{"Element": "Fire"}, library.Replacements, "the library row is left untouched")
}

// newIgnoresLevelTestTrait returns an entity holding the "Damage Resistance 3 (Limited, -40%)" case from issue #1017: a
// 3-level, 5-point-per-level trait granting DR 1 per level, carrying a single -40% modifier whose per-level features
// cancel that DR against everything and grant it back against the trait's "Cold" damage type. When the modifier does
// not take its level from the trait, it is given 2 levels of its own, deliberately different from the trait's 3, so
// that the assertions can tell which of the two the cost and features were multiplied by.
func newIgnoresLevelTestTrait(useLevelFromTrait bool) (*Entity, *Trait, *TraitModifier) {
	e := NewEntity()
	trait := NewTrait(e, nil, false)
	trait.Name = "Damage Resistance"
	trait.CanLevel = true
	trait.Levels = fxp.Three
	trait.PointsPerLevel = fxp.Five
	trait.Replacements = map[string]string{"Damage Type": "Cold"}
	baseDR := newTestDRBonus(fxp.One, AllID, AllID)
	baseDR.PerLevel = true
	trait.Features = Features{baseDR}
	mod := NewTraitModifier(e, nil, false)
	mod.Name = "Limited"
	mod.CostAdj = "-40%"
	mod.UseLevelFromTrait = useLevelFromTrait
	if !useLevelFromTrait {
		mod.Levels = fxp.Two
	}
	cancelAll := newTestDRBonus(-fxp.One, AllID, AllID)
	cancelAll.PerLevel = true
	againstType := newTestDRBonus(fxp.One, "@Damage Type@", AllID)
	againstType.PerLevel = true
	mod.Features = Features{cancelAll, againstType}
	trait.Modifiers = []*TraitModifier{mod}
	trait.SetDataOwner(e)
	e.Traits = append(e.Traits, trait)
	e.Recalculate()
	return e, trait, mod
}

// TestTraitModifierCostIgnoresLevel verifies that a leveled modifier with CostIgnoresLevel set applies its cost
// adjustment just once rather than multiplying it by its level, whether it takes that level from the trait or carries
// its own, while the modifier's level -- what its per-level features and its display key off of -- is unaffected
// either way.
func TestTraitModifierCostIgnoresLevel(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		name              string
		useLevelFromTrait bool
		level             fxp.Int
		multipliedCost    string
		drAll             int
		drCold            int
	}{
		// -40% x the trait's 3 levels is -120%, which AdjustedPoints clamps to -80%.
		{"level from the trait", true, fxp.Three, "-120%", 0, 3},
		// The modifier's own 2 levels, not the trait's 3, multiply the cost.
		{"own levels", false, fxp.Two, "-80%", 1, 2},
	} {
		e, trait, mod := newIgnoresLevelTestTrait(one.useLevelFromTrait)

		// With the flag off, the -40% is multiplied by the level, coming to -80% of 15 points: 15 - 12 = 3.
		c.Equal(fxp.Three, trait.AdjustedPoints(nil), "%s: -80%% of 15 points", one.name)
		c.Equal(one.multipliedCost, mod.CostDescription(), "%s: the cost adjustment is multiplied by the level",
			one.name)
		c.Equal(one.level, mod.CurrentLevel(), "%s: the modifier's level", one.name)

		// With the flag on, the -40% applies once: 15 - 6 = 9.
		mod.CostIgnoresLevel = true
		e.Recalculate()
		c.Equal(fxp.FromInteger(9), trait.AdjustedPoints(nil), "%s: -40%% applied once to 15 points", one.name)
		c.Equal("-40%", mod.CostDescription(), "%s: the cost adjustment is applied once", one.name)
		c.Equal(one.level, mod.CurrentLevel(), "%s: the modifier's level is unaffected", one.name)
		drMap := e.AddDRBonusesFor(TorsoID, nil, nil)
		c.Equal(one.drAll, drMap[AllID], "%s: the modifier's per-level features cancel the trait's DR against "+
			"everything by its level", one.name)
		c.Equal(one.drCold, drMap["cold"], "%s: the modifier's per-level features grant DR against the chosen damage "+
			"type by its level", one.name)
	}
}

// TestTraitModifierShowsTheLevelItWouldHaveIfEnabled verifies that a leveled modifier that is disabled, or whose trait
// is, shows the level it would have were both enabled, as a disabled trait shows its own, while a disabled modifier's
// current level, which its per-level features key off of, stays zero.
func TestTraitModifierShowsTheLevelItWouldHaveIfEnabled(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		name              string
		useLevelFromTrait bool
		disableTrait      bool
		shown             string
	}{
		{"disabled, with its own levels", false, false, "Limited 2"},
		{"disabled, with the trait's level", true, false, "Limited 3"},
		{"on a disabled trait, with the trait's level", true, true, "Limited 3"},
	} {
		e, trait, mod := newIgnoresLevelTestTrait(one.useLevelFromTrait)
		if one.disableTrait {
			trait.SetDisabled(true)
		} else {
			mod.SetDisabled(true)
		}
		e.Recalculate()
		c.Equal(one.shown, mod.String(), one.name)
		if !one.disableTrait {
			c.Equal(fxp.Int(0), mod.CurrentLevel(), "%s: the modifier's current level", one.name)
		}
	}
}

// TestTraitModifierCostDescriptionSimplifiesFraction verifies that a leveled fractional cost adjustment is reduced
// before display, so that the lists, prompts and notes show the same "x2" the editor's Total field does rather than
// "x6/3".
func TestTraitModifierCostDescriptionSimplifiesFraction(t *testing.T) {
	c := check.New(t)
	mod := NewTraitModifier(nil, nil, false)
	mod.CostAdj = "x2/3"
	mod.Levels = fxp.Three
	c.Equal("x2", mod.CostDescription(), "x2/3 at 3 levels reduces to x2")
	mod.CostAdj = "+5/2"
	mod.Levels = fxp.Two
	c.Equal("+5", mod.CostDescription(), "+5/2 at 2 levels reduces to +5")
	mod.CostAdj = "x2/3"
	c.Equal("x4/3", mod.CostDescription(), "x2/3 at 2 levels is left as a fraction")
}

// TestTraitModifierContainerHasNoCost verifies that a container never reports a cost modifier or a level, even if
// stale non-container data is present on it.
func TestTraitModifierContainerHasNoCost(t *testing.T) {
	c := check.New(t)
	container := NewTraitModifier(nil, nil, true)
	container.CostAdj = "-40%"
	container.Levels = fxp.Three
	c.False(container.IsLeveled(), "a container is never leveled")
	c.Equal(fxp.Fraction{Denominator: fxp.One}, container.CostModifierForTrait(nil), "a container has no cost modifier")
	c.Equal("", container.CostDescription(), "a container has no cost description")
}

// TestTraitModifierApplyNameableKeysKeepsAnswersForDisabledModifiers verifies that applying one modifier's
// replacements keeps the trait's others: those held for its disabled modifiers, its own, and those applied for its
// other modifiers when, as the prompt does, every modifier's keys are collected first and then applied in turn, where a
// key one modifier leaves unanswered keeps the answer given for another. A cleared key is removed, and clearing every
// replacement leaves nil.
func TestTraitModifierApplyNameableKeysKeepsAnswersForDisabledModifiers(t *testing.T) {
	c := check.New(t)
	trait := NewTrait(nil, nil, false)
	trait.Name = "Sword"
	finish := NewTraitModifier(nil, nil, false)
	finish.Name = "@Finish@ Coating"
	trait.AddModifiers(finish)
	trait.Replacements = map[string]string{"Finish": "Gilded"}
	finish.SetDisabled(true)
	color := NewTraitModifier(nil, nil, false)
	color.Name = "@Color@ Paint"
	trait.AddModifiers(color)

	colorKeys := make(map[string]string)
	color.FillWithNameableKeys(colorKeys, nil)
	colorKeys["Color"] = "Red"
	color.ApplyNameableKeys(colorKeys)

	c.Equal(map[string]string{"Finish": "Gilded", "Color": "Red"}, trait.Replacements,
		"the answer held for the switched-off modifier must survive")
	finish.SetDisabled(false)
	c.Equal("Gilded Coating", finish.NameWithReplacements(), "switching the modifier back on must find its answer")

	// Clearing every replacement leaves nil rather than an empty map.
	delete(colorKeys, "Color")
	color.ApplyNameableKeys(colorKeys)
	c.Equal(map[string]string{"Finish": "Gilded"}, trait.Replacements, "the cleared key must be removed")
	finishKeys := make(map[string]string)
	finish.FillWithNameableKeys(finishKeys, nil)
	c.Equal("Gilded", finishKeys["Finish"], "the prompt starts out on the trait's value")
	delete(finishKeys, "Finish")
	finish.ApplyNameableKeys(finishKeys)
	c.Nil(trait.Replacements, "nothing left to hold must be held as nothing")

	// As the prompt does: collect every modifier's keys first, then apply them in turn.
	trait = NewTrait(nil, nil, false)
	trait.Name = "@Foo@ Thing"
	trait.Replacements = map[string]string{"Foo": "Bar"}
	material := NewTraitModifier(nil, nil, false)
	material.Name = "@Material@ Coating"
	color = NewTraitModifier(nil, nil, false)
	color.Name = "@Color@ Paint"
	trait.AddModifiers(material, color)
	materialKeys := make(map[string]string)
	material.FillWithNameableKeys(materialKeys, nil)
	colorKeys = make(map[string]string)
	color.FillWithNameableKeys(colorKeys, nil)
	materialKeys["Material"] = "Steel"
	colorKeys["Color"] = "Red"
	material.ApplyNameableKeys(materialKeys)
	color.ApplyNameableKeys(colorKeys)
	c.Equal(map[string]string{"Foo": "Bar", "Material": "Steel", "Color": "Red"}, trait.Replacements,
		"the trait's own replacements and those applied for its other modifiers must survive")
	c.Equal("Bar Thing", trait.NameWithReplacements())
	c.Equal("Steel Coating", material.NameWithReplacements())
	c.Equal("Red Paint", color.NameWithReplacements())

	// The same, with a key shared by two modifiers left unanswered for the second.
	trait = NewTrait(nil, nil, false)
	trait.Name = "@Material@ Sword"
	trait.Replacements = map[string]string{"Material": "Iron"}
	coating := NewTraitModifier(nil, nil, false)
	coating.Name = "@Finish@ Coating"
	trim := NewTraitModifier(nil, nil, false)
	trim.Name = "@Finish@ Trim"
	trait.AddModifiers(coating, trim)
	coatingKeys := make(map[string]string)
	coating.FillWithNameableKeys(coatingKeys, nil)
	trimKeys := make(map[string]string)
	trim.FillWithNameableKeys(trimKeys, nil)
	c.Equal(nameable.Unset, trimKeys["Finish"], "the key starts out unanswered")
	coatingKeys["Finish"] = "Gilded"
	coating.ApplyNameableKeys(coatingKeys)
	trim.ApplyNameableKeys(trimKeys)
	c.Equal(map[string]string{"Material": "Iron", "Finish": "Gilded"}, trait.Replacements,
		"an unanswered key keeps the answer given for the other modifier")
}
