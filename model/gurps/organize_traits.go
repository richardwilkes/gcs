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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xstrings"
)

// traitCategory is one of the buckets a top-level trait can be filed into.
type traitCategory struct {
	// points reports whether a trait costing this many points belongs here. nil for a category that can only be
	// reached by tag.
	points func(pts fxp.Int) bool
	name   string
	tags   []string
	// overrides marks a category whose tag beats every other consideration, rather than having to be the only tag
	// present before it counts.
	overrides bool
}

// organizeTraitCategories returns the categories OrganizeTraits files traits into, in the order their containers are
// placed. This is built on each call instead of being held in a package-level variable because the names come from the
// translation catalog, which isn't loaded yet when package-level variables are initialized.
func organizeTraitCategories() []traitCategory {
	return []traitCategory{
		{
			name:   i18n.Text("Advantages"),
			tags:   []string{"Advantage", "Advantages"},
			points: func(pts fxp.Int) bool { return pts > fxp.One },
		},
		{
			name:   i18n.Text("Perks"),
			tags:   []string{"Perk", "Perks"},
			points: func(pts fxp.Int) bool { return pts == fxp.One },
		},
		{
			name:   i18n.Text("Disadvantages"),
			tags:   []string{"Disadvantage", "Disadvantages"},
			points: func(pts fxp.Int) bool { return pts < fxp.NegOne },
		},
		{
			name:   i18n.Text("Quirks"),
			tags:   []string{"Quirk", "Quirks"},
			points: func(pts fxp.Int) bool { return pts == fxp.NegOne },
		},
		{
			name:   i18n.Text("Features"),
			tags:   []string{"Feature", "Features"},
			points: func(pts fxp.Int) bool { return pts == 0 },
		},
		{
			name:      i18n.Text("Languages"),
			tags:      []string{"Language", "Languages"},
			overrides: true,
		},
	}
}

// OrganizeTraits files each top-level trait that isn't a container into a container named for its category, creating
// that container if the list doesn't already hold one by that name, then puts those containers at the front of the
// list in category order and sorts their direct children by name. A category container that ends up with no children
// is dropped. Everything else keeps its relative order behind them. Returns the new top-level list and whether it
// differs from the one passed in, so a caller can skip recording an edit for a list that was already organized.
func OrganizeTraits(owner DataOwner, list []*Trait) ([]*Trait, bool) {
	if len(list) == 0 {
		return list, false
	}
	categories := organizeTraitCategories()

	// Every trait is classified before any of them is moved. A trait's point total isn't a property of the trait
	// alone: AllModifiers() gathers the modifiers of its ancestors and EffectivelyDisabled() reports true when any
	// ancestor is disabled, so dropping a trait into a container that carries a cost modifier or is switched off would
	// change the answer for every trait filed after it. Deciding up front means each trait is judged as the user left
	// it.
	destinations := make([]int, len(list))
	for i, t := range list {
		if t.Container() {
			// Containers are left where the user put them, and their contents aren't disturbed.
			destinations[i] = -1
			continue
		}
		destinations[i] = categoryForTrait(t, categories)
	}

	// Give each category the first top-level Group container that already carries its name, so an existing container
	// and whatever the user put in it survive. Containers of any other type are passed over, since their type gives
	// them a meaning that filing unrelated traits into them would break.
	containers := make([]*Trait, len(categories))
	consumed := make([]bool, len(list))
	for ci := range categories {
		for i, t := range list {
			if !consumed[i] && isOrganizeContainerFor(t, categories[ci].name) {
				containers[ci] = t
				consumed[i] = true
				break
			}
		}
	}

	// Assemble fresh children slices rather than appending into the ones the containers already hold, so that a list
	// that turns out to need no changes is left exactly as it was found.
	existing := make([][]*Trait, len(categories))
	children := make([][]*Trait, len(categories))
	for ci, c := range containers {
		if c != nil {
			existing[ci] = slices.Clone(c.Children)
			children[ci] = slices.Clone(c.Children)
		}
	}
	for i, t := range list {
		if ci := destinations[i]; ci >= 0 {
			children[ci] = append(children[ci], t)
		}
	}

	result := make([]*Trait, 0, len(list)+len(categories))
	changed := false
	for ci := range categories {
		if len(children[ci]) == 0 {
			// A category container with nothing to hold is clutter, so it is dropped. Only a container that a category
			// claimed can be dropped this way; anything else the user made is kept, empty or not.
			continue
		}
		c := containers[ci]
		if c == nil {
			c = NewTrait(owner, nil, true)
			c.Name = categories[ci].name
		}
		slices.SortStableFunc(children[ci], compareTraitsByName)
		for _, child := range children[ci] {
			child.SetParent(c)
		}
		if !slices.Equal(existing[ci], children[ci]) {
			changed = true
		}
		c.SetChildren(children[ci])
		c.SetParent(nil)
		result = append(result, c)
	}
	for i, t := range list {
		if !consumed[i] && destinations[i] < 0 {
			t.SetParent(nil)
			result = append(result, t)
		}
	}
	if !changed && slices.Equal(list, result) {
		return list, false
	}
	return result, true
}

// categoryForTrait returns the index of the category the trait belongs to, or -1 if none of them fit, in which case
// the caller should leave the trait where it is.
func categoryForTrait(t *Trait, categories []traitCategory) int {
	// An overriding tag settles the question by itself. A language costs points like any other trait, but the user
	// filed it as a language by tagging it, and that is what they want to see.
	for i := range categories {
		if categories[i].overrides && hasAnyTag(categories[i].tags, t.Tags) {
			return i
		}
	}

	// A single category tag is a deliberate statement about where the trait goes, so it is honored even when the point
	// total disagrees. Tags for more than one category contradict each other and say nothing, so those traits are
	// filed by their cost along with the untagged ones.
	found := -1
	for i := range categories {
		if categories[i].overrides || !hasAnyTag(categories[i].tags, t.Tags) {
			continue
		}
		if found >= 0 {
			found = -1
			break
		}
		found = i
	}
	if found >= 0 {
		return found
	}

	pts := pointsForOrganizing(t)
	for i := range categories {
		if categories[i].points != nil && categories[i].points(pts) {
			return i
		}
	}
	return -1
}

// pointsForOrganizing returns the point total to file the trait by. The free AdjustedPoints function is used rather
// than the Trait.AdjustedPoints method because the method reports zero for a disabled trait, which would bury a
// switched-off 15-point advantage among the features instead of leaving it with the advantages it will rejoin as soon
// as it is switched back on.
func pointsForOrganizing(t *Trait) fxp.Int {
	return AdjustedPoints(EntityFromNode(t), t, t.CanLevel, t.BasePoints, t.Levels, t.PointsPerLevel, t.SelfControl,
		t.Frequency, t.AllModifiers(), t.RoundCostDown)
}

// hasAnyTag returns true if any of the candidates is present in tags.
func hasAnyTag(candidates, tags []string) bool {
	for _, candidate := range candidates {
		if HasTag(candidate, tags) {
			return true
		}
	}
	return false
}

// isOrganizeContainerFor returns true if the trait is a plain group container that organizing may file traits into
// under the given category name.
func isOrganizeContainerFor(t *Trait, name string) bool {
	return t.Container() && t.ContainerType == container.Group &&
		strings.EqualFold(strings.TrimSpace(t.NameWithReplacements()), name)
}

// compareTraitsByName orders two traits the way the trait table orders them when sorting by name, so that organizing a
// list and sorting the table by name agree on what "in order" looks like. Commas are dropped so that "Appearance,
// Beautiful" sorts as though it read "Appearance Beautiful", and the comparison is natural so that "Damage Resistance
// 2" comes before "Damage Resistance 10".
func compareTraitsByName(a, b *Trait) int {
	return xstrings.NaturalCmp(strings.ReplaceAll(a.String(), ",", ""), strings.ReplaceAll(b.String(), ",", ""), true)
}
