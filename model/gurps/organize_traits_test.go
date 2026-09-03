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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/toolbox/v2/check"
)

// newOrganizableTrait creates a non-container trait with a fixed point cost and the given tags.
func newOrganizableTrait(owner DataOwner, name string, points int, tags ...string) *Trait {
	trait := NewTrait(owner, nil, false)
	trait.Name = name
	trait.BasePoints = fxp.FromInteger(points)
	trait.Tags = tags
	return trait
}

// newOrganizableContainer creates a trait container of the given type.
func newOrganizableContainer(owner DataOwner, name string, containerType container.Type) *Trait {
	trait := NewTrait(owner, nil, true)
	trait.Name = name
	trait.ContainerType = containerType
	return trait
}

// organizedNames returns the display names of the traits in the list, which is what most of these tests compare.
func organizedNames(list []*Trait) []string {
	names := make([]string, len(list))
	for i, one := range list {
		names[i] = one.String()
	}
	return names
}

// TestOrganizeTraitsPointBuckets verifies that untagged traits are filed by cost alone and that the containers come out
// in the fixed category order, no matter what order the traits arrived in.
func TestOrganizeTraitsPointBuckets(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	list := []*Trait{
		newOrganizableTrait(e, "Feature", 0),
		newOrganizableTrait(e, "Quirk", -1),
		newOrganizableTrait(e, "Disadvantage", -10),
		newOrganizableTrait(e, "Perk", 1),
		newOrganizableTrait(e, "Advantage", 10),
	}
	organized, changed := OrganizeTraits(e, list)
	c.True(changed, "filing five loose traits changes the list")
	c.Equal([]string{"Advantages", "Perks", "Disadvantages", "Quirks", "Features"}, organizedNames(organized),
		"the containers appear in category order")
	c.Equal([]string{"Advantage"}, organizedNames(organized[0].Children), "10 points is an advantage")
	c.Equal([]string{"Perk"}, organizedNames(organized[1].Children), "1 point is a perk")
	c.Equal([]string{"Disadvantage"}, organizedNames(organized[2].Children), "-10 points is a disadvantage")
	c.Equal([]string{"Quirk"}, organizedNames(organized[3].Children), "-1 point is a quirk")
	c.Equal([]string{"Feature"}, organizedNames(organized[4].Children), "0 points is a feature")
}

// TestOrganizeTraitsTagOverridesPoints verifies that a single category tag decides where a trait goes even when its
// cost says otherwise, and that the tag is matched the way the rest of GCS matches tags: without regard to case, and
// against the colon-separated pieces of a compound tag.
func TestOrganizeTraitsTagOverridesPoints(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	list := []*Trait{
		newOrganizableTrait(e, "Costly Quirk", 10, "Quirk"),
		newOrganizableTrait(e, "Cheap Advantage", -1, "Advantages"),
		newOrganizableTrait(e, "Lowercase", 25, "perk"),
		newOrganizableTrait(e, "Compound", -20, "Advantage: Mental"),
	}
	organized, changed := OrganizeTraits(e, list)
	c.True(changed, "filing tagged traits changes the list")
	c.Equal([]string{"Advantages", "Perks", "Quirks"}, organizedNames(organized),
		"only the categories that received traits get containers")
	c.Equal([]string{"Cheap Advantage", "Compound"}, organizedNames(organized[0].Children),
		"a plural tag and a colon-separated subset both name the advantages")
	c.Equal([]string{"Lowercase"}, organizedNames(organized[1].Children), "tag matching ignores case")
	c.Equal([]string{"Costly Quirk"}, organizedNames(organized[2].Children), "the tag beats the point total")
}

// TestOrganizeTraitsConflictingTagsFallBackToPoints verifies that a trait claiming to belong to two categories at once
// is filed by its cost instead, since its tags cancel each other out.
func TestOrganizeTraitsConflictingTagsFallBackToPoints(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	list := []*Trait{
		newOrganizableTrait(e, "One Point", 1, "Advantage", "Perk"),
		newOrganizableTrait(e, "Minus One", -1, "Disadvantage", "Quirk"),
		newOrganizableTrait(e, "Five Points", 5, "Advantage", "Disadvantage"),
	}
	organized, changed := OrganizeTraits(e, list)
	c.True(changed, "filing the traits changes the list")
	c.Equal([]string{"Advantages", "Perks", "Quirks"}, organizedNames(organized), "three categories are used")
	c.Equal([]string{"Five Points"}, organizedNames(organized[0].Children), "5 points is an advantage")
	c.Equal([]string{"One Point"}, organizedNames(organized[1].Children), "1 point is a perk")
	c.Equal([]string{"Minus One"}, organizedNames(organized[2].Children), "-1 point is a quirk")
}

// TestOrganizeTraitsLanguageTagWins verifies that a Language tag files the trait with the languages regardless of what
// else it is tagged as or what it costs, and that the languages come last among the category containers.
func TestOrganizeTraitsLanguageTagWins(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	list := []*Trait{
		newOrganizableTrait(e, "Native Tongue", 3, "Advantage", "Language", "Mental"),
		newOrganizableTrait(e, "Broken Accent", -1, "Disadvantage", "Language", "Mental"),
		newOrganizableTrait(e, "Spoken Only", 2, "Language: Spoken"),
		newOrganizableTrait(e, "Plain Advantage", 10),
	}
	organized, changed := OrganizeTraits(e, list)
	c.True(changed, "filing the traits changes the list")
	c.Equal([]string{"Advantages", "Languages"}, organizedNames(organized), "the languages sort last")
	c.Equal([]string{"Plain Advantage"}, organizedNames(organized[0].Children), "only the untagged trait is an advantage")
	c.Equal([]string{"Broken Accent", "Native Tongue", "Spoken Only"}, organizedNames(organized[1].Children),
		"every language lands in the languages container")
}

// TestOrganizeTraitsReusesExistingContainer verifies that an existing group container is adopted rather than
// duplicated, wherever it sits in the list and whatever surrounding whitespace or casing its name has, and that what
// it already held is sorted in alongside the newly filed traits.
func TestOrganizeTraitsReusesExistingContainer(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	existing := newOrganizableContainer(e, "  advantages  ", container.Group)
	kept := newOrganizableTrait(e, "Night Vision", 9, "Advantage")
	kept.SetParent(existing)
	existing.SetChildren([]*Trait{kept})
	added := newOrganizableTrait(e, "Acute Hearing", 2)
	list := []*Trait{added, existing}
	organized, changed := OrganizeTraits(e, list)
	c.True(changed, "adopting the container changes the list")
	c.Equal(1, len(organized), "the container is the only top-level entry")
	c.True(existing == organized[0], "the existing container is reused rather than replaced")
	c.Equal([]string{"Acute Hearing", "Night Vision"}, organizedNames(organized[0].Children),
		"the trait already present is merged into the sort")
}

// TestOrganizeTraitsResortsExistingContainer verifies that a list which is already in the right shape still reports a
// change when the only thing out of place is the order of an existing container's children.
func TestOrganizeTraitsResortsExistingContainer(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	existing := newOrganizableContainer(e, "Advantages", container.Group)
	first := newOrganizableTrait(e, "Zeal", 5)
	second := newOrganizableTrait(e, "Acute Vision", 5)
	for _, one := range []*Trait{first, second} {
		one.SetParent(existing)
	}
	existing.SetChildren([]*Trait{first, second})
	organized, changed := OrganizeTraits(e, []*Trait{existing})
	c.True(changed, "reordering the children counts as a change")
	c.Equal([]string{"Acute Vision", "Zeal"}, organizedNames(organized[0].Children), "the children were sorted")
}

// TestOrganizeTraitsIgnoresNonGroupContainers verifies that a container whose type gives it a meaning of its own -- a
// meta-trait or a set of alternative abilities -- is never filed into, even when it happens to carry a category name.
// A fresh group container is made instead and the typed one keeps its contents and its place in the tail.
func TestOrganizeTraitsIgnoresNonGroupContainers(t *testing.T) {
	for _, containerType := range []container.Type{container.MetaTrait, container.AlternativeAbilities} {
		t.Run(containerType.Key(), func(_ *testing.T) {
			c := check.New(t)
			e := NewEntity()
			typed := newOrganizableContainer(e, "Advantages", containerType)
			inner := newOrganizableTrait(e, "Inner", 5)
			inner.SetParent(typed)
			typed.SetChildren([]*Trait{inner})
			emptyQuirks := newOrganizableContainer(e, "Quirks", container.MetaTrait)
			loose := newOrganizableTrait(e, "Loose", 20)
			organized, changed := OrganizeTraits(e, []*Trait{typed, emptyQuirks, loose})
			c.True(changed, "filing the loose trait changes the list")
			c.Equal([]string{"Advantages", "Advantages", "Quirks"}, organizedNames(organized),
				"a new group container leads, with the typed containers left in the tail")
			c.True(organized[0] != typed, "the typed container is not reused")
			c.Equal(container.Group, organized[0].ContainerType, "the created container is a plain group")
			c.Equal([]string{"Loose"}, organizedNames(organized[0].Children), "the loose trait is filed into the new one")
			c.True(organized[1] == typed, "the typed container keeps its relative position")
			c.Equal([]string{"Inner"}, organizedNames(typed.Children), "the typed container keeps its contents")
			c.True(organized[2] == emptyQuirks, "an empty container that was never claimed is kept")
		})
	}
}

// TestOrganizeTraitsDropsEmptyCategoryContainers verifies that a category container left with nothing in it is removed,
// while a container the user made for some other purpose is kept even when it is empty.
func TestOrganizeTraitsDropsEmptyCategoryContainers(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	emptyQuirks := newOrganizableContainer(e, "Quirks", container.Group)
	racial := newOrganizableContainer(e, "Racial", container.Group)
	loose := newOrganizableTrait(e, "Loose", 10)
	organized, changed := OrganizeTraits(e, []*Trait{emptyQuirks, racial, loose})
	c.True(changed, "dropping the empty category container changes the list")
	c.Equal([]string{"Advantages", "Racial"}, organizedNames(organized),
		"the empty category container is dropped and the other one is kept")
	c.True(organized[1] == racial, "the non-category container is the one that survived")
}

// TestOrganizeTraitsKeepsOtherContainersInOrder verifies that containers organizing has no interest in are neither
// reordered nor separated from each other.
func TestOrganizeTraitsKeepsOtherContainersInOrder(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	first := newOrganizableContainer(e, "Racial", container.Group)
	second := newOrganizableContainer(e, "Cybernetics", container.Group)
	third := newOrganizableContainer(e, "Templates", container.Group)
	for _, one := range []*Trait{first, second, third} {
		child := newOrganizableTrait(e, "Child of "+one.Name, 1)
		child.SetParent(one)
		one.SetChildren([]*Trait{child})
	}
	loose := newOrganizableTrait(e, "Loose", 10)
	organized, changed := OrganizeTraits(e, []*Trait{first, loose, second, third})
	c.True(changed, "filing the loose trait changes the list")
	c.Equal([]string{"Advantages", "Racial", "Cybernetics", "Templates"}, organizedNames(organized),
		"the untouched containers keep their relative order behind the category container")
}

// TestOrganizeTraitsSortsChildrenNaturally verifies that a container's children are ordered the way the trait table
// orders a name sort: case is ignored and embedded numbers compare as numbers, so level 2 precedes level 10.
func TestOrganizeTraitsSortsChildrenNaturally(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	drTen := newOrganizableTrait(e, "Damage Resistance", 0, "Advantage")
	drTen.CanLevel = true
	drTen.PointsPerLevel = fxp.Five
	drTen.Levels = fxp.Ten
	drTwo := newOrganizableTrait(e, "Damage Resistance", 0, "Advantage")
	drTwo.CanLevel = true
	drTwo.PointsPerLevel = fxp.Five
	drTwo.Levels = fxp.Two
	organized, changed := OrganizeTraits(e, []*Trait{
		newOrganizableTrait(e, "Zeal", 5, "Advantage"),
		newOrganizableTrait(e, "acuity", 5, "Advantage"),
		drTen,
		drTwo,
	})
	c.True(changed, "filing the traits changes the list")
	c.Equal([]string{"acuity", "Damage Resistance 2", "Damage Resistance 10", "Zeal"},
		organizedNames(organized[0].Children), "children sort naturally, ignoring case")
}

// TestOrganizeTraitsSetsParentsAndOwner verifies that the tree comes out consistent: every filed trait points at its
// new container, every container sits at the top level, and a created container is a properly formed container owned by
// the same data owner the caller passed in.
func TestOrganizeTraitsSetsParentsAndOwner(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	organized, changed := OrganizeTraits(e, []*Trait{
		newOrganizableTrait(e, "Advantage", 10),
		newOrganizableTrait(e, "Quirk", -1),
	})
	c.True(changed, "filing the traits changes the list")
	for _, one := range organized {
		c.True(one.Container(), "every top-level entry is a container")
		c.Nil(one.Parent(), "a top-level container has no parent")
		c.True(e == one.DataOwner(), "a created container belongs to the owner passed in")
		for _, child := range one.Children {
			c.True(one == child.Parent(), "a filed trait points at its container")
		}
	}
}

// TestOrganizeTraitsIsIdempotent verifies that organizing an already organized list reports no change and hands back
// the very same slice, so a caller can skip recording an undoable edit.
func TestOrganizeTraitsIsIdempotent(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	organized, changed := OrganizeTraits(e, []*Trait{
		newOrganizableTrait(e, "Advantage", 10),
		newOrganizableTrait(e, "Perk", 1),
		newOrganizableTrait(e, "Language", 1, "Language"),
	})
	c.True(changed, "the first pass changes the list")
	again, changedAgain := OrganizeTraits(e, organized)
	c.False(changedAgain, "a second pass finds nothing to do")
	c.True(slices.Equal(organized, again), "the same list is handed back untouched")
}

// TestOrganizeTraitsEmptyList verifies that there is nothing to do for a list with nothing in it.
func TestOrganizeTraitsEmptyList(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	organized, changed := OrganizeTraits(e, nil)
	c.False(changed, "a nil list is already organized")
	c.Equal(0, len(organized), "a nil list stays empty")
	organized, changed = OrganizeTraits(e, []*Trait{})
	c.False(changed, "an empty list is already organized")
	c.Equal(0, len(organized), "an empty list stays empty")
}

// TestOrganizeTraitsDisabledTrait verifies that switching a trait off doesn't change where it is filed. The
// Trait.AdjustedPoints method reports zero for a disabled trait, which would file every switched-off trait under
// Features; organizing has to look past that to what the trait costs when it is on.
func TestOrganizeTraitsDisabledTrait(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	disabled := newOrganizableTrait(e, "Sleepy", -10)
	disabled.Disabled = true
	c.Equal(fxp.Int(0), disabled.AdjustedPoints(), "a disabled trait reports no cost")
	organized, changed := OrganizeTraits(e, []*Trait{disabled})
	c.True(changed, "filing the trait changes the list")
	c.Equal([]string{"Disadvantages"}, organizedNames(organized), "a disabled disadvantage is still a disadvantage")
	c.Equal([]string{"Sleepy"}, organizedNames(organized[0].Children), "the disabled trait was filed")
}

// TestOrganizeTraitsSecondSameNamedContainerLeftAlone verifies that only one container per category is adopted; a
// duplicate keeps its contents and drops to the tail rather than being merged away.
func TestOrganizeTraitsSecondSameNamedContainerLeftAlone(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	first := newOrganizableContainer(e, "Advantages", container.Group)
	second := newOrganizableContainer(e, "Advantages", container.Group)
	stowed := newOrganizableTrait(e, "Stowed", 4)
	stowed.SetParent(second)
	second.SetChildren([]*Trait{stowed})
	loose := newOrganizableTrait(e, "Loose", 10)
	organized, changed := OrganizeTraits(e, []*Trait{first, second, loose})
	c.True(changed, "filing the loose trait changes the list")
	c.Equal(2, len(organized), "the duplicate container is kept, not merged")
	c.True(organized[0] == first, "the first matching container is the one adopted")
	c.Equal([]string{"Loose"}, organizedNames(first.Children), "the loose trait went into the adopted container")
	c.True(organized[1] == second, "the duplicate falls to the tail")
	c.Equal([]string{"Stowed"}, organizedNames(second.Children), "the duplicate keeps what it held")
}

// TestOrganizeTraitsNestedTraitsUntouched verifies that only top-level traits are re-filed. A trait the user tucked
// into a container stays where they put it, even when its cost says it belongs somewhere else.
func TestOrganizeTraitsNestedTraitsUntouched(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	existing := newOrganizableContainer(e, "Advantages", container.Group)
	nested := newOrganizableTrait(e, "Misfiled", -10)
	nested.SetParent(existing)
	existing.SetChildren([]*Trait{nested})
	organized, changed := OrganizeTraits(e, []*Trait{existing})
	c.False(changed, "there is nothing at the top level to move")
	c.Equal([]string{"Advantages"}, organizedNames(organized), "the container is left as it was")
	c.Equal([]string{"Misfiled"}, organizedNames(existing.Children), "a nested trait is never re-filed")
	c.True(existing == nested.Parent(), "the nested trait keeps its parent")
}

// TestOrganizeTraitsClassifiesBeforeMoving verifies that a trait is filed by what it cost before the move. A container
// can carry modifiers that its contents inherit, so filing a trait into one can change that trait's cost -- and would
// change where it belongs if the decision were made after the fact rather than before.
func TestOrganizeTraitsClassifiesBeforeMoving(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	existing := newOrganizableContainer(e, "Advantages", container.Group)
	mod := NewTraitModifier(e, nil, false)
	mod.Name = "Limitation"
	mod.CostAdj = "-75%"
	existing.Modifiers = []*TraitModifier{mod}
	existing.SetDataOwner(e)
	loose := newOrganizableTrait(e, "Four Points", 4)
	organized, changed := OrganizeTraits(e, []*Trait{existing, loose})
	c.True(changed, "filing the loose trait changes the list")
	c.Equal([]string{"Advantages"}, organizedNames(organized), "the trait was filed as an advantage")
	c.Equal([]string{"Four Points"}, organizedNames(organized[0].Children), "the 4-point trait went to the advantages")
	c.Equal(fxp.One, loose.AdjustedPoints(),
		"the container's limitation really does drop the trait to a perk's worth of points once it is inside")
}

// TestOrganizeTraitsLibraryShapes covers the shapes that show up in the stock libraries: language traits tagged as
// advantages, and leveled traits tagged both ways whose category depends on how many levels were bought.
func TestOrganizeTraitsLibraryShapes(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	language := newOrganizableTrait(e, "Spanish", 10, "Advantage", "Language", "Mental")
	organized, changed := OrganizeTraits(e, []*Trait{language})
	c.True(changed, "filing the language changes the list")
	c.Equal([]string{"Languages"}, organizedNames(organized), "a language tag outranks the advantage tag")

	leveled := newOrganizableTrait(e, "Extra Effort", 0, "Advantage", "Perk")
	leveled.CanLevel = true
	leveled.PointsPerLevel = fxp.One
	leveled.Levels = fxp.One
	organized, changed = OrganizeTraits(e, []*Trait{leveled})
	c.True(changed, "filing the leveled trait changes the list")
	c.Equal([]string{"Perks"}, organizedNames(organized), "one level costs one point, so it is a perk")

	leveled.SetParent(nil)
	leveled.Levels = fxp.Three
	organized, changed = OrganizeTraits(e, []*Trait{leveled})
	c.True(changed, "filing the leveled trait changes the list")
	c.Equal([]string{"Advantages"}, organizedNames(organized), "three levels cost three points, so it is an advantage")
}

// TestOrganizeTraitsWithTemplateOwner verifies that organizing works for a template, which has no entity behind its
// traits for the point calculations to consult.
func TestOrganizeTraitsWithTemplateOwner(t *testing.T) {
	c := check.New(t)
	tmpl := NewTemplate()
	list := []*Trait{
		newOrganizableTrait(tmpl, "Quirky", -1),
		newOrganizableTrait(tmpl, "Talented", 15),
	}
	organized, changed := OrganizeTraits(tmpl, list)
	c.True(changed, "filing the traits changes the list")
	c.Equal([]string{"Advantages", "Quirks"}, organizedNames(organized), "the categories are the same for a template")
	c.True(tmpl == organized[0].DataOwner(), "a created container belongs to the template")
	c.Equal([]string{"Talented"}, organizedNames(organized[0].Children), "15 points is an advantage")
	c.Equal([]string{"Quirky"}, organizedNames(organized[1].Children), "-1 point is a quirk")
}
