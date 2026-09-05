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
	"crypto/sha256"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestTraitModifierCloneDoesNotShareReplacements verifies that a copy of a modifier still carrying the legacy
// replacements map gets its own map. setTrait() installs a modifier's replacements directly into the trait it is
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
	dup.SetTargetNode(trait)
	c.Nil(dup.Replacements, "the copy hands its map off to the trait")
	c.Equal(map[string]string{"Element": "Fire"}, trait.Replacements, "the trait picks up the replacements")

	// A second modifier attaching to the same trait takes SetTargetNode's merge branch, writing into the map the trait now
	// holds. That must not reach back into the library row.
	second := NewTraitModifier(nil, nil, false)
	second.Replacements = map[string]string{"Aspect": "Heat"}
	second.SetTargetNode(trait)
	c.Equal(map[string]string{"Element": "Fire", "Aspect": "Heat"}, trait.Replacements,
		"the trait merges in the second modifier's replacements")
	c.Equal(map[string]string{"Element": "Fire"}, library.Replacements, "the library row is left untouched")
}

// newIgnoresLevelTestTrait returns an entity holding a 3-level, 5-point-per-level trait carrying a single -40%
// modifier, the shape the "Damage Resistance 3 (Limited, -40%)" case from issue #1017 takes. When the modifier does not
// take its level from the trait, it is given 2 levels of its own, deliberately different from the trait's 3, so that
// the assertions can tell which of the two the cost was multiplied by.
func newIgnoresLevelTestTrait(useLevelFromTrait bool) (*Entity, *Trait, *TraitModifier) {
	e := NewEntity()
	trait := NewTrait(e, nil, false)
	trait.Name = "Damage Resistance"
	trait.CanLevel = true
	trait.Levels = fxp.Three
	trait.PointsPerLevel = fxp.Five
	mod := NewTraitModifier(e, nil, false)
	mod.Name = "Limited"
	mod.CostAdj = "-40%"
	mod.UseLevelFromTrait = useLevelFromTrait
	if !useLevelFromTrait {
		mod.Levels = fxp.Two
	}
	trait.Modifiers = []*TraitModifier{mod}
	trait.SetDataOwner(e)
	e.Traits = append(e.Traits, trait)
	e.Recalculate()
	return e, trait, mod
}

// TestTraitModifierCostIgnoresLevel verifies that a "use level from owner" modifier with CostIgnoresLevel set applies
// its cost adjustment just once rather than multiplying it by the trait's level, while the modifier's own level -- what
// its per-level features and its display key off of -- is unaffected either way.
func TestTraitModifierCostIgnoresLevel(t *testing.T) {
	c := check.New(t)
	e, trait, mod := newIgnoresLevelTestTrait(true)

	// With the flag off, -40% is multiplied by the trait's 3 levels, and the resulting -120% is clamped to -80% by
	// AdjustedPoints: 15 - 12 = 3.
	c.Equal(fxp.Three, trait.AdjustedPoints(), "-40% x 3 levels, clamped to -80% of 15 points")
	c.Equal("-120%", mod.CostDescription(), "the cost adjustment is multiplied by the level")
	c.Equal(fxp.Three, mod.CurrentLevel(), "the modifier's level comes from the trait")

	// With the flag on, the -40% applies once: 15 - 6 = 9.
	mod.CostIgnoresLevel = true
	e.Recalculate()
	c.Equal(fxp.FromInteger(9), trait.AdjustedPoints(), "-40% applied once to 15 points")
	c.Equal("-40%", mod.CostDescription(), "the cost adjustment is applied once")
	c.Equal(fxp.Three, mod.CurrentLevel(), "the modifier's level is still the trait's, so per-level features scale")
}

// TestTraitModifierCostIgnoresLevelWithOwnLevels verifies that CostIgnoresLevel applies to a modifier carrying its own
// levels too, not just to one taking its level from the owning trait.
func TestTraitModifierCostIgnoresLevelWithOwnLevels(t *testing.T) {
	c := check.New(t)
	e, trait, mod := newIgnoresLevelTestTrait(false)

	mod.CostIgnoresLevel = true
	e.Recalculate()
	c.Equal("-40%", mod.CostDescription(), "the modifier's own 2 levels don't multiply the cost")
	c.Equal(fxp.FromInteger(9), trait.AdjustedPoints(), "-40% applied once to 15 points")
	c.Equal(fxp.Two, mod.CurrentLevel(), "the modifier keeps its own level")

	// With the flag off, the cost is multiplied by the modifier's own 2 levels, not the trait's 3, which would have
	// shown as -120%.
	mod.CostIgnoresLevel = false
	e.Recalculate()
	c.Equal("-80%", mod.CostDescription(), "with the flag off, the modifier's own levels multiply the cost")
	c.Equal(fxp.Three, trait.AdjustedPoints(), "-80% of 15 points")
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

// TestTraitModifierCostIgnoresLevelAffectsHash verifies that the flag participates in the source hash, so that a
// library row differing only in it is reported as mismatched rather than in sync.
func TestTraitModifierCostIgnoresLevelAffectsHash(t *testing.T) {
	c := check.New(t)
	mod := NewTraitModifier(nil, nil, false)
	mod.Name = "Limited"
	mod.CostAdj = "-40%"
	mod.UseLevelFromTrait = true

	h := sha256.New()
	mod.Hash(h)
	before := h.Sum(nil)

	mod.CostIgnoresLevel = true
	h.Reset()
	mod.Hash(h)
	c.NotEqual(before, h.Sum(nil), "flipping CostIgnoresLevel changes the hash")
}

// TestTraitModifierCostIgnoresLevelRoundTrip verifies that the flag is omitted from the JSON when false, so that
// existing files re-save unchanged, and survives a round-trip when true.
func TestTraitModifierCostIgnoresLevelRoundTrip(t *testing.T) {
	c := check.New(t)
	mod := NewTraitModifier(nil, nil, false)
	mod.Name = "Limited"
	mod.CostAdj = "-40%"
	mod.UseLevelFromTrait = true

	data, err := jio.Marshal(mod)
	c.NoError(err)
	c.NotContains(string(data), "cost_ignores_level", "the field is omitted when false")

	mod.CostIgnoresLevel = true
	data, err = jio.Marshal(mod)
	c.NoError(err)
	c.Contains(string(data), "cost_ignores_level", "the field is written when true")
	var restored TraitModifier
	c.NoError(jio.Unmarshal(data, &restored))
	c.True(restored.CostIgnoresLevel, "the field survives the round-trip")
}

// TestIssue1017LimitedDamageResistance covers the scenario from issue #1017: Damage Resistance 3 with a "Limited"
// modifier that cancels the trait's DR against everything and grants it back against a nameable damage type. The
// modifier's features scale with the trait's 3 levels while its -40% is charged just once.
func TestIssue1017LimitedDamageResistance(t *testing.T) {
	c := check.New(t)
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
	mod.UseLevelFromTrait = true
	mod.CostIgnoresLevel = true
	cancelAll := newTestDRBonus(-fxp.One, AllID, AllID)
	cancelAll.PerLevel = true
	againstType := newTestDRBonus(fxp.One, "@Damage Type@", AllID)
	againstType.PerLevel = true
	mod.Features = Features{cancelAll, againstType}

	trait.Modifiers = []*TraitModifier{mod}
	trait.SetDataOwner(e)
	e.Traits = append(e.Traits, trait)
	e.Recalculate()

	c.Equal(fxp.FromInteger(9), trait.AdjustedPoints(), "the -40% is charged once, not once per level")
	drMap := e.AddDRBonusesFor(TorsoID, nil, nil)
	c.Equal(0, drMap[AllID], "the modifier cancels the trait's DR against everything")
	c.Equal(3, drMap["cold"], "the modifier grants DR 3 against the chosen damage type")
}
