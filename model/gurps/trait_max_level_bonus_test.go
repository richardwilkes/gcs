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
	"encoding/json/v2"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/traitsel"
	"github.com/richardwilkes/toolbox/v2/check"
)

func newMaxLevelBonus(sel traitsel.Type, amount string) *TraitMaxLevelBonus {
	b := NewTraitMaxLevelBonus()
	b.SelectionType = sel
	b.Amount = amount
	return b
}

func newLeveledTrait(e *Entity, maxLevels string) *Trait {
	t := NewTrait(e, nil, false)
	t.CanLevel = true
	t.MaxLevels = maxLevels
	return t
}

// TestTraitMaxLevelResolution covers the base expression, the operation implied by the entered text, the Add -> % -> x
// stacking order applied by "to this trait" bonuses, and the zero-floor clamp.
func TestTraitMaxLevelResolution(t *testing.T) {
	c := check.New(t)

	// No expression and no bonuses: there is no maximum.
	trait := newLeveledTrait(nil, "")
	c.Equal(fxp.Int(0), trait.ResolvedMaxLevels(), "no maximum")

	// A plain-number expression resolves to that number.
	trait = newLeveledTrait(nil, "20")
	c.Equal(fxp.FromInteger(20), trait.ResolvedMaxLevels(), "plain number expression")

	// A script expression resolves to its numeric result.
	trait = newLeveledTrait(nil, "10 + 5")
	c.Equal(fxp.FromInteger(15), trait.ResolvedMaxLevels(), "script expression")

	// Add -> % -> x order: ((10 + 2) * 1.5) * 2 = 36.
	trait = newLeveledTrait(nil, "10")
	trait.Features = Features{
		newMaxLevelBonus(traitsel.ThisTrait, "+2"),
		newMaxLevelBonus(traitsel.ThisTrait, "50%"),
		newMaxLevelBonus(traitsel.ThisTrait, "x2"),
	}
	c.Equal(fxp.FromInteger(36), trait.ResolvedMaxLevels(), "add -> percent -> multiply")

	// Minimum clamp: 1 + (-100) = -99, clamped to 0.
	trait = newLeveledTrait(nil, "1")
	trait.Features = Features{newMaxLevelBonus(traitsel.ThisTrait, "-100")}
	c.Equal(fxp.Int(0), trait.ResolvedMaxLevels(), "clamped to minimum of 0")

	// A non-leveled trait has no maximum, regardless of expression.
	trait = newLeveledTrait(nil, "20")
	trait.CanLevel = false
	c.Equal(fxp.Int(0), trait.ResolvedMaxLevels(), "non-leveled trait has no maximum")
}

// TestTraitMaxLevelBonusFromModifier verifies that a "to this trait" bonus carried by a trait modifier is applied only
// while that modifier is enabled.
func TestTraitMaxLevelBonusFromModifier(t *testing.T) {
	c := check.New(t)
	trait := newLeveledTrait(nil, "10")
	mod := NewTraitModifier(nil, nil, false)
	mod.Features = Features{newMaxLevelBonus(traitsel.ThisTrait, "+3")}
	trait.Modifiers = []*TraitModifier{mod}

	mod.Disabled = false
	c.Equal(fxp.FromInteger(13), trait.ResolvedMaxLevels(), "enabled modifier bonus applies")

	mod.Disabled = true
	c.Equal(fxp.FromInteger(10), trait.ResolvedMaxLevels(), "disabled modifier bonus ignored")
}

// TestTraitMaxLevelBonusTraitWithName covers the entity-collected "to traits whose name" selector, including name + tag
// matching and per-level scaling from the attaching trait.
func TestTraitMaxLevelBonusTraitWithName(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	strength := newLeveledTrait(e, "20")
	strength.Name = "Strength"
	strength.Tags = []string{"Attribute"}
	e.Traits = append(e.Traits, strength)

	other := newLeveledTrait(e, "20")
	other.Name = "Fatigue"
	other.Tags = []string{"Attribute"}
	e.Traits = append(e.Traits, other)

	// A trait grants +5 maximum level to a trait named "Strength".
	bonus := newMaxLevelBonus(traitsel.TraitWithName, "+5")
	bonus.NameCriteria.Compare = criteria.IsText
	bonus.NameCriteria.Qualifier = "Strength"
	granter := NewTrait(e, nil, false)
	granter.Features = append(granter.Features, bonus)
	e.Traits = append(e.Traits, granter)
	e.Recalculate()

	c.Equal(fxp.FromInteger(25), strength.ResolvedMaxLevels(), "matching name receives the bonus")
	c.Equal(fxp.FromInteger(20), other.ResolvedMaxLevels(), "non-matching trait is unaffected")

	// Per-level scaling comes from the attaching trait's level.
	bonus.PerLevel = true
	granter.CanLevel = true
	granter.Levels = fxp.Three
	e.Recalculate()
	c.Equal(fxp.FromInteger(35), strength.ResolvedMaxLevels(), "per-level whose-name bonus scales by the granter's level: 20 + 5*3")
}

// TestTraitMaxLevelFlagging verifies that a trait whose level exceeds its resolved maximum is flagged with an
// unsatisfied reason, and that a within-range level is not flagged.
func TestTraitMaxLevelFlagging(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	trait := newLeveledTrait(e, "5")
	trait.Name = "Strength"
	trait.Levels = fxp.FromInteger(8)
	e.Traits = append(e.Traits, trait)
	e.Recalculate()
	c.NotEqual("", trait.UnsatisfiedReason, "level above the maximum is flagged")

	trait.Levels = fxp.FromInteger(3)
	e.Recalculate()
	c.Equal("", trait.UnsatisfiedReason, "level within the maximum is not flagged")

	// A trait with no maximum is never flagged, no matter how high the level.
	trait.MaxLevels = ""
	trait.Levels = fxp.FromInteger(100)
	e.Recalculate()
	c.Equal("", trait.UnsatisfiedReason, "no maximum means no flag")
}

// TestTraitMaxLevelBonusCannotCreateMaximum verifies that a trait with no declared maximum stays unlimited even when a
// bonus matches it. Bonuses used to be computed from a base of zero, so a broadly-scoped +2 turned "no maximum" into a
// maximum of 2.
func TestTraitMaxLevelBonusCannotCreateMaximum(t *testing.T) {
	c := check.New(t)

	// Each operation, applied to a trait that declares no maximum, must leave it unlimited.
	for _, amount := range []string{"+2", "50%", "x2", "-2"} {
		trait := newLeveledTrait(nil, "")
		trait.Features = Features{newMaxLevelBonus(traitsel.ThisTrait, amount)}
		c.Equal(fxp.Int(0), trait.ResolvedMaxLevels(), "a %q bonus must not create a maximum", amount)
	}

	// An expression that resolves to zero means "no maximum" just as an empty one does.
	trait := newLeveledTrait(nil, "0")
	trait.Features = Features{newMaxLevelBonus(traitsel.ThisTrait, "+2")}
	c.Equal(fxp.Int(0), trait.ResolvedMaxLevels(), "a maximum expression of 0 stays unlimited")

	// A modifier-carried bonus is subject to the same rule.
	trait = newLeveledTrait(nil, "")
	mod := NewTraitModifier(nil, nil, false)
	mod.Features = Features{newMaxLevelBonus(traitsel.ThisTrait, "+3")}
	trait.Modifiers = []*TraitModifier{mod}
	c.Equal(fxp.Int(0), trait.ResolvedMaxLevels(), "a modifier bonus must not create a maximum")
}

// TestTraitMaxLevelBonusBroadScopeLeavesUncappedTraitsAlone verifies end-to-end that a broadly-scoped "traits whose
// name" bonus raises the cap of traits that declare one without imposing a cap -- and an unsatisfied-reason flag -- on
// the ones that don't.
func TestTraitMaxLevelBonusBroadScopeLeavesUncappedTraitsAlone(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	capped := newLeveledTrait(e, "5")
	capped.Name = "Capped"
	capped.Levels = fxp.FromInteger(6)
	e.Traits = append(e.Traits, capped)

	uncapped := newLeveledTrait(e, "")
	uncapped.Name = "Uncapped"
	uncapped.Levels = fxp.FromInteger(6)
	e.Traits = append(e.Traits, uncapped)

	// A bonus that matches anything grants +2 maximum levels.
	bonus := newMaxLevelBonus(traitsel.TraitWithName, "+2")
	bonus.NameCriteria.Compare = criteria.AnyText
	granter := NewTrait(e, nil, false)
	granter.Features = append(granter.Features, bonus)
	e.Traits = append(e.Traits, granter)
	e.Recalculate()

	c.Equal(fxp.FromInteger(7), capped.ResolvedMaxLevels(), "a declared maximum of 5 is raised to 7")
	c.Equal("", capped.UnsatisfiedReason, "level 6 is within the raised maximum of 7")
	c.Equal(fxp.Int(0), uncapped.ResolvedMaxLevels(), "a trait with no maximum stays unlimited")
	c.Equal("", uncapped.UnsatisfiedReason, "an unlimited trait must not be flagged for exceeding a maximum")
}

// TestTraitMaxLevelBonusRoundTrip verifies that each selector/operation combination survives a JSON round-trip.
func TestTraitMaxLevelBonusRoundTrip(t *testing.T) {
	c := check.New(t)
	this := newMaxLevelBonus(traitsel.ThisTrait, "x2")
	byName := newMaxLevelBonus(traitsel.TraitWithName, "+50%")
	byName.PerLevel = true
	byName.NameCriteria.Compare = criteria.IsText
	byName.NameCriteria.Qualifier = "Strength"
	byName.TagsCriteria.Compare = criteria.IsText
	byName.TagsCriteria.Qualifier = "Attribute"
	original := Features{this, byName}

	data, err := json.Marshal(original)
	c.NoError(err)
	var restored Features
	c.NoError(json.Unmarshal(data, &restored))
	c.Equal(len(original), len(restored), "feature count preserved")
	again, err := json.Marshal(restored)
	c.NoError(err)
	c.Equal(string(data), string(again), "re-marshaled JSON is stable across a round-trip")
}
