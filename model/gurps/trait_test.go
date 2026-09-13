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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/traitsel"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
)

// TestHasTag verifies that a tag matches either as a whole -- even when it contains colons -- or as one of the
// colon-separated subsets within a stored tag, ignoring case and surrounding whitespace.
func TestHasTag(t *testing.T) {
	c := check.New(t)
	tags := []string{"Advantage: Mental", "Physical"}

	c.True(HasTag("Advantage: Mental", tags), "an exact match on a colon-containing tag is found")
	c.True(HasTag("  advantage: MENTAL  ", tags), "whole-tag matching ignores case and surrounding whitespace")
	c.True(HasTag("Physical", tags), "an exact match on a tag without a colon is found")

	c.True(HasTag("Advantage", tags), "the portion before the colon is found")
	c.True(HasTag("Mental", tags), "the portion after the colon is found")

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

	bonus := NewTraitBonus()
	bonus.NameCriteria.Qualifier = "Growth"
	bonus.PerLevel = true
	trait := addTraitWithFeatures(e, "Growth", bonus)
	trait.CanLevel = true
	trait.Levels = fxp.Two
	e.Recalculate()

	// 2 base levels + 1 per level, where the per-level scaling falls back to the unadjusted 2 levels.
	c.Equal(fxp.FromInteger(4), trait.CurrentLevel(), "a self-matching per-level trait bonus resolves without recursing")
}

// TestTraitLevelMutuallyReferentialBonuses verifies that two traits whose per-level bonuses adjust each other's level
// resolve to finite levels rather than recursing until the stack overflows.
func TestTraitLevelMutuallyReferentialBonuses(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	firstBonus := NewTraitBonus()
	firstBonus.NameCriteria.Qualifier = "Second"
	firstBonus.PerLevel = true
	first := addTraitWithFeatures(e, "First", firstBonus)
	first.CanLevel = true
	first.Levels = fxp.One

	secondBonus := NewTraitBonus()
	secondBonus.NameCriteria.Qualifier = "First"
	secondBonus.PerLevel = true
	second := addTraitWithFeatures(e, "Second", secondBonus)
	second.CanLevel = true
	second.Levels = fxp.One

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
	bonus := NewTraitBonus()
	bonus.NameCriteria.Qualifier = "Innate Attack"
	bonus.Amount = fxp.Two
	addTraitWithFeatures(e, "Granter", bonus)
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

// TestAlternativeAbilitiesMultipleSlots verifies that setting AlternativeSlots > 1 bills that many of the most
// expensive children at full cost and the rest at 20%.
func TestAlternativeAbilitiesMultipleSlots(t *testing.T) {
	c := check.New(t)

	newAltContainer := func(slots int, childPoints ...int) *Trait {
		parent := NewTrait(nil, nil, true)
		parent.ContainerType = container.AlternativeAbilities
		parent.AlternativeSlots = slots
		for _, points := range childPoints {
			child := NewTrait(nil, parent, false)
			child.BasePoints = fxp.FromInteger(points)
			parent.Children = append(parent.Children, child)
		}
		return parent
	}

	// A stored value of 0 means "unset" and behaves like a single slot: 20 + 2.
	c.Equal(fxp.FromInteger(22), newAltContainer(0, 20, 10).AdjustedPoints(),
		"an unset slot count resolves to a single slot")

	// With 2 slots, the two most expensive children (20 and 10) are billed in full and the rest at 20%: 20 + 10 + 1.
	c.Equal(fxp.FromInteger(31), newAltContainer(2, 20, 10, 5).AdjustedPoints(),
		"two slots bill the two most expensive children at full cost")

	// A slot count larger than the number of children bills every child in full: 20 + 10 + 5.
	c.Equal(fxp.FromInteger(35), newAltContainer(5, 20, 10, 5).AdjustedPoints(),
		"a slot count exceeding the child count bills every child at full cost")
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

	c.Equal(fxp.FromInteger(16), first.AdjustedPoints(), "10 + 2*3")
	c.Equal(fxp.FromInteger(20), second.AdjustedPoints(), "10 + 2*5")
	c.Equal(fxp.FromInteger(16), first.AdjustedPoints(), "the first child's cost is unaffected by the second's")

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

func TestTraitFixedPointsCopiesAreIndependent(t *testing.T) {
	c := check.New(t)
	source := NewTrait(nil, nil, true)
	source.ContainerType = container.FixedCost
	source.FixedPoints = new(fxp.FromInteger(5))

	clone := source.Clone(LibraryFile{}, NewTemplate(), nil, Reference)
	c.True(source.FixedPoints != clone.FixedPoints, "cloning must allocate an independent FixedPoints value")
	*clone.FixedPoints = fxp.FromInteger(9)
	c.Equal(fxp.FromInteger(5), *source.FixedPoints, "mutating a clone must not mutate its source")

	var edit TraitEditData
	edit.CopyFrom(source)
	c.True(source.FixedPoints != edit.FixedPoints, "editor data must allocate an independent FixedPoints value")
	*edit.FixedPoints = fxp.FromInteger(11)
	c.Equal(fxp.FromInteger(5), *source.FixedPoints, "mutating editor data must not mutate its source")
}

// TestTraitEditDataCapturesMigratedReplacements verifies that populating an editor from a trait whose modifiers still
// carry the legacy per-modifier replacements captures the migrated replacements in the editor's snapshot. Without that,
// applying the editor's data back writes the pre-migration snapshot and silently drops them.
func TestTraitEditDataCapturesMigratedReplacements(t *testing.T) {
	c := check.New(t)
	trait := NewTrait(nil, nil, false)
	trait.Name = "Resistance"
	mod := NewTraitModifier(nil, nil, false)
	mod.Name = "@element@ Only"
	mod.Replacements = map[string]string{"element": "Fire"}
	trait.Modifiers = []*TraitModifier{mod}

	var edit TraitEditData
	edit.CopyFrom(trait)
	c.Equal("Fire", edit.Replacements["element"], "the editor's snapshot has the migrated replacements")
	edit.Replacements["element"] = "Ice"
	c.Equal("Fire", trait.Replacements["element"], "the editor's snapshot does not share the trait's map")

	edit.ApplyTo(trait)
	c.Equal("Ice", trait.Replacements["element"], "applying the editor's data writes the edited replacements")
	c.Equal(1, len(trait.Modifiers))
	if len(trait.Modifiers) != 1 {
		return
	}
	c.Equal("Ice Only", trait.Modifiers[0].NameWithReplacements())
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

	clone := source.Clone(LibraryFile{}, nil, nil, Reference)
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
	clonedContainer := parent.Clone(LibraryFile{}, nil, nil, Reference)
	c.Equal(1, len(clonedContainer.Children), "the child was cloned")
	clonedChild := clonedContainer.Children[0]
	c.True(clonedChild.Modifiers[0].OwningTrait() == clonedChild, "a cloned child's modifier belongs to that child")
}

func TestTraitFixedPointsRoundTrip(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value *fxp.Int
	}{
		{name: "unset"},
		{name: "zero", value: new(fxp.Int(0))},
		{name: "positive", value: new(fxp.Five)},
		{name: "negative", value: new(-fxp.Two)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			trait := NewTrait(nil, nil, true)
			trait.ContainerType = container.FixedCost
			trait.FixedPoints = tc.value
			data, err := jio.Marshal(trait)
			c.NoError(err)
			if tc.value == nil {
				c.NotContains(string(data), "fixed_points")
			} else {
				c.Contains(string(data), "fixed_points")
			}
			var restored Trait
			c.NoError(jio.Unmarshal(data, &restored))
			c.Equal(container.FixedCost, restored.ContainerType)
			c.Equal(tc.value, restored.FixedPoints)
		})
	}
}

func TestTraitFixedPointsHash(t *testing.T) {
	c := check.New(t)
	trait := NewTrait(nil, nil, true)
	trait.ContainerType = container.FixedCost
	hashTrait := func() []byte {
		h := sha256.New()
		trait.Hash(h)
		return h.Sum(nil)
	}
	unset := hashTrait()
	trait.FixedPoints = new(fxp.Int(0))
	zero := hashTrait()
	c.NotEqual(unset, zero, "unset and explicit zero must differ")
	trait.FixedPoints = new(fxp.One)
	one := hashTrait()
	c.NotEqual(zero, one, "the value participates in the source hash")
	trait.FixedPoints = new(fxp.One)
	c.Equal(one, hashTrait(), "pointer identity does not affect the hash")
	for _, kind := range container.Types {
		if kind == container.FixedCost {
			continue
		}
		trait.ContainerType = kind
		trait.FixedPoints = nil
		before := hashTrait()
		trait.FixedPoints = new(fxp.One)
		c.Equal(before, hashTrait(), "unused fixed points do not affect other container hashes")
	}
}

func TestTraitClearUnusedFixedPoints(t *testing.T) {
	for _, kind := range container.Types {
		t.Run(kind.Key(), func(t *testing.T) {
			c := check.New(t)
			trait := NewTrait(nil, nil, true)
			trait.ContainerType = kind
			trait.FixedPoints = new(fxp.Int(0))
			trait.ClearUnusedFieldsForType()
			if kind == container.FixedCost {
				c.Equal(new(fxp.Int(0)), trait.FixedPoints)
			} else {
				c.True(trait.FixedPoints == nil)
			}
		})
	}
	t.Run("non-container", func(t *testing.T) {
		trait := NewTrait(nil, nil, false)
		trait.ContainerType = container.FixedCost
		trait.FixedPoints = new(fxp.One)
		trait.ClearUnusedFieldsForType()
		check.New(t).True(trait.FixedPoints == nil)
	})
}

// TestTraitFixedCostAdjustedPoints covers modeled costs independently of picker selection validation.
func TestTraitFixedCostAdjustedPoints(t *testing.T) {
	for _, tc := range []struct {
		name       string
		fixed      *fxp.Int
		pickerType picker.Type
		compare    criteria.NumericComparison
		qualifier  fxp.Int
		disabled   bool
		want       fxp.Int
	}{
		{name: "manual overrides children and picker", fixed: new(fxp.Five), pickerType: picker.Points, compare: criteria.EqualsNumber, qualifier: fxp.Ten, want: fxp.Five},
		{name: "manual without picker", fixed: new(fxp.Five), want: fxp.Five},
		{name: "explicit zero overrides picker", fixed: new(fxp.Int(0)), pickerType: picker.Points, compare: criteria.EqualsNumber, qualifier: fxp.Ten},
		{name: "negative manual value", fixed: new(-fxp.Five), want: -fxp.Five},
		{name: "exact points picker", pickerType: picker.Points, compare: criteria.EqualsNumber, qualifier: fxp.Ten, want: fxp.Ten},
		{name: "exact zero points picker", pickerType: picker.Points, compare: criteria.EqualsNumber},
		{name: "no picker sums children", want: fxp.FromInteger(30)},
		{name: "at most sums children", pickerType: picker.Points, compare: criteria.AtMostNumber, qualifier: fxp.Ten, want: fxp.FromInteger(30)},
		{name: "at least sums children", pickerType: picker.Points, compare: criteria.AtLeastNumber, qualifier: fxp.Ten, want: fxp.FromInteger(30)},
		{name: "not equals sums children", pickerType: picker.Points, compare: criteria.NotEqualsNumber, qualifier: fxp.Ten, want: fxp.FromInteger(30)},
		{name: "count picker sums children", pickerType: picker.Count, compare: criteria.EqualsNumber, qualifier: fxp.Two, want: fxp.FromInteger(30)},
		{name: "disabled manual container", fixed: new(fxp.Five), disabled: true},
		{name: "disabled picker container", pickerType: picker.Points, compare: criteria.EqualsNumber, qualifier: fxp.Ten, disabled: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			trait := NewTrait(nil, nil, true)
			trait.ContainerType = container.FixedCost
			trait.FixedPoints = tc.fixed
			trait.TemplatePicker.Type = tc.pickerType
			trait.TemplatePicker.Qualifier.Compare = tc.compare
			trait.TemplatePicker.Qualifier.Qualifier = tc.qualifier
			trait.Disabled = tc.disabled
			for _, points := range []fxp.Int{fxp.Ten, fxp.Twenty} {
				child := NewTrait(nil, trait, false)
				child.BasePoints = points
				trait.Children = append(trait.Children, child)
			}
			check.New(t).Equal(tc.want, trait.AdjustedPoints())
		})
	}
}

func TestTraitFixedCostOwnership(t *testing.T) {
	for _, tc := range []struct {
		name  string
		owner DataOwner
		want  container.Type
	}{
		{name: "character", owner: NewEntity(), want: container.Group},
		{name: "template", owner: NewTemplate(), want: container.FixedCost},
		{name: "library", want: container.FixedCost},
	} {
		for _, clone := range []bool{false, true} {
			action := "assign"
			if clone {
				action = "clone"
			}
			t.Run(tc.name+"/"+action, func(t *testing.T) {
				c := check.New(t)
				source := NewTrait(NewTemplate(), nil, true)
				source.ContainerType = container.FixedCost
				source.FixedPoints = new(fxp.Five)
				child := NewTrait(source.DataOwner(), source, true)
				child.ContainerType = container.FixedCost
				child.FixedPoints = new(fxp.Int(0))
				source.Children = []*Trait{child}
				result := source
				if clone {
					result = source.Clone(LibraryFile{}, tc.owner, nil, Reference)
					c.Equal(container.FixedCost, source.ContainerType)
					c.Equal(new(fxp.Five), source.FixedPoints)
					c.Equal(container.FixedCost, child.ContainerType)
					c.Equal(new(fxp.Int(0)), child.FixedPoints)
				} else {
					result.SetDataOwner(tc.owner)
				}
				c.Equal(tc.want, result.ContainerType)
				c.Equal(tc.want, result.Children[0].ContainerType)
				c.True(result.DataOwner() == tc.owner)
				c.True(result.Children[0].DataOwner() == tc.owner)
				if tc.want == container.Group {
					c.True(result.FixedPoints == nil)
					c.True(result.Children[0].FixedPoints == nil)
				} else {
					c.Equal(new(fxp.Five), result.FixedPoints)
					c.Equal(new(fxp.Int(0)), result.Children[0].FixedPoints)
				}
			})
		}
	}
}

func TestTraitFixedCostSourceSyncStaysGroup(t *testing.T) {
	c := check.New(t)
	owner := NewEntity()
	libFile := LibraryFile{Library: "Test Library", Path: "Test" + TraitsExt}

	source := NewTrait(nil, nil, true)
	source.ContainerType = container.FixedCost
	source.FixedPoints = new(fxp.Five)

	local := NewTrait(owner, nil, true)
	local.Source = Source{LibraryFile: libFile, TID: source.TID}
	owner.SourceMatcher().libHashes = map[LibraryFile]libSrcData{
		libFile: {dataHashes: map[tid.TID]HashAndData{source.TID: {Hash: Hash64(source), Data: source}}},
	}

	state, _ := owner.SourceMatcher().Match(local)
	c.NotEqual(srcstate.Matched, state, "precondition: the local Group has drifted from its Fixed Cost source")
	c.Equal(container.Group, local.ContainerType)
	c.True(local.FixedPoints == nil)

	local.SyncWithSource()

	c.Equal(container.Group, local.ContainerType)
	c.True(local.FixedPoints == nil)
	c.Equal(Source{}, local.Source)
}

func TestTraitFixedCostInlineTag(t *testing.T) {
	for _, tc := range []struct {
		name         string
		fixed        *fxp.Int
		pointsPicker bool
	}{
		{name: "manual", fixed: new(fxp.Five)},
		{name: "zero", fixed: new(fxp.Int(0))},
		{name: "picker", pointsPicker: true},
		{name: "children"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			trait := NewTrait(NewTemplate(), nil, true)
			trait.ContainerType = container.FixedCost
			trait.FixedPoints = tc.fixed
			if tc.pointsPicker {
				trait.TemplatePicker.Type = picker.Points
				trait.TemplatePicker.Qualifier.Compare = criteria.EqualsNumber
				trait.TemplatePicker.Qualifier.Qualifier = fxp.Five
			}
			var data CellData
			trait.CellData(TraitDescriptionColumn, &data)
			check.New(t).Equal("Fixed", data.InlineTag)
		})
	}
}
