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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/traitsel"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestHasTag verifies that a tag matches either as a whole -- even when it contains colons -- or as one of the
// colon-separated subsets within a stored tag, ignoring case and surrounding whitespace.
func TestHasTag(t *testing.T) {
	c := check.New(t)
	tags := []string{"Advantage: Mental", "Physical"}

	// The whole tag matches, including the colon-containing one.
	c.True(HasTag("Advantage: Mental", tags), "an exact match on a colon-containing tag is found")
	c.True(HasTag("  advantage: MENTAL  ", tags), "whole-tag matching ignores case and surrounding whitespace")
	c.True(HasTag("Physical", tags), "an exact match on a tag without a colon is found")

	// The colon-separated subsets still match.
	c.True(HasTag("Advantage", tags), "the portion before the colon is found")
	c.True(HasTag("Mental", tags), "the portion after the colon is found")

	// Non-matches remain non-matches.
	c.False(HasTag("Advantage: Physical", tags), "a colon-containing tag that isn't present isn't found")
	c.False(HasTag("Social", tags), "a tag that isn't present isn't found")
	c.False(HasTag("Advantage: Mental", nil), "nothing matches an empty set of tags")
}

// TestTraitLevelSelfReferentialBonus verifies that a per-level TraitBonus which matches the very trait carrying it
// resolves to a finite level instead of recursing until the stack overflows. The bonus scales by the trait's own level,
// which falls back to the unadjusted level while that level is being resolved.
func TestTraitLevelSelfReferentialBonus(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	trait := NewTrait(e, nil, false)
	trait.Name = "Growth"
	trait.CanLevel = true
	trait.Levels = fxp.Two
	bonus := NewTraitBonus()
	bonus.NameCriteria.Qualifier = "Growth"
	bonus.PerLevel = true
	trait.Features = append(trait.Features, bonus)
	e.Traits = append(e.Traits, trait)
	e.Recalculate()

	// 2 base levels + 1 per level, where the per-level scaling falls back to the unadjusted 2 levels.
	c.Equal(fxp.FromInteger(4), trait.CurrentLevel(), "a self-matching per-level trait bonus resolves without recursing")
}

// TestTraitLevelMutuallyReferentialBonuses verifies that two traits whose per-level bonuses adjust each other's level
// resolve to finite levels rather than recursing until the stack overflows.
func TestTraitLevelMutuallyReferentialBonuses(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	first := NewTrait(e, nil, false)
	first.Name = "First"
	first.CanLevel = true
	first.Levels = fxp.One
	firstBonus := NewTraitBonus()
	firstBonus.NameCriteria.Qualifier = "Second"
	firstBonus.PerLevel = true
	first.Features = append(first.Features, firstBonus)

	second := NewTrait(e, nil, false)
	second.Name = "Second"
	second.CanLevel = true
	second.Levels = fxp.One
	secondBonus := NewTraitBonus()
	secondBonus.NameCriteria.Qualifier = "First"
	secondBonus.PerLevel = true
	second.Features = append(second.Features, secondBonus)

	e.Traits = append(e.Traits, first, second)
	e.Recalculate()

	// Resolving either one drops the other's per-level scaling back to its unadjusted level: 1 + 1*(1 + 1*1) = 3.
	c.Equal(fxp.Three, first.CurrentLevel(), "mutually-referential per-level bonuses resolve without recursing")
	c.Equal(fxp.Three, second.CurrentLevel(), "mutually-referential per-level bonuses resolve without recursing")
}

// TestTraitMaxLevelPerLevelBonusOwners verifies that a per-level "this trait" max-level bonus scales by the level of
// the node it is attached to: the trait for a bonus on the trait's own features, and the modifier for a bonus carried
// by a trait modifier.
func TestTraitMaxLevelPerLevelBonusOwners(t *testing.T) {
	c := check.New(t)

	// A per-level bonus on the trait's own features scales by the trait's level: 10 + 1*5.
	trait := newLeveledTrait(nil, "10")
	trait.Levels = fxp.FromInteger(5)
	ownBonus := newMaxLevelBonus(traitsel.ThisTrait, "+1")
	ownBonus.PerLevel = true
	trait.Features = Features{ownBonus}
	c.Equal(fxp.FromInteger(15), trait.ResolvedMaxLevels(), "a per-level bonus on the trait scales by the trait's level")

	// A per-level bonus on a leveled modifier scales by the modifier's level: 10 + 1*3.
	trait = newLeveledTrait(nil, "10")
	trait.Levels = fxp.FromInteger(5)
	mod := NewTraitModifier(nil, nil, false)
	mod.Levels = fxp.Three
	modBonus := newMaxLevelBonus(traitsel.ThisTrait, "+1")
	modBonus.PerLevel = true
	mod.Features = Features{modBonus}
	trait.Modifiers = []*TraitModifier{mod}
	trait.SetDataOwner(nil)
	c.Equal(fxp.FromInteger(13), trait.ResolvedMaxLevels(),
		"a per-level bonus on a modifier scales by the modifier's level")

	// A modifier that takes its level from the trait scales by the trait's level: 10 + 1*5.
	mod.UseLevelFromTrait = true
	c.Equal(fxp.FromInteger(15), trait.ResolvedMaxLevels(),
		"a per-level bonus on a use-level-from-owner modifier scales by the trait's level")
}

// TestTraitModifierCostUsesPurchasedLevels verifies that a "use level from owner" modifier is costed against the
// trait's purchased levels rather than its bonus-adjusted current level, so that free levels granted by a TraitBonus
// don't attract enhancement percentages. The modifier's own level -- which drives its per-level features and its
// display -- still tracks the bonus-adjusted level.
func TestTraitModifierCostUsesPurchasedLevels(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	trait := NewTrait(e, nil, false)
	trait.Name = "Innate Attack"
	trait.CanLevel = true
	trait.Levels = fxp.Three
	trait.PointsPerLevel = fxp.Ten
	mod := NewTraitModifier(e, nil, false)
	mod.CostAdj = "+10%"
	mod.UseLevelFromTrait = true
	trait.Modifiers = []*TraitModifier{mod}
	trait.SetDataOwner(e)
	e.Traits = append(e.Traits, trait)
	e.Recalculate()

	// 3 purchased levels at 10 points each, enhanced by 3 * 10%: 30 + 30*30/100.
	c.Equal(fxp.FromInteger(39), trait.AdjustedPoints(), "10/level x 3 levels, +30%")
	c.Equal(fxp.Three, mod.CurrentLevel(), "the modifier's level matches the trait's")

	// A TraitBonus grants 2 free levels. Those raise the trait's current level, but not what was paid for, so neither
	// the leveled base cost nor the enhancement may move.
	granter := NewTrait(e, nil, false)
	granter.Name = "Granter"
	bonus := NewTraitBonus()
	bonus.NameCriteria.Qualifier = "Innate Attack"
	bonus.Amount = fxp.Two
	granter.Features = append(granter.Features, bonus)
	e.Traits = append(e.Traits, granter)
	e.Recalculate()

	c.Equal(fxp.Five, trait.CurrentLevel(), "the bonus raises the trait's current level to 5")
	c.Equal(fxp.FromInteger(39), trait.AdjustedPoints(),
		"bonus-granted levels are free, so the enhancement stays at +30%")
	c.Equal(fxp.Five, mod.CurrentLevel(),
		"the modifier's level still tracks the trait's current level, for its per-level features and display")
}

// TestTraitDescriptionSelfControlSuffix verifies that the description cell shows the self-control roll suffix for every
// roll other than "None required", including the "Never resist" case.
func TestTraitDescriptionSelfControlSuffix(t *testing.T) {
	c := check.New(t)
	trait := NewTrait(nil, nil, false)
	trait.Name = "Berserk"
	var data CellData

	trait.SelfControl = selfctrl.None
	trait.CellData(TraitDescriptionColumn, &data)
	c.Equal("Berserk", data.Primary, "no suffix when no self-control roll is required")

	trait.SelfControl = selfctrl.CR12
	trait.CellData(TraitDescriptionColumn, &data)
	c.Equal("Berserk (CR12)", data.Primary, "a numeric self-control roll is shown")

	trait.SelfControl = selfctrl.Always
	trait.CellData(TraitDescriptionColumn, &data)
	c.Equal("Berserk (No CR)", data.Primary, `"Never resist" is shown as "No CR"`)
}

// TestAlternativeAbilitiesCost verifies that the most expensive child of an alternative abilities container is billed
// at full cost and the rest at 20%, even when every child costs negative points.
func TestAlternativeAbilitiesCost(t *testing.T) {
	c := check.New(t)

	newAltContainer := func(childPoints ...int) *Trait {
		parent := NewTrait(nil, nil, true)
		parent.ContainerType = container.AlternativeAbilities
		for _, points := range childPoints {
			child := NewTrait(nil, parent, false)
			child.BasePoints = fxp.FromInteger(points)
			parent.Children = append(parent.Children, child)
		}
		return parent
	}

	// -5 is the most expensive, so it is billed in full and the others at 20%: -5 + -4 + -2.
	c.Equal(fxp.FromInteger(-11), newAltContainer(-20, -10, -5).AdjustedPoints(),
		"all-negative children still bill the most expensive one at full cost")

	// 20 is the most expensive, so it is billed in full and the other at 20%: 20 + 2.
	c.Equal(fxp.FromInteger(22), newAltContainer(20, 10).AdjustedPoints(),
		"positive children bill the most expensive one at full cost")

	// A mix behaves the same way: 10 + -4 + 0.
	c.Equal(fxp.FromInteger(6), newAltContainer(10, -20, 0).AdjustedPoints(),
		"mixed children bill the most expensive one at full cost")
}

// TestInheritedModifiersAreNotRepointed verifies that computing a child's points costs an inherited modifier against
// that child without re-pointing the modifier at it, leaving the container's own display of the modifier intact.
func TestInheritedModifiersAreNotRepointed(t *testing.T) {
	c := check.New(t)

	parent := NewTrait(nil, nil, true)
	parent.Name = "Container"
	parent.Replacements = map[string]string{"type": "Fire"}
	mod := NewTraitModifier(nil, nil, false)
	mod.Name = "@type@ Attack"
	mod.CostAdj = "+2"
	mod.UseLevelFromTrait = true
	parent.Modifiers = []*TraitModifier{mod}

	newChild := func(name string, levels int) *Trait {
		child := NewTrait(nil, parent, false)
		child.Name = name
		child.Replacements = map[string]string{"type": "Ice"}
		child.CanLevel = true
		child.Levels = fxp.FromInteger(levels)
		child.BasePoints = fxp.FromInteger(10)
		parent.Children = append(parent.Children, child)
		return child
	}
	first := newChild("First", 3)
	second := newChild("Second", 5)
	parent.SetDataOwner(nil)

	// The inherited +2 per level modifier is costed against the level of the child it is being applied to.
	c.Equal(fxp.FromInteger(16), first.AdjustedPoints(), "10 + 2*3")
	c.Equal(fxp.FromInteger(20), second.AdjustedPoints(), "10 + 2*5")
	c.Equal(fxp.FromInteger(16), first.AdjustedPoints(), "the first child's cost is unaffected by the second's")

	// The container's modifier still belongs to the container, so the container's display of it is unchanged.
	c.Equal(parent, mod.OwningTrait(), "the inherited modifier still belongs to the container")
	c.Equal("Fire Attack", mod.NameWithReplacements(),
		"the inherited modifier still resolves against the container's replacements")
}

// TestTraitEditDataCopyLinksModifiers verifies that the modifier copies held by an editor know the trait they belong
// to, so that a "use level from owner" modifier resolves its level there.
func TestTraitEditDataCopyLinksModifiers(t *testing.T) {
	c := check.New(t)

	trait := NewTrait(nil, nil, false)
	trait.Name = "Trait"
	trait.CanLevel = true
	trait.Levels = fxp.Three
	trait.BasePoints = fxp.FromInteger(10)
	mod := NewTraitModifier(nil, nil, false)
	mod.CostAdj = "+2"
	mod.UseLevelFromTrait = true
	trait.Modifiers = []*TraitModifier{mod}
	trait.SetDataOwner(nil)

	var data TraitEditData
	data.CopyFrom(trait)
	c.Equal(1, len(data.Modifiers), "the modifier was copied")
	c.Equal(trait, data.Modifiers[0].OwningTrait(), "the copied modifier knows its trait")
	c.Equal(fxp.Three, data.Modifiers[0].CurrentLevel(), "the copied modifier resolves its level from the trait")
	c.Equal(fxp.FromInteger(16), AdjustedPoints(nil, trait, data.CanLevel, data.BasePoints, data.Levels,
		data.PointsPerLevel, data.SelfControl, data.Frequency, data.Modifiers, data.RoundCostDown), "10 + 2*3")
}

// TestTraitCloneModifiersBelongToTheClone verifies that duplicating a trait gives the copies of its modifiers to the
// duplicate. Clone() routed through CopyFrom(), which exists for the editor and therefore points the modifier copies at
// the trait handed to it -- the trait being cloned. The duplicate's modifiers then resolved their names against the
// original's replacements and their levels from the original's level, so editing the original dragged the duplicate's
// modifiers along with it. Sheets and templates hide this by reattaching owners on rebuild, but a traits library file
// never does.
func TestTraitCloneModifiersBelongToTheClone(t *testing.T) {
	c := check.New(t)

	source := NewTrait(nil, nil, false)
	source.Name = "@element@ Resistance"
	source.Replacements = map[string]string{"element": "Fire"}
	source.CanLevel = true
	source.Levels = fxp.Three
	source.BasePoints = fxp.FromInteger(10)
	mod := NewTraitModifier(nil, nil, false)
	mod.Name = "@element@ Only"
	mod.CostAdj = "+2"
	mod.UseLevelFromTrait = true
	source.Modifiers = []*TraitModifier{mod}
	source.SetDataOwner(nil)

	clone := source.Clone(LibraryFile{}, nil, nil, false)
	c.Equal(1, len(clone.Modifiers), "the modifier was copied")
	// Compared as pointers: the two traits are equal by value at this point, so only identity distinguishes them.
	c.True(clone.Modifiers[0].OwningTrait() == clone, "the copy belongs to the clone, not the trait cloned from")
	c.True(mod.OwningTrait() == source, "the original's modifier still belongs to the original")

	// Diverging the clone must not be read through the original, and vice versa.
	clone.Replacements["element"] = "Ice"
	clone.Levels = fxp.Five
	c.Equal("Ice Only", clone.Modifiers[0].NameWithReplacements(),
		"the copy resolves names against the clone's replacements")
	c.Equal(fxp.Five, clone.Modifiers[0].CurrentLevel(), "the copy takes its level from the clone")
	c.Equal(fxp.FromInteger(20), clone.AdjustedPoints(), "10 + 2*5")
	c.Equal("Fire Only", mod.NameWithReplacements(), "the original's modifier is unaffected")
	c.Equal(fxp.Three, mod.CurrentLevel(), "the original's modifier still reports the original's level")
	c.Equal(fxp.FromInteger(16), source.AdjustedPoints(), "10 + 2*3")

	// Children are cloned the same way, so their modifiers must follow the same rule.
	parent := NewTrait(nil, nil, true)
	parent.Name = "Container"
	source.parent = parent
	parent.Children = []*Trait{source}
	parent.SetDataOwner(nil)
	clonedContainer := parent.Clone(LibraryFile{}, nil, nil, false)
	c.Equal(1, len(clonedContainer.Children), "the child was cloned")
	clonedChild := clonedContainer.Children[0]
	c.True(clonedChild.Modifiers[0].OwningTrait() == clonedChild, "a cloned child's modifier belongs to that child")
}
